package test

import (
	"encoding/json"
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

func BenchmarkEncDec(b *testing.B) {
	m := MyStruct{
		A: 1,
		B: 329,
		C: 32,
		D: 3.14,
		E: 872981721,
		F: true,
		H: Struct2{
			A: 12,
			B: "the swizz",
		},
		I: []Struct2{
			{A: 128, B: "this is it"},
			{A: 128, B: "the real thing"},
		},
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var bu []byte
		for pb.Next() {
			bu = bu[:0]
			s := m.ΦλSize()
			bu = m.ΦλAppend(bu)

			var v MyStruct
			n, err := v.ΦλUnmarshal(bu)
			if err != nil {
				b.Fatal(err)
			}
			if n != s {
				b.Errorf("sizes mismatch %d %d", s, n)
			}
		}
	})
}

func BenchmarkEncDecJSON(b *testing.B) {
	m := MyStruct{
		A: 1,
		B: 329,
		C: 32,
		D: 3.14,
		E: 872981721,
		F: true,
		H: Struct2{
			A: 12,
			B: "the swizz",
		},
		I: []Struct2{
			{A: 128, B: "this is it"},
			{A: 128, B: "the real thing"},
		},
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			data, err := json.Marshal(&m)
			if err != nil {
				b.Fatal(err)
			}

			var v MyStruct
			if err := json.Unmarshal(data, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
}
