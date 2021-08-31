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
type IntCodec struct{}

// Size returns the number of bytes needed to encode a Int
func (IntCodec) Size(ptr unsafe.Pointer) int {
	return SizeVarInt(int64(*(*int)(ptr)))
}

// Append encodes a Int
func (IntCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarInt(data, int64(*(*int)(ptr)))
}

// Read decodes a Int
func (IntCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarInt(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*int)(ptr) = int(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c IntCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new(int))
}

// WireType returns the wire type used to encode this type
func (c IntCodec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c IntCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Int64Codec is a coddec for an int
type Int64Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Int64Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarInt(int64(*(*int64)(ptr)))
}

// Append encodes a int64
func (Int64Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarInt(data, int64(*(*int64)(ptr)))
}

// Read decodes a int64
func (Int64Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarInt(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*int64)(ptr) = int64(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Int64Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(int64))
}

// WireType returns the wire type used to encode this type
func (c Int64Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Int64Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Int32Codec is a coddec for an int
type Int32Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Int32Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarInt(int64(*(*int32)(ptr)))
}

// Append encodes a int32
func (Int32Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarInt(data, int64(*(*int32)(ptr)))
}

// Read decodes a int32
func (Int32Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarInt(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}

	*(*int32)(ptr) = int32(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Int32Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(int32))
}

// WireType returns the wire type used to encode this type
func (c Int32Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Int32Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Int16Codec is a coddec for an int
type Int16Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Int16Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarInt(int64(*(*int16)(ptr)))
}

// Append encodes a int16
func (Int16Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarInt(data, int64(*(*int16)(ptr)))
}

// Read decodes a int16
func (Int16Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarInt(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}

	*(*int16)(ptr) = int16(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Int16Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(int16))
}

// WireType returns the wire type used to encode this type
func (c Int16Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Int16Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Int8Codec is a coddec for an int
type Int8Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Int8Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarInt(int64(*(*int8)(ptr)))
}

// Append encodes a int8
func (Int8Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarInt(data, int64(*(*int8)(ptr)))
}

// Read decodes a int8
func (Int8Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarInt(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}

	*(*int8)(ptr) = int8(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Int8Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(int8))
}

// WireType returns the wire type used to encode this type
func (c Int8Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Int8Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// UintCodec is a coddec for a uint
type UintCodec struct{}

// Size returns the number of bytes needed to encode a Int
func (UintCodec) Size(ptr unsafe.Pointer) int {
	return SizeVarUint(uint64(*(*uint)(ptr)))
}

// Append encodes a Int
func (UintCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarUint(data, uint64(*(*uint)(ptr)))
}

// Read decodes a Int
func (UintCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*uint)(ptr) = uint(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c UintCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new(uint))
}

// WireType returns the wire type used to encode this type
func (c UintCodec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c UintCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Uint64Codec is a coddec for a uint
type Uint64Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Uint64Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarUint(uint64(*(*uint64)(ptr)))
}

// Append encodes a Int
func (Uint64Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarUint(data, uint64(*(*uint64)(ptr)))
}

// Read decodes a Int
func (Uint64Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*uint64)(ptr) = uint64(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Uint64Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(uint64))
}

// WireType returns the wire type used to encode this type
func (c Uint64Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Uint64Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Uint32Codec is a coddec for a uint
type Uint32Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Uint32Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarUint(uint64(*(*uint32)(ptr)))
}

// Append encodes a Int
func (Uint32Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarUint(data, uint64(*(*uint32)(ptr)))
}

// Read decodes a Int
func (Uint32Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*uint32)(ptr) = uint32(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Uint32Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(uint32))
}

// WireType returns the wire type used to encode this type
func (c Uint32Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Uint32Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Uint16Codec is a coddec for a uint16
type Uint16Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Uint16Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarUint(uint64(*(*uint16)(ptr)))
}

// Append encodes a Int
func (Uint16Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarUint(data, uint64(*(*uint16)(ptr)))
}

// Read decodes a Int
func (Uint16Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*uint16)(ptr) = uint16(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Uint16Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(uint16))
}

// WireType returns the wire type used to encode this type
func (c Uint16Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Uint16Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Uint8Codec is a coddec for a uint16
type Uint8Codec struct{}

// Size returns the number of bytes needed to encode a Int
func (Uint8Codec) Size(ptr unsafe.Pointer) int {
	return SizeVarUint(uint64(*(*uint8)(ptr)))
}

// Append encodes a Int
func (Uint8Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	return AppendVarUint(data, uint64(*(*uint8)(ptr)))
}

// Read decodes a Int
func (Uint8Codec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	i, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt var int")
	}
	*(*uint8)(ptr) = uint8(i)
	return n, nil
}

// New creates a pointer to a new Int
func (c Uint8Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(uint8))
}

// WireType returns the wire type used to encode this type
func (c Uint8Codec) WireType() WireType {
	return WTVarInt
}

// Omit indicates whether this field should be omitted
func (c Uint8Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}
