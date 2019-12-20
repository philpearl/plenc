package plenc

import (
	"math"
	"strconv"
	"testing"
	"unsafe"
)

func TestFloat32(t *testing.T) {
	c := Float32Codec{}
	tests := []float32{0, 1, -1, 3.14, math.MaxFloat32, math.SmallestNonzeroFloat32}

	for _, test := range tests {
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 32), func(t *testing.T) {
			l := c.Size(unsafe.Pointer(&test))
			b := make([]byte, 0, l)
			data := c.Append(b, unsafe.Pointer(&test))
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
	}
}

func TestFloat64(t *testing.T) {
	c := Float64Codec{}
	tests := []float64{0, 1, -1, 3.14, math.MaxFloat32, math.SmallestNonzeroFloat32, math.MaxFloat64, math.SmallestNonzeroFloat64}

	for _, test := range tests {
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 64), func(t *testing.T) {
			l := c.Size(unsafe.Pointer(&test))
			b := make([]byte, 0, l)
			data := c.Append(b, unsafe.Pointer(&test))
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
	}
}
