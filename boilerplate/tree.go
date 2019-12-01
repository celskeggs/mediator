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

type TypeInfo struct {
	Path   string
	Parent string

	StructName string
	Package    string

	FoundConstructor bool
	Getters          []*GetterInfo
	Vars             []VarInfo
	Procs            []ProcInfo
}

type TreeInfo struct {
	Packages []*build.Package
	Paths    map[string]*TypeInfo
}

func NewTreeInfo() *TreeInfo {
	return &TreeInfo{
		Paths: map[string]*TypeInfo{},
	}
}
