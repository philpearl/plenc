package test

import "time"

import "github.com/philpearl/plenc"

//go:generate plencgen -pkg github.com/philpearl/plenc/cmd/plencgen/test -type Struct2
//go:generate plencgen -pkg github.com/philpearl/plenc/cmd/plencgen/test -type MyStruct

type Struct2 struct {
	A uint8
	B string
}

type FunnyInt int16

type MyStruct struct {
	A int
	B uint
	C uint32
	D float32
	E float64
	F bool
	// G *bool // Have we got this?
	H Struct2
	I []Struct2
	J *Struct2
	K []*Struct2
	L FunnyInt
	M time.Time
	N *time.Time
	Q OptInt
}

type OptInt struct {
	Valid bool
	Value int
}

func (o OptInt) ΦλSizeFull(index int) int {
	if !o.Valid {
		return 0
	}
	return plenc.SizeTag(plenc.WTVarInt, index) + plenc.SizeVarInt(int64(o.Value))
}

func (o OptInt) ΦλAppendFull(data []byte, index int) []byte {
	if !o.Valid {
		return data
	}
	data = plenc.AppendTag(data, plenc.WTVarInt, index)
	return plenc.AppendVarInt(data, int64(o.Value))
}

func (o *OptInt) ΦλUnmarshal(data []byte) (n int, err error) {
	// Since we implement FullMarshaler we need to have the same semantics in out unmarshaler
	v, n := plenc.ReadVarInt(data)
	if n > 0 {
		o.Valid = true
		o.Value = int(v)
	}
	return n, nil
}
