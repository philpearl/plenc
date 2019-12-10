package test

import (
	"github.com/philpearl/plenc"
)

// TODO: missing types
// slice of numeric ()
// slice of other
// pointers
// TODO: option whether top-level type is a pointer for marshaler

// ΦλSize works out how many bytes are needed to encode MyStruct
func (e *MyStruct) ΦλSize() (size int) {
	if e == nil {
		return 0
	}

	size += plenc.SizeTag(plenc.WTVarInt, 1)
	size += plenc.SizeVarInt(int64(e.A))

	size += plenc.SizeTag(plenc.WTVarInt, 2)
	size += plenc.SizeVarUint(uint64(e.B))

	size += plenc.SizeTag(plenc.WTVarInt, 3)
	size += plenc.SizeVarUint(uint64(e.C))

	size += plenc.SizeTag(plenc.WT32, 4)
	size += plenc.SizeFloat32(float32(e.D))

	size += plenc.SizeTag(plenc.WT64, 5)
	size += plenc.SizeFloat64(float64(e.E))

	size += plenc.SizeTag(plenc.WTVarInt, 6)
	size += plenc.SizeBool(e.F)

	if s := e.H.ΦλSize(); s != 0 {
		size += plenc.SizeTag(plenc.WTLength, 7)
		size += plenc.SizeVarUint(uint64(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			size += plenc.SizeTag(plenc.WTLength, 8)
			size += plenc.SizeVarUint(uint64(s))
			size += s
		}
	}

	if s := e.J.ΦλSize(); s != 0 {
		size += plenc.SizeTag(plenc.WTLength, 9)
		size += plenc.SizeVarUint(uint64(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.K {
		if s := e.K[i].ΦλSize(); s != 0 {
			size += plenc.SizeTag(plenc.WTLength, 10)
			size += plenc.SizeVarUint(uint64(s))
			size += s
		}
	}

	size += plenc.SizeTag(plenc.WTVarInt, 11)
	size += plenc.SizeVarInt(int64(e.L))

	{
		var t plenc.Time
		t.Set(e.M)
		if s := t.ΦλSize(); s != 0 {
			size += plenc.SizeTag(plenc.WTLength, 12)
			size += plenc.SizeVarUint(uint64(s))
			size += s
		}

	}

	if e.N != nil {
		var t plenc.Time
		t.Set(*e.N)
		if s := t.ΦλSize(); s != 0 {
			size += plenc.SizeTag(plenc.WTLength, 13)
			size += plenc.SizeVarUint(uint64(s))
			size += s
		}

	}

	return size
}

// ΦλAppend encodes MyStruct by appending to data. It returns the final slice
func (e *MyStruct) ΦλAppend(data []byte) []byte {

	data = plenc.AppendTag(data, plenc.WTVarInt, 1)
	data = plenc.AppendVarInt(data, int64(e.A))

	data = plenc.AppendTag(data, plenc.WTVarInt, 2)
	data = plenc.AppendVarUint(data, uint64(e.B))

	data = plenc.AppendTag(data, plenc.WTVarInt, 3)
	data = plenc.AppendVarUint(data, uint64(e.C))

	data = plenc.AppendTag(data, plenc.WT32, 4)
	data = plenc.AppendFloat32(data, float32(e.D))

	data = plenc.AppendTag(data, plenc.WT64, 5)
	data = plenc.AppendFloat64(data, float64(e.E))

	data = plenc.AppendTag(data, plenc.WTVarInt, 6)
	data = plenc.AppendBool(data, e.F)

	if s := e.H.ΦλSize(); s != 0 {
		data = plenc.AppendTag(data, plenc.WTLength, 7)
		data = plenc.AppendVarUint(data, uint64(s))
		data = e.H.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, 8)
			data = plenc.AppendVarUint(data, uint64(s))
			data = e.I[i].ΦλAppend(data)
		}
	}

	if s := e.J.ΦλSize(); s != 0 {
		data = plenc.AppendTag(data, plenc.WTLength, 9)
		data = plenc.AppendVarUint(data, uint64(s))
		data = e.J.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.K {
		if s := e.K[i].ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, 10)
			data = plenc.AppendVarUint(data, uint64(s))
			data = e.K[i].ΦλAppend(data)
		}
	}

	data = plenc.AppendTag(data, plenc.WTVarInt, 11)
	data = plenc.AppendVarInt(data, int64(e.L))

	{
		var t plenc.Time
		t.Set(e.M)
		if s := t.ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, 12)
			data = plenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)
		}
	}

	if e.N != nil {
		var t plenc.Time
		t.Set(*e.N)
		if s := t.ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, 13)
			data = plenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)
		}
	}

	return data
}
