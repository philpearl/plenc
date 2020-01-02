package plenc

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestJSONMap(t *testing.T) {
	type customMap map[string]interface{}
	RegisterCodec(reflect.TypeOf(customMap{}), JSONMapCodec{})

	in := customMap{
		"a": 1,
		"b": -1,
		"c": 1.1,
		"d": "hat",
		"e": map[string]interface{}{
			"f": 1,
			"a": []interface{}{1, 2, 3},
		},
		"f": nil,
		"g": true,
	}

	var (
		d   []byte
		err error
	)
	d, err = Marshal(d, &in)
	if err != nil {
		t.Fatal(err)
	}

	var out customMap
	if err := Unmarshal(d, &out); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(in, out); diff != "" {
		t.Fatalf("maps differ. %s", diff)
	}
}
