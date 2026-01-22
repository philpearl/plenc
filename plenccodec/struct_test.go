package plenccodec_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
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

func TestZeroReuseSlice(t *testing.T) {
	type s1 struct {
		A int `plenc:"1"`
		B int `plenc:"2"`
	}
	v := []s1{{A: 1}, {A: 2}}
	v = v[:0]

	w := []s1{{B: 3}, {B: 4}}
	data, err := plenc.Marshal(nil, w)
	if err != nil {
		t.Fatal(err)
	}

	if err := plenc.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(w, v); diff != "" {
		t.Fatal(diff)
	}
}

func TestZeroReuse(t *testing.T) {
	type s1 struct {
		A int `plenc:"1"`
		B int `plenc:"2"`
	}
	v := s1{A: 37, B: 42}

	w := s1{A: 0, B: 3}
	data, err := plenc.Marshal(nil, w)
	if err != nil {
		t.Fatal(err)
	}

	// We expect this to update v and not zero fields that aren't in the output.
	// Note that plenc is implicitly omitempty for some field types, so zero
	// ints in the data we're unmarshalling won't overwrite anything.
	if err := plenc.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(s1{A: 37, B: 3}, v); diff != "" {
		t.Fatal(diff)
	}
}

func TestStructDescriptor(t *testing.T) {
	type s2 struct {
		A string `plenc:"1"`
	}
	type s1 struct {
		A int `plenc:"1,flat"`
		B int `plenc:"2"`
		C s2  `plenc:"3"`
	}

	c, err := plenc.CodecForType(reflect.TypeOf(s1{}))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(plenccodec.Descriptor{
		TypeName: "s1",
		Type:     plenccodec.FieldTypeStruct,
		Elements: []plenccodec.Descriptor{
			{Index: 1, Name: "A", Type: plenccodec.FieldTypeFlatInt},
			{Index: 2, Name: "B", Type: plenccodec.FieldTypeInt},
			{
				Index:    3,
				Name:     "C",
				Type:     plenccodec.FieldTypeStruct,
				TypeName: "s2",
				Elements: []plenccodec.Descriptor{
					{Index: 1, Name: "A", Type: plenccodec.FieldTypeString},
				},
			},
		},
	}, c.Descriptor()); diff != "" {
		t.Fatal(diff)
	}
}

func BenchmarkNestedStruct(b *testing.B) {
	// Test deeply nested struct performance
	type Level3 struct {
		Value int    `plenc:"1"`
		Name  string `plenc:"2"`
	}

	type Level2 struct {
		Items []Level3 `plenc:"1"`
		Count int      `plenc:"2"`
		Label string   `plenc:"3"`
	}

	type Level1 struct {
		Children []Level2 `plenc:"1"`
		ID       int64    `plenc:"2"`
		Tags     []string `plenc:"3"`
	}

	type Root struct {
		Nodes   []Level1 `plenc:"1"`
		Version int      `plenc:"2"`
	}

	// Build a moderately complex nested structure
	in := Root{
		Version: 1,
		Nodes:   make([]Level1, 10),
	}
	for i := range in.Nodes {
		in.Nodes[i] = Level1{
			ID:       int64(i),
			Tags:     []string{"tag1", "tag2", "tag3"},
			Children: make([]Level2, 5),
		}
		for j := range in.Nodes[i].Children {
			in.Nodes[i].Children[j] = Level2{
				Count: j,
				Label: "label",
				Items: make([]Level3, 3),
			}
			for k := range in.Nodes[i].Children[j].Items {
				in.Nodes[i].Children[j].Items[k] = Level3{
					Value: k,
					Name:  "item",
				}
			}
		}
	}

	b.Run("nested", func(b *testing.B) {
		b.ReportAllocs()
		var data []byte
		for b.Loop() {
			var err error
			data, err = plenc.Marshal(data[:0], &in)
			if err != nil {
				b.Fatal(err)
			}
			var out Root
			if err := plenc.Unmarshal(data, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkFlatStruct(b *testing.B) {
	// Compare with a flat struct of similar data volume
	type Flat struct {
		V1  int    `plenc:"1"`
		V2  int    `plenc:"2"`
		V3  int    `plenc:"3"`
		V4  int64  `plenc:"4"`
		V5  string `plenc:"5"`
		V6  string `plenc:"6"`
		V7  string `plenc:"7"`
		V8  int    `plenc:"8"`
		V9  int    `plenc:"9"`
		V10 int    `plenc:"10"`
	}

	in := Flat{
		V1: 1, V2: 2, V3: 3, V4: 100,
		V5: "hello", V6: "world", V7: "test",
		V8: 8, V9: 9, V10: 10,
	}

	b.Run("flat", func(b *testing.B) {
		b.ReportAllocs()
		var data []byte
		for b.Loop() {
			var err error
			data, err = plenc.Marshal(data[:0], &in)
			if err != nil {
				b.Fatal(err)
			}
			var out Flat
			if err := plenc.Unmarshal(data, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
