package plenc_test

import (
	"fmt"

	"github.com/philpearl/plenc"
)

func Example() {
	type Hat struct {
		Type string  `plenc:"1"`
		Size float32 `plenc:"2"`
	}

	type Person struct {
		Name string `plenc:"1"`
		Age  int    `plenc:"2"`
		Hats []Hat  `plenc:"3"`
	}

	var me = Person{
		Name: "Lucy",
		Age:  25,
		Hats: []Hat{
			{Type: "Fedora", Size: 6},
			{Type: "floppy", Size: 5.5},
		},
	}

	data, err := plenc.Marshal(nil, &me)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%X\n", data)

	var out Person
	if err := plenc.Unmarshal(data, &out); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n", out)

	// Output: 260A044C75637910321A1C0D0A064665646F7261150000C0400D0A06666C6F707079150000B040
	// plenc_test.Person{Name:"Lucy", Age:25, Hats:[]plenc_test.Hat{plenc_test.Hat{Type:"Fedora", Size:6}, plenc_test.Hat{Type:"floppy", Size:5.5}}}
}
