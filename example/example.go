package example

import (
	"fmt"
	"time"
)

//go:generate plencgen -pkg github.com/philpearl/plenc/example -type example

type example struct {
	Name     string
	Age      int
	Starting time.Time
}

// Example shows how to use the generated functions
func Example() {

	e := example{
		Name:     "Simon",
		Age:      32,
		Starting: time.Date(2019, 12, 10, 18, 43, 32, 0, time.UTC),
	}

	buf := make([]byte, 0, e.ΦλSize())
	buf = e.ΦλAppend(buf)
	fmt.Printf("Encoded as %X\n", buf)

	// Decode the data into a new variable
	var e2 example

	_, err := e2.ΦλUnmarshal(buf)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(e2)
}
