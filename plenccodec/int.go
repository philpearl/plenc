package plenccodec

import (
	"fmt"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// IntCodec is a coddec for an int
type IntCodec[T int | int8 | int16 | int32 | int64] struct{}

// Size returns the number of bytes needed to encode a Int
func (IntCodec[T]) Size(ptr unsafe.Pointer) int {
	return plenccore.SizeVarInt(int64(*(*T)(ptr)))
}

// Append encodes a Int
func (IntCodec[T]) Append(data []byte, ptr unsafe.Pointer) []byte {
	return plenccore.AppendVarInt(data, int64(*(*T)(ptr)))
}

// Read decodes a Int
func (IntCodec[T]) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	i, n := plenccore.ReadVarInt(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*T)(ptr) = T(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c IntCodec[T]) New() unsafe.Pointer {
	return unsafe.Pointer(new(T))
}

// WireType returns the wire type used to encode this type
func (c IntCodec[T]) WireType() plenccore.WireType {
	return plenccore.WTVarInt
}

// Omit indicates whether this field should be omitted
func (c IntCodec[T]) Omit(ptr unsafe.Pointer) bool {
	return false
}

// UintCodec is a coddec for a uint
type UintCodec[T uint | uint8 | uint16 | uint32 | uint64] struct{}

// Size returns the number of bytes needed to encode a Int
func (UintCodec[T]) Size(ptr unsafe.Pointer) int {
	return plenccore.SizeVarUint(uint64(*(*T)(ptr)))
}

// Append encodes a Int
func (UintCodec[T]) Append(data []byte, ptr unsafe.Pointer) []byte {
	return plenccore.AppendVarUint(data, uint64(*(*T)(ptr)))
}

// Read decodes a Int
func (UintCodec[T]) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	i, n := plenccore.ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*T)(ptr) = T(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c UintCodec[T]) New() unsafe.Pointer {
	return unsafe.Pointer(new(T))
}

// WireType returns the wire type used to encode this type
func (c UintCodec[T]) WireType() plenccore.WireType {
	return plenccore.WTVarInt
}

// Omit indicates whether this field should be omitted
func (c UintCodec[T]) Omit(ptr unsafe.Pointer) bool {
	return false
}
