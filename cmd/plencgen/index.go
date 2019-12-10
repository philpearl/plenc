package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func indexPath(packagePath, structName string) string {
	return filepath.Join(packagePath, fmt.Sprintf("%s.φλ", structName))
}

func loadIndex(packagePath, structName string) (*index, error) {
	path := indexPath(packagePath, structName)
	datab, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not read index file. %w", err)
	}

	data := string(datab)
	lines := strings.Split(data, "\n")
	if len(lines) > 0 {
		lines = lines[1:]
	}

	return &index{
		path:    path,
		entries: lines,
	}, nil
}

func (idx *index) save() error {
	data := "Save this file in your Version Control System\n" + strings.Join(idx.entries, "\n")
	return ioutil.WriteFile(idx.path, []byte(data), 0600)
}

type index struct {
	path    string
	entries []string
}

func (idx *index) indexFor(name string) int {
	for i, e := range idx.entries {
		if e == name {
			return i + 1
		}
	}
	idx.entries = append(idx.entries, name)
	return len(idx.entries)
}
