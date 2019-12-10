package main

import (
	"bytes"
	"go/format"
	"strings"
	"text/template"
)

func createUnmarshaler(typeInfo data) ([]byte, error) {
	var b bytes.Buffer
	if err := decodeTmpl.Execute(&b, typeInfo); err != nil {
		return nil, err
	}

	return format.Source(b.Bytes())
}

var decodeTmplFuncs = template.FuncMap{
	"runTemplate": runDecodeTemplate,
}

func runDecodeTemplate(name string, data interface{}) (string, error) {
	var b strings.Builder
	err := decodeTmpl.ExecuteTemplate(&b, name, data)
	return b.String(), err
}

var decodeTmpl *template.Template

func init() {
	decodeTmpl = template.Must(template.New("decode").Funcs(decodeTmplFuncs).Parse(decodeTemplateText))
}

var decodeTemplateText = `
package {{ .Package }}

import (
	"fmt"
	"time"

	"github.com/philpearl/φλenc"
)

var _ time.Time

{{ define "TimeDecode" }}
	{
		var (
			t φλenc.Time
			s uint64
			err error
		)
		s, n = φλenc.ReadVarUint(data[offset:])
		offset += n
		n, err = t.ΦλUnmarshal(data[offset:offset+int(s)])
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal field %d {{.Name}} ({{.Type}}). %w", index, err)
		}	
		e.{{.Name}} = t.Standard()
	}
{{ end }}

{{ define "TimePtrDecode" }}
	{
		var (
			t φλenc.Time
			s uint64
			err error
		)
		s, n = φλenc.ReadVarUint(data[offset:])
		offset += n
		n, err = t.ΦλUnmarshal(data[offset:offset+int(s)])
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal field %d {{.Name}} ({{.Type}}). %w", index, err)
		}			
		*e.{{.Name}} = t.Standard()
	}
{{ end }}


{{ define "MethodDecode" }}
	s, n := φλenc.ReadVarUint(data[offset:])
	offset += n
	n, err := e.{{.Name}}.ΦλUnmarshal(data[offset:offset+int(s)])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal field %d {{.Name}} ({{.Type}}). %w", index, err)
	}
{{ end }}

{{ define "MethodSliceDecode" }}
	// Slice of method-y things. Good to grow the slice first in case it is large
	l := len(e.{{.Name}})
	if cap(e.{{.Name}}) > l {
		e.{{.Name}} = e.{{.Name}}[:l+1]
	} else {
		// Need to grow slice. What's the best way?!
		e.{{.Name}} = append(e.{{.Name}}, {{.Type}}{})
	}

	// Slice of method-y things
	s, n := φλenc.ReadVarUint(data[offset:])
	offset += n
	n, err := e.{{.Name}}[l].ΦλUnmarshal(data[offset:offset+int(s)])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal field %d {{.Name}} ({{.Type}}). %w", index, err)
	}
{{ end }}

{{ define "MethodPointerSliceDecode" }}
	// Slice of method-y things. Good to grow the slice first in case it is large
	l := len(e.{{.Name}})
	e.{{.Name}} = append(e.{{.Name}}, &{{.Type}}{})

	// Slice of method-y things
	s, n := φλenc.ReadVarUint(data[offset:])
	offset += n
	n, err := e.{{.Name}}[l].ΦλUnmarshal(data[offset:offset+int(s)])
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal field %d {{.Name}} ({{.Type}}). %w", index, err)
	}
{{ end }}

{{ define "IntDecode" }}
	// intdecode
	v, n := φλenc.ReadVarInt(data[offset:])
	e.{{.Name}} = {{.Type}}(v)
{{ end }}

{{ define "UintDecode" }}
	v, n := φλenc.ReadVarUint(data[offset:])
	e.{{.Name}} = {{.Type}}(v)
{{ end }}

{{ define "Float32Decode" }}
	v, n := φλenc.ReadFloat32(data[offset:])
	e.{{.Name}} = {{.Type}}(v)
{{ end }}

{{ define "Float64Decode" }}
	v, n := φλenc.ReadFloat64(data[offset:])
	e.{{.Name}} = {{.Type}}(v)
{{ end }}

{{ define "BoolDecode" }}
	v, n := φλenc.ReadBool(data[offset:])
	e.{{.Name}} = {{.Type}}(v)
{{ end }}

{{ define "StringDecode" }}
	v, n := φλenc.ReadString(data[offset:])
	e.{{.Name}} = {{.Type}}(v)
{{ end }}


func (e *{{.Name}}) ΦλUnmarshal(data []byte) (int, error) {

	var offset int
	for offset < len(data) {
		wt, index, n := φλenc.ReadTag(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		switch index {
{{ range .Fields }}			
		case {{.Index}}:

{{ if .IsPointer }}			
			if e.{{.Name}} == nil {
				e.{{.Name}} = new({{.Type}})
			}
{{ end }}			
			{{ runTemplate .DecodeTemplate . }}

			offset += n
{{ end }}			

		default:
			// Field corresponding to index does not exist
			n, err := φλenc.Skip(data[offset:], wt)
			if err != nil {
				return 0, fmt.Errorf("failed to skip field %d. %w", index, err)
			}
			offset += n
		}
	}

	return offset, nil
}
`
