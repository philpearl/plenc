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

	c, err := p.codecForType(typ)
	if err != nil {
		return nil, err
	}

	if c.Omit(ptr) {
		return nil, nil
	}
	if data == nil {
		data = make([]byte, 0, c.Size(ptr))
	}

	return c.Append(data, ptr), nil
}

func (p *Plenc) Unmarshal(data []byte, value interface{}) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("you must pass in a non-nil pointer")
	}

	c, err := p.codecForType(rv.Type().Elem())
	if err != nil {
		return err
	}

	_, err = c.Read(data, unsafe.Pointer(rv.Pointer()), c.WireType())
	return err
}

type eface struct {
	rtype unsafe.Pointer
	data  unsafe.Pointer
}

func unpackEFace(obj interface{}) *eface {
	return (*eface)(unsafe.Pointer(&obj))
}

func packEFace(ptr unsafe.Pointer, typ reflect.Type) interface{} {
	e := eface{
		rtype: unpackEFace(typ).data,
		data:  ptr,
	}
	return *(*interface{})(unsafe.Pointer(&e))
}
