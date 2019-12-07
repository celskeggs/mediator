package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"text/template"
)

var treeTemplateFile = template.New("world")

func init() {
	_, err := treeTemplateFile.Parse(treeTemplate)
	if err != nil {
		panic("could not parse template: " + err.Error())
	}
}

func WriteTree(tree *TreeInfo, filename string) error {
	buf := bytes.NewBuffer(nil)
	err := treeTemplateFile.ExecuteTemplate(buf, "world", tree)
	if err != nil {
		return err
	}
	out, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, out, 0644)
	if err != nil {
		return err
	}
	return nil
}

var treeTemplate = `// Code generated by mediator boilerplate; DO NOT EDIT.
package {{.ImplPackage}}

import (
	"github.com/celskeggs/mediator/platform/types"
)

type tree struct {}

type treeSingletons struct {
{{- range $path, $type := .Paths }}
{{- if $type.Singleton }}
	{{$type.Type}} *types.Datum
{{- end }}
{{- end }}
}

var Tree types.TypeTree = tree{}

func (tree) PopulateRealm(realm *types.Realm) {
    realm.TreePrivateState = &treeSingletons{
{{- range $path, $type := .Paths }}
{{- if $type.Singleton }}
		{{$type.Type}}: New{{$type.Type}}(realm),
{{- end }}
{{- end }}
	}
}

func (tree) Parent(path types.TypePath) types.TypePath {
	switch path {
{{- range $path, $type := .Paths }}
	case "{{$path}}":
		return "{{if ne $type.Parent "/"}}{{$type.Parent}}{{ end }}"
{{- end }}
	default:
		panic("unknown type " + path.String())
	}
}

func (tree) New(realm *types.Realm, path types.TypePath, params ...types.Value) *types.Datum {
	switch path {
{{- range $path, $type := .Paths }}
{{- if $type.Singleton }}
	case "{{$path}}":
		return realm.TreePrivateState.(*treeSingletons).{{$type.Type}}
{{- else }}
	case "{{$path}}":
		return New{{$type.Type}}(realm, params...)
{{- end }}
{{- end }}
	default:
		panic("unknown type " + path.String())
	}
}
`
