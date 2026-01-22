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
	p.codecRegistry.Store(typ, "", c)
}

func (p *Plenc) RegisterCodecWithTag(typ reflect.Type, tag string, c plenccodec.Codec) {
	p.codecRegistry.Store(typ, tag, c)
}

// RegisterDefaultCodecs sets up the default codecs for plenc. It is called
// automatically for the default plenc instance, but if you create your own
// instance of Plenc you should call this before using it.
func (p *Plenc) RegisterDefaultCodecs() {
	p.RegisterCodec(reflect.TypeFor[bool](), plenccodec.BoolCodec{})

	p.RegisterCodec(reflect.TypeFor[float64](), plenccodec.Float64Codec{})
	p.RegisterCodec(reflect.TypeFor[float32](), plenccodec.Float32Codec{})

	p.RegisterCodec(reflect.TypeFor[int](), plenccodec.IntCodec[int]{})
	p.RegisterCodec(reflect.TypeFor[int8](), plenccodec.IntCodec[int8]{})
	p.RegisterCodec(reflect.TypeFor[int16](), plenccodec.IntCodec[int16]{})
	p.RegisterCodec(reflect.TypeFor[int32](), plenccodec.IntCodec[int32]{})
	p.RegisterCodec(reflect.TypeFor[int64](), plenccodec.IntCodec[int64]{})

	p.RegisterCodecWithTag(reflect.TypeFor[int](), "flat", plenccodec.FlatIntCodec[uint]{})
	p.RegisterCodecWithTag(reflect.TypeFor[int8](), "flat", plenccodec.FlatIntCodec[uint8]{})
	p.RegisterCodecWithTag(reflect.TypeFor[int16](), "flat", plenccodec.FlatIntCodec[uint16]{})
	p.RegisterCodecWithTag(reflect.TypeFor[int32](), "flat", plenccodec.FlatIntCodec[uint32]{})
	p.RegisterCodecWithTag(reflect.TypeFor[int64](), "flat", plenccodec.FlatIntCodec[uint64]{})

	p.RegisterCodec(reflect.TypeFor[uint](), plenccodec.UintCodec[uint]{})
	p.RegisterCodec(reflect.TypeFor[uint64](), plenccodec.UintCodec[uint64]{})
	p.RegisterCodec(reflect.TypeFor[uint32](), plenccodec.UintCodec[uint32]{})
	p.RegisterCodec(reflect.TypeFor[uint16](), plenccodec.UintCodec[uint16]{})
	p.RegisterCodec(reflect.TypeFor[uint8](), plenccodec.UintCodec[uint8]{})
	p.RegisterCodec(reflect.TypeFor[string](), plenccodec.StringCodec{})
	p.RegisterCodecWithTag(reflect.TypeFor[string](), "intern", &plenccodec.InternedStringCodec{})
	p.RegisterCodec(reflect.TypeFor[[]byte](), plenccodec.BytesCodec{})
	if p.ProtoCompatibleTime {
		p.RegisterCodec(reflect.TypeFor[time.Time](), plenccodec.TimeCompatCodec{})
	} else {
		p.RegisterCodec(reflect.TypeFor[time.Time](), plenccodec.TimeCodec{})
	}
}
