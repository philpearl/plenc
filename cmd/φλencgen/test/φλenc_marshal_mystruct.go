package test

import (
	"github.com/philpearl/φλenc"
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

	size += φλenc.SizeTag(φλenc.WTVarInt, 1)
	size += φλenc.SizeVarInt(int64(e.A))

	size += φλenc.SizeTag(φλenc.WTVarInt, 2)
	size += φλenc.SizeVarUint(uint64(e.B))

	size += φλenc.SizeTag(φλenc.WTVarInt, 3)
	size += φλenc.SizeVarUint(uint64(e.C))

	size += φλenc.SizeTag(φλenc.WT32, 4)
	size += φλenc.SizeFloat32(float32(e.D))

	size += φλenc.SizeTag(φλenc.WT64, 5)
	size += φλenc.SizeFloat64(float64(e.E))

	size += φλenc.SizeTag(φλenc.WTVarInt, 6)
	size += φλenc.SizeBool(e.F)

	if s := e.H.ΦλSize(); s != 0 {
		size += φλenc.SizeTag(φλenc.WTLength, 7)
		size += φλenc.SizeVarUint(uint64(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			size += φλenc.SizeTag(φλenc.WTLength, 8)
			size += φλenc.SizeVarUint(uint64(s))
			size += s
		}
	}

	if s := e.J.ΦλSize(); s != 0 {
		size += φλenc.SizeTag(φλenc.WTLength, 9)
		size += φλenc.SizeVarUint(uint64(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.K {
		if s := e.K[i].ΦλSize(); s != 0 {
			size += φλenc.SizeTag(φλenc.WTLength, 10)
			size += φλenc.SizeVarUint(uint64(s))
			size += s
		}
	}

	size += φλenc.SizeTag(φλenc.WTVarInt, 11)
	size += φλenc.SizeVarInt(int64(e.L))

	{
		var t φλenc.Time
		t.Set(e.M)
		if s := t.ΦλSize(); s != 0 {
			size += φλenc.SizeTag(φλenc.WTLength, 12)
			size += φλenc.SizeVarUint(uint64(s))
			size += s
		}

	}

	if e.N != nil {
		var t φλenc.Time
		t.Set(*e.N)
		if s := t.ΦλSize(); s != 0 {
			size += φλenc.SizeTag(φλenc.WTLength, 13)
			size += φλenc.SizeVarUint(uint64(s))
			size += s
		}

	}

	return size
}

// ΦλAppend encodes MyStruct by appending to data. It returns the final slice
func (e *MyStruct) ΦλAppend(data []byte) []byte {

	data = φλenc.AppendTag(data, φλenc.WTVarInt, 1)
	data = φλenc.AppendVarInt(data, int64(e.A))

	data = φλenc.AppendTag(data, φλenc.WTVarInt, 2)
	data = φλenc.AppendVarUint(data, uint64(e.B))

	data = φλenc.AppendTag(data, φλenc.WTVarInt, 3)
	data = φλenc.AppendVarUint(data, uint64(e.C))

	data = φλenc.AppendTag(data, φλenc.WT32, 4)
	data = φλenc.AppendFloat32(data, float32(e.D))

	data = φλenc.AppendTag(data, φλenc.WT64, 5)
	data = φλenc.AppendFloat64(data, float64(e.E))

	data = φλenc.AppendTag(data, φλenc.WTVarInt, 6)
	data = φλenc.AppendBool(data, e.F)

	if s := e.H.ΦλSize(); s != 0 {
		data = φλenc.AppendTag(data, φλenc.WTLength, 7)
		data = φλenc.AppendVarUint(data, uint64(s))
		data = e.H.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, 8)
			data = φλenc.AppendVarUint(data, uint64(s))
			data = e.I[i].ΦλAppend(data)
		}
	}

	if s := e.J.ΦλSize(); s != 0 {
		data = φλenc.AppendTag(data, φλenc.WTLength, 9)
		data = φλenc.AppendVarUint(data, uint64(s))
		data = e.J.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.K {
		if s := e.K[i].ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, 10)
			data = φλenc.AppendVarUint(data, uint64(s))
			data = e.K[i].ΦλAppend(data)
		}
	}

	data = φλenc.AppendTag(data, φλenc.WTVarInt, 11)
	data = φλenc.AppendVarInt(data, int64(e.L))

	{
		var t φλenc.Time
		t.Set(e.M)
		if s := t.ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, 12)
			data = φλenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)
		}
	}

	if e.N != nil {
		var t φλenc.Time
		t.Set(*e.N)
		if s := t.ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, 13)
			data = φλenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)
		}
	}

	return data
}
