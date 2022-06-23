package plenc

import (
	"fmt"
	"reflect"

	"github.com/philpearl/plenc/plenccodec"
	"github.com/philpearl/plenc/plenccore"
)

var defaultPlenc Plenc

func init() {
	defaultPlenc.RegisterDefaultCodecs()
}

// RegisterCodec registers a codec with plenc so it can be used for marshaling
// and unmarshaling. If you write a custom codec then you need to register it
// before it can be used.
func RegisterCodec(typ reflect.Type, c plenccodec.Codec) {
	defaultPlenc.RegisterCodec(typ, c)
}

func (p *Plenc) codecForBasicType(typ reflect.Type) (plenccodec.Codec, error) {
	c, ok := p.codecRegistry.Load(typ)
	if ok {
		return c.(plenccodec.Codec), nil
	}
	return nil, fmt.Errorf("no codec available for %s", typ.Name())
}

// CodecForType returns a codec for the requested type. It should only be needed
// when constructing a codec based on an existing plenc codec
func CodecForType(typ reflect.Type) (plenccodec.Codec, error) {
	return defaultPlenc.CodecForType(typ)
}

// CodecForType finds an existing codec for a type or constructs a codec
func (p *Plenc) CodecForType(typ reflect.Type) (plenccodec.Codec, error) {
	c, ok := p.codecRegistry.Load(typ)
	if ok {
		return c.(plenccodec.Codec), nil
	}

	var err error

	switch typ.Kind() {
	case reflect.Ptr:
		subc, err := p.CodecForType(typ.Elem())
		if err != nil {
			return nil, err
		}
		c = plenccodec.PointerWrapper{Underlying: subc}

	case reflect.Struct:
		c, err = plenccodec.BuildStructCodec(p, typ)
		if err != nil {
			return nil, err
		}

	case reflect.Slice:
		subt := typ.Elem()
		subc, err := p.CodecForType(subt)
		if err != nil {
			return nil, err
		}
		bs := plenccodec.BaseSliceWrapper{Underlying: subc, EltSize: subt.Size(), EltType: unpackEFace(subt).data}
		switch subc.WireType() {
		case plenccore.WTVarInt:
			c = plenccodec.WTVarIntSliceWrapper{BaseSliceWrapper: bs}
		case plenccore.WT64, plenccore.WT32:
			if subt.Kind() == reflect.Ptr {
				// Can probably support these if we don't allow missing entries
				return nil, fmt.Errorf("slices of pointers to float32 & float64 are not supported")
			}
			c = plenccodec.WTFixedSliceWrapper{BaseSliceWrapper: bs}
		case plenccore.WTLength:
			c = plenccodec.WTLengthSliceWrapper{BaseSliceWrapper: bs}
		case plenccore.WTSlice:
			return nil, fmt.Errorf("slices of slices of structs or strings are not supported")
		default:
			return nil, fmt.Errorf("unexpected wire type %d for slice wrapper for type %q", subc.WireType(), typ.Name())
		}

	case reflect.Map:
		c, err = plenccodec.BuildMapCodec(p, typ)
		if err != nil {
			return nil, err
		}

	// Really expect codecs for basic types to be pre-registered, but named
	// types will have a different type for the same kind
	case reflect.Bool:
		c, err = p.codecForBasicType(reflect.TypeOf(bool(false)))
		if err != nil {
			return nil, err
		}

	case reflect.Int:
		c, err = p.codecForBasicType(reflect.TypeOf(int(0)))
		if err != nil {
			return nil, err
		}
	case reflect.Int32:
		c, err = p.codecForBasicType(reflect.TypeOf(int32(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Int64:
		c, err = p.codecForBasicType(reflect.TypeOf(int64(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint:
		c, err = p.codecForBasicType(reflect.TypeOf(uint(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Float32:
		c, err = p.codecForBasicType(reflect.TypeOf(float32(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Float64:
		c, err = p.codecForBasicType(reflect.TypeOf(float64(0)))
		if err != nil {
			return nil, err
		}

	case reflect.String:
		c, err = p.codecForBasicType(reflect.TypeOf(""))
		if err != nil {
			return nil, err
		}

	case reflect.Int8:
		c, err = p.codecForBasicType(reflect.TypeOf(int8(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Int16:
		c, err = p.codecForBasicType(reflect.TypeOf(int16(0)))
		if err != nil {
			return nil, err
		}
	case reflect.Uint8:
		c, err = p.codecForBasicType(reflect.TypeOf(uint8(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint16:
		c, err = p.codecForBasicType(reflect.TypeOf(uint16(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint32:
		c, err = p.codecForBasicType(reflect.TypeOf(uint32(0)))
		if err != nil {
			return nil, err
		}

	case reflect.Uint64:
		c, err = p.codecForBasicType(reflect.TypeOf(uint64(0)))
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
		return nil, fmt.Errorf("could not find or create a codec for %s", typ)
	}

	cv, _ := p.codecRegistry.LoadOrStore(typ, c)
	return cv.(plenccodec.Codec), nil
}
