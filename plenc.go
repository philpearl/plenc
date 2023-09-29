package plenc

import (
	"reflect"
	"time"

	"github.com/philpearl/plenc/plenccodec"
)

// Plenc is an instance of the plenc encoding and decoding system. It contains
// a registry of plenc codecs. Use it if you need fine-grain control on plenc's
// behaviour.
//
//	var p plenc.Plenc
//	p.RegisterDefaultCodecs()
//	data, err := p.Marshal(nil, mystruct)
type Plenc struct {
	// Plenc's original time handling was not compatible with protobuf's
	// google.protobuf.Timestamp. Set this to true to enable a proto compatible
	// time codec. Set it before calling RegisterDefaultCodecs.
	ProtoCompatibleTime bool
	// ProtoCompatibleArrays controls how plenc handles slices and arrays of
	// data. When set to true Plenc writes arrays that are compatible with
	// protobuf. If not true it uses a format that allows arrays to be read more
	// efficiently. Set it before calling RegisterDefaultCodecs.
	ProtoCompatibleArrays bool

	codecRegistry baseRegistry
}

func (p *Plenc) RegisterCodec(typ reflect.Type, c plenccodec.Codec) {
	p.codecRegistry.Store(typ, c)
}

// RegisterDefaultCodecs sets up the default codecs for plenc. It is called
// automatically for the default plenc instance, but if you create your own
// instance of Plenc you should call this before using it.
func (p *Plenc) RegisterDefaultCodecs() {
	p.RegisterCodec(reflect.TypeOf(false), plenccodec.BoolCodec{})
	p.RegisterCodec(reflect.TypeOf(float64(0)), plenccodec.Float64Codec{})
	p.RegisterCodec(reflect.TypeOf(float32(0)), plenccodec.Float32Codec{})
	p.RegisterCodec(reflect.TypeOf(int(0)), plenccodec.IntCodec[int]{})
	p.RegisterCodec(reflect.TypeOf(int8(0)), plenccodec.IntCodec[int8]{})
	p.RegisterCodec(reflect.TypeOf(int16(0)), plenccodec.IntCodec[int16]{})
	p.RegisterCodec(reflect.TypeOf(int32(0)), plenccodec.IntCodec[int32]{})
	p.RegisterCodec(reflect.TypeOf(int64(0)), plenccodec.IntCodec[int64]{})
	p.RegisterCodec(reflect.TypeOf(uint(0)), plenccodec.UintCodec[uint]{})
	p.RegisterCodec(reflect.TypeOf(uint64(0)), plenccodec.UintCodec[uint64]{})
	p.RegisterCodec(reflect.TypeOf(uint32(0)), plenccodec.UintCodec[uint32]{})
	p.RegisterCodec(reflect.TypeOf(uint16(0)), plenccodec.UintCodec[uint16]{})
	p.RegisterCodec(reflect.TypeOf(uint8(0)), plenccodec.UintCodec[uint8]{})
	p.RegisterCodec(reflect.TypeOf(""), plenccodec.StringCodec{})
	p.RegisterCodec(reflect.TypeOf([]byte(nil)), plenccodec.BytesCodec{})
	if p.ProtoCompatibleTime {
		p.RegisterCodec(reflect.TypeOf(time.Time{}), plenccodec.TimeCompatCodec{})
	} else {
		p.RegisterCodec(reflect.TypeOf(time.Time{}), plenccodec.TimeCodec{})
	}
}
