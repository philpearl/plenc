package plenccodec

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// Float64Codec is a coddec for a float64
type Float64Codec struct{}

// append encodes a float64
func (Float64Codec) append(data []byte, ptr unsafe.Pointer) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(*(*float64)(ptr)))
	return append(data, b[:]...)
}

// Read decodes a float64
func (Float64Codec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	if l := len(data); l < 8 {
		if l == 0 {
			*(*float64)(ptr) = 0
			return 0, nil
		}
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
func (c Float64Codec) WireType() plenccore.WireType {
	return plenccore.WT64
}

// Omit indicates whether this field should be omitted
func (c Float64Codec) Omit(ptr unsafe.Pointer) bool {
	return *(*float64)(ptr) == 0
}

func (c Float64Codec) Descriptor() Descriptor {
	return Descriptor{Type: FieldTypeFloat64}
}

func (c Float64Codec) Size(ptr unsafe.Pointer, tag []byte) int {
	return 8 + len(tag)
}

func (c Float64Codec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	data = append(data, tag...)
	return c.append(data, ptr)
}

// Float32Codec is a coddec for a float32
type Float32Codec struct{}

// append encodes a float32
func (Float32Codec) append(data []byte, ptr unsafe.Pointer) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(*(*float32)(ptr)))
	return append(data, b[:]...)
}

// Read decodes a float32
func (Float32Codec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	if l := len(data); l < 4 {
		if l == 0 {
			*(*float32)(ptr) = 0
			return 0, nil
		}
		return 0, fmt.Errorf("not enough data to read a float32. Have %d bytes", l)
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
func (c Float32Codec) WireType() plenccore.WireType {
	return plenccore.WT32
}

// Omit indicates whether this field should be omitted
func (c Float32Codec) Omit(ptr unsafe.Pointer) bool {
	return *(*float32)(ptr) == 0
}

func (c Float32Codec) Descriptor() Descriptor {
	return Descriptor{Type: FieldTypeFloat32}
}

func (c Float32Codec) Size(ptr unsafe.Pointer, tag []byte) int {
	return 4 + len(tag)
}

func (c Float32Codec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	data = append(data, tag...)
	return c.append(data, ptr)
}
