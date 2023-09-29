package plenccodec_test

import (
	"math"
	"strconv"
	"testing"
	"unsafe"

	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

func TestFloat32(t *testing.T) {
	c := plenccodec.Float32Codec{}
	tests := []float32{0, 1, -1, 3.14, math.MaxFloat32, math.SmallestNonzeroFloat32}

	for _, test := range tests {
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 32), func(t *testing.T) {
			l := c.Size(unsafe.Pointer(&test), nil)
			b := make([]byte, 0, l)
			data := c.Append(b, unsafe.Pointer(&test), nil)
			var actual float32
			n, err := c.Read(data, unsafe.Pointer(&actual), c.WireType())
			if err != nil {
				t.Fatal(err)
			}
			if n != l {
				t.Errorf("lengths differ %d %d", l, n)
			}
			if actual != test {
				t.Errorf("values differ. actual %f, expected %f", actual, test)
			}
		})

		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 32)+"_marshal", func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out float32
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if test != out {
				t.Errorf("Result incorrect for %f - got %f", test, out)
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	c := plenccodec.Float64Codec{}
	tests := []float64{0, 1, -1, 3.14, math.MaxFloat32, math.SmallestNonzeroFloat32, math.MaxFloat64, math.SmallestNonzeroFloat64}

	for _, test := range tests {
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 64), func(t *testing.T) {
			l := c.Size(unsafe.Pointer(&test), nil)
			b := make([]byte, 0, l)
			data := c.Append(b, unsafe.Pointer(&test), nil)
			var actual float64
			n, err := c.Read(data, unsafe.Pointer(&actual), c.WireType())
			if err != nil {
				t.Fatal(err)
			}
			if n != l {
				t.Errorf("lengths differ %d %d", l, n)
			}
			if actual != test {
				t.Errorf("values differ. actual %f, expected %f", actual, test)
			}
		})
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 64)+"_marshal", func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out float64
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if test != out {
				t.Errorf("Result incorrect for %f - got %f", test, out)
			}
		})
	}
}
