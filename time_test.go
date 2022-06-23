package plenc

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/philpearl/plenc/plenccore"
)

func TestTime(t *testing.T) {
	f := fuzz.New()

	type twrap struct {
		T time.Time `plenc:"1"`
		U int       `plenc:"2"`
	}

	c, err := CodecForType(reflect.TypeOf(twrap{}))
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

func BenchmarkTime(b *testing.B) {
	b.ReportAllocs()
	in := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		var data []byte
		for pb.Next() {
			var err error
			data, err = Marshal(data[:0], &in)
			if err != nil {
				b.Fatal(err)
			}
			var out time.Time
			if err := Unmarshal(data, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
