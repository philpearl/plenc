// plenctag adds plenc tags to your structs
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/fatih/structtag"
)

// config defines how tags should be modified
type config struct {
	write            bool
	excludeJSONMinus bool
	excludeSQLMinus  bool
	excludePrivate   bool

	fset *token.FileSet
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	var cfg config

	flag.BoolVar(&cfg.write, "w", true, "Write result to (source) file instead of stdout")
	flag.BoolVar(&cfg.excludeJSONMinus, "json", false, "Exclude json:\"-\"")
	flag.BoolVar(&cfg.excludeSQLMinus, "sql", true, "Exclude sql:\"-\"")
	flag.BoolVar(&cfg.excludePrivate, "private", true, "Exclude private fields (starting with lower case letter)")

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "no files specified")
		flag.Usage()
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		node, err := cfg.parse(filename)
		if err != nil {
			return err
		}

		rewrittenNode, err := cfg.rewrite(node)
		if err != nil {
			return err
		}

		if err := cfg.format(rewrittenNode, filename); err != nil {
			return err
		}
	}

	return nil
}

func (c *config) parse(filename string) (ast.Node, error) {
	c.fset = token.NewFileSet()
	return parser.ParseFile(c.fset, filename, nil, parser.ParseComments)
}

func (c *config) format(file ast.Node, filename string) error {
	var buf bytes.Buffer
	err := format.Node(&buf, c.fset, file)
	if err != nil {
		return err
	}

	if c.write {
		err = os.WriteFile(filename, buf.Bytes(), 0)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(buf.String())
	}

	return nil
}

// rewrite rewrites the node for structs
func (c *config) rewrite(node ast.Node) (ast.Node, error) {
	var errs rewriteErrors

	recordError := func(f *ast.Field, err error) {
		errs.Append(fmt.Errorf("%s:%d:%d:%s",
			c.fset.Position(f.Pos()).Filename,
			c.fset.Position(f.Pos()).Line,
			c.fset.Position(f.Pos()).Column,
			err))
	}

	rewriteFunc := func(n ast.Node) bool {
		x, ok := n.(*ast.StructType)
		if !ok {
			return true
		}

		// We make two passes through the fields. First we find the maximum existing plenc tag value. In the
		// second pass we add plenc tags starting after this max value and skipping fields that match filters
		var maxPlenc int
		for _, f := range x.Fields.List {
			if f.Tag == nil {
				continue
			}

			pl, err := plencValue(f.Tag.Value)
			if err != nil {
				recordError(f, err)
				continue
			}

			if pl > maxPlenc {
				maxPlenc = pl
			}
		}

		// Now we make updates
		for _, f := range x.Fields.List {
			if c.excludePrivate {
				r, _ := utf8.DecodeRuneInString(f.Names[0].Name)
				if unicode.IsLower(r) {
					continue
				}
			}
			if f.Tag == nil {
				f.Tag = &ast.BasicLit{}
			}

			tags, err := extractTags(f.Tag.Value)
			if err != nil {
				recordError(f, err)
				continue
			}

			if _, err := tags.Get("plenc"); err == nil {
				// This field has a plenc tag, so we can leave it alone
				continue
			}

			// No plenc tag. Either we explicitly exclude it `plenc:"-"`, or we give it a number `plenc:"12"`
			tag := structtag.Tag{Key: "plenc"}
			if c.isExcluded(tags) {
				tag.Name = "-"
			} else {
				maxPlenc++
				tag.Name = strconv.Itoa(maxPlenc)

			}
			tags.Set(&tag)

			f.Tag.Value = quote(tags.String())
		}

		return true
	}

	ast.Inspect(node, rewriteFunc)

	if errs != nil {
		return node, errs
	}

	return node, nil
}

func (c *config) isExcluded(tags *structtag.Tags) bool {
	if c.excludeSQLMinus {
		tag, err := tags.Get("sql")
		if err == nil && tag.Name == "-" {
			return true
		}
	}
	if c.excludeJSONMinus {
		tag, err := tags.Get("json")
		if err == nil && tag.Name == "-" {
			return true
		}
	}
	return false
}

func extractTags(tag string) (*structtag.Tags, error) {
	if tag == "" {
		return &structtag.Tags{}, nil
	}
	var err error
	tag, err = strconv.Unquote(tag)
	if err != nil {
		return nil, fmt.Errorf("could not unquote tags. %w", err)
	}

	return structtag.Parse(tag)
}

func plencValue(tag string) (int, error) {
	tags, err := extractTags(tag)
	if err != nil {
		return 0, err
	}

	if tags == nil {
		return 0, err
	}

	tagg, err := tags.Get("plenc")
	if err != nil {
		// Only error is if it isn't present
		return 0, nil
	}

	if tagg.Name == "-" {
		// explicitly excluded
		return 0, err
	}

	return strconv.Atoi(tagg.Name)
}

func quote(tag string) string {
	return "`" + tag + "`"
}

type rewriteErrors []error

func (r rewriteErrors) Error() string {
	var buf bytes.Buffer
	for _, e := range r {
		buf.WriteString(fmt.Sprintf("%s\n", e.Error()))
	}
	return buf.String()
}

func (r *rewriteErrors) Append(err error) {
	if err == nil {
		return
	}
	*r = append(*r, err)
}
