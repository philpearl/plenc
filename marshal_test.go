package plenc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
)

type InnerThing struct {
	A string    `plenc:"1"`
	B float64   `plenc:"2"`
	C time.Time `plenc:"3"`
}

type TestThing struct {
	A  float64     `plenc:"1"`
	B  []float64   `plenc:"2"`
	C  *float64    `plenc:"3"`
	D  float32     `plenc:"4"`
	E  []float32   `plenc:"5"`
	F  *float32    `plenc:"6"`
	G  int         `plenc:"7"`
	H  []int       `plenc:"8"`
	I  *int        `plenc:"9"`
	J  uint        `plenc:"10"`
	K  []uint      `plenc:"11"`
	L  *uint       `plenc:"12"`
	M  bool        `plenc:"13"`
	N  []bool      `plenc:"14"`
	O  *bool       `plenc:"15"`
	P  string      `plenc:"16"`
	Q  []string    `plenc:"17"`
	R  *string     `plenc:"18"`
	S  time.Time   `plenc:"19"`
	T  []time.Time `plenc:"20"`
	U  *time.Time  `plenc:"21"`
	V  int32       `plenc:"22"`
	W  []int32     `plenc:"23"`
	X  *int32      `plenc:"24"`
	Y  int64       `plenc:"25"`
	Z  []int64     `plenc:"26"`
	A1 *int64      `plenc:"27"`
	A2 int16       `plenc:"29"`
	A3 []int16     `plenc:"30"`
	A4 *int16      `plenc:"31"`
	A5 uint8       `plenc:"32"`
	A6 []uint8     `plenc:"33"`
	A7 *uint8      `plenc:"34"`

	Z1 InnerThing `plenc:"28"`
}

func TestMarshal(t *testing.T) {

	f := fuzz.New()
	for i := 0; i < 10000; i++ {
		var in TestThing
		f.Fuzz(&in)

		data, err := Marshal(nil, &in)
		if err != nil {
			t.Fatal(err)
		}

		var out TestThing
		if err := Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(in, out); diff != "" {
			t.Logf("%x", data)

			var out TestThing
			if err := Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(in, out); diff != "" {
				t.Logf("re-run differs too")
			} else {
				t.Logf("re-run does not differ")
			}

			t.Fatalf("structs differ. %s", diff)
		}
	}
}

func TestMarshalUnmarked(t *testing.T) {
	type unmarked struct {
		A string
	}

	var in unmarked
	_, err := Marshal(nil, &in)
	if err == nil {
		t.Errorf("expected an error as field has no plenc tag")
	}
	if err.Error() != "no plenc tag on field 0 A of unmarked" {
		t.Errorf("error %q not as expected", err)
	}
}

func TestMarshalDuplicate(t *testing.T) {
	type duplicate struct {
		A string `plenc:"1"`
		B string `plenc:"1"`
	}

	var in duplicate
	_, err := Marshal(nil, &in)
	if err == nil {
		t.Errorf("expected an error as fields have duplicate plenc tags")
	}
	if err.Error() != "failed building codec for duplicate. Multiple fields have index 1" {
		t.Errorf("error %q not as expected", err)
	}
}

func BenchmarkCycle(b *testing.B) {
	f := fuzz.New()
	var in TestThing
	f.Fuzz(&in)

	b.Run("plenc", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var data []byte
			for pb.Next() {
				var err error
				data, err = Marshal(data[:0], &in)
				if err != nil {
					b.Fatal(err)
				}
				var out TestThing
				if err := Unmarshal(data, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var err error
				data, err := json.Marshal(&in)
				if err != nil {
					b.Fatal(err)
				}
				var out TestThing
				if err := json.Unmarshal(data, &out); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func TestNamedTypes(t *testing.T) {
	type Bool bool
	type Int int
	type Float64 float64
	type Float32 float32
	type Uint uint

	type MyStruct struct {
		V1 Bool    `plenc:"1"`
		V2 Int     `plenc:"2"`
		V3 Float64 `plenc:"3"`
		V4 Float32 `plenc:"4"`
		V5 Uint    `plenc:"5"`
	}

	var in, out MyStruct

	f := fuzz.New()
	f.Fuzz(&in)

	data, err := Marshal(nil, &in)
	if err != nil {
		t.Fatal(err)
	}

	if err := Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(in, out); diff != "" {
		t.Fatalf("results differ. %s", diff)
	}
}
