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

func (p PointerWrapper) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	t := (*unsafe.Pointer)(ptr)
	if *t == nil {
		*t = p.Underlying.New()
	}

	return p.Underlying.Read(data, *t, wt)
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

type baseSliceWrapper struct {
	Underlying Codec
	EltSize    uintptr
	EltType    unsafe.Pointer
}

func (c baseSliceWrapper) Omit(ptr unsafe.Pointer) bool {
	h := *(*sliceHeader)(ptr)
	return h.Len == 0
}

func (c baseSliceWrapper) New() unsafe.Pointer {
	return unsafe.Pointer(&sliceHeader{})
}

func (c baseSliceWrapper) WireType() WireType {
	return WTLength
}

// WTLengthSliceWrapper is a codec for a slice of a type that's encoded using
// the WTSlice wire type.
type WTLengthSliceWrapper struct {
	baseSliceWrapper
}

func (c WTLengthSliceWrapper) Size(ptr unsafe.Pointer) int {
	h := *(*sliceHeader)(ptr)
	size := SizeVarUint(uint64(h.Len))
	for i := 0; i < h.Len; i++ {
		s := c.Underlying.Size(unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize))
		size += s + SizeVarUint(uint64(s))
	}
	return size
}

// Append encodes the slice, and appends the encoded version to data
func (c WTLengthSliceWrapper) Append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)

	// Append the count of items in the slice
	data = AppendVarUint(data, uint64(h.Len))
	// Append each of the items. They're all prefixed by their length
	for i := 0; i < h.Len; i++ {
		ptr := unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize)
		data = AppendVarUint(data, uint64(c.Underlying.Size(ptr)))
		data = c.Underlying.Append(data, ptr)
	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the corect size for the slice
func (c WTLengthSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	if wt == WTLength {
		return c.readAsWTLength(data, ptr)
	}

	// First we read the number of items in the slice
	count, n := ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt data looking for WTSlice count")
	}

	// Now make sure we have enough capacity in the slice
	h := (*sliceHeader)(ptr)
	if h.Cap < int(count) {
		// Ensure the GC knows the type of this slice.
		h.Data = unsafe_NewArray(c.EltType, int(count))
		h.Cap = int(count)
	}
	h.Len = int(count)

	offset := n
	for i := 0; i < h.Len; i++ {
		s, n := ReadVarUint(data[offset:])
		if n <= 0 {
			return 0, fmt.Errorf("invalid varint for slice entry %d", i)
		}
		offset += n
		n, err := c.Underlying.Read(data[offset:offset+int(s)], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), WTLength)
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}

// readAsWTLength is here for protobuf compatibility. protobuf writes certain array types by simply repeating the encoding for an individual field. So here we just read one underlying value and append it to the slice
func (c WTLengthSliceWrapper) readAsWTLength(data []byte, ptr unsafe.Pointer) (n int, err error) {
	h := (*sliceHeader)(ptr)
	if h.Cap == h.Len {
		// Need to make room
		cap := h.Cap * 2
		if cap == 0 {
			cap = 8
		}
		nh := sliceHeader{
			Data: unsafe_NewArray(c.EltType, int(cap)),
			Len:  h.Len,
			Cap:  cap,
		}
		if h.Len != 0 {
			// copy over the old data
			typedslicecopy(c.EltType, nh, *h)
		}
		nh.Len = h.Len
		nh.Cap = cap

		*h = nh
	}

	n, err = c.Underlying.Read(data, unsafe.Pointer(uintptr(h.Data)+uintptr(h.Len)*c.EltSize), WTLength)
	if err != nil {
		return 0, err
	}
	h.Len++
	return n, nil
}

func (c WTLengthSliceWrapper) WireType() WireType {
	return WTSlice
}

// WTFixedSliceWrapper is a codec for a type that's encoded as a fixed 32 or 64
// byte value (i.e. float32 or float64)
type WTFixedSliceWrapper struct {
	baseSliceWrapper
}

func (c WTFixedSliceWrapper) Size(ptr unsafe.Pointer) int {
	h := *(*sliceHeader)(ptr)
	return c.Underlying.Size(nil) * h.Len
}

// Append encodes the slice without the tag
func (c WTFixedSliceWrapper) Append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)
	for i := 0; i < h.Len; i++ {
		data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the corect size for the slice
func (c WTFixedSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	count := len(data) / c.Underlying.Size(nil)

	// Now make sure we have enough data in the slice
	h := (*sliceHeader)(ptr)
	if h.Cap < count {
		// Ensure the GC knows the type of this slice.
		h.Data = unsafe_NewArray(c.EltType, int(count))
		h.Cap = int(count)
	}
	h.Len = count

	var offset int
	for i := 0; i < h.Len; i++ {
		n, err := c.Underlying.Read(data[offset:], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), c.Underlying.WireType())
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}

// WTVarIntSliceWrapper is a codec for a type encoded using the WTVarInt wire
// type
type WTVarIntSliceWrapper struct {
	baseSliceWrapper
}

func (c WTVarIntSliceWrapper) Size(ptr unsafe.Pointer) int {
	h := *(*sliceHeader)(ptr)
	size := 0
	for i := 0; i < h.Len; i++ {
		size += c.Underlying.Size(unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize))
	}
	return size
}

// Append encodes the slice without the tag
func (c WTVarIntSliceWrapper) Append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)
	for i := 0; i < h.Len; i++ {
		data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize))
	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the corect size for the slice
func (c WTVarIntSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	// We step forward through out data to count how many things are in the slice
	var offset, count int
	for offset < len(data) {
		_, n := ReadVarUint(data[offset:])
		if n < 0 {
			return 0, fmt.Errorf("corrupt data")
		}
		offset += n
		count++
	}

	// Now make sure we have enough data in the slice
	h := (*sliceHeader)(ptr)
	if h.Cap < count {
		// Ensure the GC knows the type of this slice.
		h.Data = unsafe_NewArray(c.EltType, int(count))
		h.Cap = int(count)
	}
	h.Len = count

	offset = 0
	for i := 0; i < h.Len; i++ {
		n, err := c.Underlying.Read(data[offset:], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), WTVarInt)
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}
