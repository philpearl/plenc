package plenc

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

func (p *Plenc) buildStructCodec(typ reflect.Type) (Codec, error) {
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
		var postfix string
		if comma := strings.IndexByte(tag, ','); comma != -1 {
			postfix = tag[comma+1:]
			tag = tag[:comma]
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

		fc, err := p.codecForType(sf.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to find codec for field %d (%s) of %s. %w", i, sf.Name, typ.Name(), err)
		}

		if postfix == "intern" {
			if in, ok := fc.(Interner); ok {
				// Note we get an independent interner for each field
				fc = in.WithInterning()
			}
		}

		field.codec = fc
		field.tag = AppendTag(nil, fc.WireType(), field.index)
		if sf.Type.Kind() == reflect.Map {
			field.deref = true
		}
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
	deref  bool
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
	for _, field := range c.fields {
		// For most fields we have a pointer to the type, and this is the same
		// when we call these functions for types within structs or when we
		// pass an interface to Marshal. But maps are kind of pointers and
		// kind of not. When passed to Marshal via interfaces we get passed
		// the underlying map pointer. But when the map is in a struct, we
		// have a pointer to the underlying map pointer
		fptr := unsafe.Pointer(uintptr(ptr) + field.offset)
		if field.deref {
			fptr = *(*unsafe.Pointer)(fptr)
		}
		if !field.codec.Omit(fptr) {
			s := field.codec.Size(fptr)
			if field.codec.WireType() == WTLength {
				s += SizeVarUint(uint64(s))
			}
			size += len(field.tag) + s
		}
	}
	return size
}

func (c *structCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	for _, field := range c.fields {
		fptr := unsafe.Pointer(uintptr(ptr) + field.offset)
		if field.deref {
			fptr = *(*unsafe.Pointer)(fptr)
		}
		if field.codec.Omit(fptr) {
			continue
		}
		// TODO: In protobuf arrays of anything other than numbers are not
		// "packed", but are repeated tag and all. This isn't strictly necessary
		// for the protocol, but if we want to inter-operate... Well, we could
		// just be able to read that without necessarily being able to write it.
		data = append(data, field.tag...)
		if field.codec.WireType() == WTLength {
			data = AppendVarUint(data, uint64(field.codec.Size(fptr)))
		}
		data = field.codec.Append(data, fptr)
	}

	return data
}

func (c *structCodec) Read(data []byte, ptr unsafe.Pointer, wt WireType) (n int, err error) {
	l := len(data)

	var offset int
	for offset < l {
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

		fl := l
		if wt == WTLength {
			// For WTLength types we read out the length and ensure the data we
			// read the field from is the right length
			v, n := ReadVarUint(data[offset:])
			if n <= 0 {
				return 0, fmt.Errorf("varuint overflow reading field %d of %s", index, c.rtype.Name())
			}
			offset += n
			fl = int(v) + offset
			if fl > l {
				return 0, fmt.Errorf("length %d of field %d of %s exceeds data length", fl, index, c.rtype.Name())
			}
		}

		d := c.fieldsByIndex[index]
		n, err := d.codec.Read(data[offset:fl], unsafe.Pointer(uintptr(ptr)+d.offset), wt)
		if err != nil {
			return 0, fmt.Errorf("failed reading field %d of %s. %w", index, c.rtype.Name(), err)
		}
		offset += n
	}

	return offset, nil
}

func (c *structCodec) New() unsafe.Pointer {
	return unsafe.Pointer(reflect.New(c.rtype).Pointer())
}

func (c *structCodec) WireType() WireType {
	return WTLength
}
