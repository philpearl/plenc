package example

import (
	"time"
)

//go:generate plencgen -pkg github.com/philpearl/plenc/example -type example

type example struct {
	Name     string
	Age      int
	Starting time.Time
}
