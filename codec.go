package plenc

import (
	"fmt"
	"unsafe"
)

// Codec is implmented by types that can encode and decode themselves.
type Codec interface {
	Omit(ptr unsafe.Pointer) bool
	Size(ptr unsafe.Pointer) (size int)
	Append(data []byte, ptr unsafe.Pointer) []byte
	Read(data []byte, ptr unsafe.Pointer) (n int, err error)
	New() unsafe.Pointer
	WireType() WireType
}

// PointerWrapper wraps a codec so it can be used for a pointer to the type
type PointerWrapper struct {
	Underlying Codec
}

func (p PointerWrapper) Omit(ptr unsafe.Pointer) bool {
	t := *(*unsafe.Pointer)(ptr)
	if t == nil {
		return true
	}
	return p.Underlying.Omit(t)
}

func (p PointerWrapper) Size(ptr unsafe.Pointer) (size int) {
	t := *(*unsafe.Pointer)(ptr)
	if t == nil {
		return 0
	}
	return p.Underlying.Size(t)
}

func (p PointerWrapper) Append(data []byte, ptr unsafe.Pointer) []byte {
	t := *(*unsafe.Pointer)(ptr)
	if t == nil {
		return data
	}

	return p.Underlying.Append(data, t)
}

func (p PointerWrapper) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	t := (*unsafe.Pointer)(ptr)
	if *t == nil {
		*t = p.Underlying.New()
	}

	return p.Underlying.Read(data, *t)
}

func (p PointerWrapper) New() unsafe.Pointer {
	v := p.Underlying.New()
	return unsafe.Pointer(&v)
}

func (p PointerWrapper) WireType() WireType {
	return p.Underlying.WireType()
}

type SliceWrapper struct {
	Underlying Codec
	EltSize    uintptr
}

// sliceheader is a safer version of reflect.SliceHeader. The Data field here is a pointer that GC will track.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func (c SliceWrapper) Omit(ptr unsafe.Pointer) bool {
	h := *(*sliceHeader)(ptr)
	return h.Len == 0
}

func (c SliceWrapper) Size(ptr unsafe.Pointer) int {
	size := c.InternalSize(ptr)
	return SizeVarUint(uint64(size)) + size
}

func (c SliceWrapper) InternalSize(ptr unsafe.Pointer) int {
	h := *(*sliceHeader)(ptr)
	size := 0
	for i := 0; i < h.Len; i++ {
		size += c.Underlying.Size(unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize))
	}
	return size
}

// Append encodes the slice without the tag
func (c SliceWrapper) Append(data []byte, ptr unsafe.Pointer) []byte {

	// Length tags are followed by a byte count
	data = AppendVarUint(data, uint64(c.InternalSize(ptr)))

	h := *(*sliceHeader)(ptr)
	for i := 0; i < h.Len; i++ {
		// TODO: If we do strict protobuf then structs should also have the tag repeated
		data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
	}

	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
func (c SliceWrapper) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	l, n := ReadVarUint(data)
	if n <= 0 {
		return 0, fmt.Errorf("slice length is corrupt")
	}
	data = data[n:]

	// We step forward through out data to count how many things are in the slice
	var offset, count int
	for offset < int(l) {
		n, err := Skip(data[offset:], c.Underlying.WireType())
		if err != nil {
			return 0, err
		}
		offset += n
		count++
	}

	// Now make sure we have enough data in the slice
	h := (*sliceHeader)(ptr)
	if h.Cap < count {
		// Do some crazy shit so this slice is treated as if it contains pointers.
		// TODO: also try reflect.MakeSlice. Might be better if it isn't slower
		slice := make([]unsafe.Pointer, 1+count*int(c.EltSize)/int(unsafe.Sizeof(unsafe.Pointer(nil))))
		*h = *(*sliceHeader)(unsafe.Pointer(&slice))
		h.Cap = count
	}
	h.Len = count

	offset = 0
	for i := 0; i < h.Len; i++ {
		n, err := c.Underlying.Read(data[offset:], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return n + int(l), nil
}

func (c SliceWrapper) New() unsafe.Pointer {
	return unsafe.Pointer(&sliceHeader{})
}

func (c SliceWrapper) WireType() WireType {
	return WTLength
}
