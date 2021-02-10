package plenc

import "unsafe"

//go:linkname unsafe_NewArray reflect.unsafe_NewArray
func unsafe_NewArray(rtype unsafe.Pointer, length int) unsafe.Pointer

// typedslicecopy copies a slice of elemType values from src to dst,
// returning the number of elements copied.
//go:linkname typedslicecopy reflect.typedslicecopy
//go:noescape
func typedslicecopy(elemType unsafe.Pointer, dst, src sliceHeader) int

//go:linkname mapassign reflect.mapassign
//go:noescape
func mapassign(typ unsafe.Pointer, hmap unsafe.Pointer, key, val unsafe.Pointer)

// m escapes into the return value, but the caller of mapiterinit
// doesn't let the return value escape.
//go:linkname mapiterinit reflect.mapiterinit
//go:noescape
func mapiterinit(t unsafe.Pointer, m unsafe.Pointer) unsafe.Pointer

//go:linkname mapiterkey reflect.mapiterkey
//go:noescape
func mapiterkey(it unsafe.Pointer) (key unsafe.Pointer)

//go:linkname mapiterelem reflect.mapiterelem
//go:noescape
func mapiterelem(it unsafe.Pointer) (elem unsafe.Pointer)

//go:linkname mapiternext reflect.mapiternext
//go:noescape
func mapiternext(it unsafe.Pointer)

//go:linkname maplen reflect.maplen
//go:noescape
func maplen(m unsafe.Pointer) int
