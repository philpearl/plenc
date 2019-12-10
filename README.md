
# φλenc

[![GoDoc](https://godoc.org/github.com/philpearl/φλenc?status.svg)](https://godoc.org/github.com/philpearl/φλenc) 
[![Build Status](https://travis-ci.org/philpearl/φλenc.svg)](https://travis-ci.org/philpearl/φλenc)

φλenc is a serialisation library based around protobuf. It uses a very similar encoding to protobuf, but it does not use .proto files or the protobuf data definition language. Instead Go structs are used to define how messages are encoded.

φλenc works using code generation. The included φλencgen tool takes care of code generation. φλencgen generates code for the φλenc Marshaler and Unmarshaler interfaces. As well as the code, φλencgen creates a .φλ file for each struct. These track the indices allocated to each struct field over time. This means you can add and remove fields from the structs and still be able to read encoded data as the .φλ files ensures field indices are not re-used for new fields.

You should add all the generated files, both the code and the .φλ files, to your version control system.

