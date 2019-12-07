package main

import (
	"fmt"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "while removing existing")
	}
	decls, pkg, err := EnumerateDecls("header.go")
	if err != nil {
		return errors.Wrap(err, "while enumerating declarations")
	}
	if len(decls) == 0 {
		return fmt.Errorf("no mediator declarations found")
	}
	tree := NewTreeInfo(pkg)
	for _, decl := range decls {
		if decl.Path != decl.ParentPath {
			println("decl", decl.Package.ImportPath, decl.StructName, decl.Path, decl.ParentPath, strings.Join(decl.Options, ","))
			err := tree.LoadFromDecl(decl)
			if err != nil {
				return errors.Wrapf(err, "while loading from declarations for %s", decl.Path)
			}
		}
	}
	for _, decl := range decls {
		if decl.Path == decl.ParentPath {
			println("extension", decl.Package.ImportPath, decl.StructName, decl.Path)
			err := tree.LoadFromExtension(decl)
			if err != nil {
				return errors.Wrapf(err, "while loading from extensions for %s", decl.Path)
			}
		}
	}
	err = tree.CascadeOptions()
	if err != nil {
		return errors.Wrap(err, "while cascading options")
	}
	err = tree.LoadPackages()
	if err != nil {
		return errors.Wrap(err, "while loading packages")
	}
	impls, err := tree.Encode()
	if err != nil {
		return errors.Wrap(err, "while encoding")
	}
	for _, impl := range impls {
		err := WriteImpl(impl, impl.Filename())
		if err != nil {
			return errors.Wrap(err, "while writing implementations")
		}
	}
	err = WriteTree(tree, "tree.go")
	if err != nil {
		return errors.Wrap(err, "while writing tree.go")
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
