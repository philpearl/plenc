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
		I null.Int
		B null.Bool
		F null.Float
		S null.String
		T null.Time
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
	type TestThing struct {
		I  null.Int
		I2 null.Int
		B  null.Bool
		B2 null.Bool
		F  null.Float
		S  null.String
		T  null.Time
	}

	v := TestThing{
		I: null.IntFrom(42),
		B: null.BoolFrom(true),
		F: null.FloatFrom(3.14),
		S: null.StringFrom("jhsdfkjahskfhkjhsdkf"),
	}

	b.Run("plenc", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var data []byte
			for pb.Next() {
				var err error
				data, err = plenc.Marshal(data[:0], &v)
				if err != nil {
					b.Fatal(err)
				}

				var w TestThing
				if err := plenc.Unmarshal(data, &w); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var err error
				data, err := json.Marshal(&v)
				if err != nil {
					b.Fatal(err)
				}

				var w TestThing
				if err := json.Unmarshal(data, &w); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("easyjson", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var err error
				data, err := easyjson.Marshal(&v)
				if err != nil {
					b.Fatal(err)
				}

				var w TestThing
				if err := easyjson.Unmarshal(data, &w); err != nil {
					b.Fatal(err)
				}
			}
		})
	})

}
