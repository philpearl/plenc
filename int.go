package philenc

import (
	"math/bits"
)

// ReadVarUint reads a varint from data and returns it
func ReadVarUint(data []byte) (v uint64, n int) {
	// return binary.Uvarint(data)
	if len(data) == 0 {
		return 0, 0
	}

	// This is somewhat careless about overflow or very long data
	for i, d := range data {
		v |= uint64(d&0x7F) << (7 * i)
		if d&0x80 == 0 {
			return v, i + 1
		}
	}
	panic("bad data")
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
