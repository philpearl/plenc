package plenc

import "unsafe"

type eface struct {
	rtype unsafe.Pointer
	data  unsafe.Pointer
}

func unpackEFace(obj any) *eface {
	return (*eface)(unsafe.Pointer(&obj))
}
