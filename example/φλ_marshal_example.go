package example

import (
	"github.com/philpearl/plenc"
)

// TODO: missing types
// slice of numeric ()
// slice of other
// pointers
// TODO: option whether top-level type is a pointer for marshaler

// ΦλSize works out how many bytes are needed to encode example
func (e *example) ΦλSize() (size int) {
	if e == nil {
		return 0
	}

	size += plenc.SizeTag(plenc.WTLength, 1)
	size += plenc.SizeString(e.Name)

	size += plenc.SizeTag(plenc.WTVarInt, 2)
	size += plenc.SizeVarInt(int64(e.Age))

	{
		var t plenc.Time
		t.Set(e.Starting)
		if s := t.ΦλSize(); s != 0 {
			size += plenc.SizeTag(plenc.WTLength, 3)
			size += plenc.SizeVarUint(uint64(s))
			size += s
		}

	}

	return size
}

// ΦλAppend encodes example by appending to data. It returns the final slice
func (e *example) ΦλAppend(data []byte) []byte {

	data = plenc.AppendTag(data, plenc.WTLength, 1)
	data = plenc.AppendString(data, e.Name)

	data = plenc.AppendTag(data, plenc.WTVarInt, 2)
	data = plenc.AppendVarInt(data, int64(e.Age))

	{
		var t plenc.Time
		t.Set(e.Starting)
		if s := t.ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, 3)
			data = plenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)
		}
	}

	return data
}
