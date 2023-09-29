package prototest

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/philpearl/plenc"
	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type pMsg1 struct {
	V1        int64             `plenc:"1"`
	V2        string            `plenc:"2"`
	V3        string            `plenc:"3"`
	Timestamp time.Time         `plenc:"4"`
	V5        int64             `plenc:"5,flat"`
	V6        map[string]string `plenc:"6,proto"`
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
	t0 := time.Date(1970, 3, 15, 13, 17, 0, 0, time.UTC)
	ts := timestamppb.New(t0)

	protoM := Msg2{
		V1: []*Msg1{
			{
				V1:        1,
				V2:        "2",
				V3:        "3",
				Timestamp: ts,
				V5:        5,
				V6:        map[string]string{"a": "b"},
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

	plencM := pMsg2{
		V1: []pMsg1{
			{
				V1:        1,
				V2:        "2",
				V3:        "3",
				Timestamp: t0,
				V5:        5,
				V6:        map[string]string{"a": "b"},
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

	t.Run("proto->plenc", func(t *testing.T) {
		data, err := proto.Marshal(&protoM)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Printf("%X\n", data)
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
		var p plenc.Plenc
		p.ProtoCompatibleTime = true
		p.RegisterDefaultCodecs()

		if err := p.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(&plencM, &out); diff != "" {
			t.Fatalf("result not as hoped. %s", diff)
		}
	})

	t.Run("plenc->proto", func(t *testing.T) {
		var p plenc.Plenc
		p.ProtoCompatibleTime = true
		p.ProtoCompatibleArrays = true
		p.RegisterDefaultCodecs()

		data, err := p.Marshal(nil, &plencM)
		if err != nil {
			t.Fatal(err)
		}

		var out Msg2
		if err := proto.Unmarshal(data, &out); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(&protoM, &out, cmpopts.IgnoreUnexported(Msg2{}, Msg1{}, timestamppb.Timestamp{})); diff != "" {
			t.Fatalf("result not as hoped. %s", diff)
		}
	})
}
