package test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/philpearl/plenc"
)

func TestEncDec(t *testing.T) {
	f := fuzz.New().Funcs(
		func(o *[]*Struct2, c fuzz.Continue) {
			// Don't want nils in our slices of pointers
			l := c.Intn(10)
			if l == 0 {
				return
			}
			*o = make([]*Struct2, l)
			for i := range *o {
				var s Struct2
				c.Fuzz(&s)
				(*o)[i] = &s
			}
		},
		func(o *OptInt, c fuzz.Continue) {
			o.Valid = c.RandBool()
			if o.Valid {
				o.Value = c.Int()
			}
		},
	)

	for i := 0; i < 100; i++ {
		var in, out MyStruct
		f.Fuzz(&in)

		data := in.ΦλAppend(nil)
		if _, err := out.ΦλUnmarshal(data); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(in, out); diff != "" {
			t.Fatalf("structures differ. %s", diff)
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

func BenchmarkEncDecPlencMarshal(b *testing.B) {
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
		var data []byte
		for pb.Next() {
			var err error
			data, err = plenc.Marshal(data[:0], &m)
			if err != nil {
				b.Fatal(err)
			}

			var v MyStruct
			if err := plenc.Unmarshal(data, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
}
