package predefs

import (
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/dream/path"
	"strings"
)

func ToTitle(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[0:1]) + name[1:]
}

func PathToStructName(path path.TypePath) string {
	if path.IsEmpty() {
		panic("cannot convert empty path to string")
	}
	if !path.IsAbsolute {
		panic("cannot convert non-absolute path to string")
	}
	var title []string
	for _, part := range path.Segments {
		title = append(title, ToTitle(part))
	}
	return strings.Join(title, "")
}

type GlobalProcedureInfo struct {
	Name  string
	GoRef string
}

type TypeDefiner interface {
	Exists(typePath path.TypePath) bool
	ParentOf(typePath path.TypePath) path.TypePath
	Ref(typePath path.TypePath, skipOverrides bool) string
	ResolveField(typePath path.TypePath, shortName string) (definingStruct string, longName string, goType gotype.GoType, found bool)
	ResolveGlobalProcedure(name string) (GlobalProcedureInfo, bool)
}

type TypeInfo struct {
	Path    string
	Package string
	Parent  string
}

func (ti TypeInfo) StructName() string {
	return PathToStructName(path.ConstTypePath(ti.Path))
}

func (ti TypeInfo) Ref() string {
	return ti.Package + ".I" + ti.StructName()
}

type FieldInfo struct {
	ShortName string
	LongName  string
	DefPath   string
	GoType    gotype.GoType
}

var platformDefs = []TypeInfo{
	{"/datum", "datum", ""},
	{"/atom", "platform", "/datum"},
	{"/atom/movable", "platform", "/atom"},
	{"/area", "platform", "/atom"},
	{"/turf", "platform", "/atom"},
	{"/obj", "platform", "/atom/movable"},
	{"/mob", "platform", "/atom/movable"},
	{"/sound", "platform", "/datum"},
	{"/client", "platform", "/datum"},
}

var platformFields = []FieldInfo{
	{"name", "Appearance.Name", "/atom", gotype.String()},
	{"icon", "Appearance.Icon", "/atom", gotype.External("*icon.Icon")},
	{"desc", "Appearance.Desc", "/atom", gotype.String()},
	{"density", "Density", "/atom", gotype.Bool()},
	{"opacity", "Opacity", "/atom", gotype.Bool()},
}

var platformGlobalProcs = []GlobalProcedureInfo{
	{"ismob", "platform.IsMob"},
}

type platformDefiner struct {
}

var PlatformDefiner TypeDefiner = &platformDefiner{}

func (p platformDefiner) GetTypeInfo(typePath path.TypePath) *TypeInfo {
	for _, ent := range platformDefs {
		if ent.Path == typePath.String() {
			return &ent
		}
	}
	return nil
}

func (p platformDefiner) Exists(typePath path.TypePath) bool {
	return p.GetTypeInfo(typePath) != nil
}

func (p platformDefiner) ParentOf(typePath path.TypePath) path.TypePath {
	return path.ConstTypePath(p.GetTypeInfo(typePath).Parent)
}

func (p platformDefiner) Ref(typePath path.TypePath, skipOverrides bool) string {
	return p.GetTypeInfo(typePath).Ref()
}

func (p platformDefiner) StructName(typePath path.TypePath) string {
	return p.GetTypeInfo(typePath).StructName()
}

func (p platformDefiner) ResolveField(typePath path.TypePath, shortName string) (definingStruct string, longName string, goType gotype.GoType, found bool) {
	for _, field := range platformFields {
		if field.DefPath == typePath.String() && shortName == field.ShortName {
			return p.StructName(typePath), field.LongName, field.GoType, true
		}
	}
	parentPath := p.ParentOf(typePath)
	if parentPath.IsEmpty() {
		return "", "", gotype.None(), false
	}
	return p.ResolveField(parentPath, shortName)
}

func (p platformDefiner) ResolveGlobalProcedure(name string) (GlobalProcedureInfo, bool) {
	for _, proc := range platformGlobalProcs {
		if proc.Name == name {
			return proc, true
		}
	}
	return GlobalProcedureInfo{}, false
}
