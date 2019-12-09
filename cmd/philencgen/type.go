package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/loader"
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

// TODO: field indexes stored in a file (not Go) next to the type. So that we can persist indices without folk
// needing to manually tag the fields. And we can assign indices to new fields without clashing with removed
// ones.

func parseType(o *options) (d data, err error) {
	// Load up the package we're interested in
	conf := loader.Config{}
	conf.Import(o.path)

	lprog, err := conf.Load()
	if err != nil {
		return d, err
	}

	pkgInfo := lprog.Package(o.path)
	if pkgInfo == nil {
		return d, fmt.Errorf("package %s not found", o.path)
	}
	pkg := pkgInfo.Pkg
	obj := pkg.Scope().Lookup(o.structName)
	if obj == nil {
		return d, fmt.Errorf("type %s not found", o.structName)
	}

	d.Package = pkg.Name()
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

	var isSlice bool
	if slice, ok := typ.(*types.Slice); ok {
		isSlice = true
		typ = slice.Elem()
	}

	if named, ok := typ.(*types.Named); ok {
		f.Type = named.Obj().Name()
		if isSlice {
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

	if basic, ok := typ.(*types.Basic); ok {
		f.Type = basic.Name()
		bi := basic.Info()
		switch {
		case bi&types.IsBoolean != 0:
			f.DecodeTemplate = "BoolDecode"
			f.SizeTemplate = "BoolSize"
			f.AppendTemplate = "BoolAppend"
		case bi&types.IsInteger|types.IsUnsigned == types.IsInteger|types.IsUnsigned:
			if basic.Kind() == types.Byte {
				// TODO: []byte can be special
			}
			f.DecodeTemplate = "UintDecode"
			f.SizeTemplate = "UintSize"
			f.AppendTemplate = "UintAppend"
		case bi&types.IsInteger|types.IsUnsigned == types.IsInteger:
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
