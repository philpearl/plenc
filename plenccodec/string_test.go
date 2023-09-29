package plenccodec_test

import (
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

func TestString(t *testing.T) {
	tests := []string{
		"",
		"a",
		"¬πø∆˙©¥å∫˜",
		"this is a string",
		"ɚ&d珿ȨDT葚µ噘",
		"\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00",
		`¶娋搱ĚoÞB@\贞敠ơƸŜ`,
		"\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x03\x00\x00\x00",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			c, err := plenc.CodecForType(reflect.TypeOf(test))
			if err != nil {
				t.Fatal(err)
			}

			l := c.Size(unsafe.Pointer(&test), nil)

			data := c.Append(nil, unsafe.Pointer(&test), nil)

			if len(data) != l {
				t.Errorf("data not expected length. %d %d", l, len(data))
			}

			var out string
			_, err = c.Read(data, unsafe.Pointer(&out), c.WireType())
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestStringMarshal(t *testing.T) {
	tests := []string{
		"",
		"a",
		"¬πø∆˙©¥å∫˜",
		"this is a string",
		"ɚ&d珿ȨDT葚µ噘",
		"\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00",
		`¶娋搱ĚoÞB@\贞敠ơƸŜ`,
		"\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x03\x00\x00\x00",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			data, err := plenc.Marshal(nil, test)
			if err != nil {
				t.Fatal(err)
			}

			var out string
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestInternedString(t *testing.T) {
	var c plenccodec.InternedStringCodec

	values := []string{
		"hat", "cheese", "elephant", "hat", "hat", "cheese", "elephant",
	}

	var data []byte
	allocs := testing.AllocsPerRun(1000, func() {
		for _, test := range values {
			l := c.Size(unsafe.Pointer(&test), nil)
			data = c.Append(data[:0], unsafe.Pointer(&test), nil)

			if len(data) != l {
				t.Errorf("data not expected length. %d %d", l, len(data))
			}

			var out string
			_, err := c.Read(data, unsafe.Pointer(&out), c.WireType())
			if err != nil {
				t.Fatal(err)
			}

			if out != test {
				t.Fatalf("mismatch %q, %q", test, out)
			}

		}
	})

	if allocs > 0.1 {
		t.Fatal(allocs)
	}
}

func TestInternedStringParallel(t *testing.T) {
	type my struct {
		V string `plenc:"1,intern"`
	}

	values := []string{
		"hat", "cheese", "elephant", "hat", "hat", "cheese", "elephant",
	}
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			var data []byte
			var val, out my
			for i := 0; i < 1000; i++ {
				for _, test := range values {
					val.V = test
					data, err := plenc.Marshal(data[:0], &val)
					if err != nil {
						t.Error(err)
						return
					}

					if err := plenc.Unmarshal(data, &out); err != nil {
						t.Error(err)
						return
					}

					if out.V != test {
						t.Errorf("mismatch %q, %q", test, out)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func BenchmarkInternedString(b *testing.B) {
	type myb struct {
		V string `plenc:"1,intern"`
	}

	values := []string{
		"hat", "cheese", "elephant", "hat", "hat", "cheese", "elephant",
	}
	var data []byte
	var val, out myb
	for _, test := range values {
		val.V = test
		var err error
		data, err = plenc.Marshal(data[:0], &val)
		if err != nil {
			b.Fatal(err)
		}

		if err := plenc.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}

		if out.V != test {
			b.Fatalf("mismatch %q, %q", test, out)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out = myb{}
		if err := plenc.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}

		if out.V != "elephant" {
			b.Fatalf("mismatch %q, %q", "elephant", out)
		}
	}
}

func TestStringSlice(t *testing.T) {
	v := []string{
		"M珣X觻%ƾƽ9J9S腸H滩Ýk",
		"Eŕ漠",
		"织睱Ħ7õ咖Ê騄",
		"t沋晛岊ıƭ宋Yȯ¿q&",
		"Ʊãʙ#訃睩愴émė6Ɍ邔5汚鞗Ƈ",
		"桏&",
		"?Ȗ曽Ȯɕ稌!r囮ǯWQ猒÷飹嫗MJ",
		"",
	}

	c, err := plenc.CodecForType(reflect.TypeOf(v))
	if err != nil {
		t.Fatal(err)
	}

	data := c.Append(nil, unsafe.Pointer(&v), nil)

	var out []string
	_, err = c.Read(data, unsafe.Pointer(&out), c.WireType())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(v, out); diff != "" {
		t.Fatal(diff)
	}
}

func TestStringSliceCompat(t *testing.T) {
	type my struct {
		V []string `plenc:"1"`
	}
	v := my{V: []string{
		"M珣X觻%ƾƽ9J9S腸H滩Ýk",
		"Eŕ漠",
		"织睱Ħ7õ咖Ê騄",
		"t沋晛岊ıƭ宋Yȯ¿q&",
		"Ʊãʙ#訃睩愴émė6Ɍ邔5汚鞗Ƈ",
		"桏&",
		"?Ȗ曽Ȯɕ稌!r囮ǯWQ猒÷飹嫗MJ",
		"",
		"cheese it",
		"hats",
	}}

	var pl plenc.Plenc
	pl.ProtoCompatibleArrays = true
	pl.RegisterDefaultCodecs()

	data, err := pl.Marshal(nil, v)
	if err != nil {
		t.Fatal(err)
	}

	var out my

	if err := pl.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(v, out); diff != "" {
		t.Fatal(diff)
	}
}
