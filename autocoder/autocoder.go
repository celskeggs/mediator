package main

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/convert"
	"go/build"
	"os"
)

const DeclGoName = "gen_decl.go"
const ResourcePackName = "resource_pack.tgz"

func Autocode() error {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: autocoder <project.dme>\n")
		os.Exit(1)
	}
	pkg, err := build.Default.ImportDir(".", build.ImportComment)
	if err != nil {
		return err
	}
	err = convert.ConvertFiles([]string{os.Args[1]}, DeclGoName, ResourcePackName, pkg.Name)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := Autocode()
	if err != nil {
		fmt.Println("generate error:", err)
		os.Exit(1)
	}
}
