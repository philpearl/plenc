package plenc

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// Codec is what you implement to encode / decode a type. Note that codecs are separate from the types they
// encode, and that they are registered with the system via RegisterCodec
type Codec interface {
	Omit(ptr unsafe.Pointer) bool
	Size(ptr unsafe.Pointer) (size int)
	Append(data []byte, ptr unsafe.Pointer) []byte
	Read(data []byte, ptr unsafe.Pointer) (n int, err error)
	New() unsafe.Pointer
	WireType() WireType
}

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

	case reflect.Map:
		c, err = buildMapCodec(typ)
		if err != nil {
			return nil, err
		}

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
