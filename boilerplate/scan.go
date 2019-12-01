package main

import (
	"errors"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"path"
	"strings"
)

type Decl struct {
	Package    *build.Package
	StructName string
	Path       string
	ParentPath string
}

func Unquote(name string) string {
	if !strings.HasPrefix(name, "\"") || !strings.HasSuffix(name, "\"") {
		panic("expected quoted name")
	}
	return name[1 : len(name)-1]
}

func EnumeratePackages(filename string) ([]string, error) {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var imports []string
	for _, i := range ast.Imports {
		imports = append(imports, Unquote(i.Path.Value))
	}
	return imports, nil
}

func ScanDeclsInFile(filename string, pkg *build.Package) ([]Decl, error) {
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var decls []Decl
	for _, commentGroup := range ast.Comments {
		for _, comment := range commentGroup.List {
			text := strings.TrimSpace(strings.TrimLeft(comment.Text, "/*"))
			if strings.HasPrefix(text, "mediator:declare ") {
				parts := strings.Split(text, " ")
				if len(parts) != 4 {
					return nil, fmt.Errorf("mediator:declare does not have exactly three arguments in %q", text)
				}
				decls = append(decls, Decl{
					Package:    pkg,
					StructName: parts[1],
					Path:       parts[2],
					ParentPath: parts[3],
				})
			}
		}
	}
	return decls, nil
}

func ScanDeclsInPackage(pkgname string) ([]Decl, error) {
	pkg, err := build.Default.Import(pkgname, "", 0)
	if err != nil {
		return nil, err
	}
	if len(pkg.CgoFiles) > 0 {
		return nil, errors.New("nonzero number of cgo files")
	}
	if len(pkg.IgnoredGoFiles) > 0 {
		return nil, errors.New("nonzero number of ignored go files")
	}
	if len(pkg.InvalidGoFiles) > 0 {
		return nil, errors.New("nonzero number of invalid go files")
	}
	var allDecls []Decl
	for _, filename := range pkg.GoFiles {
		decls, err := ScanDeclsInFile(path.Join(pkg.Dir, filename), pkg)
		if err != nil {
			return nil, err
		}
		allDecls = append(allDecls, decls...)
	}
	return allDecls, nil
}

func ScanDecls(packages []string) ([]Decl, error) {
	var allDecls []Decl
	for _, pkg := range packages {
		decls, err := ScanDeclsInPackage(pkg)
		if err != nil {
			return nil, err
		}
		allDecls = append(allDecls, decls...)
	}
	return allDecls, nil
}

func EnumerateDecls(filename string) ([]Decl, error) {
	packages, err := EnumeratePackages("header.go")
	if err != nil {
		return nil, err
	}
	decls, err := ScanDecls(packages)
	if err != nil {
		return nil, err
	}
	return decls, nil
}
