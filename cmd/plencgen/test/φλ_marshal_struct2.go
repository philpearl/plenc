package test

import (
	"github.com/philpearl/plenc"
)

// TODO: missing types
// slice of numeric ()
// slice of other
// pointers
// TODO: option whether top-level type is a pointer for marshaler

// ΦλSize works out how many bytes are needed to encode Struct2
func (e *Struct2) ΦλSize() (size int) {
	if e == nil {
		return 0
	}

	size += plenc.SizeTag(plenc.WTVarInt, 1)
	size += plenc.SizeVarUint(uint64(e.A))

	size += plenc.SizeTag(plenc.WTLength, 2)
	size += plenc.SizeString(e.B)

	return size
}

// ΦλAppend encodes Struct2 by appending to data. It returns the final slice
func (e *Struct2) ΦλAppend(data []byte) []byte {

	data = plenc.AppendTag(data, plenc.WTVarInt, 1)
	data = plenc.AppendVarUint(data, uint64(e.A))

	data = plenc.AppendTag(data, plenc.WTLength, 2)
	data = plenc.AppendString(data, e.B)

	return data
}
