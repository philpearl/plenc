package plenc

import "fmt"

// WireType represents a protobuf wire type
type WireType int8

const (
	// WTVarInt signals a variable-length encoded integer. Signed integers are encoded with zig-zag encoding
	// first
	WTVarInt WireType = iota
	// WT64 signals a 64 bit integer. Used for float64
	WT64
	// WTLength signals length-value data. Length is encoded as a varint.
	WTLength
	wtStartGroupDeprecated
	wtEndGroupDeprecated
	// WT32 signals a 32 bit integer. Used for float32
	WT32
)

// ReadTag reads the wire type and field index from data
func ReadTag(data []byte) (wt WireType, index, n int) {
	v, n := ReadVarUint(data)
	wt = WireType(v & 0x7)
	index = int(v >> 3)

	// fmt.Println("tag", wt, index, n)
	return wt, index, n
}

// SizeTag determines the space needed to encode a tag
func SizeTag(wt WireType, index int) int {
	tag := uint64(index<<3) | uint64(wt)
	return SizeVarUint(tag)
}

// AppendTag encodes the tag and appends it to data
func AppendTag(data []byte, wt WireType, index int) []byte {
	tag := uint64(index<<3) | uint64(wt)
	return AppendVarUint(data, tag)
}

// Skip returns the size of a data item in the encoded data
func Skip(data []byte, wt WireType) (int, error) {
	switch wt {
	case WTVarInt:
		for i, v := range data {
			if v&0x80 == 0 {
				return i + 1, nil
			}
		}
		return 0, fmt.Errorf("unexpected end of data. %X", data)
	case WT64:
		return 8, nil
	case WTLength:
		l, n := ReadVarUint(data)
		return int(l) + n, nil
	case WT32:
		return 4, nil
	}
	return 0, fmt.Errorf("unsupported wire type %v", wt)
}
