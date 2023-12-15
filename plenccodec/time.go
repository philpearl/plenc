package plenccodec

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
// Note that this is not compatible with google.protobuf.Timestamp. The structure
// is the same, but they don't use zigzag encoding for the fields. They still do
// allow negative values though
type TimeCodec struct{}

// size returns the number of bytes needed to encode a Time
func (tc TimeCodec) size(ptr unsafe.Pointer) (size int) {
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)
	return IntCodec[int64]{}.Size(unsafe.Pointer(&e.Seconds), varInt1Tag) +
		IntCodec[int32]{}.Size(unsafe.Pointer(&e.Nanoseconds), varInt2Tag)
}

var (
	varInt1Tag = plenccore.AppendTag(nil, plenccore.WTVarInt, 1)
	varInt2Tag = plenccore.AppendTag(nil, plenccore.WTVarInt, 2)
)

// append encodes a Time
func (tc TimeCodec) append(data []byte, ptr unsafe.Pointer) []byte {
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)

	data = IntCodec[int64]{}.Append(data, unsafe.Pointer(&e.Seconds), varInt1Tag)
	data = IntCodec[int32]{}.Append(data, unsafe.Pointer(&e.Nanoseconds), varInt2Tag)
	return data
}

// Read decodes a Time
func (tc TimeCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	l := len(data)
	if l == 0 {
		*(*time.Time)(ptr) = time.Time{}
		return 0, nil
	}

	var e ptime
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

func (TimeCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&time.Time{})
}

func (TimeCodec) Omit(ptr unsafe.Pointer) bool {
	return (*time.Time)(ptr).IsZero()
}

func (tc TimeCodec) WireType() plenccore.WireType {
	return plenccore.WTLength
}

func (tc TimeCodec) Descriptor() Descriptor {
	return Descriptor{Type: FieldTypeTime}
}

func (c TimeCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	l := c.size(ptr)
	if len(tag) != 0 {
		l += len(tag) + plenccore.SizeVarUint(uint64(l))
	}
	return l
}

func (c TimeCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	if len(tag) != 0 {
		data = append(data, tag...)
		data = plenccore.AppendVarUint(data, uint64(c.size(ptr)))
	}

	return c.append(data, ptr)
}

type TimeCompatCodec struct {
	TimeCodec
}

// size returns the number of bytes needed to encode a Time
func (tc TimeCompatCodec) size(ptr unsafe.Pointer) (size int) {
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)
	return UintCodec[uint64]{}.Size(unsafe.Pointer(&e.Seconds), varInt1Tag) +
		+UintCodec[uint32]{}.Size(unsafe.Pointer(&e.Nanoseconds), varInt2Tag)
}

// append encodes a Time
func (tc TimeCompatCodec) append(data []byte, ptr unsafe.Pointer) []byte {
	t := *(*time.Time)(ptr)
	var e ptime
	e.Set(t)

	data = UintCodec[uint64]{}.Append(data, unsafe.Pointer(&e.Seconds), varInt1Tag)
	data = UintCodec[uint32]{}.Append(data, unsafe.Pointer(&e.Nanoseconds), varInt2Tag)
	return data
}

func (c TimeCompatCodec) Size(ptr unsafe.Pointer, tag []byte) int {
	l := c.size(ptr)
	if len(tag) != 0 {
		l += len(tag) + plenccore.SizeVarUint(uint64(l))
	}
	return l
}

func (c TimeCompatCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	if len(tag) != 0 {
		data = append(data, tag...)
		data = plenccore.AppendVarUint(data, uint64(c.size(ptr)))
	}

	return c.append(data, ptr)
}

// Read decodes a Time
func (tc TimeCompatCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	l := len(data)
	if l == 0 {
		*(*time.Time)(ptr) = time.Time{}
		return 0, nil
	}

	var e ptime
	var offset int
	for offset < l {
		wt, index, n := plenccore.ReadTag(data[offset:])
		offset += n

		switch index {
		case 1:
			n, err := UintCodec[uint64]{}.Read(data[offset:], unsafe.Pointer(&e.Seconds), wt)
			if err != nil {
				return 0, fmt.Errorf("failed reading seconds field of time. %w", err)
			}
			offset += n

		case 2:
			n, err := UintCodec[uint32]{}.Read(data[offset:], unsafe.Pointer(&e.Nanoseconds), wt)
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

// BQTimestampStruct encodes a time.Time as a flat (not zigzag) int64 of
// microseconds since the Epoch. This is how the BigQuery write API expects a
// timestamp to be encoded
type BQTimestampCodec struct {
	FlatIntCodec[uint64]
}

func (BQTimestampCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&time.Time{})
}

func (BQTimestampCodec) Omit(ptr unsafe.Pointer) bool {
	return (*time.Time)(ptr).IsZero()
}

func (c BQTimestampCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	var ts int64
	n, err = c.FlatIntCodec.Read(data, unsafe.Pointer(&ts), wt)
	if err != nil {
		return n, err
	}
	*(*time.Time)(ptr) = time.UnixMicro(ts)
	return n, nil
}

func (c BQTimestampCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	ts := (*time.Time)(ptr).UnixMicro()
	return c.FlatIntCodec.Append(data, unsafe.Pointer(&ts), tag)
}
