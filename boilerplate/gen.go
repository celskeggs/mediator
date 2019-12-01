package main

import (
	"bytes"
	"github.com/hashicorp/go-multierror"
	"go/format"
	"io/ioutil"
	"strings"
	"text/template"
)

type PreparedChunk struct {
	Package      string
	PackageShort string
	StructName   string
	Vars         []VarInfo
}

type PreparedVar struct {
	VarInfo
	StructName   string
	PackageShort string
}

type PreparedGetter struct {
	GetterInfo
	StructName string
}

type PreparedProc struct {
	ProcInfo
	StructName string
}

type PreparedImplementation struct {
	Imports    []string
	Chunks     []PreparedChunk
	Vars       []PreparedVar
	Getters    []PreparedGetter
	Procs      []PreparedProc
	TypePath   string
	Type       string
	ParentPath string
}

const GenFilenamePrefix = "impl_"
const GenFilenameSuffix = ".go"

func (i *PreparedImplementation) Filename() string {
	if len(i.TypePath) < 2 || i.TypePath[0] != '/' {
		panic("invalid type path")
	}
	return GenFilenamePrefix + strings.ReplaceAll(i.TypePath[1:], "/", "_") + GenFilenameSuffix
}

var templateFile = template.New("world")

func init() {
	_, err := templateFile.Parse(implTemplate)
	if err != nil {
		panic("could not parse template: " + err.Error())
	}
}

func WriteImpl(impl *PreparedImplementation, filename string) error {
	buf := bytes.NewBuffer(nil)
	err := templateFile.ExecuteTemplate(buf, "world", impl)
	if err != nil {
		return err
	}
	out, err := format.Source(buf.Bytes())
	if err != nil {
		// write anyway for the sake of debugging
		err2 := ioutil.WriteFile(filename, buf.Bytes(), 0644)
		return multierror.Append(err, err2)
	}
	err = ioutil.WriteFile(filename, out, 0644)
	if err != nil {
		return err
	}
	return nil
}

var implTemplate = `// Code generated by mediator boilerplate; DO NOT EDIT.
package impl

import (
{{- range .Imports}}
	"{{ . }}"
{{- end}}
)

type {{.Type}}Impl struct {
{{- range .Chunks}}
	{{.PackageShort}}.{{.StructName}}
{{- end}}
}

func New{{.Type}}(params ...types.Value) types.DatumImpl {
	return &{{.Type}}Impl{
{{- range .Chunks}}
		{{.StructName}}: {{.PackageShort}}.New{{.StructName}}(params...),
{{- end}}
	}
}

func (t *{{.Type}}Impl) Type() types.TypePath {
	return "{{.TypePath}}"
}

func (t *{{.Type}}Impl) Var(src *types.Datum, name string) (types.Value, bool) {
	switch name {
	case "type":
		return types.TypePath("{{.TypePath}}"), true
	case "parent_type":
{{- if .ParentPath}}
		return types.TypePath("{{.ParentPath}}"), true
{{- else}}
		return nil, true
{{- end}}
{{- range .Vars}}
	{{- $conv := .ConvertTo }}
	case "{{.FieldName}}":
		return {{index $conv 0}}t.{{.StructName}}.{{.LongName}}{{index $conv 1}}, true
{{- end}}
{{- range .Getters}}
	case "{{.FieldName}}":
		return t.{{.StructName}}.Get{{.LongName}}(src), true
{{- end}}
	default:
		return nil, false
	}
}

func (t *{{.Type}}Impl) SetVar(src *types.Datum, name string, value types.Value) types.SetResult {
	switch name {
	case "type":
		return types.SetResultReadOnly
	case "parent_type":
		return types.SetResultReadOnly
{{- range .Vars}}
	{{- $conv := .ConvertFrom }}
	case "{{.FieldName}}":
		t.{{.StructName}}.{{.LongName}} = {{index $conv 0}}value{{index $conv 1}}
		return types.SetResultOk
{{- end}}
{{- range .Getters}}
	case "{{.FieldName}}":
{{- if .HasSetter}}
		t.{{.StructName}}.Set{{.LongName}}(src, value)
		return types.SetResultOk
{{- else}}
		return types.SetResultReadOnly
{{- end}}
{{- end}}
	default:
		return types.SetResultNonexistent
	}
}

func (t *{{.Type}}Impl) Proc(src *types.Datum, name string, params ...types.Value) (types.Value, bool) {
	switch name {
{{- range .Procs}}
	case "{{.Name}}":
		return t.{{.StructName}}.{{.ProcName}}(src{{range .ParamNums}}, types.Param(params, {{ . }}){{end}}), true
{{- end}}
	default:
		return nil, false
	}
}

func (t *{{.Type}}Impl) Chunk(ref string) interface{} {
	switch ref {
{{- range .Chunks}}
	case "{{.Package}}.{{.StructName}}":
		return &t.{{.StructName}}
{{- end}}
	default:
		return nil
	}
}
`
