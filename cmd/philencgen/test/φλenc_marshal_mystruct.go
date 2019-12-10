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
	if e == nil {
		return 0
	}

	size += philenc.SizeTag(philenc.WTVarInt, 1)
	size += philenc.SizeVarUint(uint64(e.A))

	size += philenc.SizeTag(philenc.WTVarInt, 2)
	size += philenc.SizeVarUint(uint64(e.B))

	size += philenc.SizeTag(philenc.WTVarInt, 3)
	size += philenc.SizeVarUint(uint64(e.C))

	size += philenc.SizeTag(philenc.WT32, 4)
	size += philenc.SizeFloat32(float32(e.D))

	size += philenc.SizeTag(philenc.WT64, 5)
	size += philenc.SizeFloat64(float64(e.E))

	size += philenc.SizeTag(philenc.WTVarInt, 6)
	size += philenc.SizeBool(e.F)

	if s := e.H.ΦλSize(); s != 0 {
		size += philenc.SizeTag(philenc.WTLength, 7)
		size += philenc.SizeVarUint(uint64(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			size += philenc.SizeTag(philenc.WTLength, 8)
			size += philenc.SizeVarUint(uint64(s))
			size += s
		}
	}

	if s := e.J.ΦλSize(); s != 0 {
		size += philenc.SizeTag(philenc.WTLength, 9)
		size += philenc.SizeVarUint(uint64(s))
		size += s
	}

	// Each element of the slice is separately encoded
	for i := range e.K {
		if s := e.K[i].ΦλSize(); s != 0 {
			size += philenc.SizeTag(philenc.WTLength, 10)
			size += philenc.SizeVarUint(uint64(s))
			size += s
		}
	}

	size += philenc.SizeTag(philenc.WTVarInt, 11)
	size += philenc.SizeVarUint(uint64(e.L))

	return size
}

func (e *MyStruct) ΦλAppend(data []byte) []byte {

	data = philenc.AppendTag(data, philenc.WTVarInt, 1)
	data = philenc.AppendVarUint(data, uint64(e.A))

	data = philenc.AppendTag(data, philenc.WTVarInt, 2)
	data = philenc.AppendVarUint(data, uint64(e.B))

	data = philenc.AppendTag(data, philenc.WTVarInt, 3)
	data = philenc.AppendVarUint(data, uint64(e.C))

	data = philenc.AppendTag(data, philenc.WT32, 4)
	data = philenc.AppendFloat32(data, float32(e.D))

	data = philenc.AppendTag(data, philenc.WT64, 5)
	data = philenc.AppendFloat64(data, float64(e.E))

	data = philenc.AppendTag(data, philenc.WTVarInt, 6)
	data = philenc.AppendBool(data, e.F)

	if s := e.H.ΦλSize(); s != 0 {
		data = philenc.AppendTag(data, philenc.WTLength, 7)
		data = philenc.AppendVarUint(data, uint64(s))
		data = e.H.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.I {
		if s := e.I[i].ΦλSize(); s != 0 {
			data = philenc.AppendTag(data, philenc.WTLength, 8)
			data = philenc.AppendVarUint(data, uint64(s))
			data = e.I[i].ΦλAppend(data)
		}
	}

	if s := e.J.ΦλSize(); s != 0 {
		data = philenc.AppendTag(data, philenc.WTLength, 9)
		data = philenc.AppendVarUint(data, uint64(s))
		data = e.J.ΦλAppend(data)
	}

	// Each element of the slice is separately encoded
	for i := range e.K {
		if s := e.K[i].ΦλSize(); s != 0 {
			data = philenc.AppendTag(data, philenc.WTLength, 10)
			data = philenc.AppendVarUint(data, uint64(s))
			data = e.K[i].ΦλAppend(data)
		}
	}

	data = philenc.AppendTag(data, philenc.WTVarInt, 11)
	data = philenc.AppendVarUint(data, uint64(e.L))

	return data
}
