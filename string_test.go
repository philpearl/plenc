package plenc

import (
	"fmt"
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
			c, err := codecForType(reflect.TypeOf(test))
			if err != nil {
				t.Fatal(err)
			}

			l := c.Size(unsafe.Pointer(&test))

			data := c.Append(nil, unsafe.Pointer(&test))

			if len(data) != l {
				t.Errorf("data not expected length. %d %d", l, len(data))
			}

			var out string
			_, err = c.Read(data, unsafe.Pointer(&out))
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

	c, err := codecForType(reflect.TypeOf(v))
	if err != nil {
		t.Fatal(err)
	}

	data := c.Append(nil, unsafe.Pointer(&v))

	fmt.Printf("%X\n", data)
	/* 1B - 27
	4DE78FA3 58E8A7BB 25C6BEC6 BD394A39
	53E885B8 48E6BBA9 C39D6B
	06
	45C595E6 BCA0
	13
	E7BB87E7 9DB1C4A6 37C3B5E5 9296C38A
	E9A884
	18
	74E6B28B E6999BE5 B28AC4B1 C6ADE5AE
	8B59C8AF C2BF7126
	24
	C6B1C3A3 CA9923E8 A883E79D A9E684B4
	C3A96DC4 9736C98C E9829435 E6B19AE9
	9E97C687
	04
	E6A18F26
	23
	3FC896E6 9BBDC8AE C995E7A8 8C2172E5
	9BAEC7AF 5751E78C 92C3B7E9 A3B9E5AB
	974D4A
	*/

	var out []string
	_, err = c.Read(data, unsafe.Pointer(&out))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(v, out); diff != "" {
		t.Fatal(diff)
	}
}
