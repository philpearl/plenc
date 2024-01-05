package plenccodec

import (
	"strconv"
	"time"
)

// Outputter is an interface that a Descriptor uses to turn plenc data to some
// other output
type Outputter interface {
	StartObject()
	EndObject()
	StartArray()
	EndArray()
	NameField(name string)

	Int64(v int64)
	Uint64(v uint64)
	Float64(v float64)
	Float32(v float32)
	String(v string)
	Bool(v bool)
	Time(t time.Time)
	Raw(v string)
}

// JSONOutput converts Descriptor output to JSON.
type JSONOutput struct {
	data    []byte
	depth   int
	inField bool

	stack []stackEntry
}

type state int

const (
	stateValue state = iota
	stateKey
	stateObjValue
)

type stackEntry struct {
	state state
}

func (j *JSONOutput) Done() []byte {
	j.end()
	return j.data
}

func (j *JSONOutput) Reset() {
	j.data = j.data[:0]
	j.depth = 0
	j.inField = false
	j.stack = j.stack[:0]
}

func (j *JSONOutput) prefix() {
	if j.inField {
		j.inField = false
		return
	}
	for i := 0; i < j.depth; i++ {
		j.data = append(j.data, "  "...)
	}
}

func (j *JSONOutput) end() {
	if j.depth == 0 {
		j.data = append(j.data, '\n')
		return
	}
	j.depth--
	j.stack = j.stack[:len(j.stack)-1]
	l := len(j.data)
	if l < 2 {
		return
	}
	if j.data[l-2] == ',' && j.data[l-1] == '\n' {
		j.data[l-2] = '\n'
		j.data = j.data[:l-1]
	}
}

func (j *JSONOutput) punctuate() {
	if len(j.stack) == 0 {
		return
	}
	s := &j.stack[len(j.stack)-1]
	switch s.state {
	case stateKey:
		j.data = append(j.data, ": "...)
		s.state = stateObjValue
	case stateObjValue:
		j.data = append(j.data, ",\n"...)
		s.state = stateKey
	case stateValue:
		j.data = append(j.data, ",\n"...)
	}
}

func (j *JSONOutput) StartObject() {
	j.prefix()
	j.data = append(j.data, "{\n"...)
	j.depth++
	j.stack = append(j.stack, stackEntry{state: stateKey})
}

func (j *JSONOutput) EndObject() {
	j.end()
	j.prefix()
	j.data = append(j.data, '}')
	j.punctuate()
}

func (j *JSONOutput) StartArray() {
	j.prefix()
	j.data = append(j.data, "[\n"...)
	j.stack = append(j.stack, stackEntry{state: stateValue})
	j.depth++
}

func (j *JSONOutput) EndArray() {
	j.end()
	j.prefix()
	j.data = append(j.data, ']')
	j.punctuate()
}

func (j *JSONOutput) NameField(name string) {
	j.prefix()
	j.inField = true
	j.data = j.appendString(j.data, name)
	j.punctuate()
}

func (j *JSONOutput) Int64(v int64) {
	j.prefix()
	j.data = strconv.AppendInt(j.data, v, 10)
	j.punctuate()
}

func (j *JSONOutput) Uint64(v uint64) {
	j.prefix()
	j.data = strconv.AppendUint(j.data, v, 10)
	j.punctuate()
}

func (j *JSONOutput) Float64(v float64) {
	j.prefix()
	j.data = strconv.AppendFloat(j.data, v, 'g', -1, 64)
	j.punctuate()
}

func (j *JSONOutput) Float32(v float32) {
	j.prefix()
	j.data = strconv.AppendFloat(j.data, float64(v), 'g', -1, 64)
	j.punctuate()
}

func (j *JSONOutput) String(v string) {
	j.prefix()
	j.data = j.appendString(j.data, v)
	j.punctuate()
}

func (j *JSONOutput) Raw(v string) {
	j.prefix()
	j.data = append(j.data, v...)
	j.punctuate()
}

func (j *JSONOutput) Bool(v bool) {
	j.prefix()
	j.data = strconv.AppendBool(j.data, v)
	j.punctuate()
}

func (j *JSONOutput) Time(t time.Time) {
	j.prefix()
	j.data = t.AppendFormat(j.data, `"`+time.RFC3339Nano+`"`)
	j.punctuate()
}

const hex = "0123456789abcdef"

func (j *JSONOutput) appendString(data []byte, v string) []byte {
	data = append(data, '"')
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\\', '"':
			data = append(data, '\\', c)
		case '\n':
			data = append(data, '\\', 'n')
		case '\r':
			data = append(data, '\\', 'r')
		case '\t':
			data = append(data, '\\', 't')
		default:
			if c < 32 {
				data = append(data, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xF])
			} else {
				// append in its natural form
				data = append(data, c)
			}
		}
	}
	data = append(data, '"')
	return data
}
