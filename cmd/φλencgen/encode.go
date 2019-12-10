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
		"github.com/philpearl/φλenc"
	)
	
	// TODO: missing types
	// slice of numeric ()
	// slice of other
	// pointers
	// TODO: option whether top-level type is a pointer for marshaler
	
	{{ define "TimeSize" }}
		{
			var t φλenc.Time
			t.Set(e.{{.Name}})
			if s := t.ΦλSize(); s != 0 {
				size += φλenc.SizeTag(φλenc.WTLength, {{.Index}})
				size += φλenc.SizeVarUint(uint64(s))
				size += s		
			}
	
		}
	{{ end }}
	
	{{ define "TimeAppend" }}
	{
		var t φλenc.Time
		t.Set(e.{{.Name}})
		if 	s := t.ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, {{.Index}})
			data = φλenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)	
		}
	}
	{{ end }}

	{{ define "TimePtrSize" }}
		if e.{{.Name}} != nil {
			var t φλenc.Time
			t.Set(*e.{{.Name}})
			if s := t.ΦλSize(); s != 0 {
				size += φλenc.SizeTag(φλenc.WTLength, {{.Index}})
				size += φλenc.SizeVarUint(uint64(s))
				size += s		
			}
	
		}
	{{ end }}
	
	{{ define "TimePtrAppend" }}
	if e.{{.Name}} != nil {
		var t φλenc.Time
		t.Set(*e.{{.Name}})
		if 	s := t.ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, {{.Index}})
			data = φλenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)	
		}
	}
	{{ end }}


	{{ define "MethodSize" }}
		if s := e.{{.Name}}.ΦλSize(); s != 0 {
			size += φλenc.SizeTag(φλenc.WTLength, {{.Index}})
			size += φλenc.SizeVarUint(uint64(s))
			size += s		
		}
	{{ end }}
	
	{{ define "MethodAppend" }}
		if 	s := e.{{.Name}}.ΦλSize(); s != 0 {
			data = φλenc.AppendTag(data, φλenc.WTLength, {{.Index}})
			data = φλenc.AppendVarUint(data, uint64(s))
			data = e.{{.Name}}.ΦλAppend(data)	
		}
	{{ end }}
	
	{{ define "MethodSliceSize" }}
		// Each element of the slice is separately encoded
		for i := range e.{{.Name}} {
			if s := e.{{.Name}}[i].ΦλSize(); s != 0 {
				size += φλenc.SizeTag(φλenc.WTLength, {{.Index}})
				size += φλenc.SizeVarUint(uint64(s))
				size += s	
			}
		}
	{{ end }}
	
	{{ define "MethodSliceAppend" }}
		// Each element of the slice is separately encoded
		for i := range e.{{.Name}} {
			if s := e.{{.Name}}[i].ΦλSize(); s != 0 {
				data = φλenc.AppendTag(data, φλenc.WTLength, {{.Index}})
				data = φλenc.AppendVarUint(data, uint64(s))
				data = e.{{.Name}}[i].ΦλAppend(data)		
			}
		}
	{{ end }}
	
	{{ define "BoolSize" }}
		size += φλenc.SizeTag(φλenc.WTVarInt, {{.Index}})
		size += φλenc.SizeBool(e.{{.Name}})
	{{ end }}
	
	{{ define "BoolAppend" }}
		data = φλenc.AppendTag(data, φλenc.WTVarInt, {{.Index}})
		data = φλenc.AppendBool(data, e.{{.Name}})
	{{ end }}
	
	{{ define "Float32Size" }}
		size += φλenc.SizeTag(φλenc.WT32, {{.Index}})
		size += φλenc.SizeFloat32(float32(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float32Append" }}
		data = φλenc.AppendTag(data, φλenc.WT32, {{.Index}})
		data = φλenc.AppendFloat32(data, float32(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float64Size" }}
		size += φλenc.SizeTag(φλenc.WT64, {{.Index}})
		size += φλenc.SizeFloat64(float64(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float64Append" }}
		data = φλenc.AppendTag(data, φλenc.WT64, {{.Index}})
		data = φλenc.AppendFloat64(data, float64(e.{{.Name}}))
	{{ end }}
	
	{{ define "IntSize" }}
		size += φλenc.SizeTag(φλenc.WTVarInt, {{.Index}})
		size += φλenc.SizeVarInt(int64(e.{{.Name}}))
	{{ end }}
	
	{{ define "IntAppend" }}
		data = φλenc.AppendTag(data, φλenc.WTVarInt, {{.Index}})
		data = φλenc.AppendVarInt(data, int64(e.{{.Name}}))
	{{ end }}
	
	{{ define "UintSize" }}
		size += φλenc.SizeTag(φλenc.WTVarInt, {{.Index}})
		size += φλenc.SizeVarUint(uint64(e.{{.Name}}))
	{{ end }}
	
	{{ define "UintAppend" }}
		data = φλenc.AppendTag(data, φλenc.WTVarInt, {{.Index}})
		data = φλenc.AppendVarUint(data, uint64(e.{{.Name}}))
	{{ end }}
	
	
	{{ define "StringSize" }}
		size += φλenc.SizeTag(φλenc.WTLength, {{.Index}})
		size += φλenc.SizeString(e.{{.Name}})
	{{ end }}
	
	{{ define "StringAppend" }}
		data = φλenc.AppendTag(data, φλenc.WTLength, {{.Index}})
		data = φλenc.AppendString(data, e.{{.Name}})
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
