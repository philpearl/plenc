package plenccore

import "testing"

// FuzzVarUintRoundTrip tests that encoding and decoding uint64 values works correctly.
func FuzzVarUintRoundTrip(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(1))
	f.Add(uint64(127))
	f.Add(uint64(128))
	f.Add(uint64(1<<32 - 1))
	f.Add(uint64(1<<63 - 1))
	f.Add(uint64(1 << 63))

	f.Fuzz(func(t *testing.T, v uint64) {
		data := AppendVarUint(nil, v)

		size := SizeVarUint(v)
		if len(data) != size {
			t.Errorf("SizeVarUint(%d) = %d, but encoded length = %d", v, size, len(data))
		}

		decoded, n := ReadVarUint(data)
		if n != len(data) {
			t.Errorf("ReadVarUint consumed %d bytes, expected %d", n, len(data))
		}
		if decoded != v {
			t.Errorf("ReadVarUint = %d, want %d", decoded, v)
		}
	})
}

// FuzzVarIntRoundTrip tests that encoding and decoding int64 values works correctly.
func FuzzVarIntRoundTrip(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(int64(1<<31 - 1))
	f.Add(int64(-1 << 31))
	f.Add(int64(1<<62 - 1))
	f.Add(int64(-1 << 62))

	f.Fuzz(func(t *testing.T, v int64) {
		data := AppendVarInt(nil, v)

		size := SizeVarInt(v)
		if len(data) != size {
			t.Errorf("SizeVarInt(%d) = %d, but encoded length = %d", v, size, len(data))
		}

		decoded, n := ReadVarInt(data)
		if n != len(data) {
			t.Errorf("ReadVarInt consumed %d bytes, expected %d", n, len(data))
		}
		if decoded != v {
			t.Errorf("ReadVarInt = %d, want %d", decoded, v)
		}
	})
}

// FuzzSkip tests that Skip doesn't panic on arbitrary data.
func FuzzSkip(f *testing.F) {
	f.Add(byte(0), []byte{0x00})                          // WTVarInt
	f.Add(byte(1), []byte{0, 0, 0, 0, 0, 0, 0, 0})        // WT64
	f.Add(byte(2), []byte{0x05, 'h', 'e', 'l', 'l', 'o'}) // WTLength
	f.Add(byte(5), []byte{0, 0, 0, 0})                    // WT32
	f.Add(byte(0), []byte{})                              // empty

	f.Fuzz(func(t *testing.T, wt byte, data []byte) {
		// Clamp wire type to valid range
		wireType := WireType(wt % 6)
		// Should not panic
		_, _ = Skip(data, wireType)
	})
}

// FuzzTag tests that tag encoding/decoding works correctly.
func FuzzTagRoundTrip(f *testing.F) {
	f.Add(byte(0), 1)
	f.Add(byte(1), 100)
	f.Add(byte(2), 1000)
	f.Add(byte(5), 10000)

	f.Fuzz(func(t *testing.T, wt byte, index int) {
		if index < 0 {
			return // negative indices not valid
		}
		wireType := WireType(wt % 6)
		if wireType == 3 || wireType == 4 {
			return // deprecated wire types
		}

		data := AppendTag(nil, wireType, index)

		size := SizeTag(wireType, index)
		if len(data) != size {
			t.Errorf("SizeTag(%d, %d) = %d, but encoded length = %d", wireType, index, size, len(data))
		}

		gotWT, gotIndex, n := ReadTag(data)
		if n != len(data) {
			t.Errorf("ReadTag consumed %d bytes, expected %d", n, len(data))
		}
		if gotWT != wireType {
			t.Errorf("ReadTag wire type = %d, want %d", gotWT, wireType)
		}
		if gotIndex != index {
			t.Errorf("ReadTag index = %d, want %d", gotIndex, index)
		}
	})
}
