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
