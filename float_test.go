package φλenc

import (
	"math"
	"strconv"
	"testing"
)

func TestFloat32(t *testing.T) {

	tests := []float32{0, 1, -1, 3.14, math.MaxFloat32, math.SmallestNonzeroFloat32}

	for _, test := range tests {
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 32), func(t *testing.T) {
			l := SizeFloat32(test)
			b := make([]byte, 0, l)
			data := AppendFloat32(b, test)
			actual, n := ReadFloat32(data)

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

	tests := []float64{0, 1, -1, 3.14, math.MaxFloat32, math.SmallestNonzeroFloat32, math.MaxFloat64, math.SmallestNonzeroFloat64}

	for _, test := range tests {
		t.Run(strconv.FormatFloat(float64(test), 'g', -1, 32), func(t *testing.T) {
			l := SizeFloat64(test)
			b := make([]byte, 0, l)
			data := AppendFloat64(b, test)
			actual, n := ReadFloat64(data)

			if n != l {
				t.Errorf("lengths differ %d %d", l, n)
			}
			if actual != test {
				t.Errorf("values differ. actual %f, expected %f", actual, test)
			}
		})
	}
}

func BenchmarkAppendFloat64(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf []byte
		for pb.Next() {
			buf = AppendFloat64(buf[:0], 3.138383)
			buf = AppendFloat64(buf, 5.138383)
			buf = AppendFloat64(buf, 83.138383)
		}
	})
}
