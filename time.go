package φλenc

import (
	"fmt"
	"time"
)

// Time is a representation of time in UTC
type Time struct {
	Seconds     int64
	Nanoseconds int32
}

// Set sets the time from a time.Time
func (e *Time) Set(t time.Time) {
	e.Seconds = t.Unix()
	e.Nanoseconds = int32(t.Nanosecond())
}

func (e *Time) Standard() time.Time {
	return time.Unix(e.Seconds, int64(e.Nanoseconds))
}

func (e *Time) ΦλSize() (size int) {
	if e == nil {
		return 0
	}

	size += SizeTag(WTVarInt, 1)
	size += SizeVarInt(int64(e.Seconds))

	size += SizeTag(WTVarInt, 2)
	size += SizeVarInt(int64(e.Nanoseconds))

	return size
}

func (e *Time) ΦλAppend(data []byte) []byte {

	data = AppendTag(data, WTVarInt, 1)
	data = AppendVarInt(data, int64(e.Seconds))

	data = AppendTag(data, WTVarInt, 2)
	data = AppendVarInt(data, int64(e.Nanoseconds))

	return data
}

func (e *Time) ΦλUnmarshal(data []byte) (int, error) {

	var offset int
	for offset < len(data) {
		wt, index, n := ReadTag(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		switch index {

		case 1:

			v, n := ReadVarInt(data[offset:])
			e.Seconds = int64(v)

			offset += n

		case 2:

			v, n := ReadVarInt(data[offset:])
			e.Nanoseconds = int32(v)

			offset += n

		default:
			// Field corresponding to index does not exist
			n, err := Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d. %w", index, err)
			}
			offset += n
		}
	}

	return offset, nil
}
