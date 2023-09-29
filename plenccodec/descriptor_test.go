package plenccodec_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

//go:embed descriptor_test.json
var expTestDescriptor string

func TestDescriptor(t *testing.T) {
	type sub struct {
		A int8   `plenc:"1"`
		B string `plenc:"2"`
	}

	// maps are rendered as a list of keys and values because maps like this
	// don't have a valid JSON representation
	type mykey struct {
		A int `plenc:"1"`
		B int `plenc:"2"`
	}
	type mymap map[mykey]string

	type my struct {
		A int                 `plenc:"1"`
		B float32             `plenc:"2"`
		C string              `plenc:"3"`
		D uint                `plenc:"4"`
		E []float64           `plenc:"5"`
		F []sub               `plenc:"6"`
		G [][]uint32          `plenc:"7"`
		H [][]float32         `plenc:"8"`
		I *uint               `plenc:"9"`
		J mymap               `plenc:"10"`
		K []byte              `plenc:"11"`
		L map[float32]float32 `plenc:"12"`
		M *int                `plenc:"13"`
		N time.Time           `plenc:"14"`
		O bool                `plenc:"15" json:"elephant"`
		P map[string]any      `plenc:"16"`
	}

	plenc.RegisterCodec(reflect.TypeOf(map[string]any{}), plenccodec.JSONMapCodec{})
	plenc.RegisterCodec(reflect.TypeOf([]any{}), plenccodec.JSONArrayCodec{})

	c, err := plenc.CodecForType(reflect.TypeOf(my{}))
	if err != nil {
		t.Fatal(err)
	}
	d := c.Descriptor()

	// Check we can encode and decode a Descriptor!
	descData, err := plenc.Marshal(nil, &d)
	if err != nil {
		t.Fatal(err)
	}
	var dd plenccodec.Descriptor
	if err := plenc.Unmarshal(descData, &dd); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(d, dd); diff != "" {
		t.Fatal(diff)
	}

	// Now test we can use the descriptor
	var seven uint = 7
	in := my{
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
		J: mymap{
			mykey{A: 9, B: 4}: "nine",
		},
		K: []byte{0, 1, 2, 3},
		// Map order is random, so we'll just have one entry in the map. I have
		// tested with two!
		L: map[float32]float32{
			3.14: 13.37,
		},
		N: time.Date(1970, 3, 15, 0, 0, 0, 1337e5, time.UTC),
		O: true,
		P: map[string]any{
			"array": []any{1, 1.3, "cheese", json.Number("1337")},
		},
	}

	data, err := plenc.Marshal(nil, in)
	if err != nil {
		t.Fatal(err)
	}

	{
		// Check we can decode that plenc
		var out my
		if err := plenc.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(in, out); diff != "" {
			t.Fatal(diff)
		}
	}

	var j plenccodec.JSONOutput

	if err := d.Read(&j, data); err != nil {
		t.Fatal(err, string(j.Done()))
	}
	out := string(j.Done())

	if diff := cmp.Diff(expTestDescriptor, out); diff != "" {
		t.Log(out)
		t.Fatal(diff)
	}
}
