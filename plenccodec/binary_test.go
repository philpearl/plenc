package plenccodec_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/philpearl/plenc"
)

func TestBinaryCompatibility(t *testing.T) {
	anInt32 := int32(1234)
	// These tests ensure the wire format of the library does not change. We
	// encode data and check the encoding matches golden values we've stored.
	tests := []struct {
		name     string
		toEncode any
	}{
		{
			name:     "string",
			toEncode: "hats",
		},
		{
			name:     "string_array",
			toEncode: []string{"hats", "coats"},
		},
		{
			name:     "bytes",
			toEncode: []byte{1, 2, 3, 4},
		},
		{
			name:     "int16",
			toEncode: int16(1234),
		},
		{
			name:     "int32",
			toEncode: int32(1234),
		},
		{
			name:     "int64",
			toEncode: int64(12343453453),
		},
		{
			name:     "uint16",
			toEncode: uint16(1234),
		},
		{
			name:     "uint32",
			toEncode: uint32(1234),
		},
		{
			name:     "uint64",
			toEncode: uint64(12343453453),
		},
		{
			name:     "int_array",
			toEncode: []int{1, 2, 1337, 98, -100},
		},
		{
			name:     "float32",
			toEncode: float32(1234.5678),
		},
		{
			name:     "float64",
			toEncode: float64(1234.5678),
		},
		{
			name:     "float_array",
			toEncode: []float64{1.2, 3.4, 5.6},
		},
		{
			name:     "bool",
			toEncode: true,
		},
		{
			name:     "bool_array",
			toEncode: []bool{true, false, true},
		},

		{
			name: "struct",
			toEncode: struct {
				Name string   `plenc:"1"`
				Age  int      `plenc:"2,flat"`
				F32  float32  `plenc:"3"`
				F64  float64  `plenc:"4"`
				I    int      `plenc:"5"`
				J    []uint32 `plenc:"6"`
				K    []string `plenc:"7"`
				L    *int     `plenc:"8"`
				M    *int32   `plenc:"9"`
				// TODO: add more types
			}{
				Name: "Phil",
				Age:  1337,
				F32:  1234.5678,
				F64:  1234.5678,
				I:    -234332,
				J:    []uint32{747439, 2223, 3344},
				K:    []string{"hats", "coats"},
				L:    nil,
				M:    &anInt32,
			},
		},
		{
			name: "struct_array",
			toEncode: []struct {
				Name string `plenc:"1"`
				Age  int    `plenc:"2"`
			}{{Name: "Phil", Age: 1337}, {Name: "Bob", Age: 42}},
		},
		{
			// Sigh. The binary representation of a map is not stable as the
			// iteration order of the map is random. We'll test with one entry
			// for now.
			name: "map",
			toEncode: map[string]int{
				"Phil": 1337,
			},
		},
		{
			name:     "time",
			toEncode: time.Date(1970, 3, 15, 13, 37, 42, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode the data.
			encoded, err := plenc.Marshal(nil, test.toEncode)
			if err != nil {
				t.Fatalf("failed to encode: %v", err)
			}

			golden, err := os.ReadFile(filepath.Join("testdata", test.name+".golden"))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if err := os.WriteFile(filepath.Join("testdata", test.name+".golden"), encoded, 0o644); err != nil {
						t.Fatal(err)
					}
					return
				}
				t.Fatalf("failed to read golden file: %v", err)
			}

			if string(encoded) != string(golden) {
				t.Fatalf("encoded data does not match golden file")
			}
		})
	}
}
