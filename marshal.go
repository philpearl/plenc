package plenc

import (
	"fmt"
	"reflect"
	"unsafe"
)

func Marshal(data []byte, value interface{}) ([]byte, error) {
	return defaultPlenc.Marshal(data, value)
}

func Unmarshal(data []byte, value interface{}) error {
	return defaultPlenc.Unmarshal(data, value)
}

func (p *Plenc) Marshal(data []byte, value interface{}) ([]byte, error) {
	typ := reflect.TypeOf(value)
	ptr := unpackEFace(value).data
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()

		// When marshalling we don't want a pointer to a map as a map is a
		// pointer-ish type itself.
		if typ.Kind() == reflect.Map {
			ptr = *(*unsafe.Pointer)(ptr)
		}
	}

	c, err := p.CodecForType(typ)
	if err != nil {
		return nil, err
	}

	if c.Omit(ptr) {
		return nil, nil
	}
	if data == nil {
		data = make([]byte, 0, c.Size(ptr, nil))
	}

	return c.Append(data, ptr, nil), nil
}

func (p *Plenc) Unmarshal(data []byte, value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("you must pass in a non-nil pointer")
	}

	c, err := p.CodecForType(rv.Type().Elem())
	if err != nil {
		return err
	}

	_, err = c.Read(data, unsafe.Pointer(rv.Pointer()), c.WireType())
	return err
}
