package null

import "github.com/unravelin/null"

//go:generate easyjson -no_std_marshalers -pkg

//easyjson:json
type BenchThing struct {
	I  null.Int    `plenc:"1"`
	I2 null.Int    `plenc:"2"`
	B  null.Bool   `plenc:"3"`
	B2 null.Bool   `plenc:"4"`
	F  null.Float  `plenc:"5"`
	S  null.String `plenc:"6"`
	T  null.Time   `plenc:"7"`
}
