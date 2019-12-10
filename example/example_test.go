package example

import (
	"fmt"
	"time"
)

// Example shows how to use the generated functions to serialise and deserialise a struct
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

	// Output: Encoded as 0A0553696D6F6E10401A0808A89DFFDE0B1000
	// {Simon 32 2019-12-10 18:43:32 +0000 GMT}
}
