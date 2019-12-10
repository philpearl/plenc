package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

var intf *types.Interface

func getInterface() (*types.Interface, error) {
	if intf != nil {
		return intf, nil
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

	obj := pkgs[0].Types.Scope().Lookup("Marshaler")
	if obj == nil {
		return nil, fmt.Errorf("could not find marshaler")
	}

	intf = obj.Type().Underlying().(*types.Interface)

	return intf, nil
}
