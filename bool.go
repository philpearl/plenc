package plenc

import (
	"fmt"
	"reflect"
	"unsafe"
)

// ReadBool reads a bool from data and returns it
func ReadBool(data []byte) (v bool, n int) {
	uv, n := ReadVarUint(data)
	return uv != 0, n
}

// SizeBool determines how many bytes it would take to encode the bool v
func SizeBool(v bool) int {
	return 1
}

// AppendBool appends a varint encoding of v to data. It returns the resulting slice
func AppendBool(data []byte, v bool) []byte {
	var uv uint64
	if v {
		uv = 1
	}
	return AppendVarUint(data, uv)
}

func init() {
	registerCodec(reflect.TypeOf(false), BoolCodec{})
}

// BoolCodec is a coddec for an bool
type BoolCodec struct{}

// Size returns the number of bytes needed to encode a bool
func (BoolCodec) Size(ptr unsafe.Pointer) int {
	return 1
}

// Append encodes a bool
func (BoolCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendBool(data, *(*bool)(ptr))
}

// Read decodes a bool
func (BoolCodec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	b, n := ReadBool(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*bool)(ptr) = b
	return n, nil
}

// New creates a pointer to a new bool
func (c BoolCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new(bool))
}

// WireType returns the wire type used to encode this type
func (c BoolCodec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c BoolCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}
