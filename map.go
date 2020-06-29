package plenc

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
)

// TODO: mapCodec doesn't work as it stands. Might be easier to do specific codecs for particular types
type mapCodec struct {
	keyCodec   Codec
	valueCodec Codec
	rtype      reflect.Type
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

	return mapCodec{keyCodec: kc, valueCodec: vc, rtype: typ}, nil
}

func (c mapCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}

func (c mapCodec) Size(ptr unsafe.Pointer) (size int) {
	val := reflect.NewAt(c.rtype, ptr).Elem()
	iter := val.MapRange()
	for iter.Next() {
		size += c.keyCodec.Size(unsafe.Pointer(iter.Key().Pointer()))
		size += c.valueCodec.Size(unsafe.Pointer(iter.Value().Pointer()))
	}
	return size
}

func (c mapCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	lOrig := len(data)

	// We avoid calculating the size of the data we need to add by guessing it will fit in 1 byte and
	// shuffling if not.
	data = append(data, 0)

	val := reflect.NewAt(c.rtype, ptr).Elem()
	iter := val.MapRange()
	for iter.Next() {

		data = c.keyCodec.Append(data, unsafe.Pointer(iter.Key().Addr().Pointer()))
		data = c.valueCodec.Append(data, unsafe.Pointer(iter.Value().Addr().Pointer()))
	}

	if s := len(data) - lOrig - 1; s > 0x7F {
		// Need to shuffle data as our size is longer
		data = moveForward(data, lOrig+1, SizeVarUint(uint64(s))-1)
		binary.PutUvarint(data[lOrig:], uint64(s))
	} else {
		data[lOrig] = byte(s)
	}

	return data
}

func (c mapCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	l, n := ReadVarUint(data)
	if n <= 0 {
		return 0, fmt.Errorf("varuint overflow reading %s", c.rtype.Name())
	}
	data = data[n:]
	if len(data) < int(l) {
		return 0, fmt.Errorf("not enough data to read %s. Have %d bytes, need %d", c.rtype.Name(), len(data), l)
	}

	val := reflect.NewAt(c.rtype, ptr)

	var offset int
	for offset < int(l) {
		k := c.keyCodec.New()
		v := c.valueCodec.New()

		n, err = c.keyCodec.Read(data[offset:], k, c.keyCodec.WireType())
		if err != nil {
			return 0, err
		}
		offset += n
		n, err = c.valueCodec.Read(data[offset:], v, c.valueCodec.WireType())
		if err != nil {
			return 0, err
		}
		offset += n

		val.SetMapIndex(reflect.NewAt(c.rtype.Key(), k), reflect.NewAt(c.rtype.Elem(), v))
	}

	return offset + n, nil
}
func (c mapCodec) New() unsafe.Pointer {
	return unsafe.Pointer(reflect.New(c.rtype).Pointer())
}
func (c mapCodec) WireType() WireType {
	return WTLength
}

func moveForward(data []byte, from, dist int) []byte {
	// Ensure we have enough space
	l := len(data)
	if cap(data)-l < dist {
		data = append(data, make([]byte, dist)...)
		data = data[:l]
	}

	data = data[:l+dist]
	for i := len(data) - 1; i >= from+dist; i-- {
		data[i] = data[i-dist]
	}
	return data
}
