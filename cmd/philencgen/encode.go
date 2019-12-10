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
		"github.com/philpearl/philenc"
	)
	
	// TODO: missing types
	// slice of numeric ()
	// slice of other
	// pointers
	// TODO: option whether top-level type is a pointer for marshaler
	
	{{ define "TimeSize" }}
		{
			var t philenc.Time
			t.Set(e.{{.Name}})
			if s := t.ΦλSize(); s != 0 {
				size += philenc.SizeTag(philenc.WTLength, {{.Index}})
				size += philenc.SizeVarUint(uint64(s))
				size += s		
			}
	
		}
	{{ end }}
	
	{{ define "TimeAppend" }}
	{
		var t philenc.Time
		t.Set(e.{{.Name}})
		if 	s := t.ΦλSize(); s != 0 {
			data = philenc.AppendTag(data, philenc.WTLength, {{.Index}})
			data = philenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)	
		}
	}
	{{ end }}

	{{ define "TimePtrSize" }}
		if e.{{.Name}} != nil {
			var t philenc.Time
			t.Set(*e.{{.Name}})
			if s := t.ΦλSize(); s != 0 {
				size += philenc.SizeTag(philenc.WTLength, {{.Index}})
				size += philenc.SizeVarUint(uint64(s))
				size += s		
			}
	
		}
	{{ end }}
	
	{{ define "TimePtrAppend" }}
	if e.{{.Name}} != nil {
		var t philenc.Time
		t.Set(*e.{{.Name}})
		if 	s := t.ΦλSize(); s != 0 {
			data = philenc.AppendTag(data, philenc.WTLength, {{.Index}})
			data = philenc.AppendVarUint(data, uint64(s))
			data = t.ΦλAppend(data)	
		}
	}
	{{ end }}


	{{ define "MethodSize" }}
		if s := e.{{.Name}}.ΦλSize(); s != 0 {
			size += philenc.SizeTag(philenc.WTLength, {{.Index}})
			size += philenc.SizeVarUint(uint64(s))
			size += s		
		}
	{{ end }}
	
	{{ define "MethodAppend" }}
		if 	s := e.{{.Name}}.ΦλSize(); s != 0 {
			data = philenc.AppendTag(data, philenc.WTLength, {{.Index}})
			data = philenc.AppendVarUint(data, uint64(s))
			data = e.{{.Name}}.ΦλAppend(data)	
		}
	{{ end }}
	
	{{ define "MethodSliceSize" }}
		// Each element of the slice is separately encoded
		for i := range e.{{.Name}} {
			if s := e.{{.Name}}[i].ΦλSize(); s != 0 {
				size += philenc.SizeTag(philenc.WTLength, {{.Index}})
				size += philenc.SizeVarUint(uint64(s))
				size += s	
			}
		}
	{{ end }}
	
	{{ define "MethodSliceAppend" }}
		// Each element of the slice is separately encoded
		for i := range e.{{.Name}} {
			if s := e.{{.Name}}[i].ΦλSize(); s != 0 {
				data = philenc.AppendTag(data, philenc.WTLength, {{.Index}})
				data = philenc.AppendVarUint(data, uint64(s))
				data = e.{{.Name}}[i].ΦλAppend(data)		
			}
		}
	{{ end }}
	
	{{ define "BoolSize" }}
		size += philenc.SizeTag(philenc.WTVarInt, {{.Index}})
		size += philenc.SizeBool(e.{{.Name}})
	{{ end }}
	
	{{ define "BoolAppend" }}
		data = philenc.AppendTag(data, philenc.WTVarInt, {{.Index}})
		data = philenc.AppendBool(data, e.{{.Name}})
	{{ end }}
	
	{{ define "Float32Size" }}
		size += philenc.SizeTag(philenc.WT32, {{.Index}})
		size += philenc.SizeFloat32(float32(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float32Append" }}
		data = philenc.AppendTag(data, philenc.WT32, {{.Index}})
		data = philenc.AppendFloat32(data, float32(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float64Size" }}
		size += philenc.SizeTag(philenc.WT64, {{.Index}})
		size += philenc.SizeFloat64(float64(e.{{.Name}}))
	{{ end }}
	
	{{ define "Float64Append" }}
		data = philenc.AppendTag(data, philenc.WT64, {{.Index}})
		data = philenc.AppendFloat64(data, float64(e.{{.Name}}))
	{{ end }}
	
	{{ define "IntSize" }}
		size += philenc.SizeTag(philenc.WTVarInt, {{.Index}})
		size += philenc.SizeVarInt(int64(e.{{.Name}}))
	{{ end }}
	
	{{ define "IntAppend" }}
		data = philenc.AppendTag(data, philenc.WTVarInt, {{.Index}})
		data = philenc.AppendVarInt(data, int64(e.{{.Name}}))
	{{ end }}
	
	{{ define "UintSize" }}
		size += philenc.SizeTag(philenc.WTVarInt, {{.Index}})
		size += philenc.SizeVarUint(uint64(e.{{.Name}}))
	{{ end }}
	
	{{ define "UintAppend" }}
		data = philenc.AppendTag(data, philenc.WTVarInt, {{.Index}})
		data = philenc.AppendVarUint(data, uint64(e.{{.Name}}))
	{{ end }}
	
	
	{{ define "StringSize" }}
		size += philenc.SizeTag(philenc.WTLength, {{.Index}})
		size += philenc.SizeString(e.{{.Name}})
	{{ end }}
	
	{{ define "StringAppend" }}
		data = philenc.AppendTag(data, philenc.WTLength, {{.Index}})
		data = philenc.AppendString(data, e.{{.Name}})
	{{ end }}
	
	
	func (e *{{ .Name }}) ΦλSize() (size int) {
		if e == nil {
			return 0
		}
	
	{{ range .Fields }}
		{{ runTemplate .SizeTemplate . }}
	{{ end }}
	
		return size
	}
	
	func (e *{{ .Name }}) ΦλAppend(data []byte) []byte {
	
	{{ range .Fields }}
		{{ runTemplate .AppendTemplate . }}
	{{ end }}
	
		return data
	}
	`))

}
