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
			l := SizeString(test)
			b := make([]byte, 0, l)
			b = AppendString(b, test)
			actual, n := ReadString(b)
			if n != l {
				t.Errorf("length %d is not as expected (%d)", n, l)
			}
			if actual != test {
				t.Errorf("result %q is not as expected %q", actual, test)
			}
		})
	}
}

func TestStringSlice(t *testing.T) {
	for i := 0; i < 100000; i++ {
		v := []string{
			"M珣X觻%ƾƽ9J9S腸H滩Ýk",
			"Eŕ漠",
			"织睱Ħ7õ咖Ê騄",
			"t沋晛岊ıƭ宋Yȯ¿q&",
			"Ʊãʙ#訃睩愴émė6Ɍ邔5汚鞗Ƈ",
			"桏&",
			"?Ȗ曽Ȯɕ稌!r囮ǯWQ猒÷飹嫗MJ",
		}

		c, err := codecForType(reflect.TypeOf(v))
		if err != nil {
			t.Fatal(err)
		}

		data := c.Append(nil, unsafe.Pointer(&v))

		var out []string
		_, err = c.Read(data, unsafe.Pointer(&out))
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(v, out); diff != "" {
			t.Fatal(diff)
		}
	}
}
