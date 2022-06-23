package plencjson

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/philpearl/plenc/plenccodec"
	"github.com/philpearl/plenc/plenccore"
)

// One map we handle is one that can deal with JSON map[string]interface{}. In
// this case the value is either nil, string, integer, float64, bool, array (of
// these types) or object (another map[string]interface). We would need to
// encode the value type as the standard wire types of proto don't encode enough
// info. Encode as a list of key value pairs with a value type field. There's
// some unfortunate overlap between the types we need and wire types, but it's
// probably better to make consistent use of the proto encoding?? Why is this
// better than just serialising with JSON and dumping the bytes? It's probably a
// bit denser has faster to parse.

// JSONMapCodec is for serialising JSON maps encoded in Go as
// map[string]interface{}. To use this codec you must register it for use with
// map[string]interface{} or a named map[string]interface{} type
type JSONMapCodec struct{}

// JSONArrayCodec is for serialising JSON arrays encoded as []interface{}
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

func (c JSONMapCodec) Size(ptr unsafe.Pointer) (size int) {
	// this is just a map pointer here!
	var m map[string]interface{}
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

func (c JSONMapCodec) sizeKV(k string, v interface{}) (size int) {
	size += plenccore.SizeTag(plenccodec.StringCodec{}.WireType(), 1)
	size += plenccore.SizeVarUint(uint64(len(k)))
	size += len(k)
	return size + sizeJSONValue(v)
}

func (c JSONMapCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	// this is just a map pointer here!
	var m map[string]interface{}
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

func (c JSONMapCodec) appendKV(data []byte, k string, v interface{}) []byte {
	data = plenccore.AppendTag(data, plenccodec.StringCodec{}.WireType(), 1)
	data = plenccore.AppendVarUint(data, uint64(len(k)))
	data = plenccodec.StringCodec{}.Append(data, unsafe.Pointer(&k))
	return appendJSONValue(data, v)
}

func (c JSONMapCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	count, n := plenccore.ReadVarUint(data)
	if n == 0 {
		return 0, nil
	}
	offset := n

	m := *(*map[string]interface{})(ptr)
	if m == nil {
		m = make(map[string]interface{}, count)
		*(*map[string]interface{})(ptr) = m
	}

	for ; count > 0; count-- {
		l, n := plenccore.ReadVarUint(data[offset:])
		if n < 0 {
			return 0, fmt.Errorf("bad length in map")
		}
		offset += n
		var key string
		var val interface{}

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
	m := make(map[string]interface{})
	return unsafe.Pointer(&m)
}

func (c JSONMapCodec) WireType() plenccore.WireType { return plenccore.WTSlice }

func (c JSONArrayCodec) Omit(ptr unsafe.Pointer) bool {
	return (ptr == nil) || (len(*(*[]interface{})(ptr)) == 0)
}

func (c JSONArrayCodec) Size(ptr unsafe.Pointer) (size int) {
	a := *(*[]interface{})(ptr)
	size = plenccore.SizeVarUint(uint64(len(a)))
	// Each entry is encoded preceeded by its length
	for _, val := range a {
		itemSize := sizeJSONValue(val)
		size += plenccore.SizeVarUint(uint64(itemSize)) + itemSize
	}
	return size
}

func (c JSONArrayCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	a := *(*[]interface{})(ptr)
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

	a := *(*[]interface{})(ptr)
	if a == nil {
		a = make([]interface{}, count)
		*(*[]interface{})(ptr) = a
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
	return unsafe.Pointer(&[]interface{}{})
}

func (c JSONArrayCodec) WireType() plenccore.WireType { return plenccore.WTSlice }

func sizeJSONValue(v interface{}) (size int) {
	size += plenccore.SizeTag(plenccore.WTVarInt, 2)
	switch v := v.(type) {
	case nil:
		size += plenccore.SizeVarUint(uint64(jsonTypeNil))
	case string:
		size += plenccore.SizeVarUint(uint64(jsonTypeString))

		size += plenccore.SizeTag(plenccore.WTLength, 3)
		size += plenccore.SizeVarUint(uint64(len(v)))
		size += plenccodec.StringCodec{}.Size(unsafe.Pointer(&v))
	case int:
		size += plenccore.SizeVarUint(uint64(jsonTypeInt))
		size += plenccore.SizeTag(plenccore.WTVarInt, 3)
		size += plenccodec.IntCodec[int]{}.Size(unsafe.Pointer(&v))
	case float64:
		size += plenccore.SizeVarUint(uint64(jsonTypeFloat))
		size += plenccore.SizeTag(plenccore.WT64, 3)
		size += plenccodec.Float64Codec{}.Size(unsafe.Pointer(&v))
	case bool:
		size += plenccore.SizeVarUint(uint64(jsonTypeBool))
		size += plenccore.SizeTag(plenccore.WTVarInt, 3)
		size += plenccodec.BoolCodec{}.Size(unsafe.Pointer(&v))
	case []interface{}:
		size += plenccore.SizeVarUint(uint64(jsonTypeArray))
		size += plenccore.SizeTag(JSONArrayCodec{}.WireType(), 3)
		size += JSONArrayCodec{}.Size(unsafe.Pointer(&v))
	case map[string]interface{}:
		size += plenccore.SizeVarUint(uint64(jsonTypeObject))
		size += plenccore.SizeTag(JSONMapCodec{}.WireType(), 3)
		size += JSONMapCodec{}.Size(unsafe.Pointer(unpackEFace(v).data))
	case json.Number:
		// Save this as a string
		size += plenccore.SizeVarUint(uint64(jsonTypeNumber))
		size += plenccore.SizeTag(plenccore.WTLength, 3)
		size += plenccore.SizeVarUint(uint64(len(v)))
		size += plenccodec.StringCodec{}.Size(unsafe.Pointer(&v))
	default:
		panic(fmt.Sprintf("unexpected json type %T", v))
	}

	return size
}

func appendJSONValue(data []byte, v interface{}) []byte {
	data = plenccore.AppendTag(data, plenccore.WTVarInt, 2)
	switch v := v.(type) {
	case nil:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeNil))
	case string:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeString))
		data = plenccore.AppendTag(data, plenccore.WTLength, 3)
		data = plenccore.AppendVarUint(data, uint64(len(v)))
		data = plenccodec.StringCodec{}.Append(data, unsafe.Pointer(&v))
	case int:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeInt))
		data = plenccore.AppendTag(data, plenccore.WTVarInt, 3)
		data = plenccodec.IntCodec[int]{}.Append(data, unsafe.Pointer(&v))
	case float64:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeFloat))
		data = plenccore.AppendTag(data, plenccore.WT64, 3)
		data = plenccodec.Float64Codec{}.Append(data, unsafe.Pointer(&v))
	case bool:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeBool))
		data = plenccore.AppendTag(data, plenccore.WTVarInt, 3)
		data = plenccodec.BoolCodec{}.Append(data, unsafe.Pointer(&v))
	case []interface{}:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeArray))
		data = plenccore.AppendTag(data, JSONArrayCodec{}.WireType(), 3)
		data = JSONArrayCodec{}.Append(data, unsafe.Pointer(&v))
	case map[string]interface{}:
		data = plenccore.AppendVarUint(data, uint64(jsonTypeObject))
		data = plenccore.AppendTag(data, JSONMapCodec{}.WireType(), 3)
		data = JSONMapCodec{}.Append(data, unsafe.Pointer(unpackEFace(v).data))
	case json.Number:
		// Save this as a string
		data = plenccore.AppendVarUint(data, uint64(jsonTypeNumber))
		data = plenccore.AppendTag(data, plenccore.WTLength, 3)
		data = plenccore.AppendVarUint(data, uint64(len(v)))
		data = plenccodec.StringCodec{}.Append(data, unsafe.Pointer(&v))
	default:
		panic(fmt.Sprintf("unexpected json type %T", v))
	}

	return data
}

func readJSONKV(data []byte, key *string, val *interface{}) (n int, err error) {
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
				return 0, fmt.Errorf("Bad length on string field")
			}
			offset += n

			n, err := plenccodec.StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(key), wt)
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
					return 0, fmt.Errorf("Bad length on string field")
				}
				offset += n
				var v string
				n, err := plenccodec.StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				*val = v
				offset += n

			case jsonTypeInt:
				var v int
				n, err := plenccodec.IntCodec[int]{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeFloat:
				var v float64
				n, err := plenccodec.Float64Codec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeBool:
				var v bool
				n, err := plenccodec.BoolCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeArray:
				var v []interface{}
				n, err := JSONArrayCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
				if err != nil {
					return 0, err
				}
				offset += n
				*val = v

			case jsonTypeObject:
				var v map[string]interface{}
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
				n, err := plenccodec.StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(&v), wt)
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

type eface struct {
	rtype unsafe.Pointer
	data  unsafe.Pointer
}

func unpackEFace(obj interface{}) *eface {
	return (*eface)(unsafe.Pointer(&obj))
}
