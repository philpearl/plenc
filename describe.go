package plenc

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

var (
	codecCache     sync.Map
	codecCacheLock sync.RWMutex
)

// RegisterCodec makes a codec available for a type
func RegisterCodec(typ reflect.Type, c Codec) {
	registerCodec(typ, c)
}

// registerCodec makes a codec available for a type
func registerCodec(typ reflect.Type, c Codec) {
	codecCache.Store(typ, c)
}

// codecForType finds an existing codec for a type or constructs a codec
func codecForType(typ reflect.Type) (Codec, error) {
	c, ok := codecCache.Load(typ)
	if ok {
		return c.(Codec), nil
	}

	var err error

	switch typ.Kind() {
	case reflect.Ptr:
		subc, err := codecForType(typ.Elem())
		if err != nil {
			return nil, err
		}
		c = PointerWrapper{Underlying: subc}

	case reflect.Struct:
		c, err = buildStructCodec(typ)
		if err != nil {
			return nil, err
		}

	case reflect.Slice:
		subt := typ.Elem()
		subc, err := codecForType(subt)
		if err != nil {
			return nil, err
		}
		c = SliceWrapper{Underlying: subc, EltSize: subt.Size()}

	// Really expect codecs for basic types to be pre-registered, but named types will have a different type for
	// the same kind
	case reflect.Bool:
		c, err = codecForType(reflect.TypeOf(bool(false)))
		if err != nil {
			return nil, err
		}

	case reflect.Int:
		c, err = codecForType(reflect.TypeOf(int(0)))
		if err != nil {
			return nil, err
		}
	case reflect.Int32:
		c, err = codecForType(reflect.TypeOf(int32(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Int64:
		c, err = codecForType(reflect.TypeOf(int64(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint:
		c, err = codecForType(reflect.TypeOf(uint(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Float32:
		c, err = codecForType(reflect.TypeOf(float32(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Float64:
		c, err = codecForType(reflect.TypeOf(float64(0)))
		if err != nil {
			return nil, err
		}

	case reflect.String:
		c, err = codecForType(reflect.TypeOf(""))
		if err != nil {
			return nil, err
		}

	case reflect.Int8:
		c, err = codecForType(reflect.TypeOf(int8(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Int16:
		c, err = codecForType(reflect.TypeOf(int16(0)))
		if err != nil {
			return nil, err
		}
	case reflect.Uint8:
		c, err = codecForType(reflect.TypeOf(uint8(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint16:
		c, err = codecForType(reflect.TypeOf(uint16(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint32:
		c, err = codecForType(reflect.TypeOf(uint32(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint64:
		c, err = codecForType(reflect.TypeOf(uint64(0)))
		if err != nil {
			return nil, err
		}

		// These are cases we can't do yet (or ever?)
		// case reflect.Uintptr:
		// case reflect.Complex64:
		// case reflect.Complex128:
		// case reflect.Array:
		// case reflect.Map:
		// case reflect.Interface:
		// case reflect.Chan:
		// case reflect.Func:
		// case reflect.UnsafePointer:
	}

	if c == nil {
		return nil, fmt.Errorf("could not find or create a codec for %s", typ.Name())
	}

	cv, _ := codecCache.LoadOrStore(typ, c)
	return cv.(Codec), nil
}

func buildStructCodec(typ reflect.Type) (Codec, error) {
	c := structCodec{
		rtype:  typ,
		fields: make([]description, typ.NumField()),
	}

	var maxIndex int
	for i := range c.fields {
		field := &c.fields[i]

		sf := typ.Field(i)

		field.offset = sf.Offset
		field.index = i + 1 // TODO: must do better! use a tag?
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

	c.fieldsByIndex = make([]shortDesc, maxIndex+1)
	for _, f := range c.fields {

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
