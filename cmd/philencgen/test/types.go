package test

import "time"

//go:generate philencgen -pkg github.com/philpearl/philenc/cmd/philencgen/test -type Struct2 -out la
//go:generate philencgen -pkg github.com/philpearl/philenc/cmd/philencgen/test -type MyStruct -out la

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
}
