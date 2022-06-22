package plenc

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"unsafe"
)

// ReadVarUint reads a varint from data and returns it
func ReadVarUint(data []byte) (v uint64, n int) {
	return binary.Uvarint(data)
}

// SizeVarUint determines how many bytes it would take to encode the int v
func SizeVarUint(v uint64) int {
	if v < 0x80 {
		return 1
	}
	bits := bits.Len64(v)
	return (bits + 6) / 7
}

// AppendVarUint appends a varint encoding of v to data. It returns the resulting slice
func AppendVarUint(data []byte, v uint64) []byte {
	for v >= 0x80 {
		data = append(data, byte(v)|0x80)
		v >>= 7
	}
	return append(data, byte(v))
}

// ZigZag performs zig-zag encoding of an int into a uint. 0->0 -1->1 1->2, etc. So positive numbers are
// doubled
func ZigZag(v int64) uint64 {
	return uint64((v << 1) ^ (v >> 63))
}

// ZagZig reverses ZigZag
func ZagZig(v uint64) int64 {
	return int64(v>>1) ^ -int64(v&1)
}

// ReadVarInt reads a signed int value from data
func ReadVarInt(data []byte) (v int64, n int) {
	u, n := ReadVarUint(data)
	return ZagZig(u), n
}

// SizeVarInt returns the number of bytes needed to encode v
func SizeVarInt(v int64) int {
	return SizeVarUint(ZigZag(v))
}

// AppendVarInt encodes v as a varint and appends the result to data
func AppendVarInt(data []byte, v int64) []byte {
	return AppendVarUint(data, ZigZag(v))
}

// IntCodec is a coddec for an int
type IntCodec[T int | int8 | int16 | int32 | int64] struct{}

// Size returns the number of bytes needed to encode a Int
func (IntCodec[T]) Size(ptr unsafe.Pointer) int {
	return SizeVarInt(int64(*(*T)(ptr)))
}

// Append encodes a Int
func (IntCodec[T]) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarInt(data, int64(*(*T)(ptr)))
}

// Read decodes a Int
func (IntCodec[T]) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarInt(data)
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
func (c IntCodec[T]) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c IntCodec[T]) Omit(ptr unsafe.Pointer) bool {
	return false
}

// UintCodec is a coddec for a uint
type UintCodec[T uint | uint8 | uint16 | uint32 | uint64] struct{}

// Size returns the number of bytes needed to encode a Int
func (UintCodec[T]) Size(ptr unsafe.Pointer) int {
	return SizeVarUint(uint64(*(*T)(ptr)))
}

// Append encodes a Int
func (UintCodec[T]) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarUint(data, uint64(*(*T)(ptr)))
}

// Read decodes a Int
func (UintCodec[T]) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarUint(data)
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
func (c UintCodec[T]) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c UintCodec[T]) Omit(ptr unsafe.Pointer) bool {
	return false
}
