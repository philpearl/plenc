package philenc

import "testing"

func TestString(t *testing.T) {
	tests := []string{
		"",
		"a",
		"¬πø∆˙©¥å∫˜",
		"this is a string",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			l := SizeString(test)
			b := make([]byte, 0, l)
			b = AppendString(b, test)
			actual, n := ReadString(b)
			if n != l {
				t.Errorf("length %d is not as expected (%d)", n, l)
			}
			if actual != test {
				t.Errorf("result %q is not as expected %q", actual, test)
			}
		})
	}
}
