package plenc

import (
	"fmt"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// BoolCodec is a codec for a bool
type BoolCodec struct{}

// Size returns the number of bytes needed to encode a bool
func (BoolCodec) Size(ptr unsafe.Pointer) int {
	return 1
}

// Append encodes a bool
func (BoolCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	var uv uint64
	if *(*bool)(ptr) {
		uv = 1
	}
	return plenccore.AppendVarUint(data, uv)
}

// Read decodes a bool
func (BoolCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	uv, n := plenccore.ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*bool)(ptr) = (uv != 0)
	return n, nil
}

// New creates a pointer to a new bool
func (c BoolCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new(bool))
}

// WireType returns the wire type used to encode this type
func (c BoolCodec) WireType() plenccore.WireType {
	return plenccore.WTVarInt
}

// Omit indicates whether this field should be omitted
func (c BoolCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}
