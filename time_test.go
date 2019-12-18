package plenc

import (
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
)

func TestTime(t *testing.T) {
	f := fuzz.New()

	type twrap struct {
		T time.Time `plenc:"1"`
		U int       `plenc:"2"`
	}

	c, err := codecForType(reflect.TypeOf(twrap{}))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100000; i++ {
		var t0 twrap
		f.Fuzz(&t0)

		data := c.Append(nil, unsafe.Pointer(&t0))

		var t1 twrap
		n, err := c.Read(data, unsafe.Pointer(&t1))
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
