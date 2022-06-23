package plenccodec_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc/plenccodec"
)

func TestJSONOutput(t *testing.T) {
	tests := []struct {
		sequence func(j *plenccodec.JSONOutput)
		exp      string
	}{
		{
			sequence: func(j *plenccodec.JSONOutput) {
				j.StartObject()
				j.NameField("fred")
				j.Bool(true)
				j.NameField("brian")
				j.StartArray()
				j.Int64(1)
				j.Int64(2)
				j.Int64(3)
				j.EndArray()
				j.NameField("sheila")
				j.StartObject()
				j.EndObject()
				j.EndObject()
			},
			exp: `{
  "fred": true,
  "brian": [
    1,
    2,
    3
  ],
  "sheila": {
  }
}
`,
		},
		{
			sequence: func(j *plenccodec.JSONOutput) {
				j.String(`	£∞§¶•ªº˙©"ƒ∂µµµPhi\l`)
			},
			exp: `"\t£∞§¶•ªº˙©\"ƒ∂µµµPhi\\l"
`,
		},

		{
			sequence: func(j *plenccodec.JSONOutput) {
				j.Bool(true)
			},
			exp: "true\n",
		},
		{
			sequence: func(j *plenccodec.JSONOutput) {
				j.Time(time.Date(1970, 3, 15, 0, 0, 0, 0, time.UTC))
			},
			exp: "\"1970-03-15T00:00:00Z\"\n",
		},

		{
			sequence: func(j *plenccodec.JSONOutput) {
				j.StartObject()
				j.NameField(`a§a§"a`)
				j.String("a˙©ƒ∂†¥˚˙")
				j.EndObject()
			},
			exp: `{
  "a§a§\"a": "a˙©ƒ∂†¥˚˙"
}
`,
		},
	}

	var j plenccodec.JSONOutput

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			j.Reset()
			test.sequence(&j)
			out := j.Done()

			if diff := cmp.Diff(test.exp, string(out)); diff != "" {
				t.Fatal(diff)
			}

			if !json.Valid(out) {
				t.Fatal("output is not valid JSON")
			}
		})
	}
}
