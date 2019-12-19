package plenc

import (
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
	return len(*(*string)(ptr))
}

// Append encodes a string
func (StringCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	s := *(*string)(ptr)
	return append(data, s...)
}

// Read decodes a string
func (StringCodec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	*(*string)(ptr) = string(data)
	return len(data), nil
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
