package plenc

import "testing"

func TestBool(t *testing.T) {
	for _, b := range []bool{false, true} {
		s := SizeBool(b)
		data := AppendBool(nil, b)
		a, n := ReadBool(data)
		if s != n {
			t.Errorf("size mismatch %d %d", s, n)
		}
		if a != b {
			t.Errorf("Result incorrect for %t", b)
		}
	}
}
