package lib

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"io"
	"github.com/hashicorp/go-multierror"
	"strings"
	"fmt"
	"github.com/celskeggs/mediator/autocoder/indent"
)

type generator struct {
	Package string
	Imports map[string]iface.Package
	Errors  []error
	Writer  *indent.Writer
}

var _ iface.Gen = &generator{}

func Generate(pkg string, ac iface.AutocodeSource, output io.Writer) error {
	gen := &generator{
		Package: pkg,
		Imports: make(map[string]iface.Package),
		Writer:  indent.NewWriter(output),
	}
	imports, err := collectImports(ac)
	if err != nil {
		return err
	}
	for _, path := range imports {
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			return fmt.Errorf("invalid package path: %s", path)
		}
		gen.Imports[path] = packageRef{
			Path:    path,
			Package: parts[len(parts)-1],
		}
	}
	gen.writeHeader()
	ac(gen, &writeSource{G: gen})
	if len(gen.Errors) == 0 {
		return nil
	} else if len(gen.Errors) == 1 {
		return gen.Errors[0]
	} else {
		return multierror.Append(nil, gen.Errors...)
	}
}

func (g *generator) Indent() {
	g.Writer.Indent()
}

func (g *generator) Unindent() {
	g.Writer.Unindent()
}

func (g *generator) Write(format string, args ...interface{}) {
	_, err := fmt.Fprintf(g.Writer, format, args...)
	g.AddError(err)
}

func (g *generator) writeHeader() {
	g.Write("package %s\n\nimport (\n", g.Package)
	for path := range g.Imports {
		g.Write( "\t%s\n", escapeString(path))
	}
	g.Write( ")\n\n")
}

func (g *generator) AddError(err error) {
	if err != nil {
		g.Errors = append(g.Errors, err)
	}
}

func (g *generator) Import(path string) iface.Package {
	imp, found := g.Imports[path]
	if !found {
		panic(fmt.Sprintf(
			"asked for unexpected import: '%s'; this should be impossible," +
				"unless calling the autocoder twice produces different results", path))
	}
	return imp
}

// NOTE: additional expression-related methods on generator are defined in expr.go
