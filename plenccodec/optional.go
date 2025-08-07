package plenccodec

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// Optional is a type that can be used to represent an optional value, without
// resorting to pointers to indicate presence.
//
// Optional should be used within structs, not as a top-level type.
type Optional[T any] struct {
	Set   bool
	Value T
}

// OptionalOf creates an Optional[T] with the given value.
// It is a convenience function to avoid having to create an Optional[T] struct manually.
// This is useful for creating optional values in a more readable way.
//
// Example usage:
//
//	opt := OptionalOf(42) // Creates an Optional[int] with Set=true and Value=42
//	opt := OptionalOf("hello") // Creates an Optional[string] with Set=true and Value="hello"
//
// Note: This function is generic and works with any type T.
func OptionalOf[T any](value T) Optional[T] {
	return Optional[T]{
		Set:   true,
		Value: value,
	}
}

type optionalHeader struct {
	Set bool
}

func BuildOptionalCodec(p CodecBuilder, registry CodecRegistry, typ reflect.Type, tag string) (Codec, error) {
	valueField := typ.Field(1)
	offset := valueField.Offset
	underlying, err := p.CodecForTypeRegistry(registry, valueField.Type, tag)
	if err != nil {
		return nil, fmt.Errorf("building codec for underlying type %s: %w", typ.Name(), err)
	}

	return OptionalCodec{
		underlying: underlying,
		offset:     offset,
		typ:        typ,
	}, nil
}

// OptionalCodec is a codec for Optional[T]
type OptionalCodec struct {
	underlying Codec
	offset     uintptr
	typ        reflect.Type
}

func (p OptionalCodec) Omit(ptr unsafe.Pointer) bool {
	t := (*optionalHeader)(ptr)
	return !t.Set
}

func (p OptionalCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	t := (*optionalHeader)(ptr)
	// Need offset of the value, which depends in its alignment
	n, err = p.underlying.Read(data, unsafe.Add(ptr, p.offset), wt)
	if err != nil {
		return n, err
	}
	t.Set = true
	return n, nil
}

func (p OptionalCodec) New() unsafe.Pointer {
	return unsafe.Pointer(reflect.New(p.typ).Pointer())
}

func (p OptionalCodec) WireType() plenccore.WireType {
	return p.underlying.WireType()
}

func (p OptionalCodec) Descriptor() Descriptor {
	d := p.underlying.Descriptor()
	d.ExplicitPresence = true
	return d
}

func (p OptionalCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	// This should never be called if Omit returns true
	t := (*optionalHeader)(ptr)
	if !t.Set {
		return 0
	}
	return p.underlying.Size(unsafe.Add(ptr, p.offset), tag)
}

func (p OptionalCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	// This should never be called if Omit returns true
	t := (*optionalHeader)(ptr)
	if !t.Set {
		return data
	}
	return p.underlying.Append(data, unsafe.Add(ptr, p.offset), tag)
}
