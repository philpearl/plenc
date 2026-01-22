package plenc

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

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

// RegisterCodecWithTag is like RegisterCodec, but it registers the codec to be
// used only when the given tag is specified in the plenc struct tag
//
// For example:
//
//	type MyStruct struct {
//		A int `plenc:"1"`
//		B int `plenc:"2,flat"`
//	}
//
// The default codecs include a flat codec for ints. This does not use the ZigZag
// encoding as is more efficient when you know the values are positive.
func RegisterCodecWithTag(typ reflect.Type, tag string, c plenccodec.Codec) {
	defaultPlenc.RegisterCodecWithTag(typ, tag, c)
}

// CodecForType returns a codec for the requested type. It should only be needed
// when constructing a codec based on an existing plenc codec
func CodecForType(typ reflect.Type) (plenccodec.Codec, error) {
	return defaultPlenc.CodecForType(typ)
}

// CodecForTypeWithTag returns a codec for the requested type and tag. See
// RegisterCodecWithTag.
func CodecForTypeWithTag(typ reflect.Type, tag string) (plenccodec.Codec, error) {
	return defaultPlenc.CodecForTypeWithTag(typ, tag)
}

type baseRegistry struct {
	codecRegistry sync.Map
}

type registryKey struct {
	typ reflect.Type
	tag string
}

func (br *baseRegistry) Load(typ reflect.Type, tag string) plenccodec.Codec {
	c, ok := br.codecRegistry.Load(registryKey{typ: typ, tag: tag})
	if !ok {
		return nil
	}
	return c.(plenccodec.Codec)
}

func (br *baseRegistry) Store(typ reflect.Type, tag string, c plenccodec.Codec) {
	br.codecRegistry.Store(registryKey{typ: typ, tag: tag}, c)
}

func (br *baseRegistry) StoreOrSwap(typ reflect.Type, tag string, c plenccodec.Codec) plenccodec.Codec {
	cv, _ := br.codecRegistry.LoadOrStore(registryKey{typ: typ, tag: tag}, c)
	return cv.(plenccodec.Codec)
}

// CodecForType finds an existing codec for a type or constructs a codec. It
// calls CodecForTypeRegistry using the internal registry on p
func (p *Plenc) CodecForType(typ reflect.Type) (plenccodec.Codec, error) {
	return p.CodecForTypeRegistry(&p.codecRegistry, typ, "")
}

func (p *Plenc) CodecForTypeWithTag(typ reflect.Type, tag string) (plenccodec.Codec, error) {
	return p.CodecForTypeRegistry(&p.codecRegistry, typ, tag)
}

// localRegistry allows us to build up a set of codecs locally before
// making them visible in the main registry. This should mean we don't make any
// partial updates visible before they're complete.
type localRegistry struct {
	codecRegistry plenccodec.CodecRegistry
	local         map[registryKey]plenccodec.Codec
}

func (br *localRegistry) Load(typ reflect.Type, tag string) plenccodec.Codec {
	if c := br.codecRegistry.Load(typ, tag); c != nil {
		return c
	}
	return br.local[registryKey{typ: typ, tag: tag}]
}

func (br *localRegistry) Store(typ reflect.Type, tag string, c plenccodec.Codec) {
	br.local[registryKey{typ: typ, tag: tag}] = c
}

func (br *localRegistry) StoreOrSwap(typ reflect.Type, tag string, c plenccodec.Codec) plenccodec.Codec {
	if c2 := br.Load(typ, tag); c2 != nil {
		return c2
	}
	br.local[registryKey{typ: typ, tag: tag}] = c
	return c
}

func (br *localRegistry) Commit() {
	// There's a chance that codecs we've built will reference different
	// instances of codecs than the ones we're registering here. I don't think
	// that currently matters.
	for k, v := range br.local {
		br.codecRegistry.StoreOrSwap(k.typ, k.tag, v)
	}
}

// CodecForTypeRegistry builds a new codec for the requested type, consulting
// registry for any existing codecs needed
func (p *Plenc) CodecForTypeRegistry(registry plenccodec.CodecRegistry, typ reflect.Type, tag string) (plenccodec.Codec, error) {
	lr := &localRegistry{local: make(map[registryKey]plenccodec.Codec), codecRegistry: registry}

	icb := internalCodecBuilder{codecRegistry: lr, ProtoCompatibleArrays: p.ProtoCompatibleArrays}

	c, err := icb.CodecForTypeRegistry(lr, typ, tag)
	if err != nil {
		return nil, err
	}
	lr.Commit()
	return c, nil
}

// internalCodecBuilder is uses so we can wrap the codec registry with a
// local registry just once, then use this to build codecs as needed.
type internalCodecBuilder struct {
	codecRegistry         plenccodec.CodecRegistry
	ProtoCompatibleArrays bool
}

// codecForBasicType shortcuts the local registry. Basic types should be pre-registered
// with the underlying registry - we can't build them here.
func (p internalCodecBuilder) codecForBasicType(typ reflect.Type, tag string) (plenccodec.Codec, error) {
	c := p.codecRegistry.Load(typ, tag)
	if c == nil {
		return nil, fmt.Errorf("no codec available for %s", typ.Name())
	}
	return c, nil
}

func (p internalCodecBuilder) CodecForTypeRegistry(registry plenccodec.CodecRegistry, typ reflect.Type, tag string) (plenccodec.Codec, error) {
	c := registry.Load(typ, tag)
	if c != nil {
		return c, nil
	}

	var err error

	switch typ.Kind() {
	case reflect.Pointer:
		subc, err := p.CodecForTypeRegistry(registry, typ.Elem(), tag)
		if err != nil {
			return nil, err
		}
		c = plenccodec.PointerWrapper{Underlying: subc}

	case reflect.Struct:
		// Is this an Optional? The reflect package doesn't have a great way for
		// us to tell yet. So we just hack around with the name and package path.
		// Improve this when reflect fully supports generics.
		if strings.HasSuffix(typ.PkgPath(), "plenc/plenccodec") &&
			strings.HasPrefix(typ.Name(), "Optional[") {
			c, err = plenccodec.BuildOptionalCodec(p, registry, typ, tag)
			if err != nil {
				return nil, err
			}
		} else {
			c, err = plenccodec.BuildStructCodec(p, registry, typ, tag)
			if err != nil {
				return nil, err
			}
		}

	case reflect.Slice:
		subt := typ.Elem()
		// We assume for now that any tag here will be selecting the array
		// treatment, not the registry for the underlying type.
		subc, err := p.CodecForTypeRegistry(registry, subt, "")
		if err != nil {
			return nil, err
		}
		bs := plenccodec.BaseSliceWrapper{Underlying: subc, EltSize: subt.Size(), EltType: unpackEFace(subt).data}
		switch subc.WireType() {
		case plenccore.WTVarInt:
			c = plenccodec.WTVarIntSliceWrapper{BaseSliceWrapper: bs}
		case plenccore.WT64, plenccore.WT32:
			if subt.Kind() == reflect.Pointer {
				// Can probably support these if we don't allow missing entries
				return nil, fmt.Errorf("slices of pointers to float32 & float64 are not supported")
			}
			c = plenccodec.WTFixedSliceWrapper{BaseSliceWrapper: bs}
		case plenccore.WTLength:
			if p.ProtoCompatibleArrays || tag == "proto" {
				// When writing we just want to repeat the encoding of an
				// individual element within the slice as if it was a separate
				// element.
				// When reading we'll read elements repeatedly and append them to the array
				c = plenccodec.ProtoSliceWrapper{BaseSliceWrapper: bs}
			} else {
				c = plenccodec.WTLengthSliceWrapper{BaseSliceWrapper: bs}
			}
		case plenccore.WTSlice:
			return nil, fmt.Errorf("slices of slices of structs or strings are not supported")
		default:
			return nil, fmt.Errorf("unexpected wire type %d for slice wrapper for type %q", subc.WireType(), typ.Name())
		}

	case reflect.Map:
		c, err = plenccodec.BuildMapCodec(p, registry, typ, tag)
		if err != nil {
			return nil, err
		}

	// Really expect codecs for basic types to be pre-registered, but named
	// types will have a different type for the same kind
	case reflect.Bool:
		c, err = p.codecForBasicType(reflect.TypeFor[bool](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Int:
		c, err = p.codecForBasicType(reflect.TypeFor[int](), tag)
		if err != nil {
			return nil, err
		}
	case reflect.Int32:
		c, err = p.codecForBasicType(reflect.TypeFor[int32](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Int64:
		c, err = p.codecForBasicType(reflect.TypeFor[int64](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Uint:
		c, err = p.codecForBasicType(reflect.TypeFor[uint](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Float32:
		c, err = p.codecForBasicType(reflect.TypeFor[float32](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Float64:
		c, err = p.codecForBasicType(reflect.TypeFor[float64](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.String:
		c, err = p.codecForBasicType(reflect.TypeFor[string](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Int8:
		c, err = p.codecForBasicType(reflect.TypeFor[int8](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Int16:
		c, err = p.codecForBasicType(reflect.TypeFor[int16](), tag)
		if err != nil {
			return nil, err
		}
	case reflect.Uint8:
		c, err = p.codecForBasicType(reflect.TypeFor[uint8](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Uint16:
		c, err = p.codecForBasicType(reflect.TypeFor[uint16](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Uint32:
		c, err = p.codecForBasicType(reflect.TypeFor[uint32](), tag)
		if err != nil {
			return nil, err
		}

	case reflect.Uint64:
		c, err = p.codecForBasicType(reflect.TypeFor[uint64](), tag)
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

	return registry.StoreOrSwap(typ, tag, c), nil
}
