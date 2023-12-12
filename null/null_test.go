package null

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
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

func TestNullEmpty(t *testing.T) {
	type TestThing struct {
		I null.Int    `plenc:"1"`
		B null.Bool   `plenc:"2"`
		F null.Float  `plenc:"3"`
		S null.String `plenc:"4"`
		T null.Time   `plenc:"5"`
	}
	var v TestThing

	data, err := plenc.Marshal(nil, &v)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("expected no data, got %x", data)
	}
}

func TestNullExplicit(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		out  interface{}
		exp  interface{}
	}{
		{
			name: "empty valid string",
			in:   &null.String{sql.NullString{Valid: true}},
			out:  &null.String{},
			exp:  &null.String{sql.NullString{Valid: true}},
		},
		{
			name: "non-empty valid string",
			in:   &null.String{sql.NullString{Valid: true, String: "a"}},
			out:  &null.String{},
			exp:  &null.String{sql.NullString{Valid: true, String: "a"}},
		},
		{
			name: "zero valid int",
			in:   &null.Int{sql.NullInt64{Valid: true, Int64: 0}},
			out:  &null.Int{},
			exp:  &null.Int{sql.NullInt64{Valid: true, Int64: 0}},
		},
		{
			name: "positive valid int",
			in:   &null.Int{sql.NullInt64{Valid: true, Int64: 1}},
			out:  &null.Int{},
			exp:  &null.Int{sql.NullInt64{Valid: true, Int64: 1}},
		},
		{
			name: "negative valid int",
			in:   &null.Int{sql.NullInt64{Valid: true, Int64: -1}},
			out:  &null.Int{},
			exp:  &null.Int{sql.NullInt64{Valid: true, Int64: -1}},
		},
		{
			name: "zero valid float",
			in:   &null.Float{sql.NullFloat64{Valid: true, Float64: 0}},
			out:  &null.Float{},
			exp:  &null.Float{sql.NullFloat64{Valid: true, Float64: 0}},
		},
		{
			name: "positive valid float",
			in:   &null.Float{sql.NullFloat64{Valid: true, Float64: 1}},
			out:  &null.Float{},
			exp:  &null.Float{sql.NullFloat64{Valid: true, Float64: 1}},
		},
		{
			name: "negative valid float",
			in:   &null.Float{sql.NullFloat64{Valid: true, Float64: -1}},
			out:  &null.Float{},
			exp:  &null.Float{sql.NullFloat64{Valid: true, Float64: -1}},
		},
		// Can't test NAN because NAN != NAN!
		{
			name: "false valid bool",
			in:   &null.Bool{sql.NullBool{Valid: true, Bool: false}},
			out:  &null.Bool{},
			exp:  &null.Bool{sql.NullBool{Valid: true, Bool: false}},
		},
		{
			name: "true valid bool",
			in:   &null.Bool{sql.NullBool{Valid: true, Bool: true}},
			out:  &null.Bool{},
			exp:  &null.Bool{sql.NullBool{Valid: true, Bool: true}},
		},
		{
			name: "zero valid time",
			in:   &null.Time{Valid: true},
			out:  &null.Time{},
			exp:  &null.Time{Valid: true},
		},
		{
			name: "non-zero valid time",
			in:   &null.Time{Valid: true, Time: time.Date(1970, 3, 15, 0, 0, 0, 0, time.UTC)},
			out:  &null.Time{},
			exp:  &null.Time{Valid: true, Time: time.Date(1970, 3, 15, 0, 0, 0, 0, time.UTC)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := plenc.Marshal(nil, test.in)
			if err != nil {
				t.Fatal(err)
			}

			if err := plenc.Unmarshal(data, test.out); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.exp, test.out); diff != "" {
				t.Logf("%x", data)
				t.Fatalf("structs differ. %s", diff)
			}
		})
	}
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

func TestNullDescription(t *testing.T) {
	type TestThing struct {
		I null.Int    `plenc:"1"`
		B null.Bool   `plenc:"2"`
		F null.Float  `plenc:"3"`
		S null.String `plenc:"4"`
		T null.Time   `plenc:"5"`
	}

	v := TestThing{
		I: null.IntFrom(77),
		S: null.StringFrom("cheese"),
	}

	data, err := plenc.Marshal(nil, v)
	if err != nil {
		t.Fatal(err)
	}

	c, err := plenc.CodecForType(reflect.TypeOf(v))
	if err != nil {
		t.Fatal(err)
	}
	d := c.Descriptor()
	var j plenccodec.JSONOutput
	if err := d.Read(&j, data); err != nil {
		t.Fatal(err)
	}

	out := string(j.Done())
	if out != "{\n  \"I\": 77,\n  \"S\": \"cheese\"\n}\n" {
		t.Fatal(out)
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
}
