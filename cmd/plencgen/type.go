package main

import (
	"fmt"
	"go/types"
	"path/filepath"
)

type data struct {
	Package string
	Name    string
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

func parseType(o *options) (d data, err error) {
	// Load up the package we're interested in
	pkg, err := getPackage(o.path)
	if err != nil {
		return d, fmt.Errorf("failed to load package %s. %w", o.path, err)
	}

	obj := pkg.Types.Scope().Lookup(o.structName)
	if obj == nil {
		return d, fmt.Errorf("type %s not found", o.structName)
	}

	// There may be a file containing index information for each field
	idx, err := loadIndex(filepath.Dir(pkg.GoFiles[0]), o.structName)
	if err != nil {
		return d, err
	}

	d.Package = pkg.Name
	d.Name = o.structName

	// The type we're generating methods for should be a named struct
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
		typ, name := v.Type(), v.Name()
		f := field{
			Name:  name,
			Index: idx.indexFor(name),
		}
		if err := parseField(typ, &f); err != nil {
			return d, fmt.Errorf("failed to parse field %d %s. %w", i, f.Name, err)
		}
		d.Fields = append(d.Fields, f)
	}

	return d, idx.save()
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
			f.Type = "time.Time"
			if f.IsPointer {
				f.DecodeTemplate = "TimePtrDecode"
				f.SizeTemplate = "TimePtrSize"
				f.AppendTemplate = "TimePtrAppend"
			} else {
				f.DecodeTemplate = "TimeDecode"
				f.SizeTemplate = "TimeSize"
				f.AppendTemplate = "TimeAppend"
			}
			return nil
		}

		intfs, err := getInterface()
		if err != nil {
			return fmt.Errorf("failed to get Marshaler interface. %w", err)
		}

		if types.Implements(named, intfs.fullMarshaler) {
			if isPointerSlice {
				f.DecodeTemplate = "FullPointerSliceDecode"
				f.SizeTemplate = "FullSliceSize"
				f.AppendTemplate = "FullSliceAppend"
			} else if isSlice {
				f.DecodeTemplate = "FullSliceDecode"
				f.SizeTemplate = "FullSliceSize"
				f.AppendTemplate = "FullSliceAppend"
			} else {
				f.DecodeTemplate = "FullDecode"
				f.SizeTemplate = "FullSize"
				f.AppendTemplate = "FullAppend"
			}
			return nil
		}

		_, isStruct := named.Underlying().(*types.Struct)
		if isStruct || types.Implements(named, intfs.marshaler) {
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
