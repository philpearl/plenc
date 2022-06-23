package plenc

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/philpearl/plenc/plenccore"
)

// ptime is a representation of time in UTC. It is used to encode time.Time
type ptime struct {
	Seconds     int64 `plenc:"1"`
	Nanoseconds int32 `plenc:"2"`
}

// Set sets the time from a time.Time
func (e *ptime) Set(t time.Time) {
	e.Seconds = t.Unix()
	n := t.Nanosecond()
	e.Nanoseconds = int32(n)
}

func (e *ptime) Standard() time.Time {
	return time.Unix(e.Seconds, int64(e.Nanoseconds)).UTC()
}

// TimeCodec is a codec for Time
type TimeCodec struct{}

// Size returns the number of bytes needed to encode a Time
func (tc *TimeCodec) Size(ptr unsafe.Pointer) (size int) {
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)
	sl := IntCodec[int64]{}.Size(unsafe.Pointer(&e.Seconds))
	nl := IntCodec[int32]{}.Size(unsafe.Pointer(&e.Nanoseconds))
	return plenccore.SizeTag(plenccore.WTVarInt, 1) + sl + plenccore.SizeTag(plenccore.WTVarInt, 2) + nl
}

// Append encodes a Time
func (tc *TimeCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)

	data = plenccore.AppendTag(data, plenccore.WTVarInt, 1)
	data = IntCodec[int64]{}.Append(data, unsafe.Pointer(&e.Seconds))
	data = plenccore.AppendTag(data, plenccore.WTVarInt, 2)
	data = IntCodec[int32]{}.Append(data, unsafe.Pointer(&e.Nanoseconds))
	return data
}

// Read decodes a Time
func (tc *TimeCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	var e ptime
	l := len(data)

	var offset int
	for offset < l {
		wt, index, n := plenccore.ReadTag(data[offset:])
		offset += n

		switch index {
		case 1:
			n, err := IntCodec[int64]{}.Read(data[offset:], unsafe.Pointer(&e.Seconds), wt)
			if err != nil {
				return 0, fmt.Errorf("failed reading seconds field of time. %w", err)
			}
			offset += n

		case 2:
			n, err := IntCodec[int32]{}.Read(data[offset:], unsafe.Pointer(&e.Nanoseconds), wt)
			if err != nil {
				return 0, fmt.Errorf("failed reading nanoseconds field of time. %w", err)
			}
			offset += n

		default:
			// Field corresponding to index does not exist
			n, err := plenccore.Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d of time. %w", index, err)
			}
			offset += n
		}
	}

	*(*time.Time)(ptr) = e.Standard()

	return offset, nil
}

func (*TimeCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&time.Time{})
}

func (*TimeCodec) Omit(ptr unsafe.Pointer) bool {
	return (*time.Time)(ptr).IsZero()
}

func (tc *TimeCodec) WireType() plenccore.WireType {
	return plenccore.WTLength
}
