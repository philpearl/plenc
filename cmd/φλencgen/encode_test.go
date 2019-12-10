package main

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {

	typeInfo := data{
		Package: "cheese",
		Name:    "Cheese",
		Fields: []field{
			{
				Name:           "Size",
				Index:          1,
				SizeTemplate:   "IntSize",
				AppendTemplate: "IntAppend",
			},
			{
				Name:           "Eyes",
				Index:          2,
				SizeTemplate:   "MethodSize",
				AppendTemplate: "MethodAppend",
			},
			{
				Name:           "Name",
				Index:          3,
				SizeTemplate:   "StringSize",
				AppendTemplate: "StringAppend",
			},
		},
	}

	file, err := createMarshaler(typeInfo)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(file))
}
