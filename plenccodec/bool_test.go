package plenccodec

import (
	"testing"
	"unsafe"
)

func TestBool(t *testing.T) {
	c := BoolCodec{}
	for _, b := range []bool{false, true} {
		s := c.Size(unsafe.Pointer(&b))
		data := c.Append(nil, unsafe.Pointer(&b))
		var a bool
		n, err := c.Read(data, unsafe.Pointer(&a), c.WireType())
		if err != nil {
			t.Fatal(err)
		}
		if s != n {
			t.Errorf("size mismatch %d %d", s, n)
		}
		if a != b {
			t.Errorf("Result incorrect for %t", b)
		}
	}
}
