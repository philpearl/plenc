package plenc

import (
	"fmt"
	"unsafe"
)

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

// SliceWrapper is a codec for a slice of a type.
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
	h := *(*sliceHeader)(ptr)
	size := 0
	if c.Underlying.WireType() == WTLength {
		for i := 0; i < h.Len; i++ {
			ptr := unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize)
			s := c.Underlying.Size(ptr)
			size += s + SizeVarUint(uint64(s))
		}
	} else {
		for i := 0; i < h.Len; i++ {
			size += c.Underlying.Size(unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize))
		}
	}

	return size
}

// Append encodes the slice without the tag
func (c SliceWrapper) Append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)

	if c.Underlying.WireType() == WTLength {
		for i := 0; i < h.Len; i++ {
			// TODO: If we do strict protobuf then structs should also have the tag repeated.
			// This is a half-way house!
			ptr := unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize)
			data = AppendVarUint(data, uint64(c.Underlying.Size(ptr)))
			data = c.Underlying.Append(data, ptr)
		}
	} else {
		for i := 0; i < h.Len; i++ {
			data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
		}

	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the corect size for the slice
func (c SliceWrapper) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	l := len(data)

	// We step forward through out data to count how many things are in the slice
	wt := c.Underlying.WireType()
	var offset, count int
	switch wt {
	case WTVarInt:
		for offset < l {
			_, n := ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("corrupt data")
			}
			offset += n
			count++
		}
	case WT64:
		count = l / 8
	case WTLength:
		for offset < l {
			ll, n := ReadVarUint(data[offset:])
			if n < 0 {
				return 0, fmt.Errorf("corrupt data")
			}
			offset += int(ll) + n
			count++
		}
	case WT32:
		count = l / 4
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
	if wt == WTLength {
		for i := 0; i < h.Len; i++ {
			s, n := ReadVarUint(data[offset:])
			if n <= 0 {
				return 0, fmt.Errorf("invalid varint for slice entry %d", i)
			}
			offset += n
			n, err := c.Underlying.Read(data[offset:offset+int(s)], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
			if err != nil {
				return 0, err
			}
			offset += n
		}
	} else {
		for i := 0; i < h.Len; i++ {
			n, err := c.Underlying.Read(data[offset:], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
			if err != nil {
				return 0, err
			}
			offset += n
		}
	}

	return offset, nil
}

func (c SliceWrapper) New() unsafe.Pointer {
	return unsafe.Pointer(&sliceHeader{})
}

func (c SliceWrapper) WireType() WireType {
	return WTLength
}
