package plenccodec_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
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
	data, err := plenc.Marshal(nil, &in)
	if err != nil {
		t.Fatal(err)
	}

	var out S2
	if err := plenc.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.B != 2 {
		t.Fatalf("Unmarshal unexpected value for B. %d", out.B)
	}
}

func TestRecursiveStruct(t *testing.T) {
	type s0 struct {
		A []s0 `plenc:"1"`
		B int  `plenc:"2"`
	}

	a := s0{
		A: []s0{{A: []s0{{B: 1}}}, {A: []s0{{B: 1}}, B: 3}},
	}

	data, err := plenc.Marshal(nil, &a)
	if err != nil {
		t.Fatal(err)
	}

	var out s0
	if err := plenc.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(a, out); diff != "" {
		t.Fatalf("Not as expected. %s\n%x", diff, data)
	}
}

func TestSliceStructPtr(t *testing.T) {
	type S2 struct {
		A int `plenc:"1"`
	}
	type S1 struct {
		A []*S2 `plenc:"1"`
	}

	tests := []S1{
		{},
		{A: []*S2{
			{A: 1},
			{A: 2},
		}},
		{A: []*S2{
			{A: 1},
			{},
			{A: 2},
		}},
	}

	/*
		00001011
		0b 03
		  02 0802
		  00
		  02 0804
	*/
	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test), func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}

			var out S1
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatalf("Not as expected. %s\n%x", diff, data)
			}
		})
	}
}

func TestSliceStructPtrNil(t *testing.T) {
	type S2 struct {
		A int `plenc:"1"`
	}
	type S1 struct {
		A []*S2 `plenc:"1"`
	}

	in := S1{A: []*S2{
		{A: 1},
		nil,
		{A: 2},
	}}

	exp := S1{A: []*S2{
		{A: 1},
		{},
		{A: 2},
	}}

	data, err := plenc.Marshal(nil, &in)
	if err != nil {
		t.Fatal(err)
	}

	var out S1
	if err := plenc.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(exp, out); diff != "" {
		t.Fatalf("Not as expected. %s\n%x", diff, data)
	}
}
