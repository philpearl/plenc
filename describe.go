package plenc

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

var (
	codecCache     = make(map[reflect.Type]Codec)
	codecCacheLock sync.RWMutex
)

// RegisterCodec makes a codec available for a type
func RegisterCodec(typ reflect.Type, c Codec) {
	registerCodec(typ, c)
}

// registerCodec makes a codec available for a type
func registerCodec(typ reflect.Type, c Codec) {
	codecCacheLock.Lock()
	defer codecCacheLock.Unlock()
	codecCache[typ] = c
}

// codecForType finds an existing codec for a type or constructs a codec
func codecForType(typ reflect.Type) (Codec, error) {
	codecCacheLock.RLock()
	c, ok := codecCache[typ]
	if ok {
		codecCacheLock.RUnlock()
		return c, nil
	}
	codecCacheLock.RUnlock()

	codecCacheLock.Lock()
	defer codecCacheLock.Unlock()
	return codecForTypeLocked(typ)
}

func codecForTypeLocked(typ reflect.Type) (Codec, error) {

	c, ok := codecCache[typ]
	if ok {
		return c, nil
	}

	switch typ.Kind() {
	case reflect.Ptr:
		subc, err := codecForTypeLocked(typ.Elem())
		if err != nil {
			return nil, err
		}
		c = PointerWrapper{Underlying: subc}

	case reflect.Struct:
		var err error
		c, err = buildStructCodec(typ)
		if err != nil {
			return nil, err
		}

	case reflect.Slice:
		subt := typ.Elem()
		subc, err := codecForTypeLocked(subt)
		if err != nil {
			return nil, err
		}
		c = SliceWrapper{Underlying: subc, EltSize: subt.Size()}

	// Really expect codecs for basic types to be pre-registered
	case reflect.Bool:
	case reflect.Int:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Float32:
	case reflect.Float64:
	case reflect.String:

	case reflect.Int8:
	case reflect.Int16:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
	case reflect.Complex64:
	case reflect.Complex128:

		// These are cases we can't do yet (or ever?)
		//	case reflect.Array:
		// case reflect.Map:
		// case reflect.Interface:
		// case reflect.Chan:
		// case reflect.Func:
		// case reflect.UnsafePointer:
	}

	if c == nil {
		return nil, fmt.Errorf("could not create a codec for %s", typ.Name())
	}

	codecCache[typ] = c
	return c, nil
}

func buildStructCodec(typ reflect.Type) (Codec, error) {
	c := structCodec{
		rtype:  typ,
		fields: make([]description, typ.NumField()),
	}

	for i := range c.fields {
		field := &c.fields[i]

		sf := typ.Field(i)

		field.offset = sf.Offset
		field.rtype = sf.Type
		field.index = i + 1 // TODO: must do better! use a tag?

		fc, err := codecForTypeLocked(field.rtype)
		if err != nil {
			return nil, fmt.Errorf("failed to find codec for field %d (%s) of %s. %w", i, sf.Name, typ.Name(), err)
		}
		field.codec = fc
	}

	return &c, nil
}

type description struct {
	offset uintptr
	codec  Codec
	index  int
	rtype  reflect.Type
}

type structCodec struct {
	rtype  reflect.Type
	fields []description
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
		size += SizeAsField(field.codec, unsafe.Pointer(uintptr(ptr)+field.offset), field.index)
	}
	return size
}

func (c *structCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	s := c.SizeInternal(ptr)
	data = AppendVarUint(data, uint64(s))

	for _, field := range c.fields {
		data = AppendAsField(data, field.codec, unsafe.Pointer(uintptr(ptr)+field.offset), field.index)
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
		if n == 0 {
			break
		}
		offset += n

		found := false
		for _, field := range c.fields {
			if field.index == index {
				n, err := field.codec.Read(data[offset:], unsafe.Pointer(uintptr(ptr)+field.offset))
				if err != nil {
					return 0, fmt.Errorf("failed reading field %d of %s. %w", index, c.rtype.Name(), err)
				}
				offset += n
				found = true
				break
			}
		}

		if !found {
			// Field corresponding to index does not exist
			n, err := Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d in %s. %w", index, c.rtype.Name(), err)
			}
			offset += n
		}
	}

	return offset + n, nil
}
func (c *structCodec) New() unsafe.Pointer {
	return unsafe.Pointer(reflect.New(c.rtype).Pointer())
}
func (c *structCodec) WireType() WireType {
	return WTLength
}
