package gen

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/autocoder/predefs"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

type DefinedField struct {
	Name string
	Type gotype.GoType
}

func (d DefinedField) LongName() string {
	return predefs.ToTitle(d.Name)
}

type DefinedInit struct {
	ShortName string
	Value     string
	SourceLoc tokenizer.SourceLocation

	definingStruct string
	longName       string
	isOverride     bool
}

func (d DefinedInit) IsOverride() bool {
	return d.isOverride
}

func (d DefinedInit) DefiningStruct() string {
	return d.definingStruct
}

func (d DefinedInit) LongName() string {
	return d.longName
}

type DefinedParam struct {
	Name string
	Type string
}

type DefinedFunc struct {
	Name   string
	This   string
	Params []DefinedParam
	Body   string
}

type DefinedType struct {
	TypePath path.TypePath
	BasePath path.TypePath

	Fields []DefinedField
	Funcs  []DefinedFunc
	Inits  []DefinedInit

	context *DefinedTree
}

// collects additional information required for actually setting the initialized fields
func (d DefinedType) addContext(dt *DefinedTree) (DefinedType, error) {
	dPtr := &d
	dPtr.context = dt
	origInits := dPtr.Inits
	dPtr.Inits = make([]DefinedInit, len(origInits))
	copy(dPtr.Inits, origInits)
	for i, orig := range origInits {
		var found bool
		dPtr.Inits[i].definingStruct, dPtr.Inits[i].longName, _, found = dt.ResolveField(d.TypePath, orig.ShortName)
		if !found {
			return DefinedType{}, fmt.Errorf("no such field %s on %s at %v", orig.ShortName, d.TypePath, orig.SourceLoc)
		}
		oType := dt.GetType(dPtr.Inits[i].definingStruct)
		dPtr.Inits[i].isOverride = oType != nil && oType.IsOverride()
	}
	return d, nil
}

func (d *DefinedType) IsDefined() bool {
	return len(d.Fields) > 0 || len(d.Funcs) > 0 || d.IsOverride()
}

func (d *DefinedType) IsOverride() bool {
	return d.TypePath.Equals(d.ParentPath())
}

func (d *DefinedType) StructName() string {
	result := predefs.PathToStructName(d.TypePath)
	if d.IsOverride() {
		result = "Custom" + result
	}
	return result
}

func (d *DefinedType) InterfaceName() string {
	return "I" + d.StructName()
}

func (d *DefinedType) ParentBase() string {
	name := d.ParentName()
	if name[0] != 'I' || !unicode.IsUpper(([]rune)(name)[1]) {
		panic("invalid parent name; not an interface")
	}
	return name[1:]
}

func (d *DefinedType) ParentName() string {
	parts := strings.Split(d.ParentRef(), ".")
	return parts[len(parts)-1]
}

func (d *DefinedType) ParentRef() string {
	return d.context.Ref(d.ParentPath(), true)
}

func (d *DefinedType) RealParentRef() string {
	realParent := d.context.ParentOf(d.TypePath)
	return d.context.Ref(realParent, true)
}

func (d *DefinedType) ParentPath() path.TypePath {
	if !d.BasePath.IsEmpty() {
		return d.BasePath
	} else {
		parent, _, err := d.TypePath.SplitLast()
		if err != nil {
			panic("cannot autocompute parent path for " + d.TypePath.String() + ": " + err.Error())
		}
		if len(parent.Segments) == 0 {
			panic("cannot autocompute parent path for " + d.TypePath.String() + ": root is not a parent")
		}
		return parent
	}
}

type DefinedTree struct {
	Types     []DefinedType
	WorldName string
	WorldMob  path.TypePath
	Imports   []string
}

var _ predefs.TypeDefiner = &DefinedTree{}

func (t DefinedTree) addContext() (*DefinedTree, error) {
	tPtr := &t
	newTypes := make([]DefinedType, len(tPtr.Types))
	for i, ot := range tPtr.Types {
		var err error
		newTypes[i], err = ot.addContext(tPtr)
		if err != nil {
			return nil, err
		}
	}
	tPtr.Types = newTypes
	return tPtr, nil
}

func (t *DefinedTree) AddImport(name string) {
	for _, imp := range t.AllImports() {
		if imp == name {
			return
		}
	}
	t.Imports = append(t.Imports, name)
}

func (t *DefinedTree) Exists(path path.TypePath) bool {
	return predefs.PlatformDefiner.Exists(path) || t.GetTypeByPath(path) != nil
}

func (t *DefinedTree) ParentOf(path path.TypePath) path.TypePath {
	if predefs.PlatformDefiner.Exists(path) {
		return predefs.PlatformDefiner.ParentOf(path)
	}
	return t.GetTypeByPath(path).ParentPath()
}

func (t *DefinedTree) Ref(path path.TypePath, skipOverrides bool) (ref string) {
	if predefs.PlatformDefiner.Exists(path) {
		ref = predefs.PlatformDefiner.Ref(path, skipOverrides)
	}
	if ref == "" || !skipOverrides {
		defType := t.GetTypeByPath(path)
		if defType != nil {
			ref = defType.InterfaceName()
		}
	}
	if ref == "" {
		panic("could not find ref: " + path.String())
	}
	return ref
}

func (t *DefinedTree) ResolveField(typePath path.TypePath, shortName string) (definingStruct string, longName string, goType gotype.GoType, found bool) {
	defType := t.GetTypeByPath(typePath)
	if defType == nil {
		return predefs.PlatformDefiner.ResolveField(typePath, shortName)
	}
	for _, field := range defType.Fields {
		if field.Name == shortName {
			return defType.StructName(), field.LongName(), field.Type, true
		}
	}
	return t.ResolveField(t.ParentOf(typePath), shortName)
}

func (t DefinedTree) ResolveProcedure(typePath path.TypePath, shortName string) (predefs.ProcedureInfo, bool) {
	defType := t.GetTypeByPath(typePath)
	if defType == nil {
		return predefs.PlatformDefiner.ResolveProcedure(typePath, shortName)
	}
	util.FIXME("when we actually have proc declarations, and not just implementations, search them here")
	return t.ResolveProcedure(t.ParentOf(typePath), shortName)
}

func (t DefinedTree) ResolveGlobalProcedure(name string) (predefs.GlobalProcedureInfo, bool) {
	return predefs.PlatformDefiner.ResolveGlobalProcedure(name)
}

func (t *DefinedTree) GetTypeByPath(path path.TypePath) *DefinedType {
	for i, dType := range t.Types {
		if dType.TypePath.Equals(path) {
			return &t.Types[i]
		}
	}
	return nil
}

func (t *DefinedTree) GetType(name string) *DefinedType {
	for _, dType := range t.Types {
		if dType.StructName() == name {
			return &dType
		}
	}
	return nil
}

func (d *DefinedTree) Extends(subpath path.TypePath, superpath path.TypePath) bool {
	for !subpath.IsEmpty() {
		if superpath.Equals(subpath) {
			return true
		}
		subpath = d.ParentOf(subpath)
	}
	return false
}

func containsStr(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

var necessaryImports = []string{
	"github.com/celskeggs/mediator/platform",
	"github.com/celskeggs/mediator/platform/datum",
	"github.com/celskeggs/mediator/platform/icon",
}

func (d *DefinedTree) AllImports() (imports []string) {
	imports = make([]string, len(d.Imports))
	copy(imports, d.Imports)
	for _, necessaryImport := range necessaryImports {
		if !containsStr(necessaryImport, imports) {
			imports = append(imports, necessaryImport)
		}
	}
	sort.Strings(imports)
	return imports
}

var templateFile = template.New("world")

func init() {
	_, err := templateFile.Parse(templateText)
	if err != nil {
		panic("could not parse template: " + err.Error())
	}
}

func Generate(tree *DefinedTree, out io.Writer) error {
	treeC, err := tree.addContext()
	if err != nil {
		return err
	}
	err = templateFile.ExecuteTemplate(out, "world", treeC)
	util.FIXME("make sure that string escaping is done correctly; or at least validation")
	return err
}

func GenerateTo(tree *DefinedTree, outPath string) (err error) {
	target, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		err2 := target.Close()
		if err2 != nil && err == nil {
			err = err2
		}
	}()
	err = Generate(tree, target)
	return err
}

var templateText = `// Code generated by mediator autocoder; DO NOT EDIT.
package main

import (
{{- range .AllImports}}
	"{{ . }}"
{{- end}}
)

type DefinedWorld struct {
	platform.BaseTreeDefiner
}

{{range .Types -}}
{{- if .IsDefined -}}
{{- $type := . -}}
///// ***** {{.StructName}}

type {{.InterfaceName}} interface {
	{{.ParentRef}}
	As{{.StructName}}() *{{.StructName}}
}

type {{.StructName}} struct {
	{{.ParentRef}}
	{{- range .Fields}}
	{{.LongName}} {{.Type}}
	{{- end}}
}

var _ {{.InterfaceName}} = &{{.StructName}}{}

func (d {{.StructName}}) RawClone() datum.IDatum {
	d.{{.ParentName}} = d.{{.ParentName}}.RawClone().({{.ParentRef}})
	return &d
}

func (d *{{.StructName}}) As{{.StructName}}() *{{.StructName}} {
	return d
}

{{if .IsOverride -}}
func (d *{{.StructName}}) NextOverride() (this datum.IDatum, next datum.IDatum) {
	return d, d.{{.ParentName}}
}

func Cast{{.StructName}}(base {{.ParentRef}}) {{.InterfaceName}} {
	datum.AssertConsistent(base)

	var iter datum.IDatum = base
	for {
		cur, next := iter.NextOverride()
		if ca, ok := cur.({{.InterfaceName}}); ok {
			return ca
		}
		iter = next
		if iter == nil {
			panic("type does not implement {{.StructName}}")
		}
	}
}

func (d DefinedWorld) {{.ParentBase}}Template(parent {{.RealParentRef}}) {{.ParentRef}} {
	base := d.BaseTreeDefiner.{{.ParentBase}}Template(parent)
	return &{{.StructName}}{
		{{.ParentName}}: base,
		{{- range .Inits}}
		{{.LongName}}: {{.Value}},
		{{- end}}
	}
}

{{end -}}

{{- range .Funcs -}}
func ({{.This}} *{{$type.StructName}}) {{.Name}}({{range $index, $element := .Params}}{{if ne $index 0}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
{{.Body}}
}

{{end -}}

{{- end -}}
{{- end -}}

func (DefinedWorld) ElaborateTree(tree *datum.TypeTree, icons *icon.IconCache) {
	{{- range .Types -}}
	{{- if not .IsOverride -}}
	{{- $type := . -}}

	{{- if .IsDefined}}
	prototype{{.StructName}} := &{{.StructName}}{
		{{.ParentName}}: tree.DeriveNew("{{.ParentPath}}").({{.ParentRef}}),
		{{- range .Fields}}
		{{.LongName}}: {{.Default}},
		{{- end}}
	}
	{{- else}}
	prototype{{.StructName}} := tree.Derive("{{.ParentPath}}", "{{.TypePath}}").({{.ParentRef}})
	{{- end}}

	{{- range .Inits -}}
	{{- if .IsOverride}}
	Cast{{.DefiningStruct}}(prototype{{$type.StructName}}).As{{.DefiningStruct}}().{{.LongName}} = {{.Value}}
	{{- else}}
	prototype{{$type.StructName}}.As{{.DefiningStruct}}().{{.LongName}} = {{.Value}}
	{{- end -}}
	{{- else}}
	_ = prototype{{$type.StructName}}
	{{- end -}}
{{if .IsDefined}}
	tree.RegisterStruct("{{.TypePath}}", prototype{{.StructName}})
{{- end}}
{{end -}}
{{- end -}}
}

func (DefinedWorld) BeforeMap(world *platform.World) {
	world.Name = "{{.WorldName}}"
	world.Mob = "{{.WorldMob}}"
}

func (d DefinedWorld) Definer() platform.TreeDefiner {
	return d
}
`
