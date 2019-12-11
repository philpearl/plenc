package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type interfaces struct {
	marshaler     *types.Interface
	unmarshaler   *types.Interface
	fullMarshaler *types.Interface
}

var intfs *interfaces

func getInterface() (*interfaces, error) {
	if intfs != nil {
		return intfs, nil
	}
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}

	pkgs, err := packages.Load(cfg, "github.com/philpearl/plenc")
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package has errors")
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages")
	}

	objM := pkgs[0].Types.Scope().Lookup("Marshaler")
	if objM == nil {
		return nil, fmt.Errorf("could not find marshaler")
	}
	objU := pkgs[0].Types.Scope().Lookup("Unmarshaler")
	if objU == nil {
		return nil, fmt.Errorf("could not find Unmarshaler")
	}
	objF := pkgs[0].Types.Scope().Lookup("FullMarshaler")
	if objF == nil {
		return nil, fmt.Errorf("could not find FullMarshaler")
	}

	intfs = &interfaces{
		marshaler:     objM.Type().Underlying().(*types.Interface),
		unmarshaler:   objU.Type().Underlying().(*types.Interface),
		fullMarshaler: objF.Type().Underlying().(*types.Interface),
	}

	return intfs, nil
}
