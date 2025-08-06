package plenccodec

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

//go:generate stringer -type FieldType
type FieldType int

const (
	FieldTypeInt FieldType = iota
	FieldTypeUint
	FieldTypeFloat32
	FieldTypeFloat64
	FieldTypeString
	FieldTypeSlice
	FieldTypeStruct
	FieldTypeBool
	FieldTypeTime
	FieldTypeJSONObject
	FieldTypeJSONArray
	// Not zig-zag encoded, but expected to be signed. Don't use if negative
	// numbers are likely.
	FieldTypeFlatInt
	// Do we want int32 types?
	// Do we want fixed size int types?
	// Do we want a separate bytes type?
	// Do we want an ENUM type? How would we encode it?
)

//go:generate stringer -type LogicalType
type LogicalType int

const (
	LogicalTypeNone LogicalType = iota
	LogicalTypeTimestamp
	LogicalTypeDate
	LogicalTypeTime
	LogicalTypeMap
	LogicalTypeMapEntry
)

// Descriptor describes how a type is plenc-encoded. It contains enough
// information to decode plenc data marshalled from the described type.
type Descriptor struct {
	// Index is the plenc index of this field
	Index int `plenc:"1"`
	// Name is the name of the field
	Name string `plenc:"2"`
	// Type is the type of the field
	Type FieldType `plenc:"3"`
	// TypeName is used for struct types and is the name of the struct.
	TypeName string `plenc:"5"`
	// Elements is valid for FieldTypeSlice, FieldTypeStruct & FieldTypeMap. For
	// FieldTypeSlice we expect one entry that describes the elements of the
	// slice. For FieldTypeStruct we expect an entry for each field in the
	// struct. For FieldTypeMap we expect two entries. The first is for the key
	// type and the second is for the map type
	Elements []Descriptor `plenc:"4"`

	// ExplicitPresence is set if the field has a mechanism to distinguish when
	// it is not present. So either a pointer type or something from the null
	// package. If this is not set then a missing value indicates the zero
	// value, not a null or nil entry.
	ExplicitPresence bool `plenc:"6"`

	// The logical type of the field. This is used to indicate if the field has
	// any special meaning - e.g. if a long or string indicates a timestamp.
	LogicalType LogicalType `plenc:"7"`
}

func (d *Descriptor) Read(out Outputter, data []byte) (err error) {
	_, err = d.read(out, data)
	return err
}

func (d *Descriptor) read(out Outputter, data []byte) (n int, err error) {
	switch d.Type {
	case FieldTypeInt:
		var v int64
		n, err = IntCodec[int64]{}.Read(data, unsafe.Pointer(&v), plenccore.WTVarInt)
		out.Int64(v)
		return n, err

	case FieldTypeFlatInt:
		switch d.LogicalType {
		case LogicalTypeTimestamp:
			var v time.Time
			n, err = BQTimestampCodec{}.Read(data, unsafe.Pointer(&v), plenccore.WTVarInt)
			out.Time(v)
		default:
			var v int64
			n, err = FlatIntCodec[uint64]{}.Read(data, unsafe.Pointer(&v), plenccore.WTVarInt)
			out.Int64(v)
		}
		return n, err

	case FieldTypeUint:
		var v uint64
		n, err = UintCodec[uint64]{}.Read(data, unsafe.Pointer(&v), plenccore.WTVarInt)
		out.Uint64(v)
		return n, err

	case FieldTypeFloat32:
		var v float32
		n, err = Float32Codec{}.Read(data, unsafe.Pointer(&v), plenccore.WT32)
		out.Float32(v)
		return n, err

	case FieldTypeFloat64:
		var v float64
		n, err = Float64Codec{}.Read(data, unsafe.Pointer(&v), plenccore.WT64)
		out.Float64(v)
		return n, err

	case FieldTypeString:
		var v string
		n, err = StringCodec{}.Read(data, unsafe.Pointer(&v), plenccore.WTLength)
		out.String(v)
		return n, err

	case FieldTypeBool:
		var v bool
		n, err = BoolCodec{}.Read(data, unsafe.Pointer(&v), plenccore.WTLength)
		out.Bool(v)
		return n, err

	case FieldTypeTime:
		var v time.Time
		n, err = TimeCodec{}.Read(data, unsafe.Pointer(&v), plenccore.WTLength)
		out.Time(v)
		return n, err

	case FieldTypeSlice:
		if d.isValidJSONMap() {
			out.StartObject()
			defer out.EndObject()
		} else {
			out.StartArray()
			defer out.EndArray()
		}
		return d.readAsSlice(out, data)

	case FieldTypeStruct:
		if d.isValidJSONMapEntry() {
			return d.readAsMapEntry(out, data)
		}
		out.StartObject()
		defer out.EndObject()
		return d.readAsStruct(out, data)

	case FieldTypeJSONObject:
		out.StartObject()
		defer out.EndObject()
		return d.readAsJSON(out, data)

	case FieldTypeJSONArray:
		out.StartArray()
		defer out.EndArray()
		return d.readAsJSON(out, data)
	}

	return 0, fmt.Errorf("unrecognised field type %s", d.Type)
}

func (d *Descriptor) isValidJSONMap() bool {
	if d.Type != FieldTypeSlice || d.LogicalType != LogicalTypeMap {
		return false
	}
	if len(d.Elements) != 1 {
		return false
	}
	return d.Elements[0].isValidJSONMapEntry()
}

func (d *Descriptor) isValidJSONMapEntry() bool {
	if d.Type != FieldTypeStruct || d.LogicalType != LogicalTypeMapEntry {
		return false
	}
	if len(d.Elements) != 2 {
		return false
	}
	key := &d.Elements[0]
	return key.Type == FieldTypeString
}

func (d *Descriptor) readAsSlice(out Outputter, data []byte) (n int, err error) {
	elt := &d.Elements[0]
	switch elt.Type {
	case FieldTypeFloat32, FieldTypeFloat64, FieldTypeInt, FieldTypeUint:
		// If data is generated by protobuf this could be an element of a slice.
		// We won't support that for now. So this is either a float64 or float32
		offset := 0
		for offset < len(data) {
			n, err := elt.read(out, data[offset:])
			if err != nil {
				return 0, err
			}
			offset += n
		}
		return offset, nil

	case FieldTypeStruct, FieldTypeSlice, FieldTypeString:
		count, n := plenccore.ReadVarUint(data)
		if n < 0 {
			return 0, fmt.Errorf("corrupt data looking for WTSlice count")
		}
		offset := n
		for i := range int(count) {
			if offset >= len(data) {
				return 0, fmt.Errorf("corrupt data looking for length of slice entry %d", i)
			}
			s, n := plenccore.ReadVarUint(data[offset:])
			if n <= 0 {
				return 0, fmt.Errorf("invalid varint for slice entry %d", i)
			}
			offset += n
			if s == 0 {
				continue
			}
			end := offset + int(s)
			if end > len(data) {
				return 0, fmt.Errorf("corrupt data reading slice entry %d", i)
			}

			n, err := elt.read(out, data[offset:offset+int(s)])
			if err != nil {
				return 0, err
			}
			offset += n
		}

		return offset, nil

	default:
		return 0, fmt.Errorf("slice of unexpected element types %s", elt.Type)
	}
}

func (d *Descriptor) readAsMapEntry(out Outputter, data []byte) (n int, err error) {
	if d.Elements[0].Type != FieldTypeString {
		// map keys have to be strings to be valid JSON. So we'll output as a
		// struct instead
		return
	}

	l := len(data)

	var offset int
	for offset < l {
		wt, index, n := plenccore.ReadTag(data[offset:])
		offset += n

		var elt *Descriptor
		for i := range d.Elements {
			candidate := &d.Elements[i]
			if candidate.Index == index {
				elt = candidate
				break
			}
		}

		if elt == nil {
			// Field corresponding to index does not exist
			n, err := plenccore.Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d in %s: %w", index, d.Name, err)
			}
			offset += n
			continue
		}

		fl := l
		if wt == plenccore.WTLength {
			// For WTLength types we read out the length and ensure the data we
			// read the field from is the right length
			v, n := plenccore.ReadVarUint(data[offset:])
			if n <= 0 {
				return 0, fmt.Errorf("varuint overflow reading field %d of %s", index, d.Name)
			}
			offset += n
			fl = int(v) + offset
			if fl > l {
				return 0, fmt.Errorf("length %d of field %d of %s exceeds data length", fl, index, d.Name)
			}
		}

		n, err := elt.read(out, data[offset:fl])
		if err != nil {
			return 0, fmt.Errorf("failed reading field %d(%s) of %s. %w", index, elt.Name, d.Name, err)
		}
		offset += n
	}

	return offset, nil
}

func (d *Descriptor) readAsStruct(out Outputter, data []byte) (n int, err error) {
	l := len(data)

	var offset int
	for offset < l {
		wt, index, n := plenccore.ReadTag(data[offset:])
		offset += n

		var elt *Descriptor
		for i := range d.Elements {
			candidate := &d.Elements[i]
			if candidate.Index == index {
				elt = candidate
				break
			}
		}

		if elt == nil {
			// Field corresponding to index does not exist
			n, err := plenccore.Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d in %s: %w", index, d.Name, err)
			}
			offset += n
			continue
		}

		fl := l
		if wt == plenccore.WTLength {
			// For WTLength types we read out the length and ensure the data we
			// read the field from is the right length
			v, n := plenccore.ReadVarUint(data[offset:])
			if n <= 0 {
				return 0, fmt.Errorf("varuint overflow reading field %d of %s", index, d.Name)
			}
			offset += n
			fl = int(v) + offset
			if fl > l {
				return 0, fmt.Errorf("length %d of field %d of %s exceeds data length", fl, index, d.Name)
			}
		}

		out.NameField(elt.Name)
		n, err := elt.read(out, data[offset:fl])
		if err != nil {
			return 0, fmt.Errorf("failed reading field %d(%s) of %s. %w", index, elt.Name, d.Name, err)
		}
		offset += n
	}

	return offset, nil
}

// readAsJSON reads data from JSON objects and arrays. Both are implemented as
// slices of structs. The structs are name, value type and value. In the array
// case the name is omitted from each entry
func (d *Descriptor) readAsJSON(out Outputter, data []byte) (n int, err error) {
	count, n := plenccore.ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt data looking for WTSlice count")
	}
	offset := n
	for i := range int(count) {
		// For each entry we have a string key, a value type and a value
		s, n := plenccore.ReadVarUint(data[offset:])
		if n <= 0 {
			return 0, fmt.Errorf("invalid varint for slice entry %d", i)
		}
		offset += n
		if s == 0 {
			continue
		}

		n, err := d.readJSONObjectKV(out, data[offset:offset+int(s)])
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}

func (d *Descriptor) readJSONObjectKV(out Outputter, data []byte) (n int, err error) {
	var (
		jType  jsonType
		offset int
	)

	for offset < len(data) {
		wt, index, n := plenccore.ReadTag(data[offset:])
		offset += n
		switch index {
		case 1:
			// When using this for reading arrays we simply don't see this index
			l, n := plenccore.ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("bad length on string field")
			}
			offset += n
			var key string

			n, err := StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(&key), wt)
			if err != nil {
				return 0, err
			}
			out.NameField(key)
			offset += n
		case 2:
			v, n := plenccore.ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("invalid map type field")
			}
			jType = jsonType(v)
			offset += n
		case 3:
			switch jType {
			case jsonTypeString:
				l, n := plenccore.ReadVarUint(data[offset:])
				if n < 0 {
					return 0, fmt.Errorf("bad length on string field")
				}
				offset += n
				var v string
				n, err := StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				out.String(v)
				offset += n

			case jsonTypeInt:
				var v int64
				n, err := IntCodec[int64]{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				out.Int64(v)

			case jsonTypeFloat:
				var v float64
				n, err := Float64Codec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				out.Float64(v)

			case jsonTypeBool:
				var v bool
				n, err := BoolCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				out.Bool(v)

			case jsonTypeArray:
				d := Descriptor{Type: FieldTypeJSONArray}
				n, err := d.read(out, data[offset:])
				if err != nil {
					return 0, err
				}
				offset += n

			case jsonTypeObject:
				d := Descriptor{Type: FieldTypeJSONObject}
				n, err := d.read(out, data[offset:])
				if err != nil {
					return 0, err
				}
				offset += n

			case jsonTypeNumber:
				l, n := plenccore.ReadVarUint(data[offset:])
				if n < 0 {
					return 0, fmt.Errorf("bad length on JSON number field")
				}
				offset += n
				var v json.Number
				n, err := StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				out.Raw(v.String())
				offset += n

			default:
				return 0, fmt.Errorf("unexpected json type %d", jType)
			}
		default:
			return 0, fmt.Errorf("unexpected json field index %d", index)
		}
	}

	return offset, nil
}
