package philenc

// ReadBool reads a bool from data and returns it
func ReadBool(data []byte) (v bool, n int) {
	uv, n := ReadVarUint(data)
	return uv != 0, n
}

// SizeBool determines how many bytes it would take to encode the bool v
func SizeBool(v bool) int {
	return 1
}

// AppendBool appends a varint encoding of v to data. It returns the resulting slice
func AppendBool(data []byte, v bool) []byte {
	var uv uint
	if v {
		uv = 1
	}
	return AppendVarUint(data, uv)
}
