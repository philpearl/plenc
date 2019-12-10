
# plenc

[![GoDoc](https://godoc.org/github.com/philpearl/plenc?status.svg)](https://godoc.org/github.com/philpearl/plenc) 
[![Build Status](https://travis-ci.org/philpearl/plenc.svg)](https://travis-ci.org/philpearl/plenc)

plenc is a serialisation library based around protobuf. It uses a very similar encoding to protobuf, but it does not use .proto files or the protobuf data definition language. Instead Go structs are used to define how messages are encoded.

plenc works using code generation. The included plencgen tool takes care of code generation. plencgen generates code for the plenc Marshaler and Unmarshaler interfaces. As well as the code, plencgen creates a .φλ file for each struct. These files track the indices that are allocated to each struct field as the struct changes over time. This means you can add and remove fields from the structs and still be able to read encoded data. The .φλ files ensures field indices are not re-used for new fields.

You should add all the generated files - both the code and the .φλ files - to your version control system.

I wanted to call it φλenc, but github doesn't seem to allow non-latin repo names!

## Example
Create a struct just as you normally would. Add a go:generate line to document the plencgen command needed.

```go
//go:generate plencgen -pkg github.com/philpearl/plenc/example -type example

type example struct {
	Name     string
	Age      int
	Starting time.Time
}
```

This will create 3 methods. `ΦλSize` & `ΦλAppend` are used to serialise the struct. `ΦλUnmarshal` is used to de-serialise.

You can serialise data as follows.

```go
	e := example{
		Name:     "Simon",
		Age:      32,
		Starting: time.Date(2019, 12, 10, 18, 43, 32, 0, time.UTC),
	}

	buf := make([]byte, 0, e.ΦλSize())
	buf = e.ΦλAppend(buf)

```

Deserialising is also straightforward.

```go
	var e2 example

	_, err := e2.ΦλUnmarshal(buf)
	if err != nil {
		fmt.Println(err)
	}
```