package predefs

import (
	"strings"
)

func PathToStructName(path string) string {
	if path == "" || path[0] != '/' {
		panic("invalid type path: " + path)
	}
	parts := strings.Split(path[1:], "/")
	title := make([]string, len(parts))
	for i, part := range parts {
		if part == "" {
			panic("invalid type path: " + path)
		}
		title[i] = strings.ToUpper(part[0:1]) + part[1:]
	}
	return strings.Join(title, "")
}

type TypeDefiner interface {
	Exists(typePath string) bool
	ParentOf(typePath string) string
	Ref(typePath string, skipOverrides bool) string
	ResolveField(typePath string, shortName string) (definingStruct string, longName string, goType string)
}

type TypeInfo struct {
	Path    string
	Package string
	Parent  string
}

func (ti TypeInfo) StructName() string {
	return PathToStructName(ti.Path)
}

func (ti TypeInfo) Ref() string {
	return ti.Package + ".I" + ti.StructName()
}

type FieldInfo struct {
	ShortName string
	LongName  string
	DefPath   string
	GoType    string
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
	{"name", "Appearance.Name", "/atom", "string"},
	{"icon", "Appearance.Icon", "/atom", "icon.Icon"},
	{"desc", "Appearance.Desc", "/atom", "string"},
	{"density", "Density", "/atom", "bool"},
	{"opacity", "Opacity", "/atom", "bool"},
}

type platformDefiner struct {
}

var PlatformDefiner TypeDefiner = &platformDefiner{}

func (p platformDefiner) GetTypeInfo(typePath string) *TypeInfo {
	for _, ent := range platformDefs {
		if ent.Path == typePath {
			return &ent
		}
	}
	return nil
}

func (p platformDefiner) Exists(typePath string) bool {
	return p.GetTypeInfo(typePath) != nil
}

func (p platformDefiner) ParentOf(typePath string) string {
	return p.GetTypeInfo(typePath).Parent
}

func (p platformDefiner) Ref(typePath string, skipOverrides bool) string {
	return p.GetTypeInfo(typePath).Ref()
}

func (p platformDefiner) StructName(typePath string) string {
	return p.GetTypeInfo(typePath).StructName()
}

func (p platformDefiner) ResolveField(typePath string, shortName string) (definingStruct string, longName string, goType string) {
	for _, field := range platformFields {
		if field.DefPath == typePath && shortName == field.ShortName {
			return p.StructName(field.DefPath), field.LongName, field.GoType
		}
	}
	parentPath := p.ParentOf(typePath)
	if parentPath == "" {
		panic("no such field " + shortName)
	}
	return p.ResolveField(parentPath, shortName)
}
