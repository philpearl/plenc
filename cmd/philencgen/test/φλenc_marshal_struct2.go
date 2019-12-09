package test

import (
	"github.com/philpearl/philenc"
)

// TODO: missing types
// slice of numeric ()
// slice of other
// pointers
// TODO: option whether top-level type is a pointer for marshaler

func (e *Struct2) ΦλSize() (size int) {

	size += philenc.SizeTag(philenc.WTVarInt, 1)
	size += philenc.SizeVarUint(uint(e.A))

	size += philenc.SizeTag(philenc.WTLength, 2)
	size += philenc.SizeString(e.B)

	return size
}

func (e *Struct2) ΦλAppend(data []byte) []byte {

	data = philenc.AppendTag(data, philenc.WTVarInt, 1)
	data = philenc.AppendVarUint(data, uint(e.A))

	data = philenc.AppendTag(data, philenc.WTLength, 2)
	data = philenc.AppendString(data, e.B)

	return data
}
