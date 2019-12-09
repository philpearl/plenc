package test

import (
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestEncDec(t *testing.T) {
	f := fuzz.New()
	for i := 0; i < 100; i++ {
		var in, out MyStruct
		f.Fuzz(&in)
		data := in.ΦλAppend(nil)
		if _, err := out.ΦλUnmarshal(data); err != nil {
			t.Fatal(err)
		}
	}
}
