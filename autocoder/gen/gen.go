package gen

import (
	"text/template"
	"io"
	"strings"
	"unicode"
	"github.com/celskeggs/mediator/util"
	"os"
	"github.com/celskeggs/mediator/autocoder/predefs"
)

type DefinedField struct {
	Name    string
	Type    string
	Default string
}

func ToTitle(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[0:1]) + name[1:]
}

func (d DefinedField) LongName() string {
	return ToTitle(d.Name)
}

type DefinedInit struct {
	ShortName string
	Value     string

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
	Params []DefinedParam
	Body   string
}

func (d DefinedFunc) Trimmed() string {
	return "\t" + strings.Replace(strings.TrimSpace(d.Body), "\n", "\n\t", -1)
}

type DefinedType struct {
	TypePath string
	BasePath string

	Fields []DefinedField
	Funcs  []DefinedFunc
	Inits  []DefinedInit

	context *DefinedTree
}

func (d DefinedType) addContext(dt *DefinedTree) DefinedType {
	dPtr := &d
	dPtr.context = dt
	origInits := dPtr.Inits
	dPtr.Inits = make([]DefinedInit, len(origInits))
	copy(dPtr.Inits, origInits)
	for i, orig := range origInits {
		dPtr.Inits[i].definingStruct, dPtr.Inits[i].longName, _ = dt.ResolveField(d.TypePath, orig.ShortName)
		oType := dt.GetType(dPtr.Inits[i].definingStruct)
		dPtr.Inits[i].isOverride = oType != nil && oType.IsOverride()
	}
	return d
}

func (d *DefinedType) IsDefined() bool {
	return len(d.Fields) > 0 || len(d.Funcs) > 0 || d.IsOverride()
}

func (d *DefinedType) IsOverride() bool {
	return d.TypePath == d.ParentPath()
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

func (d *DefinedType) ParentPath() string {
	if d.BasePath != "" {
		return d.BasePath
	} else {
		parts := strings.Split(d.TypePath, "/")
		if len(parts) < 3 || parts[0] != "" {
			panic("cannot autocompute parent path for " + d.TypePath)
		}
		return strings.Join(parts[:len(parts)-1], "/")
	}
}

type DefinedTree struct {
	Types     []DefinedType
	WorldName string
	WorldMob  string
	WorldMap  string

	DefaultCoreResourcesDir string
	DefaultIconsDir         string
}

var _ predefs.TypeDefiner = &DefinedTree{}

func (t DefinedTree) addContext() *DefinedTree {
	tPtr := &t
	newTypes := make([]DefinedType, len(tPtr.Types))
	for i, ot := range tPtr.Types {
		newTypes[i] = ot.addContext(tPtr)
	}
	tPtr.Types = newTypes
	return tPtr
}

func (t *DefinedTree) Exists(path string) bool {
	return predefs.PlatformDefiner.Exists(path) || t.GetTypeByPath(path) != nil
}

func (t *DefinedTree) ParentOf(path string) string {
	if predefs.PlatformDefiner.Exists(path) {
		return predefs.PlatformDefiner.ParentOf(path)
	}
	return t.GetTypeByPath(path).ParentPath()
}

func (t *DefinedTree) Ref(path string, skipOverrides bool) (ref string) {
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
		panic("could not find ref: " + path)
	}
	return ref
}

func (t *DefinedTree) ResolveField(typePath string, shortName string) (definingStruct string, longName string, goType string) {
	defType := t.GetTypeByPath(typePath)
	if defType == nil {
		return predefs.PlatformDefiner.ResolveField(typePath, shortName)
	}
	for _, field := range defType.Fields {
		if field.Name == shortName {
			return defType.StructName(), field.LongName(), field.Type
		}
	}
	return t.ResolveField(t.ParentOf(typePath), shortName)
}

func (t *DefinedTree) GetTypeByPath(path string) *DefinedType {
	for i, dType := range t.Types {
		if dType.TypePath == path {
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

func (d *DefinedTree) Extends(subpath string, superpath string) bool {
	for subpath != "" {
		if superpath == subpath {
			return true
		}
		subpath = d.ParentOf(subpath)
	}
	return false
}

var templateFile = template.New("world")

func init() {
	_, err := templateFile.Parse(templateText)
	if err != nil {
		panic("could not parse template: " + err.Error())
	}
}

func Generate(tree *DefinedTree, out io.Writer) error {
	err := templateFile.ExecuteTemplate(out, "world", tree.addContext())
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

var templateText = `// Code generated by mediator; DO NOT EDIT.
package main

import (
	"github.com/celskeggs/mediator/platform"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/framework"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/format"
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

func (d DefinedWorld) {{.ParentBase}}Template(parent platform.IAtom) {{.ParentRef}} {
	base := d.BaseTreeDefiner.{{.ParentBase}}Template(parent)
	return &{{.StructName}}{
		{{.ParentName}}: base,
		{{- range .Fields}}
		{{.LongName}}: {{.Default}},
		{{- end}}
	}
}

{{end -}}

{{- range .Funcs -}}
func (this *{{$type.StructName}}) {{.Name}}({{range $index, $element := .Params}}{{if ne $index 0}}, {{end}}{{.Name}} {{.Type}}{{end}}) {
{{.Trimmed}}
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

func main() {
	framework.Launch(DefinedWorld{}, framework.ResourceDefaults{
		CoreResourcesDir: "{{.DefaultCoreResourcesDir}}",
		IconsDir:         "{{.DefaultIconsDir}}",
		MapPath:          "{{.WorldMap}}",
	})
}
`
