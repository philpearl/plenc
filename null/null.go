// Package null contains plenc codecs for the types in github.com/unravelin/null.
// Call RegisterCodecs to make these codecs available to plenc
package null

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plenccodec"
	"github.com/philpearl/plenc/plenccore"
	"github.com/unravelin/null"
)

// RegisterCodecs registers the codecs from this package and makes them
// available to plenc.
func RegisterCodecs() {
	plenc.RegisterCodec(reflect.TypeOf(null.Int{}), nullIntCodec{})
	plenc.RegisterCodec(reflect.TypeOf(null.Bool{}), nullBoolCodec{})
	plenc.RegisterCodec(reflect.TypeOf(null.Float{}), nullFloatCodec{})
	plenc.RegisterCodec(reflect.TypeOf(null.String{}), nullStringCodec{})
	plenc.RegisterCodec(reflect.TypeOf(null.Time{}), &nullTimeCodec{})
}

// AddCodecs registers the codecs from this package and makes them
// available to the given plenc instance
func AddCodecs(p *plenc.Plenc) {
	p.RegisterCodec(reflect.TypeOf(null.Int{}), nullIntCodec{})
	p.RegisterCodec(reflect.TypeOf(null.Bool{}), nullBoolCodec{})
	p.RegisterCodec(reflect.TypeOf(null.Float{}), nullFloatCodec{})
	p.RegisterCodec(reflect.TypeOf(null.String{}), nullStringCodec{})
	p.RegisterCodec(reflect.TypeOf(null.Time{}), &nullTimeCodec{})
}

type nullIntCodec struct {
	plenccodec.IntCodec[int64]
}

func (c nullIntCodec) Omit(ptr unsafe.Pointer) bool {
	n := *(*null.Int)(ptr)
	return !n.Valid
}

func (c nullIntCodec) Size(ptr unsafe.Pointer, tag []byte) (size int) {
	ni := (*null.Int)(ptr)
	return c.IntCodec.Size(unsafe.Pointer(&ni.Int64), tag)
}

func (c nullIntCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	ni := (*null.Int)(ptr)
	return c.IntCodec.Append(data, unsafe.Pointer(&ni.Int64), tag)
}

func (c nullIntCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	var i int64
	n, err = c.IntCodec.Read(data, unsafe.Pointer(&i), wt)
	if err != nil {
		return n, err
	}
	ni := (*null.Int)(ptr)
	ni.SetValid(i)
	return n, err
}

func (c nullIntCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&null.Int{})
}

func (c nullIntCodec) Descriptor() plenccodec.Descriptor {
	d := c.IntCodec.Descriptor()
	d.ExplicitPresence = true
	return d
}

type nullBoolCodec struct {
	plenccodec.BoolCodec
}

func (c nullBoolCodec) Omit(ptr unsafe.Pointer) bool {
	n := *(*null.Bool)(ptr)
	return !n.Valid
}

func (c nullBoolCodec) Size(ptr unsafe.Pointer, tag []byte) (size int) {
	ni := (*null.Bool)(ptr)
	return c.BoolCodec.Size(unsafe.Pointer(&ni.Bool), tag)
}

func (c nullBoolCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	ni := (*null.Bool)(ptr)
	return c.BoolCodec.Append(data, unsafe.Pointer(&ni.Bool), tag)
}

func (c nullBoolCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	var b bool
	n, err = c.BoolCodec.Read(data, unsafe.Pointer(&b), wt)
	if err != nil {
		return n, err
	}
	nb := (*null.Bool)(ptr)
	nb.SetValid(b)
	return n, err
}

func (c nullBoolCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&null.Bool{})
}

func (c nullBoolCodec) Descriptor() plenccodec.Descriptor {
	d := c.BoolCodec.Descriptor()
	d.ExplicitPresence = true
	return d
}

type nullFloatCodec struct {
	plenccodec.Float64Codec
}

func (c nullFloatCodec) Omit(ptr unsafe.Pointer) bool {
	n := *(*null.Float)(ptr)
	return !n.Valid
}

func (c nullFloatCodec) Size(ptr unsafe.Pointer, tag []byte) (size int) {
	nf := (*null.Float)(ptr)
	return c.Float64Codec.Size(unsafe.Pointer(&nf.Float64), tag)
}

func (c nullFloatCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	nf := (*null.Float)(ptr)
	return c.Float64Codec.Append(data, unsafe.Pointer(&nf.Float64), tag)
}

func (c nullFloatCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	var f float64
	n, err = c.Float64Codec.Read(data, unsafe.Pointer(&f), wt)
	if err != nil {
		return n, err
	}
	nf := (*null.Float)(ptr)
	nf.Valid = true
	nf.Float64 = f
	return n, err
}

func (c nullFloatCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&null.Float{})
}

func (c nullFloatCodec) Descriptor() plenccodec.Descriptor {
	d := c.Float64Codec.Descriptor()
	d.ExplicitPresence = true
	return d
}

type nullStringCodec struct {
	plenccodec.StringCodec
}

func (c nullStringCodec) Omit(ptr unsafe.Pointer) bool {
	n := *(*null.String)(ptr)
	return !n.Valid
}

func (c nullStringCodec) Size(ptr unsafe.Pointer, tag []byte) (size int) {
	ns := (*null.String)(ptr)
	return c.StringCodec.Size(unsafe.Pointer(&ns.String), tag)
}

func (c nullStringCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	ns := (*null.String)(ptr)
	return c.StringCodec.Append(data, unsafe.Pointer(&ns.String), tag)
}

func (c nullStringCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	ns := (*null.String)(ptr)
	n, err = c.StringCodec.Read(data, unsafe.Pointer(&ns.String), wt)
	if err != nil {
		return n, err
	}
	ns.Valid = true
	return n, err
}

func (c nullStringCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&null.String{})
}

func (c nullStringCodec) Descriptor() plenccodec.Descriptor {
	d := c.StringCodec.Descriptor()
	d.ExplicitPresence = true
	return d
}

func (nullStringCodec) WithInterning() plenccodec.Codec {
	c := plenccodec.StringCodec{}.WithInterning()
	return &internedNullStringCodec{
		internedCodec: c,
	}
}

type internedNullStringCodec struct {
	nullStringCodec
	internedCodec plenccodec.Codec
}

func (c *internedNullStringCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	ns := (*null.String)(ptr)
	n, err = c.internedCodec.Read(data, unsafe.Pointer(&ns.String), wt)
	if err != nil {
		return n, err
	}
	ns.Valid = true
	return n, err
}

type nullTimeCodec struct {
	plenccodec.TimeCodec
}

func (c *nullTimeCodec) Omit(ptr unsafe.Pointer) bool {
	n := *(*null.Time)(ptr)
	return !n.Valid
}

func (c *nullTimeCodec) Size(ptr unsafe.Pointer, tag []byte) (size int) {
	nt := (*null.Time)(ptr)
	return c.TimeCodec.Size(unsafe.Pointer(&nt.Time), tag)
}

func (c *nullTimeCodec) Append(data []byte, ptr unsafe.Pointer, tag []byte) []byte {
	nt := (*null.Time)(ptr)
	return c.TimeCodec.Append(data, unsafe.Pointer(&nt.Time), tag)
}

func (c *nullTimeCodec) Read(data []byte, ptr unsafe.Pointer, wt plenccore.WireType) (n int, err error) {
	var t time.Time
	n, err = c.TimeCodec.Read(data, unsafe.Pointer(&t), wt)
	if err != nil {
		return n, err
	}
	nt := (*null.Time)(ptr)
	nt.Valid = true
	nt.Time = t
	return n, err
}

func (c *nullTimeCodec) New() unsafe.Pointer {
	return unsafe.Pointer(&null.Time{})
}

func (c nullTimeCodec) Descriptor() plenccodec.Descriptor {
	d := c.TimeCodec.Descriptor()
	d.ExplicitPresence = true
	return d
}
