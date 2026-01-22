package plenccodec_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
)

func TestTime(t *testing.T) {
	var p1, p2, p3 plenc.Plenc
	p1.ProtoCompatibleTime = true
	p1.RegisterDefaultCodecs()
	p2.RegisterDefaultCodecs()
	p3.RegisterDefaultCodecs()
	p3.RegisterCodec(reflect.TypeFor[time.Time](), plenccodec.BQTimestampCodec{})

	f := fuzz.New()

	type twrap struct {
		T time.Time `plenc:"1"`
		U int       `plenc:"2"`
	}

	for j, p := range []*plenc.Plenc{&p1, &p2, &p3} {
		for range 100_000 {
			var t0 twrap
			f.Fuzz(&t0)

			if j == 2 {
				t0.T = t0.T.Truncate(time.Microsecond)
			}

			data, err := p.Marshal(nil, &t0)
			if err != nil {
				t.Fatal(err)
			}

			var t1 twrap
			if err := p.Unmarshal(data, &t1); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(t0, t1); diff != "" {
				t.Logf("unix %d %d", t0.T.Unix(), t0.T.Nanosecond())
				t.Fatalf("differs. %s", diff)
			}
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
