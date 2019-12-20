package plenc

import (
	"reflect"
	"unsafe"
)

func init() {
	registerCodec(reflect.TypeOf(""), StringCodec{})
	registerCodec(reflect.TypeOf([]byte(nil)), BytesCodec{})
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
func (StringCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
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

// BytesCodec is a coddec for a byte slice
type BytesCodec struct{}

// Size returns the number of bytes needed to encode a string
func (BytesCodec) Size(ptr unsafe.Pointer) int {
	return len(*(*[]byte)(ptr))
}

// Append encodes a []byte
func (BytesCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	s := *(*[]byte)(ptr)
	return append(data, s...)
}

// Read decodes a []byte
func (BytesCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	// really must copy this data to be safe from the underlying buffer changing
	// later
	*(*[]byte)(ptr) = append([]byte(nil), data...)
	return len(data), nil
}

// New creates a pointer to a new bool
func (c BytesCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new([]byte))
}

// WireType returns the wire type used to encode this type
func (c BytesCodec) WireType() WireType {
	return WTLength
}

// Omit indicates whether this field should be omitted
func (c BytesCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}
