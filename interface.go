package philenc

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
