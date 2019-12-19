package plenc

import (
	"bytes"
	"testing"
)

func TestMoveForward(t *testing.T) {

	tests := []struct {
		from int
		dist int
		exp  []byte
	}{
		{
			from: 1,
			dist: 2,
			exp:  []byte{0, 1, 2, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			from: 5,
			dist: 5,
			exp:  []byte{0, 1, 2, 3, 4, 5, 6, 7, 0, 0, 5, 6, 7},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			in := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}[:8]
			out := moveForward(in, test.from, test.dist)

			if bytes.Compare(out, test.exp) != 0 {
				t.Fatalf("Results not as expectec. Have %X, expected %X", out, test.exp)
			}
		})
	}
}

func TestFieldRemoval(t *testing.T) {
	type S1 struct {
		A int     `plenc:"1"`
		B int     `plenc:"2"`
		C string  `plenc:"3"`
		D float32 `plenc:"4"`
		E float64 `plenc:"5"`
	}
	type S2 struct {
		B int `plenc:"2"`
	}

	in := S1{A: 1, B: 2, C: "3", D: 4, E: 5}
	data, err := Marshal(nil, &in)
	if err != nil {
		t.Fatal(err)
	}

	var out S2
	if err := Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.B != 2 {
		t.Fatalf("Unmarshal unexpected value for B. %d", out.B)
	}
}
