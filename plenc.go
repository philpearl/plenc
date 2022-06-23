package plenc

import (
	"reflect"
	"sync"
	"time"

	"github.com/philpearl/plenc/plenccodec"
)

type Plenc struct {
	codecRegistry sync.Map
}

func (p *Plenc) RegisterCodec(typ reflect.Type, c plenccodec.Codec) {
	p.codecRegistry.Store(typ, c)
}

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
	p.RegisterCodec(reflect.TypeOf(time.Time{}), plenccodec.TimeCodec{})
}
