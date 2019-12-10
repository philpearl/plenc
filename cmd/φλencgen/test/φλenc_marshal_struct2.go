package test

import (
	"github.com/philpearl/φλenc"
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

	size += φλenc.SizeTag(φλenc.WTVarInt, 1)
	size += φλenc.SizeVarUint(uint64(e.A))

	size += φλenc.SizeTag(φλenc.WTLength, 2)
	size += φλenc.SizeString(e.B)

	return size
}

// ΦλAppend encodes Struct2 by appending to data. It returns the final slice
func (e *Struct2) ΦλAppend(data []byte) []byte {

	data = φλenc.AppendTag(data, φλenc.WTVarInt, 1)
	data = φλenc.AppendVarUint(data, uint64(e.A))

	data = φλenc.AppendTag(data, φλenc.WTLength, 2)
	data = φλenc.AppendString(data, e.B)

	return data
}
