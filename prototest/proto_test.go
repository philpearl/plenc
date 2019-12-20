package prototest

import (
	fmt "fmt"
	"testing"

	proto "github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/philpearl/plenc"
)

type pMsg1 struct {
	V1 int64  `plenc:"1"`
	V2 string `plenc:"2"`
	V3 string `plenc:"3"`
}
type pMsg2 struct {
	V1 []pMsg1   `plenc:"1"`
	V2 []uint64  `plenc:"2"`
	V3 []string  `plenc:"3"`
	V4 []float32 `plenc:"4"`
	V5 []float64 `plenc:"5"`
	V6 []uint32  `plenc:"6"`
	V7 []int32   `plenc:"7"`
	V8 []byte    `plenc:"8"`
}

func TestProto(t *testing.T) {
	// Here we're testing that we can decode proto-encoded data.
	// Note that going the other way is not an aim.
	m := Msg2{
		V1: []*Msg1{
			{
				V1: 1,
				V2: "2",
				V3: "3",
			},
			{
				V1: 5,
				V2: "6",
				V3: "7",
			},
		},
		V2: []uint64{4, 7, 128},
		V3: []string{"a", "b", "cd"},
		V4: []float32{1, 2, 3, 4},
		V5: []float64{5, 6, 7, 8},
		V6: []uint32{32, 22, 3444, 443},
		V7: []int32{-17, 12, 198},
		V8: []byte{234, 9, 17},
	}

	data, err := proto.Marshal(&m)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%X\n", data)
	/*
	 0A0808021201321A01330A08080A1201361A0137120404078001
	 0A (1, WTlength)
	   08 0801 120132 1A0133
	 0A (1, WTlength)
	   08 0805 120136 1A0137
	 12 (2, WTLength)
	   04 04 07 8001

	 0A
	   08 0802 120132 1A0133
	 0A
	   08 080A 120136 1A0137
	 12
	   04 04 07 8001
	 1A
	   01 61
	 1A
	   01 62
	 1A
	   02 63 64

	   Packed WT32 WT64 WTVarInt
	   Not Packed WTLENGTH

	   0A0808021201321A01330A08080A1201361A0137120404078001
	         021201321A01330A
	*/

	var out pMsg2
	if err := plenc.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	exp := pMsg2{
		V1: []pMsg1{
			{
				V1: 1,
				V2: "2",
				V3: "3",
			},
			{
				V1: 5,
				V2: "6",
				V3: "7",
			},
		},
		V2: []uint64{4, 7, 128},
		V3: []string{"a", "b", "cd"},
		V4: []float32{1, 2, 3, 4},
		V5: []float64{5, 6, 7, 8},
		V6: []uint32{32, 22, 3444, 443},
		V7: []int32{-17, 12, 198},
		V8: []byte{234, 9, 17},
	}

	if diff := cmp.Diff(&exp, &out); diff != "" {
		t.Fatalf("result not as hoped. %s", diff)
	}
}
