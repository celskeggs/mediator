package main

import (
	"go/ast"
	"go/build"
	"go/token"
)

type GetterInfo struct {
	FieldName string
	LongName  string
	HasGetter bool
	HasSetter bool
}

type VarInfo struct {
	FieldName       string
	LongName        string
	Type            ast.Expr
	FileSet         *token.FileSet
	DefiningImports []*ast.ImportSpec
}

type ProcInfo struct {
	Name       string
	ParamCount int
}

type SourceInfo struct {
	StructName   string
	Package      string
	PackageShort string

	FoundConstructor bool
	Getters          []*GetterInfo
	Vars             []VarInfo
	Procs            []ProcInfo
}

type TypeInfo struct {
	Path   string
	Parent string

	Sources []*SourceInfo
}

type TreeInfo struct {
	ImplPackage string
	Packages    []*build.Package
	Paths       map[string]*TypeInfo
	PkgNames    map[string]string
}

func NewTreeInfo(pkg string) *TreeInfo {
	return &TreeInfo{
		ImplPackage: pkg,
		Paths:       map[string]*TypeInfo{},
		PkgNames:    map[string]string{},
	}
}
