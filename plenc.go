package plenc

import (
	"reflect"
	"sync"
	"time"
)

type Plenc struct {
	codecRegistry sync.Map
}

func (p *Plenc) RegisterCodec(typ reflect.Type, c Codec) {
	p.codecRegistry.Store(typ, c)
}

func (p *Plenc) RegisterDefaultCodecs() {
	p.RegisterCodec(reflect.TypeOf(false), BoolCodec{})
	p.RegisterCodec(reflect.TypeOf(float64(0)), Float64Codec{})
	p.RegisterCodec(reflect.TypeOf(float32(0)), Float32Codec{})
	p.RegisterCodec(reflect.TypeOf(int(0)), IntCodec[int]{})
	p.RegisterCodec(reflect.TypeOf(int8(0)), IntCodec[int8]{})
	p.RegisterCodec(reflect.TypeOf(int16(0)), IntCodec[int16]{})
	p.RegisterCodec(reflect.TypeOf(int32(0)), IntCodec[int32]{})
	p.RegisterCodec(reflect.TypeOf(int64(0)), IntCodec[int64]{})
	p.RegisterCodec(reflect.TypeOf(uint(0)), UintCodec[uint]{})
	p.RegisterCodec(reflect.TypeOf(uint64(0)), UintCodec[uint64]{})
	p.RegisterCodec(reflect.TypeOf(uint32(0)), UintCodec[uint32]{})
	p.RegisterCodec(reflect.TypeOf(uint16(0)), UintCodec[uint16]{})
	p.RegisterCodec(reflect.TypeOf(uint8(0)), UintCodec[uint8]{})
	p.RegisterCodec(reflect.TypeOf(""), StringCodec{})
	p.RegisterCodec(reflect.TypeOf([]byte(nil)), BytesCodec{})
	p.RegisterCodec(reflect.TypeOf(time.Time{}), &TimeCodec{})
}
