package main

import (
	"bytes"
	"go/format"
	"strings"
	"text/template"
)

func createMarshaler(typeInfo data) ([]byte, error) {
	var b bytes.Buffer
	if err := encodeTmpl.Execute(&b, typeInfo); err != nil {
		return nil, err
	}

	return format.Source(b.Bytes())
}

var encodeTmplFuncs = template.FuncMap{
	"runTemplate": runEncodeTemplate,
}

func runEncodeTemplate(name string, data interface{}) (string, error) {
	var b strings.Builder
	err := encodeTmpl.ExecuteTemplate(&b, name, data)
	return b.String(), err
}

var encodeTmpl *template.Template

func init() {
	encodeTmpl = template.Must(template.New("encode").Funcs(encodeTmplFuncs).Parse(`
	package {{ .Package }}
	
	import (
		"github.com/philpearl/plenc"
	)
	
	// TODO: missing types
	// slice of numeric ()
	// slice of other
	// pointers
	// TODO: option whether top-level type is a pointer for marshaler
	
	{{ define "TimeSize" }}
		{
			var t plenc.Time
			t.Set(e.{{.Name}})
			if s := t.ΦλSize(); s != 0 {
				size += plenc.SizeTag(plenc.WTLength, {{.Index}})
				size += plenc.SizeVarUint(uint64(s))
				size += s		
			}
	
		}
	{{ end }}
	
	{{ define "TimeAppend" }}
	{
		var t plenc.Time
		t.Set(e.{{.Name}})
		if 	s := t.ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, {{.Index}})
			data = plenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)	
		}
	}
	{{ end }}

	{{ define "TimePtrSize" }}
		if e.{{.Name}} != nil {
			var t plenc.Time
			t.Set(*e.{{.Name}})
			if s := t.ΦλSize(); s != 0 {
				size += plenc.SizeTag(plenc.WTLength, {{.Index}})
				size += plenc.SizeVarUint(uint64(s))
				size += s		
			}
	
		}
	{{ end }}
	
	{{ define "TimePtrAppend" }}
	if e.{{.Name}} != nil {
		var t plenc.Time
		t.Set(*e.{{.Name}})
		if 	s := t.ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, {{.Index}})
			data = plenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)	
		}
	}
	{{ end }}


	{{ define "MethodSize" }}
		if s := e.{{.Name}}.ΦλSize(); s != 0 {
			size += plenc.SizeTag(plenc.WTLength, {{.Index}})
			size += plenc.SizeVarUint(uint64(s))
			size += s		
		}
	{{ end }}
	
	{{ define "MethodAppend" }}
		if 	s := e.{{.Name}}.ΦλSize(); s != 0 {
			data = plenc.AppendTag(data, plenc.WTLength, {{.Index}})
			data = plenc.AppendVarUint(data, uint64(s))
			data = e.{{.Name}}.ΦλAppend(data)	
		}
	{{ end }}
	
	{{ define "MethodSliceSize" }}
		// Each element of the slice is separately encoded
		for i := range e.{{.Name}} {
			if s := e.{{.Name}}[i].ΦλSize(); s != 0 {
				size += plenc.SizeTag(plenc.WTLength, {{.Index}})
				size += plenc.SizeVarUint(uint64(s))
				size += s	
			}
		}
	{{ end }}
	
	{{ define "MethodSliceAppend" }}
		// Each element of the slice is separately encoded
		for i := range e.{{.Name}} {
			if s := e.{{.Name}}[i].ΦλSize(); s != 0 {
				data = plenc.AppendTag(data, plenc.WTLength, {{.Index}})
				data = plenc.AppendVarUint(data, uint64(s))
				data = e.{{.Name}}[i].ΦλAppend(data)		
			}
		}
	{{ end }}
	
	{{ define "BoolSize" }}
		size += plenc.SizeTag(plenc.WTVarInt, {{.Index}})
		size += plenc.SizeBool(e.{{.Name}})
	{{ end }}
	
	{{ define "BoolAppend" }}
		data = plenc.AppendTag(data, plenc.WTVarInt, {{.Index}})
		data = plenc.AppendBool(data, e.{{.Name}})
	{{ end }}
	
	{{ define "Float32Size" }}
		size += plenc.SizeTag(plenc.WT32, {{.Index}})
		size += plenc.SizeFloat32(float32(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float32Append" }}
		data = plenc.AppendTag(data, plenc.WT32, {{.Index}})
		data = plenc.AppendFloat32(data, float32(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float64Size" }}
		size += plenc.SizeTag(plenc.WT64, {{.Index}})
		size += plenc.SizeFloat64(float64(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float64Append" }}
		data = plenc.AppendTag(data, plenc.WT64, {{.Index}})
		data = plenc.AppendFloat64(data, float64(e.{{.Name}}))
	{{ end }}
	
	{{ define "IntSize" }}
		size += plenc.SizeTag(plenc.WTVarInt, {{.Index}})
		size += plenc.SizeVarInt(int64(e.{{.Name}}))
	{{ end }}
	
	{{ define "IntAppend" }}
		data = plenc.AppendTag(data, plenc.WTVarInt, {{.Index}})
		data = plenc.AppendVarInt(data, int64(e.{{.Name}}))
	{{ end }}
	
	{{ define "UintSize" }}
		size += plenc.SizeTag(plenc.WTVarInt, {{.Index}})
		size += plenc.SizeVarUint(uint64(e.{{.Name}}))
	{{ end }}
	
	{{ define "UintAppend" }}
		data = plenc.AppendTag(data, plenc.WTVarInt, {{.Index}})
		data = plenc.AppendVarUint(data, uint64(e.{{.Name}}))
	{{ end }}
	
	
	{{ define "StringSize" }}
		size += plenc.SizeTag(plenc.WTLength, {{.Index}})
		size += plenc.SizeString(e.{{.Name}})
	{{ end }}
	
	{{ define "StringAppend" }}
		data = plenc.AppendTag(data, plenc.WTLength, {{.Index}})
		data = plenc.AppendString(data, e.{{.Name}})
	{{ end }}
	
	
	// ΦλSize works out how many bytes are needed to encode {{.Name}}
	func (e *{{ .Name }}) ΦλSize() (size int) {
		if e == nil {
			return 0
		}
	
	{{ range .Fields }}
		{{ runTemplate .SizeTemplate . }}
	{{ end }}
	
		return size
	}
	
	// ΦλAppend encodes {{.Name}} by appending to data. It returns the final slice
	func (e *{{ .Name }}) ΦλAppend(data []byte) []byte {
	
	{{ range .Fields }}
		{{ runTemplate .AppendTemplate . }}
	{{ end }}
	
		return data
	}
	`))

}
