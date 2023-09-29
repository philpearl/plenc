package plenccodec

import (
	"fmt"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// PointerWrapper wraps a codec so it can be used for a pointer to the type
type PointerWrapper struct {
	Underlying Codec
}

func (p PointerWrapper) Omit(ptr unsafe.Pointer) bool {
	t := *(*unsafe.Pointer)(ptr)
	return t == nil
}

func (p PointerWrapper) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
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

func (p PointerWrapper) WireType() plenccore.WireType {
	return p.Underlying.WireType()
}

func (p PointerWrapper) Descriptor() Descriptor {
	return p.Underlying.Descriptor()
}

func (p PointerWrapper) Size(ptr unsafe.Pointer, tag []byte) int {
	// This should never be called if Omit returns true
	t := *(*unsafe.Pointer)(ptr)
	if t == nil {
		return 0
	}
	return p.Underlying.Size(t, tag)
}

func (p PointerWrapper) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	// This should never be called if Omit returns true
	t := *(*unsafe.Pointer)(ptr)
	if t == nil {
		return data
	}
	return p.Underlying.Append(data, t, tag)
}

type BaseSliceWrapper struct {
	Underlying Codec
	EltSize    uintptr
	EltType    unsafe.Pointer
}

func (c BaseSliceWrapper) Omit(ptr unsafe.Pointer) bool {
	h := *(*sliceHeader)(ptr)
	return h.Len == 0
}

func (c BaseSliceWrapper) New() unsafe.Pointer {
	return unsafe.Pointer(&sliceHeader{})
}

func (c BaseSliceWrapper) WireType() plenccore.WireType {
	return plenccore.WTLength
}

func (c BaseSliceWrapper) Descriptor() Descriptor {
	return Descriptor{
		Type: FieldTypeSlice,
		Elements: []Descriptor{
			c.Underlying.Descriptor(),
		},
	}
}

// WTLengthSliceWrapper is a codec for a slice of a type that's encoded using
// the WTLength wire type. It uses the WTSlice wire type for the slice itself.
type WTLengthSliceWrapper struct {
	BaseSliceWrapper
}

func (c WTLengthSliceWrapper) size(ptr unsafe.Pointer) int {
	h := *(*sliceHeader)(ptr)
	size := plenccore.SizeVarUint(uint64(h.Len))
	for i := 0; i < h.Len; i++ {
		s := c.Underlying.Size(unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), nil)
		size += s + plenccore.SizeVarUint(uint64(s))
	}
	return size
}

// append encodes the slice, and appends the encoded version to data
func (c WTLengthSliceWrapper) append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)

	// Append the count of items in the slice
	data = plenccore.AppendVarUint(data, uint64(h.Len))
	// Append each of the items. They're all prefixed by their length
	for i := 0; i < h.Len; i++ {
		ptr := unsafe.Pointer(uintptr(h.Data) + uintptr(i)*c.EltSize)
		data = plenccore.AppendVarUint(data, uint64(c.Underlying.Size(ptr, nil)))
		data = c.Underlying.Append(data, ptr, nil)
	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the corect size for the slice
func (c WTLengthSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	if wt == plenccore.WTLength {
		return c.readAsWTLength(data, ptr)
	}

	// First we read the number of items in the slice
	count, n := plenccore.ReadVarUint(data)
	if n < 0 {
		return 0, fmt.Errorf("corrupt data looking for WTSlice count")
	}

	// Now make sure we have enough capacity in the slice
	h := (*sliceHeader)(ptr)
	if h.Cap < int(count) {
		// Ensure the GC knows the type of this slice.
		h.Data = unsafe_NewArray(c.EltType, int(count))
		h.Cap = int(count)
	} else {
		// We're going to re-use the backing array. It's going to be surprising
		// if we don't start from zeros so we zero everything.
		for i := 0; i < int(count); i++ {
			// We'll only write to fields if the data is present, so start by zeroing
			// the target
			ptr := unsafe.Add(h.Data, i*int(c.EltSize))
			typedmemclr(unpackEFace(c.EltType).data, ptr)
		}
	}
	h.Len = int(count)

	offset := n
	for i := 0; i < h.Len; i++ {
		s, n := plenccore.ReadVarUint(data[offset:])
		if n <= 0 {
			return 0, fmt.Errorf("invalid varint for slice entry %d", i)
		}
		offset += n

		ptr := unsafe.Add(h.Data, i*int(c.EltSize))
		n, err := c.Underlying.Read(data[offset:offset+int(s)], ptr, plenccore.WTLength)
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}

// readAsWTLength is here for protobuf compatibility. protobuf writes certain
// array types by simply repeating the encoding for an individual field. So here
// we just read one underlying value and append it to the slice
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

	dptr := unsafe.Add(h.Data, h.Len*int(c.EltSize))
	typedmemclr(unpackEFace(c.EltType).data, dptr)
	n, err = c.Underlying.Read(data, dptr, plenccore.WTLength)
	if err != nil {
		return 0, err
	}
	h.Len++
	return n, nil
}

func (c WTLengthSliceWrapper) WireType() plenccore.WireType {
	return plenccore.WTSlice
}

func (c WTLengthSliceWrapper) Size(ptr unsafe.Pointer, tag []byte) int {
	return c.size(ptr) + len(tag)
}

func (c WTLengthSliceWrapper) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	data = append(data, tag...)
	return c.append(data, ptr)
}

// WTFixedSliceWrapper is a codec for a type that's encoded as a fixed 32 or 64
// byte value (i.e. float32 or float64)
type WTFixedSliceWrapper struct {
	BaseSliceWrapper
}

// append encodes the slice without the tag
func (c WTFixedSliceWrapper) append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)
	for i := 0; i < h.Len; i++ {
		data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), nil)
	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the corect size for the slice
func (c WTFixedSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	count := len(data) / c.Underlying.Size(nil, nil)

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

func (p WTFixedSliceWrapper) Size(ptr unsafe.Pointer, tag []byte) int {
	h := *(*sliceHeader)(ptr)
	l := p.Underlying.Size(nil, nil) * h.Len
	if len(tag) > 0 {
		l += len(tag) + plenccore.SizeVarUint(uint64(l))
	}
	return l
}

func (p WTFixedSliceWrapper) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	if len(tag) > 0 {
		data = append(data, tag...)
		data = plenccore.AppendVarUint(data, uint64(p.Size(ptr, nil)))
	}
	return p.append(data, ptr)
}

// WTVarIntSliceWrapper is a codec for a type encoded using the WTVarInt wire
// type.
type WTVarIntSliceWrapper struct {
	BaseSliceWrapper
}

func (c WTVarIntSliceWrapper) size(ptr unsafe.Pointer) int {
	h := *(*sliceHeader)(ptr)
	size := 0
	for i := 0; i < h.Len; i++ {
		size += c.Underlying.Size(unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), nil)
	}
	return size
}

// append encodes the slice without the tag
func (c WTVarIntSliceWrapper) append(data []byte, ptr unsafe.Pointer) []byte {
	h := *(*sliceHeader)(ptr)
	for i := 0; i < h.Len; i++ {
		data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), nil)
	}
	return data
}

// Read decodes a slice. It assumes the WTLength tag has already been decoded
// and that the data slice is the correct size for the slice
func (c WTVarIntSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	// We step forward through out data to count how many things are in the slice
	var offset, count int
	for offset < len(data) {
		_, n := plenccore.ReadVarUint(data[offset:])
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
		n, err := c.Underlying.Read(data[offset:], unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), plenccore.WTVarInt)
		if err != nil {
			return 0, err
		}
		offset += n
	}

	return offset, nil
}

func (p WTVarIntSliceWrapper) Size(ptr unsafe.Pointer, tag []byte) int {
	l := p.size(ptr)
	if len(tag) > 0 {
		l += len(tag) + plenccore.SizeVarUint(uint64(l))
	}
	return l
}

func (p WTVarIntSliceWrapper) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	if len(tag) > 0 {
		data = append(data, tag...)
		data = plenccore.AppendVarUint(data, uint64(p.size(ptr)))
	}
	return p.append(data, ptr)
}

// ProtoSliceWrapper is for encoding slices of WTLength encoded objects how
// protobuf does it. When writing the elements of the slice are simply repeated.
// When reading each element is treated separately and appended to the slice.
//
// Note this does not work outside of a struct as we don't add tags and
// therefore the length of each element is not known
type ProtoSliceWrapper struct {
	BaseSliceWrapper
}

// Size calculates the amount of data needed for this slice, including
// repeated tags and lengths
func (c ProtoSliceWrapper) Size(ptr unsafe.Pointer, tag []byte) int {
	h := *(*sliceHeader)(ptr)
	var l int
	for i := 0; i < h.Len; i++ {
		l += c.Underlying.Size(unsafe.Add(h.Data, uintptr(i)*c.EltSize), tag)
	}
	return l
}

// Append appends the data for this slice, including repeated tags and
// lengths for each element
func (c ProtoSliceWrapper) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	h := *(*sliceHeader)(ptr)
	for i := 0; i < h.Len; i++ {
		data = c.Underlying.Append(data, unsafe.Pointer(uintptr(h.Data)+uintptr(i)*c.EltSize), tag)
	}
	return data
}

func (c ProtoSliceWrapper) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
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

	dptr := unsafe.Add(h.Data, h.Len*int(c.EltSize))
	typedmemclr(unpackEFace(c.EltType).data, dptr)
	n, err = c.Underlying.Read(data, dptr, plenccore.WTLength)
	if err != nil {
		return 0, err
	}
	h.Len++
	return n, nil
}
