package plenc

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/philpearl/plenc/plenccodec"
)

// Marshal serialises value, appending it to data. If data is nil a new slice is
// created.
func Marshal(data []byte, value any) ([]byte, error) {
	return defaultPlenc.Marshal(data, value)
}

// Unmarshal deserialises data into value.
func Unmarshal(data []byte, value any) error {
	return defaultPlenc.Unmarshal(data, value)
}

// Size returns the number of bytes required to marshal the value.
func Size(value any) (int, error) {
	return defaultPlenc.Size(value)
}

func (p *Plenc) getCodec(value any) (plenccodec.Codec, error) {
	typ := reflect.TypeOf(value)
	if typ == nil {
		// A really very nil nil
		return nil, nil
	}
	// If our value is a pointer, the pointer is directly in data. If it is a
	// value, then Go allocates a space on the heap, and gives us a pointer to
	// that. So in most cases, whether we receive a pointer or not, we get a
	// pointer here.
	//
	// maps are essentially pointers. if we're passed a map by value it isn't
	// put into an allocated space on the heap. We're given the map pointer
	// directly. So we need to treat the pointer and non-pointer cases
	// differently.
	//
	// This also applies if we're passed a struct that contains just a map.
	// That's treated like we were just passed a map. And that applies if the
	// map is wrapped in several layers of struct. As long as it's just a map,
	// that's what we get.
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	c, err := p.CodecForType(typ)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (p *Plenc) Marshal(data []byte, value any) ([]byte, error) {
	c, err := p.getCodec(value)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}

	ptr := unpackEFace(value).data
	if typ := reflect.TypeOf(value); shouldWrapPtr(typ) {
		ptrWrap := ptr
		ptr = noEscape(unsafe.Pointer(&ptrWrap))
	}

	if ptr == nil || c.Omit(ptr) {
		return nil, nil
	}
	if data == nil {
		data = make([]byte, 0, c.Size(ptr, nil))
	}

	return c.Append(data, ptr, nil), nil
}

func shouldWrapPtr(typ reflect.Type) bool {
	if typ.Kind() == reflect.Pointer {
		return false
	}
	return isBasicallyMapOrPointer(typ)
}

func isBasicallyMapOrPointer(typ reflect.Type) bool {
	if typ.Kind() == reflect.Map || typ.Kind() == reflect.Pointer {
		return true
	}
	if typ.Kind() == reflect.Struct && typ.NumField() == 1 && isBasicallyMapOrPointer(typ.Field(0).Type) {
		return true
	}
	return false
}

func (p *Plenc) Unmarshal(data []byte, value any) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("you must pass in a non-nil pointer")
	}

	c, err := p.CodecForType(rv.Type().Elem())
	if err != nil {
		return err
	}

	_, err = c.Read(data, unsafe.Pointer(rv.Pointer()), c.WireType())
	return err
}

func (p *Plenc) Size(value any) (int, error) {
	c, err := p.getCodec(value)
	if err != nil {
		return 0, err
	}
	if c == nil {
		return 0, nil
	}

	ptr := unpackEFace(value).data
	if typ := reflect.TypeOf(value); shouldWrapPtr(typ) {
		ptrWrap := ptr
		ptr = noEscape(unsafe.Pointer(&ptrWrap))
	}

	if c.Omit(ptr) {
		return 0, nil
	}

	return c.Size(ptr, nil), nil
}
