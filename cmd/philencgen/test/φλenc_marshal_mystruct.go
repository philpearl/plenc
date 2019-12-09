package test

import (
	"github.com/philpearl/philenc"
)

// TODO: missing types
// slice of numeric ()
// slice of other
// pointers
// TODO: option whether top-level type is a pointer for marshaler

func (e *MyStruct) ΦλSize() (size int) {

	size += philenc.SizeTag(philenc.WTVarInt, 1)
	size += philenc.SizeVarUint(uint(e.A))

	size += philenc.SizeTag(philenc.WTVarInt, 2)
	size += philenc.SizeVarUint(uint(e.B))

	size += philenc.SizeTag(philenc.WTVarInt, 3)
	size += philenc.SizeVarUint(uint(e.C))

	size += philenc.SizeTag(philenc.WT32, 4)
	size += philenc.SizeFloat32(float32(e.D))

	size += philenc.SizeTag(philenc.WT64, 5)
	size += philenc.SizeFloat64(float64(e.E))

	size += philenc.SizeTag(philenc.WTVarInt, 6)
	size += philenc.SizeBool(e.F)

	if s := e.H.ΦλSize(); s != 0 {
		size += philenc.SizeTag(philenc.WTLength, 7)
		size += philenc.SizeVarUint(uint(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			size += philenc.SizeTag(philenc.WTLength, 8)
			size += philenc.SizeVarUint(uint(s))
			size += s
		}
	}

	return size
}

func (e *MyStruct) ΦλAppend(data []byte) []byte {

	data = philenc.AppendTag(data, philenc.WTVarInt, 1)
	data = philenc.AppendVarUint(data, uint(e.A))

	data = philenc.AppendTag(data, philenc.WTVarInt, 2)
	data = philenc.AppendVarUint(data, uint(e.B))

	data = philenc.AppendTag(data, philenc.WTVarInt, 3)
	data = philenc.AppendVarUint(data, uint(e.C))

	data = philenc.AppendTag(data, philenc.WT32, 4)
	data = philenc.AppendFloat32(data, float32(e.D))

	data = philenc.AppendTag(data, philenc.WT64, 5)
	data = philenc.AppendFloat64(data, float64(e.E))

	data = philenc.AppendTag(data, philenc.WTVarInt, 6)
	data = philenc.AppendBool(data, e.F)

	if s := e.H.ΦλSize(); s != 0 {
		data = philenc.AppendTag(data, philenc.WTLength, 7)
		data = philenc.AppendVarUint(data, uint(s))
		data = e.H.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			data = philenc.AppendTag(data, philenc.WTLength, 8)
			data = philenc.AppendVarUint(data, uint(s))
			data = e.I[i].ΦλAppend(data)
		}
	}

	return data
}
