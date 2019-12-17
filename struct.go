package plenc

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

func buildStructCodec(typ reflect.Type) (Codec, error) {
	c := structCodec{
		rtype:  typ,
		fields: make([]description, typ.NumField()),
	}

	var maxIndex int
	var count int
	for i := range c.fields {
		sf := typ.Field(i)

		r, _ := utf8.DecodeRuneInString(sf.Name)
		if unicode.IsLower(r) {
			continue
		}

		tag := sf.Tag.Get("plenc")
		if tag == "" {
			return nil, fmt.Errorf("no plenc tag on field %d %s of %s", i, sf.Name, typ.Name())
		}
		if tag == "-" {
			continue
		}
		index, err := strconv.Atoi(tag)
		if err != nil {
			return nil, fmt.Errorf("could not parse plenc tag on field %d %s of %s. %w", i, sf.Name, typ.Name(), err)
		}

		field := &c.fields[count]
		count++
		field.offset = sf.Offset
		field.index = index
		if field.index > maxIndex {
			maxIndex = field.index
		}

		fc, err := codecForType(sf.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to find codec for field %d (%s) of %s. %w", i, sf.Name, typ.Name(), err)
		}
		field.codec = fc
		field.tag = AppendTag(nil, fc.WireType(), field.index)
	}
	c.fields = c.fields[:count]

	c.fieldsByIndex = make([]shortDesc, maxIndex+1)
	for _, f := range c.fields {
		if c.fieldsByIndex[f.index].codec != nil {
			return nil, fmt.Errorf("failed building codec for %s. Multiple fields have index %d", typ.Name(), f.index)
		}
		c.fieldsByIndex[f.index] = shortDesc{
			codec:  f.codec,
			offset: f.offset,
		}
	}

	return &c, nil
}

type description struct {
	offset uintptr
	codec  Codec
	index  int
	tag    []byte
}

type shortDesc struct {
	codec  Codec
	offset uintptr
}

type structCodec struct {
	rtype         reflect.Type
	fields        []description
	fieldsByIndex []shortDesc
}

func (c *structCodec) Omit(ptr unsafe.Pointer) bool {
	return false
}

func (c *structCodec) Size(ptr unsafe.Pointer) (size int) {
	size = c.SizeInternal(ptr)
	return SizeVarUint(uint64(size)) + size
}

func (c *structCodec) SizeInternal(ptr unsafe.Pointer) (size int) {
	for _, field := range c.fields {
		fptr := unsafe.Pointer(uintptr(ptr) + field.offset)
		if !field.codec.Omit(fptr) {
			size += len(field.tag) + field.codec.Size(fptr)
		}
	}
	return size
}

func (c *structCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	lOrig := len(data)

	// We avoid calculating the size of the data we need to add by guessing it will fit in 1 byte and
	// shuffling if not.
	data = append(data, 0)

	for _, field := range c.fields {
		fptr := unsafe.Pointer(uintptr(ptr) + field.offset)
		if field.codec.Omit(fptr) {
			continue
		}
		data = append(data, field.tag...)
		data = field.codec.Append(data, fptr)
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

func (c *structCodec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	l, n := ReadVarUint(data)
	if n <= 0 {
		return 0, fmt.Errorf("varuint overflow reading %s", c.rtype.Name())
	}
	data = data[n:]
	if len(data) < int(l) {
		return 0, fmt.Errorf("not enough data to read %s. Have %d bytes, need %d", c.rtype.Name(), len(data), l)
	}

	var offset int
	for offset < int(l) {
		wt, index, n := ReadTag(data[offset:])
		offset += n

		if index >= len(c.fieldsByIndex) || c.fieldsByIndex[index].codec == nil {
			// Field corresponding to index does not exist
			n, err := Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d in %s. %w", index, c.rtype.Name(), err)
			}
			offset += n
			continue
		}

		d := c.fieldsByIndex[index]
		n, err := d.codec.Read(data[offset:], unsafe.Pointer(uintptr(ptr)+d.offset))
		if err != nil {
			return 0, fmt.Errorf("failed reading field %d of %s. %w", index, c.rtype.Name(), err)
		}
		offset += n
	}

	return offset + n, nil
}
func (c *structCodec) New() unsafe.Pointer {
	return unsafe.Pointer(reflect.New(c.rtype).Pointer())
}
func (c *structCodec) WireType() WireType {
	return WTLength
}
