package plenc

import (
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestConcurrentMarshalUnmarshal tests that concurrent Marshal/Unmarshal
// operations on the same type are safe.
func TestConcurrentMarshalUnmarshal(t *testing.T) {
	type Inner struct {
		X int    `plenc:"1"`
		Y string `plenc:"2"`
	}
	type Data struct {
		A int     `plenc:"1"`
		B string  `plenc:"2"`
		C float64 `plenc:"3"`
		D []int   `plenc:"4"`
		E []Inner `plenc:"5"`
		F *int    `plenc:"6"`
	}

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup

	// Use a channel to collect any errors
	errCh := make(chan error, goroutines)

	for id := range goroutines {
		wg.Go(func() {
			seven := 7
			in := Data{
				B: "test string",
				C: 3.14159,
				D: []int{1, 2, 3, 4, 5},
				E: []Inner{{X: 1, Y: "one"}, {X: 2, Y: "two"}},
				F: &seven,
			}
			for i := range iterations {
				in.A = id*1000 + i

				data, err := Marshal(nil, &in)
				if err != nil {
					errCh <- err
					return
				}

				var out Data
				if err := Unmarshal(data, &out); err != nil {
					errCh <- err
					return
				}

				if diff := cmp.Diff(in, out); diff != "" {
					t.Errorf("goroutine %d iteration %d: mismatch (-want +got):\n%s", id, i, diff)
					return
				}
			}
		})
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent error: %v", err)
	}
}

// TestConcurrentDifferentTypes tests concurrent Marshal/Unmarshal of
// different types simultaneously.
func TestConcurrentDifferentTypes(t *testing.T) {
	type TypeA struct {
		X int    `plenc:"1"`
		Y string `plenc:"2"`
	}
	type TypeB struct {
		M float64 `plenc:"1"`
		N bool    `plenc:"2"`
	}
	type TypeC struct {
		P []int  `plenc:"1"`
		Q string `plenc:"2"`
	}

	const iterations = 500

	var wg sync.WaitGroup

	// Goroutine for TypeA
	wg.Go(func() {
		for i := range iterations {
			in := TypeA{X: i, Y: "type-a"}
			data, err := Marshal(nil, &in)
			if err != nil {
				t.Errorf("TypeA marshal error: %v", err)
				return
			}
			var out TypeA
			if err := Unmarshal(data, &out); err != nil {
				t.Errorf("TypeA unmarshal error: %v", err)
				return
			}
			if diff := cmp.Diff(in, out); diff != "" {
				t.Errorf("TypeA mismatch: %s", diff)
				return
			}
		}
	})

	// Goroutine for TypeB
	wg.Go(func() {
		for i := range iterations {
			in := TypeB{M: float64(i) * 1.5, N: i%2 == 0}
			data, err := Marshal(nil, &in)
			if err != nil {
				t.Errorf("TypeB marshal error: %v", err)
				return
			}
			var out TypeB
			if err := Unmarshal(data, &out); err != nil {
				t.Errorf("TypeB unmarshal error: %v", err)
				return
			}
			if diff := cmp.Diff(in, out); diff != "" {
				t.Errorf("TypeB mismatch: %s", diff)
				return
			}
		}
	})

	// Goroutine for TypeC
	wg.Go(func() {
		for i := range iterations {
			in := TypeC{P: []int{i, i + 1, i + 2}, Q: "type-c"}
			data, err := Marshal(nil, &in)
			if err != nil {
				t.Errorf("TypeC marshal error: %v", err)
				return
			}
			var out TypeC
			if err := Unmarshal(data, &out); err != nil {
				t.Errorf("TypeC unmarshal error: %v", err)
				return
			}
			if diff := cmp.Diff(in, out); diff != "" {
				t.Errorf("TypeC mismatch: %s", diff)
				return
			}
		}
	})

	wg.Wait()
}
