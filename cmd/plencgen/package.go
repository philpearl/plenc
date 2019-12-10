package main

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

// getPackage loads up package information for our target package
func getPackage(path string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedName | packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("package has errors")
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages")
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("expected to load one package. Have %d packages", len(pkgs))
	}

	return pkgs[0], nil
}
