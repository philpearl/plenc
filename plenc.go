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
	p.RegisterCodec(reflect.TypeOf(int(0)), IntCodec{})
	p.RegisterCodec(reflect.TypeOf(int8(0)), Int8Codec{})
	p.RegisterCodec(reflect.TypeOf(int16(0)), Int16Codec{})
	p.RegisterCodec(reflect.TypeOf(int32(0)), Int32Codec{})
	p.RegisterCodec(reflect.TypeOf(int64(0)), Int64Codec{})
	p.RegisterCodec(reflect.TypeOf(uint(0)), UintCodec{})
	p.RegisterCodec(reflect.TypeOf(uint64(0)), Uint64Codec{})
	p.RegisterCodec(reflect.TypeOf(uint32(0)), Uint32Codec{})
	p.RegisterCodec(reflect.TypeOf(uint16(0)), Uint16Codec{})
	p.RegisterCodec(reflect.TypeOf(uint8(0)), Uint8Codec{})
	p.RegisterCodec(reflect.TypeOf(""), StringCodec{})
	p.RegisterCodec(reflect.TypeOf([]byte(nil)), BytesCodec{})
	p.RegisterCodec(reflect.TypeOf(time.Time{}), &TimeCodec{})
}
