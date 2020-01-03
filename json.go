package plenc

import (
	"encoding/json"
	"fmt"
	"unsafe"
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
type JSONMapCodec struct {
}

// JSONArrayCodec is for serialising JSON arrays encoded as []interface{}
type JSONArrayCodec struct {
}

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
	m := *(*map[string]interface{})(ptr)
	// We'll use the WTSlice wire type, so first is the number of items
	size = SizeVarUint(uint64(len(m)))
	for k, v := range m {
		// With WTSlice each item is preceeded by its length
		itemSize := c.sizeKV(k, v)
		size += SizeVarUint(uint64(itemSize)) + itemSize
	}
	return size
}

func (c JSONMapCodec) sizeKV(k string, v interface{}) (size int) {
	size += SizeTag(StringCodec{}.WireType(), 1)
	size += SizeVarUint(uint64(len(k)))
	size += len(k)
	return size + sizeJSONValue(v)
}

func (c JSONMapCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	m := *(*map[string]interface{})(ptr)

	// First the number of items
	data = AppendVarUint(data, uint64(len(m)))

	// Next each item preceeded by its length
	for k, v := range m {
		s := c.sizeKV(k, v)
		data = AppendVarUint(data, uint64(s))
		data = c.appendKV(data, k, v)
	}

	return data
}

func (c JSONMapCodec) appendKV(data []byte, k string, v interface{}) []byte {
	data = AppendTag(data, StringCodec{}.WireType(), 1)
	data = AppendVarUint(data, uint64(len(k)))
	data = StringCodec{}.Append(data, unsafe.Pointer(&k))
	return appendJSONValue(data, v)
}

func (c JSONMapCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	count, n := ReadVarUint(data)
	offset := n

	m := *(*map[string]interface{})(ptr)
	if m == nil {
		m = make(map[string]interface{}, count)
		*(*map[string]interface{})(ptr) = m
	}

	for ; count > 0; count-- {
		l, n := ReadVarUint(data[offset:])
		if n < 0 {
			return 0, fmt.Errorf("Bad length in map")
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

func (c JSONMapCodec) WireType() WireType { return WTSlice }

func (c JSONArrayCodec) Omit(ptr unsafe.Pointer) bool {
	return (ptr == nil) || (len(*(*[]interface{})(ptr)) == 0)
}
func (c JSONArrayCodec) Size(ptr unsafe.Pointer) (size int) {
	a := *(*[]interface{})(ptr)
	size = SizeVarUint(uint64(len(a)))
	// Each entry is encoded preceeded by its length
	for _, val := range a {
		itemSize := sizeJSONValue(val)
		size += SizeVarUint(uint64(itemSize)) + itemSize
	}
	return size
}

func (c JSONArrayCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	a := *(*[]interface{})(ptr)
	data = AppendVarUint(data, uint64(len(a)))
	// Each entry is encoded preceeded by its length
	for _, val := range a {
		itemSize := sizeJSONValue(val)
		data = AppendVarUint(data, uint64(itemSize))
		data = appendJSONValue(data, val)
	}
	return data
}

func (c JSONArrayCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	count, n := ReadVarUint(data)
	offset := n

	a := *(*[]interface{})(ptr)
	if a == nil {
		a = make([]interface{}, count)
		*(*[]interface{})(ptr) = a
	}

	for i := range a {
		l, n := ReadVarUint(data[offset:])
		if n < 0 {
			return 0, fmt.Errorf("Bad length in map")
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

func (c JSONArrayCodec) WireType() WireType { return WTSlice }

func sizeJSONValue(v interface{}) (size int) {
	size += SizeTag(WTVarInt, 2)
	switch v := v.(type) {
	case nil:
		size += SizeVarUint(uint64(jsonTypeNil))
	case string:
		size += SizeVarUint(uint64(jsonTypeString))

		size += SizeTag(WTLength, 3)
		size += SizeVarUint(uint64(len(v)))
		size += StringCodec{}.Size(unsafe.Pointer(&v))
	case int:
		size += SizeVarUint(uint64(jsonTypeInt))
		size += SizeTag(WTVarInt, 3)
		size += IntCodec{}.Size(unsafe.Pointer(&v))
	case float64:
		size += SizeVarUint(uint64(jsonTypeFloat))
		size += SizeTag(WT64, 3)
		size += Float64Codec{}.Size(unsafe.Pointer(&v))
	case bool:
		size += SizeVarUint(uint64(jsonTypeBool))
		size += SizeTag(WTVarInt, 3)
		size += BoolCodec{}.Size(unsafe.Pointer(&v))
	case []interface{}:
		size += SizeVarUint(uint64(jsonTypeArray))
		size += SizeTag(JSONArrayCodec{}.WireType(), 3)
		size += JSONArrayCodec{}.Size(unsafe.Pointer(&v))
	case map[string]interface{}:
		size += SizeVarUint(uint64(jsonTypeObject))
		size += SizeTag(JSONMapCodec{}.WireType(), 3)
		size += JSONMapCodec{}.Size(unsafe.Pointer(&v))
	case json.Number:
		// Save this as a string
		size += SizeVarUint(uint64(jsonTypeNumber))
		size += SizeTag(WTLength, 3)
		size += SizeVarUint(uint64(len(v)))
		size += StringCodec{}.Size(unsafe.Pointer(&v))
	default:
		panic(fmt.Sprintf("unexpected json type %T", v))
	}

	return size
}

func appendJSONValue(data []byte, v interface{}) []byte {
	data = AppendTag(data, WTVarInt, 2)
	switch v := v.(type) {
	case nil:
		data = AppendVarUint(data, uint64(jsonTypeNil))
	case string:
		data = AppendVarUint(data, uint64(jsonTypeString))
		data = AppendTag(data, WTLength, 3)
		data = AppendVarUint(data, uint64(len(v)))
		data = StringCodec{}.Append(data, unsafe.Pointer(&v))
	case int:
		data = AppendVarUint(data, uint64(jsonTypeInt))
		data = AppendTag(data, WTVarInt, 3)
		data = IntCodec{}.Append(data, unsafe.Pointer(&v))
	case float64:
		data = AppendVarUint(data, uint64(jsonTypeFloat))
		data = AppendTag(data, WT64, 3)
		data = Float64Codec{}.Append(data, unsafe.Pointer(&v))
	case bool:
		data = AppendVarUint(data, uint64(jsonTypeBool))
		data = AppendTag(data, WTVarInt, 3)
		data = BoolCodec{}.Append(data, unsafe.Pointer(&v))
	case []interface{}:
		data = AppendVarUint(data, uint64(jsonTypeArray))
		data = AppendTag(data, JSONArrayCodec{}.WireType(), 3)
		data = JSONArrayCodec{}.Append(data, unsafe.Pointer(&v))
	case map[string]interface{}:
		data = AppendVarUint(data, uint64(jsonTypeObject))
		data = AppendTag(data, JSONMapCodec{}.WireType(), 3)
		data = JSONMapCodec{}.Append(data, unsafe.Pointer(&v))
	case json.Number:
		// Save this as a string
		data = AppendVarUint(data, uint64(jsonTypeNumber))
		data = AppendTag(data, WTLength, 3)
		data = AppendVarUint(data, uint64(len(v)))
		data = StringCodec{}.Append(data, unsafe.Pointer(&v))
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
		wt, index, n := ReadTag(data[offset:])
		offset += n
		switch index {
		case 1:
			// When using this for reading arrays we simply don't see this index
			l, n := ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("Bad length on string field")
			}
			offset += n

			n, err := StringCodec{}.Read(data[offset:offset+int(l)], unsafe.Pointer(key), wt)
			if err != nil {
				return 0, err
			}
			offset += n
		case 2:
			v, n := ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("invalid map type field")
			}
			jType = jsonType(v)
			offset += n
		case 3:
			switch jType {
			case jsonTypeString:
				l, n := ReadVarUint(data[offset:])
				if n < 0 {
					return 0, fmt.Errorf("Bad length on string field")
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
				n, err := IntCodec{}.Read(data[offset:], unsafe.Pointer(&v), wt)
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
				l, n := ReadVarUint(data[offset:])
				if n < 0 {
					return 0, fmt.Errorf("Bad length on JSON number field")
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
