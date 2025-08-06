package plenccodec

import (
	"unique"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// StringCodec is a codec for an string
type StringCodec struct{}

// size returns the number of bytes needed to encode a string
func (StringCodec) size(ptr unsafe.Pointer) int {
	return len(*(*string)(ptr))
}

// append encodes a string
func (StringCodec) append(data []byte, ptr unsafe.Pointer) []byte {
	s := *(*string)(ptr)
	return append(data, s...)
}

// Read decodes a string
func (StringCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	*(*string)(ptr) = string(data)
	return len(data), nil
}

// New creates a pointer to a new string header
func (StringCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new(string))
}

// WireType returns the wire type used to encode this type
func (StringCodec) WireType() plenccore.WireType {
	return plenccore.WTLength
}

// Omit indicates whether this field should be omitted
func (StringCodec) Omit(ptr unsafe.Pointer) bool {
	return len(*(*string)(ptr)) == 0
}

func (StringCodec) Descriptor() Descriptor {
	return Descriptor{Type: FieldTypeString}
}

func (c StringCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	l := c.size(ptr)
	if len(tag) > 0 {
		l += len(tag) + plenccore.SizeVarUint(uint64(l))
	}
	return l
}

func (c StringCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	if len(tag) != 0 {
		data = append(data, tag...)
		data = plenccore.AppendVarUint(data, uint64(c.size(ptr)))
	}
	return c.append(data, ptr)
}

// BytesCodec is a codec for a byte slice
type BytesCodec struct{}

// size returns the number of bytes needed to encode a string
func (BytesCodec) size(ptr unsafe.Pointer) int {
	return len(*(*[]byte)(ptr))
}

// append encodes a []byte
func (BytesCodec) append(data []byte, ptr unsafe.Pointer) []byte {
	s := *(*[]byte)(ptr)
	return append(data, s...)
}

// Read decodes a []byte
func (BytesCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	// really must copy this data to be safe from the underlying buffer changing
	// later
	*(*[]byte)(ptr) = append([]byte(nil), data...)
	return len(data), nil
}

// New creates a pointer to a new bool
func (c BytesCodec) New() unsafe.Pointer {
	return unsafe.Pointer(new([]byte))
}

// WireType returns the wire type used to encode this type
func (c BytesCodec) WireType() plenccore.WireType {
	return plenccore.WTLength
}

// Omit indicates whether this field should be omitted
func (c BytesCodec) Omit(ptr unsafe.Pointer) bool {
	return len(*(*[]byte)(ptr)) == 0
}

func (BytesCodec) Descriptor() Descriptor {
	return Descriptor{Type: FieldTypeString}
}

func (c BytesCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	l := c.size(ptr)
	if len(tag) != 0 {
		l += len(tag) + plenccore.SizeVarUint(uint64(l))
	}
	return l
}

func (c BytesCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	if len(tag) != 0 {
		data = append(data, tag...)
		data = plenccore.AppendVarUint(data, uint64(c.size(ptr)))
	}
	return c.append(data, ptr)
}

type Interner interface {
	WithInterning() Codec
}

type InternedStringCodec struct {
	StringCodec
}

func (c InternedStringCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	// Note this will copy the string if it stores it, so we can do this unsafe trick
	// without worrying about the underlying data changing.
	s := unique.Make(unsafe.String(unsafe.SliceData(data), len(data))).Value()

	*(*string)(ptr) = s
	return len(data), nil
}

var icCodec Codec = InternedStringCodec{}

func (c StringCodec) WithInterning() Codec {
	return icCodec
}
