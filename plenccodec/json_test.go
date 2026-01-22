package plenccodec_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

func TestJSONMap(t *testing.T) {
	type customMap map[string]any
	plenc.RegisterCodec(reflect.TypeFor[customMap](), plenccodec.JSONMapCodec{})

	tests := []customMap{
		{
			"a": 1,
			"b": -1,
			"c": 1.1,
			"d": "hat",
			"e": map[string]any{
				"f": 1,
				"a": []any{1, 2, 3},
				"b": []any{1, 1.3, "haddock"},
			},
			"f": nil,
			"g": true,
			"h": json.Number("3.1415"),
		},
		{},
		nil,
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			var (
				d   []byte
				err error
			)
			d, err = plenc.Marshal(d, &test)
			if err != nil {
				t.Fatal(err)
			}

			var out customMap
			if err := plenc.Unmarshal(d, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatalf("maps differ. %s", diff)
			}
		})
	}
}

func TestJSONMapStruct(t *testing.T) {
	type customMap map[string]any
	plenc.RegisterCodec(reflect.TypeFor[customMap](), plenccodec.JSONMapCodec{})

	type my struct {
		A customMap `plenc:"1"`
		B customMap `plenc:"2"`
	}

	in := my{
		A: make(customMap),
		B: nil,
	}

	d, err := plenc.Marshal(nil, &in)
	if err != nil {
		t.Fatal(err)
	}

	var out my
	if err := plenc.Unmarshal(d, &out); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(in, out); diff != "" {
		t.Fatalf("maps differ. %s", diff)
	}
}

func TestJSONMapSkip(t *testing.T) {
	type customMap map[string]any
	plenc.RegisterCodec(reflect.TypeFor[customMap](), plenccodec.JSONMapCodec{})

	type my struct {
		A int       `plenc:"1"`
		B customMap `plenc:"2"`
		C string    `plenc:"3"`
	}

	in := my{
		A: 37,
		B: customMap{
			"a": 1,
			"b": -1,
			"c": 1.1,
			"d": "hat",
			"e": map[string]any{
				"f": 1,
				"a": []any{1, 2, 3},
			},
			"f": nil,
			"g": true,
			"h": json.Number("3.1415"),
		},
		C: "hello",
	}

	var (
		d   []byte
		err error
	)
	d, err = plenc.Marshal(d, &in)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		A int    `plenc:"1"`
		C string `plenc:"3"`
	}
	if err := plenc.Unmarshal(d, &out); err != nil {
		t.Fatal(err)
	}

	if out.A != 37 && out.C != "hello" {
		t.Fatalf("output not as expected. %#v", out)
	}
}
