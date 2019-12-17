package plenc

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
)

func TestMarshal(t *testing.T) {

	type TestThing struct {
		A  float64     `plenc:"1"`
		B  []float64   `plenc:"2"`
		C  *float64    `plenc:"3"`
		D  float32     `plenc:"4"`
		E  []float32   `plenc:"5"`
		F  *float32    `plenc:"6"`
		G  int         `plenc:"7"`
		H  []int       `plenc:"8"`
		I  *int        `plenc:"9"`
		J  uint        `plenc:"10"`
		K  []uint      `plenc:"11"`
		L  *uint       `plenc:"12"`
		M  bool        `plenc:"13"`
		N  []bool      `plenc:"14"`
		O  *bool       `plenc:"15"`
		P  string      `plenc:"16"`
		Q  []string    `plenc:"17"`
		R  *string     `plenc:"18"`
		S  time.Time   `plenc:"19"`
		T  []time.Time `plenc:"20"`
		U  *time.Time  `plenc:"21"`
		V  int32       `plenc:"22"`
		W  []int32     `plenc:"23"`
		X  *int32      `plenc:"24"`
		Y  int64       `plenc:"25"`
		Z  []int64     `plenc:"26"`
		A1 *int64      `plenc:"27"`
		// A2 map[string]string `plenc:"28"`
	}

	f := fuzz.New()
	for i := 0; i < 10000; i++ {
		var in TestThing
		f.Fuzz(&in)

		data, err := Marshal(nil, &in)
		if err != nil {
			t.Fatal(err)
		}

		var out TestThing
		if err := Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(in, out); diff != "" {
			t.Logf("%x", data)

			var out TestThing
			if err := Unmarshal(data, &out); err != nil {
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

func TestMarshal2(t *testing.T) {
	type TestThing struct {
		A float64   `plenc:"1"`
		B []float64 `plenc:"2"`
		C *float64  `plenc:"3"`
		D float32   `plenc:"4"`
		E []float32 `plenc:"5"`
		F *float32  `plenc:"6"`
		G int       `plenc:"7"`
		H []int     `plenc:"8"`
		I *int      `plenc:"9"`
		J uint      `plenc:"10"`
		K []uint    `plenc:"11"`
		L *uint     `plenc:"12"`
		M bool      `plenc:"13"`
		N []bool    `plenc:"14"`
		O *bool     `plenc:"15"`
		P string    `plenc:"16"`
		Q []string  `plenc:"17"`
		R *string   `plenc:"18"`
		// S time.Time
		// T []time.Time
		// U *time.Time
	}

	exp := []string{
		"K煐ǂś瘑Ŧ擋燕鼪沭沲脞{.ĺ±",
		"螪鿄}佺-ǃƣ蕩柈Ĕ憑h9",
		"Ŵ>.Ő醌",
		"鍲闝鯕遜閼xƇEwĢɠ鈱",
	}

	var buf []byte
	fmt.Sscanf(
		"d0030915eaa21353a3dc3f1230ea90e7b64bdd913f5f86fcdb93d9d03faf0d40a30d51eb3fae5ea001fbebd83f5d28a64a13f7dc3f2f8b562fa6ceeb3f19c235b45b58d8eb3f2548132b3e2a20f036263ee9df703f3ac2103ddd010c3e6db5d53e69940b3f97d34d3f0081b33e3889ee9e9d9ea8b68528422eed8a8996828ef48a10d5fbced2c7b8b49a0ff8ce92c590e787949f01c4caf4a796e9adef24fbcfd2f4c2c5e5b21248c5ecdbe1d2b6dec6095084effbfb9db29ce2da015a5697b7e7c5b8a9b5abfd019dc7def0d7d4c2c14395b6aef1f28b91d805e5fff5f8dcd7fdaf31b4ae85ebabe1e383ba01fab6bba9b796a9b89c01b5b48cc69af4ea9ba901d3ffc1e584f2e1b29b01a1de83ecbba2f2ec0a60e795edf8eab0ccff4a6800720400000100780082012132443d7659e792b8e6ab80e69ea7caa2603cc3a4c5a75564e6858453e6bf8bc58d8a0169254be78590c782c59be79891c5a6e6938be78795e9bcaae6b2ade6b2b2e8849e7b2ec4bac2b11ce89eaae9bf847de4bdba2dc783c6a3e895a9e69f88c494e68691683909c5b43e2ec590e9868c1be98db2e9979de9af95e9819ce996bc78c6874577c4a2c9a0e988b1920117e6ab924027e89aa9c5b8474ac5bd7de8b280453ce6aeb0",
		"%X",
		&buf,
	)

	for i := 0; i < 100000; i++ {
		var out TestThing
		if err := Unmarshal(buf, &out); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(exp, out.Q); diff != "" {
			t.Fatal(diff)
		}
	}
}
