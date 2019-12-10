package plenc

import "testing"
import fuzz "github.com/google/gofuzz"

func TestTime(t *testing.T) {

	f := fuzz.New()
	for i := 0; i < 100; i++ {
		var tt, uu Time
		f.Fuzz(&tt)

		l := tt.ΦλSize()
		b := make([]byte, 0, l)
		b = tt.ΦλAppend(b)

		n, err := uu.ΦλUnmarshal(b)
		if err != nil {
			t.Fatal(err)
		}
		if n != l {
			t.Errorf("lengths not as expected %d %d (%d)", n, l, len(b))
		}
	}
}
