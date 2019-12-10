
# plenc

[![GoDoc](https://godoc.org/github.com/philpearl/plenc?status.svg)](https://godoc.org/github.com/philpearl/plenc) 
[![Build Status](https://travis-ci.org/philpearl/plenc.svg)](https://travis-ci.org/philpearl/plenc)

plenc is a serialisation library based around protobuf. It uses a very similar encoding to protobuf, but it does not use .proto files or the protobuf data definition language. Instead Go structs are used to define how messages are encoded.

plenc works using code generation. The included plencgen tool takes care of code generation. plencgen generates code for the plenc Marshaler and Unmarshaler interfaces. As well as the code, plencgen creates a .φλ file for each struct. These track the indices allocated to each struct field over time. This means you can add and remove fields from the structs and still be able to read encoded data as the .φλ files ensures field indices are not re-used for new fields.

You should add all the generated files, both the code and the .φλ files, to your version control system.

I wanted to call it φλenc, but github doesn't seem to allow non-latin repo names!

