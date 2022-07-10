package plenccodec_test

import (
	"testing"
	"unsafe"

	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

func TestBool(t *testing.T) {
	c := plenccodec.BoolCodec{}
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

func TestBoolMarshal(t *testing.T) {
	for _, b := range []bool{false, true} {
		data, err := plenc.Marshal(nil, &b)
		if err != nil {
			t.Fatal(err)
		}
		var out bool
		if err := plenc.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if b != out {
			t.Errorf("Result incorrect for %t", b)
		}
	}
}
