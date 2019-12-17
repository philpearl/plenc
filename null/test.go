package null

import "github.com/unravelin/null"

//go:generate easyjson -no_std_marshalers -pkg

//easyjson:json
type BenchThing struct {
	I  null.Int
	I2 null.Int
	B  null.Bool
	B2 null.Bool
	F  null.Float
	S  null.String
	T  null.Time
}
