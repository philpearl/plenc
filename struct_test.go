package plenc

import (
	"testing"
)

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
