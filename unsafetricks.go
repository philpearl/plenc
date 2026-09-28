package plenc

import "unsafe"

type eface struct {
	rtype unsafe.Pointer
	data  unsafe.Pointer
}

func unpackEFace(obj any) *eface {
	return (*eface)(unsafe.Pointer(&obj))
}

// NoEscape hides the pointer p from escape analysis, preventing it
// from escaping to the heap. It compiles down to nothing.
//
// WARNING: This is very subtle to use correctly. The caller must
// ensure that it's truly safe for p to not escape to the heap by
// maintaining runtime pointer invariants (for example, that globals
// and the heap may not generally point into a stack).
//
//go:nosplit
//go:nocheckptr
func noEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
