// Package plenccodec provides the core Codecs for plenc.
//
// You shouldn't need to interact with this package directly unless you're
// building your own custom codecs.
//
// The exception to this is the Descriptor endpoint on a Codec. This is a
// serialisable description that allows you to decode plenc data without the
// type that the data was encoded from.
package plenccodec

import (
	"reflect"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// Codec is what you implement to encode / decode a type. Note that codecs are
// separate from the types they encode, and that they are registered with the
// system via RegisterCodec.
//
// It isn't normally necessary to build a codec for a struct. Codecs for structs
// are generated automatically when plenc first sees them and then are re-used
// for the life of the program.
type Codec interface {
	Omit(ptr unsafe.Pointer) bool
	Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error)
	New() unsafe.Pointer
	WireType() plenccore.WireType

	// Descriptor returns a descriptor for the type correspondng to the Codec.
	// The descriptor can be used to interpret plenc data without access to the
	// original type. The descriptor can also be serialised (either as JSON or
	// plenc), so can be stored or communicated with another system
	Descriptor() Descriptor

	// Size returns the size of the encoded data including the tag and
	// for WTLength types the varint encoded length of the data. If the tag is
	// nil then it is not included in the size and neither is the length for
	// WTLength types
	Size(ptr unsafe.Pointer, tag []byte) int

	// Append appends the encoded data including the tag and for WTLength
	// types the varint encoded length of the data. If the tag is nil then it is
	// not included in the data and neither is the length for WTLength types
	Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte
}

// CodecRegistry is a repository of pre-existing Codecs
type CodecRegistry interface {
	// Load loads the codec from the registry. It returns nil if no codec exists
	Load(typ reflect.Type, tag string) Codec
	// StoreOrSwap adds the codec to the registry. It may return a different
	// codec if the codec has been built on another goroutine
	StoreOrSwap(typ reflect.Type, tag string, c Codec) Codec
}

// CodecBuilder either builds a new codec for a type, or finds an existing codec
type CodecBuilder interface {
	// CodecForTypeRegistry builds a new codec for the requested type,
	// consulting registry for any existing codecs needed
	CodecForTypeRegistry(registry CodecRegistry, typ reflect.Type, tag string) (Codec, error)
}
