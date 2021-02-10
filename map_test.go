package plenc

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMap(t *testing.T) {

	one := 1
	_ = one

	type mk struct {
		A int `plenc:"1"`
		B int `plenc:"2"`
	}

	type thing struct {
		M map[string]string `plenc:"1"`
		I int               `plenc:"2"`
	}

	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "string,string",
			data: map[string]string{
				"hat": "hut",
				"it":  "sit",
			},
		},
		{
			name: "string,string empty",
			data: map[string]string{},
		},
		{
			name: "string,string empty strings",
			data: map[string]string{
				"hat": "",
				"it":  "",
				"":    "elephant",
			},
		},
		{
			name: "string,int",
			data: map[string]int{
				"hat": 0,
				"it":  32,
			},
		},
		{
			name: "string,*int",
			data: map[string]*int{
				"hat": nil,
				"it":  &one,
			},
		},
		{
			name: "string,*int just nil",
			data: map[string]*int{
				"hat": nil,
			},
		},
		{
			name: "float,*int",
			data: map[float32]*int{
				0:     nil,
				1.001: &one,
			},
		},
		{
			name: "struct,*int",
			data: map[mk]*int{
				{A: 1, B: 7}: nil,
				{A: 1, B: 3}: &one,
			},
		},
		{
			name: "nil map",
			data: map[string]string(nil),
		},
		{
			name: "map in struct nil",
			data: thing{},
		},
		{
			name: "map in struct",
			data: thing{
				M: map[string]string{"A": "B"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := Marshal(nil, test.data)
			if err != nil {
				t.Fatal(err)
			}

			mv := reflect.New(reflect.TypeOf(test.data))
			if err := Unmarshal(data, mv.Interface()); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.data, mv.Elem().Interface()); diff != "" {
				t.Fatalf("result differs: %s", diff)
			}
		})
	}
}

func BenchmarkMap(b *testing.B) {

	m := map[string]string{
		"AAA":  "AAA",
		"AAA1": "AAA",
		"AAA2": "AAA",
		"AAA3": "AAA",
		"AAA4": "AAA",
		"AAA5": "AAA",
		"AAA6": "AAA",
		"AAA7": "AAA",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data, err := Marshal(nil, m)
		if err != nil {
			b.Fatal(err)
		}
		var o map[string]string
		if err := Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}
