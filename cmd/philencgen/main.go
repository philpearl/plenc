package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	var o options
	o.setup()
	flag.Parse()
	o.validate()

	if err := run(&o); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run(o *options) error {
	d, err := parseType(o)
	if err != nil {
		return err
	}

	marshaler, err := createMarshaler(d)
	if err != nil {
		return fmt.Errorf("failed creating Marshaler. %w", err)
	}

	// Output file should be next to definition. Not sure how we do that.
	filename := fmt.Sprintf("φλenc_marshal_%s.go", strings.ToLower(d.Name))
	if err := ioutil.WriteFile(filename, marshaler, 0666); err != nil {
		return fmt.Errorf("failed writing marshaler. %w", err)
	}

	unmarshaler, err := createUnmarshaler(d)
	if err != nil {
		return fmt.Errorf("failed creating Unmarshaler. %w", err)
	}

	// Output file should be next to definition. Not sure how we do that.
	filename = fmt.Sprintf("φλenc_unmarshal_%s.go", strings.ToLower(d.Name))
	if err := ioutil.WriteFile(filename, unmarshaler, 0666); err != nil {
		return fmt.Errorf("failed writing unmarshaler. %w", err)
	}

	return err
}

type options struct {
	// Name of file to create
	outfile string
	// Package to examine
	path       string
	structName string
}

func (o *options) setup() {
	flag.StringVar(&o.outfile, "out", "", "Output file")
	flag.StringVar(&o.path, "pkg", "", "Package name to look at")
	flag.StringVar(&o.structName, "type", "", "Struct type name to process")
}

func (o *options) validate() bool {
	if o.outfile == "" {
		fmt.Fprintf(os.Stderr, "You must specify an output file\n")
		flag.Usage()
		os.Exit(1)
	}

	if o.path == "" {
		fmt.Fprintf(os.Stderr, "You must specify a package to look at\n")
		flag.Usage()
		os.Exit(1)
	}

	if o.structName == "" {
		fmt.Fprintf(os.Stderr, "You must specify a structure to process\n")
		flag.Usage()
		os.Exit(1)
	}

	return true
}
