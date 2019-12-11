package main

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/convert"
	"go/build"
	"os"
)

func Autocode() error {
	if len(os.Args) < 3 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: autocoder <input.dm> ... <input.dm> <output.go>")
		os.Exit(1)
	}
	pkg, err := build.Default.ImportDir(".", build.ImportComment)
	if err != nil {
		return err
	}
	err = convert.ConvertFiles(os.Args[1:len(os.Args)-1], os.Args[len(os.Args)-1], pkg.Name)
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
