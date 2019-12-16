package plenc

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

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

func init() {
	registerCodec(reflect.TypeOf(float64(0)), Float64Codec{})
	registerCodec(reflect.TypeOf(float32(0)), Float32Codec{})
}

// Alternate approach. A bit like JSONiter.
// - Can create adapters for types in other packages
// - Might be able to write neater templates. Can do everything by selecting the thingType
// - Perhaps can use reflection in a speedy way with offset tables
// - Could extend these to include array types
//
// - But not great for writing custom codecs on a type as can't just dispatch codec calls. Would need some kind
//   of registry
// - Can we use a type as a map key for selecting codecs?
// - Could be awkward for smaller int types. I guess we just need a codec for each one
// - Still have the full control/partial issue.
// - If we're building a struct we need to know a separate size for the fields within in
//
// thingType.Size(ptr unsafe.Pointer) int
// thingType.Append(data []byte, ptr unsafe.Pointer) []data
// thingType.Read(data []byte, ptr unsafe.Pointer) (n int, err error)
// thingType.New() unsafe.Pointer

// Float64Codec is a coddec for a float64
type Float64Codec struct{}

// Size returns the number of bytes needed to encode a float64
func (Float64Codec) Size(ptr unsafe.Pointer) int {
	return 8
}

// Append encodes a float64
func (Float64Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(*(*float64)(ptr)))
	return append(data, b[:]...)
}

// Read decodes a float64
func (Float64Codec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("not enough data to read a float64. Have %d bytes", len(data))
	}
	bits := binary.LittleEndian.Uint64(data)
	*(*float64)(ptr) = math.Float64frombits(bits)
	return 8, nil
}

// New creates a pointer to a new float64
func (c Float64Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(float64))
}

// WireType returns the wire type used to encode this type
func (c Float64Codec) WireType() WireType {
	return WT64
}

// Omit indicates whether this field should be omitted
func (c Float64Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}

// Float32Codec is a coddec for a float32
type Float32Codec struct{}

// Size returns the number of bytes needed to encode a float32
func (Float32Codec) Size(ptr unsafe.Pointer) int {
	return 4
}

// Append encodes a float32
func (Float32Codec) Append(data []byte, ptr unsafe.Pointer) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(*(*float32)(ptr)))
	return append(data, b[:]...)
}

// Read decodes a float32
func (Float32Codec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("not enough data to read a float32. Have %d bytes", len(data))
	}
	bits := binary.LittleEndian.Uint32(data)
	*(*float32)(ptr) = math.Float32frombits(bits)
	return 4, nil
}

// New creates a pointer to a new float32
func (c Float32Codec) New() unsafe.Pointer {
	return unsafe.Pointer(new(float32))
}

// WireType returns the wire type used to encode this type
func (c Float32Codec) WireType() WireType {
	return WT32
}

// Omit indicates whether this field should be omitted
func (c Float32Codec) Omit(ptr unsafe.Pointer) bool {
	return false
}
