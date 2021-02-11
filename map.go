package plenc

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// mapCodec is a codec for maps. We treat it as a slice of structs with the key
// and value as the fields in the structs.
type mapCodec struct {
	keyCodec      Codec
	valueCodec    Codec
	rtype         reflect.Type
	keyTag        []byte
	valueTag      []byte
	keyIsWTLength bool
	valIsWTLength bool
	kPool         sync.Pool
	kZero         unsafe.Pointer
	vZero         unsafe.Pointer
}

func buildMapCodec(typ reflect.Type) (Codec, error) {
	kc, err := codecForType(typ.Key())
	if err != nil {
		return nil, fmt.Errorf("failed to find codec for map key %s. %w", typ.Key().Name(), err)
	}
	vc, err := codecForType(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to find codec for map value %s. %w", typ.Elem().Name(), err)
	}

	c := &mapCodec{
		keyCodec:      kc,
		valueCodec:    vc,
		rtype:         typ,
		keyTag:        AppendTag(nil, kc.WireType(), 1),
		valueTag:      AppendTag(nil, vc.WireType(), 2),
		keyIsWTLength: kc.WireType() == WTLength,
		valIsWTLength: vc.WireType() == WTLength,
	}

	c.kPool.New = c.newKey
	if l := int(typ.Key().Size()); l <= len(zero) {
		c.kZero = unsafe.Pointer(&zero)
	} else {
		z := make([]byte, l)
		c.kZero = unsafe.Pointer(&z[0])
	}

	if l := int(typ.Elem().Size()); l <= len(zero) {
		c.vZero = unsafe.Pointer(&zero)
	} else {
		z := make([]byte, l)
		c.vZero = unsafe.Pointer(&z[0])
	}

	return c, nil
}

func (c *mapCodec) newKey() interface{} {
	return c.keyCodec.New()
}

// When we're writing ptr is a map pointer. When reading it is a pointer to a
// map pointer

func (c *mapCodec) Omit(ptr unsafe.Pointer) bool {
	return ptr == nil
}

func (c *mapCodec) Size(ptr unsafe.Pointer) (size int) {
	size = SizeVarUint(uint64(maplen(ptr)))

	var iterM mapiter
	iter := (unsafe.Pointer)(&iterM)
	mapiterinit(unpackEFace(c.rtype).data, ptr, iter)
	for {
		k := mapiterkey(iter)
		if k == nil {
			break
		}
		v := mapiterelem(iter)

		s := c.sizeForEntry(k, v)
		size += SizeVarUint(uint64(s)) + s

		mapiternext(iter)
	}
	return size
}

func (c *mapCodec) sizeForEntry(k, v unsafe.Pointer) int {
	s := c.sizeFor(c.keyCodec, k, c.keyTag, c.keyIsWTLength)
	s += c.sizeFor(c.valueCodec, v, c.valueTag, c.valIsWTLength)
	return s
}

func (*mapCodec) sizeFor(c Codec, ptr unsafe.Pointer, tag []byte, useLength bool) int {
	if c.Omit(ptr) {
		return 0
	}
	s := c.Size(ptr)
	if useLength {
		s += SizeVarUint(uint64(s))
	}
	return s + len(tag)
}

func (c *mapCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	add := func(c Codec, ptr unsafe.Pointer, tag []byte, useLength bool) {
		if !c.Omit(ptr) {
			data = append(data, tag...)
			if useLength {
				data = AppendVarUint(data, uint64(c.Size(ptr)))
			}
			data = c.Append(data, ptr)
		}
	}

	// First add the count of entries
	data = AppendVarUint(data, uint64(maplen(ptr)))

	var iterM mapiter
	iter := (unsafe.Pointer)(&iterM)
	mapiterinit(unpackEFace(c.rtype).data, ptr, iter)
	for {
		k := mapiterkey(iter)
		if k == nil {
			break
		}
		v := mapiterelem(iter)

		// Add the length of each entry, then the key and value
		data = AppendVarUint(data, uint64(c.sizeForEntry(k, v)))
		add(c.keyCodec, k, c.keyTag, c.keyIsWTLength)
		add(c.valueCodec, v, c.valueTag, c.valIsWTLength)

		mapiternext(iter)
	}

	return data
}

var zero [1024]byte

func (c *mapCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	// We start with a count of entries
	count, n := ReadVarUint(data)
	if n <= 0 {
		return 0, fmt.Errorf("failed to read map size")
	}

	// ptr is a pointer to a map pointer
	if *(*unsafe.Pointer)(ptr) == nil {
		*(*unsafe.Pointer)(ptr) = unsafe.Pointer(reflect.MakeMapWithSize(c.rtype, int(count)).Pointer())
	}
	mp := *(*unsafe.Pointer)(ptr)

	// We need some space to hold keys and values as we read them out. We can
	// re-use the space on each iteration as the data is copied into the map
	// We also save some memory & time if we cache them in some pools
	k := c.kPool.Get().(unsafe.Pointer)
	defer c.kPool.Put(k)
	offset := int(n)
	for count > 0 {
		// Each entry starts with a length
		entryLength, n := ReadVarUint(data[offset:])
		if n <= 0 {
			return 0, fmt.Errorf("failed to read map entry length")
		}
		offset += n
		n, err := c.readMapEntry(mp, k, data[offset:offset+int(entryLength)])
		if err != nil {
			return 0, err
		}
		offset += n
		count--
	}

	return offset, nil
}

// readMapEntry reads out a single map entry. mp is the map pointer. k is an
// area to read key values into. data is the raw data for this map entry
func (c *mapCodec) readMapEntry(mp, k unsafe.Pointer, data []byte) (int, error) {
	offset, fieldEnd, index, wt, err := c.readTagAndLength(data, 0)
	if err != nil {
		return 0, err
	}

	if index == 1 {
		// Key is present - read it
		n, err := c.keyCodec.Read(data[offset:fieldEnd], k, wt)
		if err != nil {
			return 0, fmt.Errorf("failed reading key field of %s. %w", c.rtype.Name(), err)
		}
		offset += n
	} else {
		k = c.kZero
	}

	// Assign/find a place in the map for this key. Val is a pointer to where
	// the value should be. We're going to unmarshal into this directly
	val := mapassign(unpackEFace(c.rtype).data, mp, k)

	if offset < len(data) {
		if index == 1 {
			offset, fieldEnd, index, wt, err = c.readTagAndLength(data, offset)
			if err != nil {
				return 0, err
			}
		}

		n, err := c.valueCodec.Read(data[offset:fieldEnd], val, wt)
		if err != nil {
			return 0, fmt.Errorf("failed reading value field of %s. %w", c.rtype.Name(), err)
		}
		offset += n
	} else {
		// No value - use the nil value.
		typedmemmove(unpackEFace(c.rtype.Elem()).data, val, c.vZero)
	}

	return offset, nil
}

func (c *mapCodec) readTagAndLength(data []byte, offset int) (offset2, fieldEnd, index int, wt WireType, err error) {
	wt, index, n := ReadTag(data[offset:])
	offset += n
	fieldEnd = len(data)
	if wt == WTLength {
		// For WTLength types we read out the length and ensure the data we
		// read the field from is the right length
		fieldLen, n := ReadVarUint(data[offset:])
		if n <= 0 {
			return 0, 0, 0, wt, fmt.Errorf("varuint overflow reading %d of %s", index, c.rtype.Name())
		}
		offset += n
		fieldEnd = int(fieldLen) + offset
		if fieldEnd > len(data) {
			return 0, 0, 0, wt, fmt.Errorf("length %d of field %d of %s exceeds data length %d", fieldLen, index, c.rtype.Name(), len(data)-offset)
		}
	}

	return offset, fieldEnd, index, wt, nil
}

func (c *mapCodec) New() unsafe.Pointer {
	return unsafe.Pointer(reflect.MakeMap(c.rtype).Pointer())
}

func (c *mapCodec) WireType() WireType {
	return WTSlice
}
