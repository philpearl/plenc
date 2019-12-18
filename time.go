package plenc

import (
	"reflect"
	"sync"
	"time"
	"unsafe"
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
	return time.Unix(e.Seconds, int64(e.Nanoseconds))
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
		tc.Codec, err = codecForType(reflect.TypeOf(ptime{}))
		if err != nil {
			panic(err)
		}
	})
}

// Size returns the number of bytes needed to encode a Time
func (tc *TimeCodec) Size(ptr unsafe.Pointer) (size int) {
	tc.init()
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)

	return tc.Codec.Size(unsafe.Pointer(&e))
}

// Append encodes a Time
func (tc *TimeCodec) Append(data []byte, ptr unsafe.Pointer) []byte {
	tc.init()
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)

	return tc.Codec.Append(data, unsafe.Pointer(&e))
}

// Read decodes a Time
func (tc *TimeCodec) Read(data []byte, ptr unsafe.Pointer) (n int, err error) {
	tc.init()
	var e ptime
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
