package philenc

// ReadVarUint reads a varint from data and returns it
func ReadVarUint(data []byte) (v uint, n int) {
	for i, d := range data {
		v |= uint(d&0x7F) << (7 * i)
		if d&0x80 == 0 {
			return v, i + 1
		}
	}
	panic("bad data")
}

// SizeVarUint determines how many bytes it would take to encode the int v
func SizeVarUint(v uint) int {
	var size int
	for {
		size++
		v >>= 7
		if v == 0 {
			break
		}
	}
	return size
}

// AppendVarUint appends a varint encoding of v to data. It returns the resulting slice
func AppendVarUint(data []byte, v uint) []byte {
	b := byte(v & 0x7F)
	v >>= 7
	for v != 0 {
		b |= 0x80
		data = append(data, b)
		b = byte(v & 0x7F)
		v >>= 7
	}
	return append(data, b)
}

// ZigZag performs zig-zag encoding of an int into a uint. 0->0 -1->1 1->2, etc. So positive numbers are
// doubled
func ZigZag(v int) uint {
	return uint((v << 1) ^ (v >> 63))
}

// ZagZig reverses ZigZag
func ZagZig(v uint) int {
	return int(v>>1) ^ -int(v&1)
}

// ReadVarInt reads a signed int value from data
func ReadVarInt(data []byte) (v, n int) {
	u, n := ReadVarUint(data)
	return ZagZig(u), n
}

// SizeVarInt returns the number of bytes needed to encode v
func SizeVarInt(v int) int {
	return SizeVarUint(ZigZag(v))
}

// AppendVarInt encodes v as a varint and appends the result to data
func AppendVarInt(data []byte, v int) []byte {
	return AppendVarUint(data, ZigZag(v))
}
