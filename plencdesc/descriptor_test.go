package plencdesc_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plencdesc"
)

func TestDescriptor(t *testing.T) {
	type sub struct {
		A int8   `plenc:"1"`
		B string `plenc:"2"`
	}

	type my struct {
		A int         `plenc:"1"`
		B float32     `plenc:"2"`
		C string      `plenc:"3"`
		D uint        `plenc:"4"`
		E []float64   `plenc:"5"`
		F []sub       `plenc:"6"`
		G [][]uint32  `plenc:"7"`
		H [][]float32 `plenc:"8"`
		I *uint       `plenc:"9"`
	}

	d := plencdesc.Descriptor{
		Type: plencdesc.FieldTypeStruct,
		Elements: []plencdesc.Descriptor{
			{Name: "A", Type: plencdesc.FieldTypeInt, Index: 1},
			{Name: "B", Type: plencdesc.FieldTypeFloat32, Index: 2},
			{Name: "C", Type: plencdesc.FieldTypeString, Index: 3},
			{Name: "D", Type: plencdesc.FieldTypeUint, Index: 4},
			{Name: "E", Type: plencdesc.FieldTypeSlice, Elements: []plencdesc.Descriptor{{Type: plencdesc.FieldTypeFloat64}}, Index: 5},
			{
				Name: "F",
				Type: plencdesc.FieldTypeSlice,
				Elements: []plencdesc.Descriptor{
					{
						Type: plencdesc.FieldTypeStruct,
						Elements: []plencdesc.Descriptor{
							{Name: "A", Type: plencdesc.FieldTypeInt, Index: 1},
							{Name: "B", Type: plencdesc.FieldTypeString, Index: 2},
						},
					},
				},
				Index: 6,
			},
			{
				Name: "G",
				Type: plencdesc.FieldTypeSlice,
				Elements: []plencdesc.Descriptor{
					{
						Type:     plencdesc.FieldTypeSlice,
						Elements: []plencdesc.Descriptor{{Type: plencdesc.FieldTypeUint}},
					},
				},
				Index: 7,
			},
			{
				Name: "H",
				Type: plencdesc.FieldTypeSlice,
				Elements: []plencdesc.Descriptor{
					{
						Type:     plencdesc.FieldTypeSlice,
						Elements: []plencdesc.Descriptor{{Type: plencdesc.FieldTypeFloat32}},
					},
				},
				Index: 8,
			},
			{Name: "I", Type: plencdesc.FieldTypeUint, Index: 9},
		},
	}

	var seven uint = 7
	data, err := plenc.Marshal(nil, my{
		A: 1,
		B: 3.7,
		C: "this is my hat",
		D: 9898,
		E: []float64{1, 2.3, 3.7},
		F: []sub{
			{A: 1, B: "one"},
			{A: 2, B: "two"},
			{A: 3},
		},
		G: [][]uint32{
			{1, 2},
			{3, 4},
		},
		H: [][]float32{
			{1, 2},
			{3, 4},
		},
		I: &seven,
	})
	if err != nil {
		t.Fatal(err)
	}

	out, err := d.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]interface{}{
		"A": int64(1),
		"B": float32(3.700000047683716),
		"C": "this is my hat",
		"D": uint64(9898),
		"E": []interface{}{1.0, 2.3, 3.7},
		"F": []interface{}{
			map[string]interface{}{"A": int64(1), "B": "one"},
			map[string]interface{}{"A": int64(2), "B": "two"},
			map[string]interface{}{"A": int64(3), "B": ""},
		},
		"G": []interface{}{
			[]interface{}{uint64(1), uint64(2)},
			[]interface{}{uint64(3), uint64(4)},
		},
		"H": []interface{}{
			[]interface{}{float32(1), float32(2)},
			[]interface{}{float32(3), float32(4)},
		},
		"I": uint64(7),
	}, out); diff != "" {
		t.Fatal(diff)
	}
}
