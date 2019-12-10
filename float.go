package φλenc

import "math"

import "encoding/binary"

// ReadFloat32 reads a float32
func ReadFloat32(data []byte) (float32, int) {
	bits := binary.LittleEndian.Uint32(data)
	return math.Float32frombits(bits), 4
}

// SizeFloat32 works out the size of an encoded float32
func SizeFloat32(v float32) int {
	return 4
}

// AppendFloat32 encodes the v and appends to data
func AppendFloat32(data []byte, v float32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(v))
	return append(data, b[:]...)
}

// ReadFloat64 reads a float64
func ReadFloat64(data []byte) (float64, int) {
	bits := binary.LittleEndian.Uint64(data)
	return math.Float64frombits(bits), 8
}

// SizeFloat64 works out the size of an encoded float64
func SizeFloat64(v float64) int {
	return 8
}

// AppendFloat64 encodes the v and appends to data
func AppendFloat64(data []byte, v float64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(v))
	return append(data, b[:]...)
}
