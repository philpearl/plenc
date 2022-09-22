package plenccodec_test

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccore"
)

func TestTime(t *testing.T) {
	f := fuzz.New()

	type twrap struct {
		T time.Time `plenc:"1"`
		U int       `plenc:"2"`
	}

	c, err := plenc.CodecForType(reflect.TypeOf(twrap{}))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100000; i++ {
		var t0 twrap
		f.Fuzz(&t0)

		data := c.Append(nil, unsafe.Pointer(&t0))

		var t1 twrap
		n, err := c.Read(data, unsafe.Pointer(&t1), plenccore.WTLength)
		if n != len(data) {
			t.Errorf("not all data read. %d", n)
		}
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(t0, t1); diff != "" {
			t.Logf("unix %d %d", t0.T.Unix(), t0.T.Nanosecond())
			t.Fatalf("differs. %s", diff)
		}
	}
}

func TestTimeMarshal(t *testing.T) {
	tests := []time.Time{
		{},
		time.Date(1970, 3, 15, 0, 0, 0, 0, time.UTC),
		time.Now(),
	}

	for _, test := range tests {
		t.Run(test.String(), func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out time.Time
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if !out.Equal(test) {
				t.Fatalf("times %s and %s differ", out, test)
			}
		})
	}
}

func TestTimeInStructMarshal(t *testing.T) {
	type twrap struct {
		T time.Time `plenc:"1"`
		U int       `plenc:"2"`
	}

	tests := []twrap{
		{},
		{U: 1},
		{T: time.Date(1970, 3, 15, 0, 0, 0, 0, time.UTC)},
		{T: time.Now()},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			data, err := plenc.Marshal(nil, &test)
			if err != nil {
				t.Fatal(err)
			}
			var out twrap
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func BenchmarkTime(b *testing.B) {
	b.ReportAllocs()
	in := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		var data []byte
		var out time.Time
		for pb.Next() {
			var err error
			data, err = plenc.Marshal(data[:0], &in)
			if err != nil {
				b.Fatal(err)
			}
			if err := plenc.Unmarshal(data, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
