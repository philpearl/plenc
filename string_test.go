package plenc

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
)

func TestString(t *testing.T) {
	tests := []string{
		"",
		"a",
		"¬πø∆˙©¥å∫˜",
		"this is a string",
		"ɚ&d珿ȨDT葚µ噘",
		"\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00",
		`¶娋搱ĚoÞB@\贞敠ơƸŜ`,
		"\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00\x00\x00\x00\x03\x00\x00\x00",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			c, err := CodecForType(reflect.TypeOf(test))
			if err != nil {
				t.Fatal(err)
			}

			l := c.Size(unsafe.Pointer(&test))

			data := c.Append(nil, unsafe.Pointer(&test))

			if len(data) != l {
				t.Errorf("data not expected length. %d %d", l, len(data))
			}

			var out string
			_, err = c.Read(data, unsafe.Pointer(&out), c.WireType())
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test, out); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestStringSlice(t *testing.T) {
	v := []string{
		"M珣X觻%ƾƽ9J9S腸H滩Ýk",
		"Eŕ漠",
		"织睱Ħ7õ咖Ê騄",
		"t沋晛岊ıƭ宋Yȯ¿q&",
		"Ʊãʙ#訃睩愴émė6Ɍ邔5汚鞗Ƈ",
		"桏&",
		"?Ȗ曽Ȯɕ稌!r囮ǯWQ猒÷飹嫗MJ",
	}

	c, err := CodecForType(reflect.TypeOf(v))
	if err != nil {
		t.Fatal(err)
	}

	data := c.Append(nil, unsafe.Pointer(&v))

	var out []string
	_, err = c.Read(data, unsafe.Pointer(&out), c.WireType())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(v, out); diff != "" {
		t.Fatal(diff)
	}
}
