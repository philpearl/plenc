package plenc

import (
	"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

// Time is a representation of time in UTC. It is used to encode time.Time
type Time struct {
	Seconds     int64 `plenc:"1"`
	Nanoseconds int32 `plenc:"2"`
}

// Set sets the time from a time.Time
func (e *Time) Set(t time.Time) {
	e.Seconds = t.Unix()
	n := t.Nanosecond()
	e.Nanoseconds = int32(n)
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

func init() {
	registerCodec(reflect.TypeOf(time.Time{}), &TimeCodec{})
}

// TimeCodec is a codec for Time
type TimeCodec struct {
	Codec `plenc:"1"`
	once  sync.Once `plenc:"2"`
}

func (tc *TimeCodec) init() {
	tc.once.Do(func() {
		var err error
		tc.Codec, err = codecForType(reflect.TypeOf(Time{}))
		if err != nil {
			panic(err)
		}
	})
}

// Size returns the number of bytes needed to encode a Time
func (tc *TimeCodec) Size(ptr unsafe.Pointer) (size int) {
	tc.init()
	t := *(*time.Time)(ptr)
	var e Time
	e.Set(t)

	return tc.Codec.Size(unsafe.Pointer(&e))
}

// Append encodes a Time
func (tc *TimeCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	tc.init()
	t := *(*time.Time)(ptr)
	var e Time
	e.Set(t)

	return tc.Codec.Append(data, unsafe.Pointer(&e))
}

// Read decodes a Time
func (tc *TimeCodec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	tc.init()
	var e Time
	n, err = tc.Codec.Read(data, unsafe.Pointer(&e))
	if err != nil {
		return n, err
	}
	*(*time.Time)(ptr) = e.Standard()
	return n, nil
}

func (*TimeCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&time.Time{})
}

func (*TimeCodec) Omit(ptr unsafe.Pointer) bool {
	return (*time.Time)(ptr).IsZero()
}

func (tc *TimeCodec) WireType() WireType {
	tc.init()
	return tc.Codec.WireType()
}
