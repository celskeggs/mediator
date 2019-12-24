package gen

import (
	"bytes"
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/autocoder/predefs"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/hashicorp/go-multierror"
	"go/format"
	"io"
	"io/ioutil"
	"sort"
	"text/template"
)

type DefinedField struct {
	Name string
	Type dtype.DType
}

func (d DefinedField) LongName() string {
	return predefs.ToTitle(d.Name)
}

type DefinedInit struct {
	Name      string
	Value     string
	SourceLoc tokenizer.SourceLocation
}

type DefinedImpl struct {
	Name     string
	This     string
	Usr      string
	Params   []string
	Body     string
	Settings types.ProcSettings
}

type DefinedProc struct {
	Name string
}

type DefinedType struct {
	TypePath path.TypePath
	BasePath path.TypePath

	Fields []DefinedField
	Procs  []DefinedProc
	Impls  []DefinedImpl
	Inits  []DefinedInit

	Verbs []string

	context *DefinedTree
}

// collects additional information required for actually setting the initialized fields
func (d DefinedType) addContext(dt *DefinedTree) (DefinedType, error) {
	dPtr := &d
	dPtr.context = dt
	for _, orig := range dPtr.Inits {
		var found bool
		_, found = dt.ResolveField(d.TypePath, orig.Name)
		if !found {
			return DefinedType{}, fmt.Errorf("no such field %s on %s at %v", orig.Name, d.TypePath, orig.SourceLoc)
		}
	}
	return d, nil
}

func (d *DefinedType) IsDefined() bool {
	// defined unless it's a null override
	return !d.IsOverride() || len(d.Inits) > 0 || len(d.Fields) > 0 || len(d.Impls) > 0
}

func (d *DefinedType) IsOverride() bool {
	return d.TypePath.Equals(d.ParentPath())
}

func (d *DefinedType) DataStructName() string {
	if d.IsOverride() {
		return "Ext" + predefs.PathToDataStructName(d.TypePath)
	} else {
		return predefs.PathToDataStructName(d.TypePath)
	}
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
	Package       string
	PackageImport string
	Types         []DefinedType
	WorldName     string
	WorldMob      path.TypePath
	Imports       []string
	Maps          []string
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

func (t *DefinedTree) ResolveFieldExact(typePath path.TypePath, name string) (dType dtype.DType, found bool) {
	defType := t.GetTypeByPath(typePath)
	if defType != nil {
		for _, field := range defType.Fields {
			if field.Name == name {
				return field.Type, true
			}
		}
	}
	return predefs.PlatformDefiner.ResolveFieldExact(typePath, name)
}

func (t *DefinedTree) ResolveField(typePath path.TypePath, name string) (dType dtype.DType, found bool) {
	dType, found = t.ResolveFieldExact(typePath, name)
	if found {
		return dType, true
	}
	parent := t.ParentOf(typePath)
	if parent.IsEmpty() {
		return dtype.None(), false
	}
	return t.ResolveField(parent, name)
}

func (t DefinedTree) ResolveProcedureExact(typePath path.TypePath, name string) (predefs.ProcedureInfo, bool) {
	defType := t.GetTypeByPath(typePath)
	if defType != nil {
		for _, proc := range defType.Procs {
			if proc.Name == name {
				return predefs.ProcedureInfo{
					Name:    proc.Name,
					DefPath: defType.TypePath,
				}, true
			}
		}
	}
	return predefs.PlatformDefiner.ResolveProcedureExact(typePath, name)
}

func (t DefinedTree) ResolveProcedure(typePath path.TypePath, name string) (predefs.ProcedureInfo, bool) {
	info, found := t.ResolveProcedureExact(typePath, name)
	if found {
		return info, true
	}
	parent := t.ParentOf(typePath)
	if parent.IsEmpty() {
		return predefs.ProcedureInfo{}, false
	}
	return t.ResolveProcedure(parent, name)
}

func (t DefinedTree) GlobalProcedureExists(name string) bool {
	return predefs.PlatformDefiner.GlobalProcedureExists(name)
}

func (t *DefinedTree) GetTypeByPath(path path.TypePath) *DefinedType {
	for i, dType := range t.Types {
		if dType.TypePath.Equals(path) {
			return &t.Types[i]
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
	"github.com/celskeggs/mediator/platform/framework",
	"github.com/celskeggs/mediator/platform/types",
	"github.com/celskeggs/mediator/platform/world",
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
	buf := bytes.NewBuffer(nil)
	err = Generate(tree, buf)
	if err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// make sure we've written *something* for the sake of debugging
		err2 := ioutil.WriteFile(outPath, buf.Bytes(), 0755)
		return multierror.Append(err, err2)
	}
	err = ioutil.WriteFile(outPath, formatted, 0755)
	if err != nil {
		return err
	}
	return nil
}

var templateText = `// Code generated by mediator autocoder; DO NOT EDIT.
package {{.Package}}

import (
{{- range .AllImports}}
	"{{ . }}"
{{- end}}
)

{{- range .Types}}
{{- if .IsDefined}}
{{- $type := .}}

{{if .IsOverride -}}
//mediator:extend {{.DataStructName}} {{.TypePath}}
type {{.DataStructName}} struct {
	{{- range .Fields}}
	Var{{.LongName}} types.Value
	{{- end}}
}
{{- else -}}
//mediator:declare {{.DataStructName}} {{.TypePath}} {{.ParentPath}}
type {{.DataStructName}} struct {
	{{- range .Fields}}
	Var{{.LongName}} types.Value
	{{- end}}
}
{{- end}}

func New{{.DataStructName}}(src *types.Datum, _ *{{.DataStructName}}, _ ...types.Value) {
	{{- range .Inits}}
	src.SetVar("{{.Name}}", {{.Value}})
	{{- end}}
	{{- range .Verbs}}
	src.SetVar("verbs", src.Var("verbs").Invoke(nil, "+", atoms.NewVerb({{printf "%q, %q, %q" . $type.TypePath .}})))
	{{- end}}
}

{{range .Impls -}}
func (*{{$type.DataStructName}}) Proc{{.Name}}({{.This}} *types.Datum, {{.Usr}} *types.Datum{{range .Params}}, {{.}} types.Value{{end}}) types.Value {
{{.Body}}
}

{{ if not .Settings.IsZero }}
func (*{{$type.DataStructName}}) SettingsForProc{{.Name}}() types.ProcSettings {
	return types.ProcSettings{
{{- if not .Settings.Src.IsZero }}
		Src: types.SrcSetting{
			Type: types.{{.Settings.Src.Type}},
{{- if .Settings.Src.Dist }}
			Dist: {{.Settings.Src.Dist}},
{{- end }}
{{- if .Settings.Src.In }}
			In: {{.Settings.Src.In}},
{{- end }}
		},
{{- end }}
	}
}
{{end -}}
{{end -}}

{{- end -}}
{{- end}}

func BeforeMap(world *world.World) []string {
	world.Name = "{{.WorldName}}"
	world.Mob = "{{.WorldMob}}"
	return []string{
{{range .Maps -}}
		"{{.}}",
{{end -}}
	}
}

func BuildWorld() *world.World {
	world, _ := framework.BuildWorld(Tree, BeforeMap)
	return world
}

{{if eq .Package "main" -}}
func main() {
	framework.Launch(Tree, BeforeMap)
}
{{end -}}
`
