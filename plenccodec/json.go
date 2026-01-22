package plenccodec

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// One map we handle is one that can deal with JSON map[string]any. In
// this case the value is either nil, string, integer, float64, bool, array (of
// these types) or object (another map[string]interface). We would need to
// encode the value type as the standard wire types of proto don't encode enough
// info. Encode as a list of key value pairs with a value type field. There's
// some unfortunate overlap between the types we need and wire types, but it's
// probably better to make consistent use of the proto encoding?? Why is this
// better than just serialising with JSON and dumping the bytes? It's probably a
// bit denser has faster to parse.

// JSONMapCodec is for serialising JSON maps encoded in Go as
// map[string]any. To use this codec you must register it for use with
// map[string]any or a named map[string]any type
type JSONMapCodec struct{}

// JSONArrayCodec is for serialising JSON arrays encoded as []any
type JSONArrayCodec struct{}

type jsonType uint

const (
	jsonTypeNil jsonType = iota
	jsonTypeString
	jsonTypeInt
	jsonTypeFloat
	jsonTypeBool
	jsonTypeArray
	jsonTypeObject
	jsonTypeNumber
)

func (JSONMapCodec) Omit(ptr unsafe.Pointer) bool {
	return ptr == nil
}

func (c JSONMapCodec) size(ptr unsafe.Pointer) (size int) {
	// this is just a map pointer here!
	var m map[string]any
	*(*unsafe.Pointer)((unsafe.Pointer)(&m)) = ptr

	// We'll use the WTSlice wire type, so first is the number of items
	size = plenccore.SizeVarUint(uint64(len(m)))
	for k, v := range m {
		// With WTSlice each item is preceeded by its length
		itemSize := c.sizeKV(k, v)
		size += plenccore.SizeVarUint(uint64(itemSize)) + itemSize
	}
	return size
}

func (c JSONMapCodec) sizeKV(k string, v any) (size int) {
	size += plenccore.SizeTag(StringCodec{}.WireType(), 1)
	size += plenccore.SizeVarUint(uint64(len(k)))
	size += len(k)
	return size + sizeJSONValue(v)
}

func (c JSONMapCodec) append(data []byte, ptr unsafe.Pointer) []byte {
	// this is just a map pointer here!
	var m map[string]any
	*(*unsafe.Pointer)((unsafe.Pointer)(&m)) = ptr

	// First the number of items
	data = plenccore.AppendVarUint(data, uint64(len(m)))

	// Next each item preceeded by its length
	for k, v := range m {
		s := c.sizeKV(k, v)
		data = plenccore.AppendVarUint(data, uint64(s))
		data = c.appendKV(data, k, v)
	}

	return data
}

var keyTag = plenccore.AppendTag(nil, StringCodec{}.WireType(), 1)

func (c JSONMapCodec) appendKV(data []byte, k string, v any) []byte {
	data = StringCodec{}.Append(data, unsafe.Pointer(&k), keyTag)
	return appendJSONValue(data, v)
}

func (c JSONMapCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	count, n := plenccore.ReadVarUint(data)
	if n == 0 {
		return 0, nil
	}
	offset := n

	m := *(*map[string]any)(ptr)
	if m == nil {
		m = make(map[string]any, count)
		*(*map[string]any)(ptr) = m
	}

	for ; count > 0; count-- {
		l, n := plenccore.ReadVarUint(data[offset:])
		if n < 0 {
			return 0, fmt.Errorf("bad length in map")
		}
		offset += n
		var key string
		var val any

		n, err := readJSONKV(data[offset:offset+int(l)], &key, &val)
		if err != nil {
			return 0, err
		}
		offset += n
		m[key] = val
	}

	return offset, nil
}

func (c JSONMapCodec) New() unsafe.Pointer {
	m := make(map[string]any)
	return unsafe.Pointer(&m)
}

func (c JSONMapCodec) WireType() plenccore.WireType { return plenccore.WTSlice }

func (c JSONMapCodec) Descriptor() Descriptor {
	// This needs to be a special descriptor!
	return Descriptor{Type: FieldTypeJSONObject}
}

func (c JSONMapCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	return c.size(ptr) + len(tag)
}

func (c JSONMapCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	data = append(data, tag...)
	return c.append(data, ptr)
}

func (c JSONArrayCodec) Omit(ptr unsafe.Pointer) bool {
	return (ptr == nil) || (len(*(*[]any)(ptr)) == 0)
}

func (c JSONArrayCodec) size(ptr unsafe.Pointer) (size int) {
	a := *(*[]any)(ptr)
	size = plenccore.SizeVarUint(uint64(len(a)))
	// Each entry is encoded preceeded by its length
	for _, val := range a {
		itemSize := sizeJSONValue(val)
		size += plenccore.SizeVarUint(uint64(itemSize)) + itemSize
	}
	return size
}

func (c JSONArrayCodec) append(data []byte, ptr unsafe.Pointer) []byte {
	a := *(*[]any)(ptr)
	data = plenccore.AppendVarUint(data, uint64(len(a)))
	// Each entry is encoded preceeded by its length
	for _, val := range a {
		itemSize := sizeJSONValue(val)
		data = plenccore.AppendVarUint(data, uint64(itemSize))
		data = appendJSONValue(data, val)
	}
	return data
}

func (c JSONArrayCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	count, n := plenccore.ReadVarUint(data)
	offset := n

	a := *(*[]any)(ptr)
	if a == nil {
		a = make([]any, count)
		*(*[]any)(ptr) = a
	}

	for i := range a {
		l, n := plenccore.ReadVarUint(data[offset:])
		if n < 0 {
			return 0, fmt.Errorf("bad length in map")
		}
		offset += n

		n, err := readJSONKV(data[offset:offset+int(l)], nil, &a[i])
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}

func (c JSONArrayCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&[]any{})
}

func (c JSONArrayCodec) WireType() plenccore.WireType { return plenccore.WTSlice }

func (c JSONArrayCodec) Descriptor() Descriptor {
	// This needs to be a special descriptor!
	return Descriptor{Type: FieldTypeJSONArray}
}

func (c JSONArrayCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	return c.size(ptr) + len(tag)
}

func (c JSONArrayCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	data = append(data, tag...)
	return c.append(data, ptr)
}

func sizeJSONValue(v any) (size int) {
	size += plenccore.SizeTag(plenccore.WTVarInt, 2)
	switch v := v.(type) {
	case nil:
		size += plenccore.SizeVarUint(uint64(jsonTypeNil))
	case string:
		size += plenccore.SizeVarUint(uint64(jsonTypeString))

		size += StringCodec{}.Size(unsafe.Pointer(&v), valueWTLTag)
	case int:
		size += plenccore.SizeVarUint(uint64(jsonTypeInt))
		size += IntCodec[int]{}.Size(unsafe.Pointer(&v), valueWTVITag)
	case float64:
		size += plenccore.SizeVarUint(uint64(jsonTypeFloat))
		size += Float64Codec{}.Size(unsafe.Pointer(&v), valueWT64Tag)
	case bool:
		size += plenccore.SizeVarUint(uint64(jsonTypeBool))
		size += BoolCodec{}.Size(unsafe.Pointer(&v), valueWTVITag)
	case []any:
		size += plenccore.SizeVarUint(uint64(jsonTypeArray))
		size += JSONArrayCodec{}.Size(unsafe.Pointer(&v), valueWTSliceTag)
	case map[string]any:
		size += plenccore.SizeVarUint(uint64(jsonTypeObject))
		size += JSONMapCodec{}.Size(unsafe.Pointer(unpackEFace(v).data), valueWTSliceTag)
	case json.Number:
		// Save this as a string
		size += plenccore.SizeVarUint(uint64(jsonTypeNumber))
		size += StringCodec{}.Size(unsafe.Pointer(&v), valueWTLTag)
	default:
		panic(fmt.Sprintf("unexpected json type %T", v))
	}

	return size
}

var (
	valueWTLTag     = plenccore.AppendTag(nil, plenccore.WTLength, 3)
	valueWTVITag    = plenccore.AppendTag(nil, plenccore.WTVarInt, 3)
	valueWT64Tag    = plenccore.AppendTag(nil, plenccore.WT64, 3)
	valueWTSliceTag = plenccore.AppendTag(nil, plenccore.WTSlice, 3)
)

func appendJSONValue(data []byte, v any) []byte {
	data = plenccore.AppendTag(data, plenccore.WTVarInt, 2)
	switch v := v.(type) {
	case nil:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeNil))
	case string:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeString))
		data = StringCodec{}.Append(data, unsafe.Pointer(&v), valueWTLTag)
	case int:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeInt))
		data = IntCodec[int]{}.Append(data, unsafe.Pointer(&v), valueWTVITag)
	case float64:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeFloat))
		data = Float64Codec{}.Append(data, unsafe.Pointer(&v), valueWT64Tag)
	case bool:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeBool))
		data = BoolCodec{}.Append(data, unsafe.Pointer(&v), valueWTVITag)
	case []any:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeArray))
		data = JSONArrayCodec{}.Append(data, unsafe.Pointer(&v), valueWTSliceTag)
	case map[string]any:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeObject))
		data = JSONMapCodec{}.Append(data, unsafe.Pointer(unpackEFace(v).data), valueWTSliceTag)
	case json.Number:
		// Save this as a string
		data = plenccore.AppendVarUint(data, uint64(jsonTypeNumber))
		data = StringCodec{}.Append(data, unsafe.Pointer(&v), valueWTLTag)
	default:
		panic(fmt.Sprintf("unexpected json type %T", v))
	}

	return data
}

func readJSONKV(data []byte, key *string, val *any) (n int, err error) {
	var (
		jType  jsonType
		offset int
	)

	for offset < len(data) {
		wt, index, n := plenccore.ReadTag(data[offset:])
		if n <= 0 {
			return 0, fmt.Errorf("failed to read tag for json key-value")
		}
		offset += n
		switch index {
		case 1:
			// When using this for reading arrays we simply don't see this index
			l, n := plenccore.ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("bad length on string field")
			}
			offset += n

			n, err := StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(key), wt)
			if err != nil {
				return 0, err
			}
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
				*val = v
				offset += n

			case jsonTypeInt:
				var v int
				n, err := IntCodec[int]{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeFloat:
				var v float64
				n, err := Float64Codec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeBool:
				var v bool
				n, err := BoolCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeArray:
				var v []any
				n, err := JSONArrayCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeObject:
				var v map[string]any
				n, err := JSONMapCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

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
				*val = v
				offset += n

			default:
				return 0, fmt.Errorf("unexpected json type %d", jType)
			}
		}
	}

	return offset, nil
}
