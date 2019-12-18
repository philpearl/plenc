package plenc

import (
	"fmt"
	"reflect"
	"unsafe"
)

func init() {
	registerCodec(reflect.TypeOf(""), StringCodec{})
}

// StringCodec is a coddec for an string
type StringCodec struct{}

// Size returns the number of bytes needed to encode a string
func (StringCodec) Size(ptr unsafe.Pointer) int {
	l := len(*(*string)(ptr))
	return SizeVarUint(uint64(l)) + l
}

// Append encodes a string
func (StringCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	s := *(*string)(ptr)
	data = AppendVarUint(data, uint64(len(s)))
	return append(data, s...)
}

// Read decodes a string
func (StringCodec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	l, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*string)(ptr) = string(data[n : n+int(l)])
	return n + int(l), nil
}

// New creates a pointer to a new bool
func (c StringCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new(string))
}

// WireType returns the wire type used to encode this type
func (c StringCodec) WireType() WireType {
	return WTLength
}

// Omit indicates whether this field should be omitted
func (c StringCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}
