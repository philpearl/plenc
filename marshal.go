package plenc

import (
	"fmt"
	"reflect"
	"unsafe"
)

func Marshal(data []byte, value interface{}) ([]byte, error) {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	c, err := codecForType(typ)
	if err != nil {
		return nil, err
	}

	return c.Append(data, unsafe.Pointer(reflect.ValueOf(value).Pointer())), nil
}

func Unmarshal(data []byte, value interface{}) error {
	typ := reflect.TypeOf(value)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("you must pass in a pointer")
	}

	c, err := codecForType(typ.Elem())
	if err != nil {
		return err
	}

	_, err = c.Read(data, unsafe.Pointer(reflect.ValueOf(value).Pointer()))
	return err
}
