package plenccodec

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// StringCodec is a codec for an string
type StringCodec struct{}

// Size returns the number of bytes needed to encode a string
func (StringCodec) Size(ptr unsafe.Pointer) int {
	return len(*(*string)(ptr))
}

// Append encodes a string
func (StringCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
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
	return false
}

// BytesCodec is a codec for a byte slice
type BytesCodec struct{}

// Size returns the number of bytes needed to encode a string
func (BytesCodec) Size(ptr unsafe.Pointer) int {
	return len(*(*[]byte)(ptr))
}

// Append encodes a []byte
func (BytesCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
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
	return false
}

type Interner interface {
	WithInterning() Codec
}

type InternedStringCodec struct {
	sync.Mutex
	strings unsafe.Pointer
	StringCodec
}

func (c *InternedStringCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	p := atomic.LoadPointer(&c.strings)
	m := *(*map[string]string)((unsafe.Pointer)(&p))

	s, ok := m[string(data)]
	if !ok {
		s = c.addString(data)
	}

	*(*string)(ptr) = s
	return len(data), nil
}

func (c *InternedStringCodec) addString(data []byte) string {
	c.Lock()
	defer c.Unlock()
	p := atomic.LoadPointer(&c.strings)
	m := *(*map[string]string)((unsafe.Pointer)(&p))

	s, ok := m[string(data)]
	if !ok {
		// We completely replace the map with a new one, so the old one can
		// be read without locks. We're expecting the number of different values
		// to be small, so that this is a reasonable thing to do.
		m2 := make(map[string]string, len(m)+1)
		for k, v := range m {
			m2[k] = v
		}
		s = string(data)
		m2[s] = s

		atomic.StorePointer(&c.strings, *(*unsafe.Pointer)(unsafe.Pointer(&m2)))
	}
	return s
}

func (c *InternedStringCodec) len() int {
	p := atomic.LoadPointer(&c.strings)
	m := *(*map[string]string)((unsafe.Pointer)(&p))
	return len(m)
}

func (c StringCodec) WithInterning() Codec {
	ic := &InternedStringCodec{}
	m := map[string]string{}
	atomic.StorePointer(&ic.strings, *(*unsafe.Pointer)(unsafe.Pointer(&m)))
	return ic
}
