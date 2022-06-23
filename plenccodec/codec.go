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
	Size(ptr unsafe.Pointer) (size int)
	Append(data []byte, ptr unsafe.Pointer) []byte
	Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error)
	New() unsafe.Pointer
	WireType() plenccore.WireType

	// Descriptor returns a descriptor for the type correspondng to the Codec.
	// The descriptor can be used to interpret plenc data without access to the
	// original type. The descriptor can also be serialised (either as JSON or
	// plenc), so can be stored or communicated with another system
	Descriptor() Descriptor
}
