package plenccodec_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
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
		t.Run(strconv.FormatUint(uint64(test), 10)+"_marshal", func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out uint64
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if test != out {
				t.Errorf("Result incorrect for %d- got %d", test, out)
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
		t.Run(strconv.FormatInt(int64(test), 10)+"_marshal", func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out int64
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if test != out {
				t.Errorf("Result incorrect for %d- got %d", test, out)
			}
		})
	}
}

func TestInts(t *testing.T) {
	type my struct {
		A int  `plenc:"1"`
		B int  `plenc:"2,flat"`
		C uint `plenc:"3"`
	}

	tests := []my{
		{A: 1, B: 2, C: 3},
		{A: 0, B: 0, C: 0},
		{A: -1, B: -2, C: 0},
		{A: math.MaxInt64, B: math.MaxInt64, C: math.MaxUint64},
		{A: math.MinInt64, B: math.MinInt64, C: 0},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out my
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatal(diff)
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
