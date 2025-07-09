package plenccodec_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

func TestOptionalRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		in       any
		expbytes []byte
	}{
		{
			name:     "zero",
			in:       plenccodec.Optional[int]{Set: true, Value: 0},
			expbytes: []byte{0x00},
		},
		{
			name:     "1",
			in:       plenccodec.Optional[int]{Set: true, Value: 1},
			expbytes: []byte{0x02},
		},
		{
			name: "empty string",
			in:   plenccodec.Optional[string]{Set: true, Value: ""},
		},
		{
			name: "struct",
			in: struct {
				A plenccodec.Optional[int]     `plenc:"1"`
				B plenccodec.Optional[string]  `plenc:"2"`
				C plenccodec.Optional[float64] `plenc:"3"`
			}{
				A: plenccodec.Optional[int]{Set: true, Value: 42},
				B: plenccodec.Optional[string]{Set: true, Value: "hello"},
			},
			expbytes: []byte{0x08, 0x54, 0x12, 0x05, 'h', 'e', 'l', 'l', 'o'},
		},
		{
			name: "struct all set",
			in: struct {
				A plenccodec.Optional[int]     `plenc:"1"`
				B plenccodec.Optional[string]  `plenc:"2"`
				C plenccodec.Optional[float64] `plenc:"3"`
			}{
				A: plenccodec.Optional[int]{Set: true, Value: 42},
				B: plenccodec.Optional[string]{Set: true, Value: "hello"},
				C: plenccodec.Optional[float64]{Set: true, Value: 3.14},
			},
			expbytes: []byte{0x08, 0x54,
				0x12, 0x05, 'h', 'e', 'l', 'l', 'o',
				0x19, 0x1f, 0x85, 0xeb, 0x51, 0xb8, 0x1e, 0x09, 0x40},
		},
		{
			name: "struct zero values",
			in: struct {
				A plenccodec.Optional[int]     `plenc:"1"`
				B plenccodec.Optional[string]  `plenc:"2"`
				C plenccodec.Optional[float64] `plenc:"3"`
			}{
				A: plenccodec.Optional[int]{Set: true, Value: 0},
				B: plenccodec.Optional[string]{Set: true, Value: ""},
				C: plenccodec.Optional[float64]{Set: true, Value: 0.0},
			},

			expbytes: []byte{0x08, 0x00,
				0x12, 0x00,
				0x19, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "struct empty",
			in: struct {
				A plenccodec.Optional[int]     `plenc:"1"`
				B plenccodec.Optional[string]  `plenc:"2"`
				C plenccodec.Optional[float64] `plenc:"3"`
			}{},

			expbytes: []byte{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := plenc.Marshal(nil, test.in)
			if err != nil {
				t.Fatal(err)
			}

			if string(test.expbytes) != string(data) {
				t.Errorf("Expected bytes %x, got %x", test.expbytes, data)
			}

			out := reflect.New(reflect.TypeOf(test.in))
			if err := plenc.Unmarshal(data, out.Interface()); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.in, out.Elem().Interface()); diff != "" {
				t.Errorf("Round trip failed: %s", diff)
			}
		})
	}
}
