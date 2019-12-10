package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type data struct {
	Package string
	Name    string
	Dir     string
	Fields  []field
}

type field struct {
	Name           string
	Index          int
	Type           string
	DecodeTemplate string
	SizeTemplate   string
	AppendTemplate string
	IsPointer      bool
}

var intf *types.Interface

func getInterface() (*types.Interface, error) {
	if intf != nil {
		return intf, nil
	}
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedName | packages.NeedImports}

	pkgs, err := packages.Load(cfg, "github.com/philpearl/philenc")
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

// TODO: field indexes stored in a file (not Go) next to the type. So that we can persist indices without folk
// needing to manually tag the fields. And we can assign indices to new fields without clashing with removed
// ones.

func parseType(o *options) (d data, err error) {
	// Load up the package we're interested in
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedName | packages.NeedImports}

	pkgs, err := packages.Load(cfg, o.path)
	if err != nil {
		return d, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return d, fmt.Errorf("package has errors")
	}
	if len(pkgs) == 0 {
		return d, fmt.Errorf("no packages")
	}

	pkg := pkgs[0]
	obj := pkg.Types.Scope().Lookup(o.structName)
	if obj == nil {
		return d, fmt.Errorf("type %s not found", o.structName)
	}

	d.Package = pkg.Name
	d.Name = o.structName

	// We hope this is a structure
	s, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return d, fmt.Errorf("type %s is not a struct", d.Name)
	}

	d.Fields = make([]field, 0, s.NumFields())

	for i := 0; i < s.NumFields(); i++ {
		v := s.Field(i)
		if !v.Exported() {
			continue
		}
		typ := v.Type()
		f := field{
			Name:  v.Name(),
			Index: i + 1, // TODO: NO!!! Need something that is resistant to changes
		}
		if err := parseField(typ, &f); err != nil {
			return d, fmt.Errorf("failed to parse field %d %s. %w", i, f.Name, err)
		}
		d.Fields = append(d.Fields, f)
	}

	return d, nil
}

func parseField(typ types.Type, f *field) error {

	if ptr, ok := typ.(*types.Pointer); ok {
		f.IsPointer = true
		typ = ptr.Elem()
	}

	var isSlice, isPointerSlice bool
	if slice, ok := typ.(*types.Slice); ok {
		isSlice = true
		typ = slice.Elem()

		if ptr, ok := typ.(*types.Pointer); ok {
			isPointerSlice = true
			typ = ptr.Elem()
		}
	}

	if named, ok := typ.(*types.Named); ok {
		// If we have a named type, then if it is a struct we assume it will implement our interfaces eventually.
		// If the underlying type is a basic type then we check whether it implements the interfaces
		obj := named.Obj()
		f.Type = obj.Name()

		if obj.Name() == "Time" && obj.Pkg().Name() == "time" {
			f.DecodeTemplate = "TimeDecode"
			f.SizeTemplate = "TimeSize"
			f.AppendTemplate = "TimeAppend"
			return nil
		}

		intf, err := getInterface()
		if err != nil {
			return fmt.Errorf("failed to get Marshaler interface. %w", err)
		}

		_, isStruct := named.Underlying().(*types.Struct)
		if isStruct || types.Implements(named, intf) {
			if isPointerSlice {
				f.DecodeTemplate = "MethodPointerSliceDecode"
				f.SizeTemplate = "MethodSliceSize"
				f.AppendTemplate = "MethodSliceAppend"
			} else if isSlice {
				f.DecodeTemplate = "MethodSliceDecode"
				f.SizeTemplate = "MethodSliceSize"
				f.AppendTemplate = "MethodSliceAppend"
			} else {
				f.DecodeTemplate = "MethodDecode"
				f.SizeTemplate = "MethodSize"
				f.AppendTemplate = "MethodAppend"
			}
			return nil
		}

		typ = named.Underlying()
	}

	if basic, ok := typ.(*types.Basic); ok {
		if f.Type == "" {
			f.Type = basic.Name()
		}
		bi := basic.Info()
		switch {
		case bi&types.IsBoolean != 0:
			f.DecodeTemplate = "BoolDecode"
			f.SizeTemplate = "BoolSize"
			f.AppendTemplate = "BoolAppend"
		case bi&(types.IsInteger|types.IsUnsigned) == types.IsInteger|types.IsUnsigned:
			if basic.Kind() == types.Byte {
				// TODO: []byte can be special
			}
			f.DecodeTemplate = "UintDecode"
			f.SizeTemplate = "UintSize"
			f.AppendTemplate = "UintAppend"
		case bi&(types.IsInteger|types.IsUnsigned) == types.IsInteger:
			f.DecodeTemplate = "IntDecode"
			f.SizeTemplate = "IntSize"
			f.AppendTemplate = "IntAppend"
		case bi&types.IsFloat == types.IsFloat:
			if basic.Kind() == types.Float32 {
				f.DecodeTemplate = "Float32Decode"
				f.SizeTemplate = "Float32Size"
				f.AppendTemplate = "Float32Append"
			} else {
				f.DecodeTemplate = "Float64Decode"
				f.SizeTemplate = "Float64Size"
				f.AppendTemplate = "Float64Append"
			}
		case bi&types.IsString == types.IsString:
			f.DecodeTemplate = "StringDecode"
			f.SizeTemplate = "StringSize"
			f.AppendTemplate = "StringAppend"

		default:
			return fmt.Errorf("unsupported basic type %s", basic.Name())
		}
		if isSlice {
			return fmt.Errorf("unsupprted slice of basic type %s", basic.Name())
		}
		if f.IsPointer {
			return fmt.Errorf("unsupported pointer to basic type %s", basic.Name())
		}
		return nil
	}

	return fmt.Errorf("unsupported type %s. %T", typ.String(), typ)
}
