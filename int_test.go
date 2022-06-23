package plenc

import (
	"math"
	"strconv"
	"testing"

	"github.com/philpearl/plenc/plenccore"
)

func TestVarUint(t *testing.T) {
	tests := []uint64{
		0, 1, 2, 126, 127, 128, 255, 256, 257, 1024, 2048, 4096, 8192, 457239847, 27384971293, 98235472397459, math.MaxUint64,
	}

	for _, test := range tests {
		t.Run(strconv.FormatUint(uint64(test), 10), func(t *testing.T) {
			s := plenccore.SizeVarUint(test)
			b := make([]byte, 0, s)
			b = plenccore.AppendVarUint(b, test)
			actual, l := plenccore.ReadVarUint(b)

			if l != s {
				t.Errorf("read %d bytes, expected %d", l, s)
			}

			if actual != test {
				t.Errorf("actual %d does not match expected %d. %v", actual, test, b)
			}
		})
	}
}

func TestVarInt(t *testing.T) {
	tests := []int64{
		0, 1, -1, 2, -1, 126, -126, 127, -127, 128, -128, 254, 255, 256, 257, -254, -255, -256, -257, 1024, 2048, 4096, 8192, 457239847, 27384971293, 98235472397459, math.MaxInt64, math.MinInt64, 25018898,
	}

	for _, test := range tests {
		t.Run(strconv.FormatInt(int64(test), 10), func(t *testing.T) {
			s := plenccore.SizeVarInt(test)
			b := make([]byte, 0, s)
			b = plenccore.AppendVarInt(b, test)
			actual, l := plenccore.ReadVarInt(b)

			if l != s {
				t.Errorf("read %d bytes, expected %d", l, s)
			}

			if actual != test {
				t.Errorf("actual %d does not match expected %d. %v", actual, test, b)
			}
		})
	}
}

func TestZigZag(t *testing.T) {
	tests := []struct {
		in  int64
		exp uint64
	}{
		{0, 0},
		{-1, 1},
		{1, 2},
		{-2, 3},
		{2, 4},
		{-2147483648, 4294967295},
		{12509449, 25018898},
	}

	for _, test := range tests {
		t.Run(strconv.FormatInt(test.in, 10), func(t *testing.T) {
			if z := plenccore.ZigZag(test.in); z != test.exp {
				t.Errorf("Expected %d got %d", test.exp, z)
			}
		})
	}
}
