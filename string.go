package plenc

import (
	"fmt"
	"reflect"
	"unsafe"
)

// ReadString reads a string from data
func ReadString(data []byte) (string, int) {
	l, n := ReadVarUint(data)
	if n < 0 {
		return "", n
	}
	return string(data[n : n+int(l)]), n + int(l)
}

// SizeString works out how much room you need to store a string
func SizeString(v string) int {
	return SizeVarUint(uint64(len(v))) + len(v)
}

// AppendString appends a string encoding to data
func AppendString(data []byte, v string) []byte {
	data = AppendVarUint(data, uint64(len(v)))
	return append(data, v...)
}

func init() {
	registerCodec(reflect.TypeOf(""), StringCodec{})
}

// StringCodec is a coddec for an string
type StringCodec struct{}

// Size returns the number of bytes needed to encode a string
func (StringCodec) Size(ptr unsafe.Pointer) int {
	return SizeString(*(*string)(ptr))
}

// Append encodes a string
func (StringCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendString(data, *(*string)(ptr))
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
