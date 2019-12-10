package φλenc

// ReadString reads a string from data
func ReadString(data []byte) (string, int) {
	l, n := ReadVarUint(data)
	return string(data[n : n+int(l)]), n + int(l)
}

// SizeString works out how much room you need to store a string
func SizeString(v string) int {
	return SizeVarUint(uint64(len(v))) + len(v)
}

// AppendString appends a string encoding to data
func AppendString(data []byte, v string) []byte {
	data = AppendVarUint(data, uint64(len(v)))
	return append(data, v...)
}
