package plenccodec_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
)

func TestSliceIntPtr(t *testing.T) {
	one := 1
	two := 2
	tests := []struct {
		name  string
		input []*int
		exp   []*int
	}{
		{name: "empty slices look nil", input: []*int{}, exp: nil}, // empty slices come back as nil
		{name: "slices work", input: []*int{&one, &two}, exp: []*int{&one, &two}},
		// TODO: is it right that these get dropped? or should they error?
		// We've basically chosen to silently drop these to avoid checking!
		{name: "nil pointers not allowed", input: []*int{&one, nil}, exp: []*int{&one}},
		{name: "nil pointers not allowed 2", input: []*int{&one, nil, &two}, exp: []*int{&one, &two}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test.input)
			if err != nil {
				t.Fatal(err)
			}

			var out []*int
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.exp, out); diff != "" {
				t.Fatalf("Not as expected. %s\n%x", diff, data)
			}
		})
	}
}

func TestSliceInt(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		exp   []int
	}{
		{name: "empty slices look nil", input: []int{}, exp: nil}, // empty slices come back as nil
		{name: "slices work", input: []int{1, 2}, exp: []int{1, 2}},
		// TODO: is it right that these get dropped? or should they error?
		// We've basically chosen to silently drop these to avoid checking!
		{name: "nil pointers not allowed", input: []int{1, 0}, exp: []int{1, 0}},
		{name: "nil pointers not allowed 2", input: []int{1, 0, 2}, exp: []int{1, 0, 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test.input)
			if err != nil {
				t.Fatal(err)
			}

			var out []int
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.exp, out); diff != "" {
				t.Fatalf("Not as expected. %s\n%x", diff, data)
			}
		})
	}
}

func TestSliceFloat(t *testing.T) {
	tests := []struct {
		name  string
		input []float64
		exp   []float64
	}{
		{name: "empty slices look nil", input: []float64{}, exp: nil}, // empty slices come back as nil
		{name: "slices work", input: []float64{1, 2}, exp: []float64{1, 2}},
		// TODO: is it right that these get dropped? or should they error?
		// We've basically chosen to silently drop these to avoid checking!
		{name: "nil pointers not allowed", input: []float64{1, 0}, exp: []float64{1, 0}},
		{name: "nil pointers not allowed 2", input: []float64{1, 0, 2}, exp: []float64{1, 0, 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test.input)
			if err != nil {
				t.Fatal(err)
			}

			var out []float64
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.exp, out); diff != "" {
				t.Fatalf("Not as expected. %s\n%x", diff, data)
			}
		})
	}
}
