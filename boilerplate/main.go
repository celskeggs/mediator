package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func shouldRemove(name string) bool {
	return name == "tree.go" || (strings.HasPrefix(name, GenFilenamePrefix) && strings.HasSuffix(name, GenFilenameSuffix))
}

func RemoveExisting() error {
	existing, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
	for _, fi := range existing {
		if shouldRemove(fi.Name()) && !fi.IsDir() {
			err := os.Remove(fi.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Generate() error {
	err := RemoveExisting()
	if err != nil {
		return err
	}
	decls, err := EnumerateDecls("header.go")
	if err != nil {
		return err
	}
	if len(decls) == 0 {
		return fmt.Errorf("no mediator declarations found")
	}
	tree := NewTreeInfo()
	for _, decl := range decls {
		println("decl", decl.Package.ImportPath, decl.StructName, decl.Path, decl.ParentPath)
		err := tree.LoadFromDecl(decl)
		if err != nil {
			return err
		}
	}
	err = tree.LoadPackages()
	if err != nil {
		return err
	}
	impls, err := tree.Encode()
	if err != nil {
		return err
	}
	for _, impl := range impls {
		err := WriteImpl(impl, impl.Filename())
		if err != nil {
			return err
		}
	}
	err = WriteTree(tree, "tree.go")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) != 1 {
		fmt.Printf("usage: boilerplate\n")
		os.Exit(1)
	}
	if err := Generate(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
