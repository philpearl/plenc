package null

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/mailru/easyjson"
	"github.com/philpearl/plenc"
	"github.com/unravelin/null"
)

func init() {
	RegisterCodecs()
}

var fuzzFuncs = []interface{}{
	func(a *null.Bool, c fuzz.Continue) {
		a.Valid = c.RandBool()
		if a.Valid {
			a.Bool = c.RandBool()
		}
	},
	func(a *null.Float, c fuzz.Continue) {
		a.Valid = c.RandBool()
		if a.Valid {
			c.Fuzz(&a.Float64)
		}
	},
	func(a *null.Int, c fuzz.Continue) {
		a.Valid = c.RandBool()
		if a.Valid {
			c.Fuzz(&a.Int64)
		}
	},
	func(a *null.String, c fuzz.Continue) {
		a.Valid = c.RandBool()
		if a.Valid {
			c.Fuzz(&a.String)
		}
	},
	func(a *null.Time, c fuzz.Continue) {
		a.Valid = c.RandBool()
		if a.Valid {
			c.Fuzz(&a.Time)
		}
	},
}

func TestNull(t *testing.T) {
	type TestThing struct {
		I null.Int    `plenc:"1"`
		B null.Bool   `plenc:"2"`
		F null.Float  `plenc:"3"`
		S null.String `plenc:"4"`
		T null.Time   `plenc:"5"`
	}

	f := fuzz.New().Funcs(fuzzFuncs...)
	for i := 0; i < 10000; i++ {
		var in TestThing
		f.Fuzz(&in)

		data, err := plenc.Marshal(nil, &in)
		if err != nil {
			t.Fatal(err)
		}

		var out TestThing
		if err := plenc.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(in, out); diff != "" {
			t.Logf("%x", data)

			var out TestThing
			if err := plenc.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(in, out); diff != "" {
				t.Logf("re-run differs too")
			} else {
				t.Logf("re-run does not differ")
			}

			t.Fatalf("structs differ. %s", diff)
		}
	}
}

func BenchmarkNull(b *testing.B) {
	v := benchThing{
		I: null.IntFrom(42),
		B: null.BoolFrom(true),
		F: null.FloatFrom(3.14),
		S: null.StringFrom("jhsdfkjahskfhkjhsdkf"),
		U: null.StringFrom("hat"),
	}

	b.Run("plenc", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var data []byte
			var w benchThing
			for pb.Next() {
				var err error
				data, err = plenc.Marshal(data[:0], &v)
				if err != nil {
					b.Fatal(err)
				}

				w = benchThing{}
				if err := plenc.Unmarshal(data, &w); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var w benchThing
			for pb.Next() {
				var err error
				data, err := json.Marshal(&v)
				if err != nil {
					b.Fatal(err)
				}

				w = benchThing{}
				if err := json.Unmarshal(data, &w); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("easyjson", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var w benchThing
			for pb.Next() {
				var err error
				data, err := easyjson.Marshal(&v)
				if err != nil {
					b.Fatal(err)
				}

				w = benchThing{}
				if err := easyjson.Unmarshal(data, &w); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
