package plenc

import (
	"math"
	"testing"
	"time"
)

// FuzzUnmarshalSimple tests that Unmarshal doesn't panic on arbitrary input
// for a simple struct.
func FuzzUnmarshalSimple(f *testing.F) {
	type Simple struct {
		A int    `plenc:"1"`
		B string `plenc:"2"`
		C bool   `plenc:"3"`
	}

	// Add seed corpus from valid encodings
	var s Simple
	if data, err := Marshal(nil, &s); err == nil {
		f.Add(data)
	}
	s = Simple{A: 42, B: "hello", C: true}
	if data, err := Marshal(nil, &s); err == nil {
		f.Add(data)
	}
	// Add some edge cases
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add([]byte{0x08, 0x54}) // varint field
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01})

	f.Fuzz(func(t *testing.T, data []byte) {
		var out Simple
		// Should not panic - errors are acceptable
		_ = Unmarshal(data, &out)
	})
}

// FuzzUnmarshalComplex tests unmarshaling into a complex nested struct.
func FuzzUnmarshalComplex(f *testing.F) {
	type Inner struct {
		X int    `plenc:"1"`
		Y string `plenc:"2"`
	}
	type Complex struct {
		A int      `plenc:"1"`
		B string   `plenc:"2"`
		C float64  `plenc:"3"`
		D []int    `plenc:"4"`
		E []string `plenc:"5"`
		F *int     `plenc:"6"`
		G Inner    `plenc:"7"`
		H []Inner  `plenc:"8"`
	}

	// Seed with valid data
	seven := 7
	c := Complex{
		A: 1, B: "test", C: 3.14,
		D: []int{1, 2, 3},
		E: []string{"a", "b"},
		F: &seven,
		G: Inner{X: 10, Y: "inner"},
		H: []Inner{{X: 1, Y: "one"}, {X: 2, Y: "two"}},
	}
	if data, err := Marshal(nil, &c); err == nil {
		f.Add(data)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var out Complex
		_ = Unmarshal(data, &out)
	})
}

// FuzzUnmarshalMap tests unmarshaling into a map type.
func FuzzUnmarshalMap(f *testing.F) {
	type WithMap struct {
		M map[string]int `plenc:"1"`
		N int            `plenc:"2"`
	}

	w := WithMap{
		M: map[string]int{"a": 1, "b": 2},
		N: 42,
	}
	if data, err := Marshal(nil, &w); err == nil {
		f.Add(data)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var out WithMap
		_ = Unmarshal(data, &out)
	})
}

// FuzzUnmarshalTime tests unmarshaling time.Time fields.
func FuzzUnmarshalTime(f *testing.F) {
	type WithTime struct {
		T time.Time   `plenc:"1"`
		U []time.Time `plenc:"2"`
	}

	w := WithTime{
		T: time.Now(),
		U: []time.Time{time.Now(), time.Unix(0, 0)},
	}
	if data, err := Marshal(nil, &w); err == nil {
		f.Add(data)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var out WithTime
		_ = Unmarshal(data, &out)
	})
}

func FuzzUnmarshalTimeCompat(f *testing.F) {
	type WithTime struct {
		T time.Time   `plenc:"1"`
		U []time.Time `plenc:"2"`
	}

	w := WithTime{
		T: time.Now(),
		U: []time.Time{time.Now(), time.Unix(0, 0)},
	}
	var p Plenc
	p.ProtoCompatibleTime = true
	p.ProtoCompatibleArrays = true
	p.RegisterDefaultCodecs()
	if data, err := p.Marshal(nil, &w); err == nil {
		f.Add(data)
	}
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		var out WithTime
		_ = p.Unmarshal(data, &out)
	})
}

// FuzzRoundTrip tests that marshaled data can be unmarshaled back.
func FuzzRoundTrip(f *testing.F) {
	type Data struct {
		I int     `plenc:"1"`
		S string  `plenc:"2"`
		F float64 `plenc:"3"`
		B bool    `plenc:"4"`
	}

	f.Add(0, "", 1.32, false)
	f.Add(42, "hello", -101010110011001.7, true)
	f.Add(-1, "世界", 0.0, false)
	f.Add(1<<30, "test\x00with\x00nulls", math.MaxFloat64, true)
	f.Add(1<<30, "sfkjhaskjdfhaskjdfhaksjdhfaksjdhfkjasdhfkjashdfkjasdhfkjasdhfk", -math.MaxFloat64, true)

	f.Fuzz(func(t *testing.T, i int, s string, f float64, b bool) {
		in := Data{I: i, S: s, F: f, B: b}

		data, err := Marshal(nil, &in)
		if err != nil {
			return // Some inputs might not be marshalable
		}

		var out Data
		if err := Unmarshal(data, &out); err != nil {
			t.Errorf("failed to unmarshal valid data: %v", err)
			return
		}

		if out.I != in.I || out.S != in.S || out.F != in.F || out.B != in.B {
			t.Errorf("round trip failed: got %+v, want %+v", out, in)
		}
	})
}
