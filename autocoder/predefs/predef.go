package predefs

import (
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/path"
	"strings"
)

func ToTitle(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[0:1]) + name[1:]
}

func PathToDataStructName(path path.TypePath) string {
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
	title = append(title, "Data")
	return strings.Join(title, "")
}

type ProcedureInfo struct {
	Name    string
	DefPath path.TypePath
	IsVerb  bool
}

type TypeDefiner interface {
	Exists(typePath path.TypePath) bool
	ParentOf(typePath path.TypePath) path.TypePath
	ResolveField(typePath path.TypePath, name string) (dtype dtype.DType, found bool)
	GlobalProcedureExists(name string) bool
	ResolveProcedure(typePath path.TypePath, shortName string) (ProcedureInfo, bool)
}

type TypeInfo struct {
	Path    string
	Package string
	Parent  string
}

type FieldInfo struct {
	Name    string
	DefPath string
	Type    dtype.DType
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
	{"name", "/atom", dtype.String()},
	{"icon", "/atom", dtype.ConstPath("/icon")},
	{"desc", "/atom", dtype.String()},
	{"density", "/atom", dtype.Integer()},
	{"opacity", "/atom", dtype.Integer()},
}

var platformProcs = []ProcedureInfo{
	{"Entered", path.ConstTypePath("/atom"), false},
	{"Bump", path.ConstTypePath("/atom/movable"), false},
}

var platformGlobalProcs = []string{
	"ismob",
	"sound",
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

func (p platformDefiner) ResolveField(typePath path.TypePath, name string) (dType dtype.DType, found bool) {
	for _, field := range platformFields {
		if field.DefPath == typePath.String() && name == field.Name {
			return field.Type, true
		}
	}
	parentPath := p.ParentOf(typePath)
	if parentPath.IsEmpty() {
		return dtype.None(), false
	}
	return p.ResolveField(parentPath, name)
}

func (p platformDefiner) ResolveProcedure(typePath path.TypePath, shortName string) (ProcedureInfo, bool) {
	for _, proc := range platformProcs {
		if proc.DefPath.Equals(typePath) && shortName == proc.Name {
			return proc, true
		}
	}
	return ProcedureInfo{}, false
}

func (p platformDefiner) GlobalProcedureExists(name string) bool {
	for _, proc := range platformGlobalProcs {
		if proc == name {
			return true
		}
	}
	return false
}
