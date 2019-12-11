package plenc

// Marshaler is implemented by types that can encode themselves
type Marshaler interface {
	// ΦλSize returns the size of the encoded data
	ΦλSize() int
	// ΦλAppend encodes the object and appends it to data
	ΦλAppend(data []byte) []byte
}

// Unmarshaler is implemented by types that can decode themselves
type Unmarshaler interface {
	// ΦλUnmarshal unmarshals itself from data. It returns the number of bytes consumed
	ΦλUnmarshal(data []byte) (n int, err error)
}

// FullMarshaler allows a type to take full control over how they are unmarshalled
type FullMarshaler interface {
	// ΦλAppendFull encodes both this type itself, and it's tag (wire type and index) and any preamble.
	ΦλAppendFull(data []byte, index int) []byte

	// ΦλSizeFull determines the full size of this encoded object, including tag and any preamble
	ΦλSizeFull(index int) int
}
