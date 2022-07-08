
# plenc

[![GoDoc](https://godoc.org/github.com/philpearl/plenc?status.svg)](https://godoc.org/github.com/philpearl/plenc) 

plenc is a serialisation library based around protobuf. It uses a very similar encoding to protobuf, but it does not use .proto files or the protobuf data definition language. Instead Go structs are used to define how messages are encoded.

plenc needs you to annotate your structs with a plenc tag on each field. The tag either indicates that the field should not be encoded, or provides a persistent index number that's used for that field in the encoding. The indexes within a struct must all be unique and should not be changed. You may remove fields, but you should not re-use the index number of a removed field. You should not change the type of a field. You can change field names as these are not used in the encoding.

Tags look like the following.

```go
type mystruct struct {
	A  int     `plenc:"1"`
	B  string  `plenc:"-"` // This field is not encoded
	C  float64 `plenc:"2"`
	// The values of this field are interned. This reduces allocations if
	// there are a limited set of distinct values used.
	D  string  `plenc:"3,intern"`
}
```

The `plenctag` tool will add tags to structs for you.

plenc only encodes fields that are exported - ones where the field name begins with a capital letter.

Once you've added plenc tags to your structs then encoding and decoding looks very like the JSON standard library. The one difference is that the Marshal function allows you to append encoded data to an existing slice.

```go
var in mystruct 
data, err := plenc.Marshal(data[:0], &in)
if err != nil {
	return err
}

var out mystruct
if err := plenc.Unmarshal(data, &out); err != nil {
	return err
}
```

## Why do this?

The idea behind plenc is to unlock the performance of protobuf for folk who don't like the Go structs generated by the protobuf compiler and don't want the hassle of creating .proto files. It is for people who want to retrofit better serialisation to a system that's started with JSON.

Here's a rough benchmark to show the kind of gains you could get using plenc.

```
BenchmarkCycle/plenc-16    1369533     881 ns/op    1400 B/op     38 allocs/op
BenchmarkCycle/json-16      214154    5620 ns/op    5211 B/op    120 allocs/op
```

## Is this protobuf?
No, as it stands it is not quite protobuf. It is largely protobuf and has a soft aim to at least be able to read standard protobuf, but there are differences.

The big difference is that plenc uses its own encoding for slices of types that are implemented with WTLength. Plenc introduces a new wire-type for these - WTSlice (3). The Tag byte is followed by a unsigned varint containing the number of elements in the slice, then each element is encoded with its length as an unsigned varint then the element encoding. This encoding means the receiver easily knows the length of the slice and can allocate it in a single operation.

Plenc does aim to be able to read standard protobuf. It can read slices encoded with the standard protobuf encoding. There may be gaps in support. 

In particular using fixed32 and fixed64 encodings for integer types is not currently supported. I think we could support that via an option on the plenc tag that would select a different codec.

## Slices
Neither plenc nor protobuf distinuguish between empty and nil slices. 
If you write an empty slice it will read back as a nil slice.

Slices of pointers to floats are not allowed.

Nils within slices of pointers are not supported. 
Nils in slices of pointer to integers will be omitted. 
Nils in slices of pointers to structs will be converted to empty structs.