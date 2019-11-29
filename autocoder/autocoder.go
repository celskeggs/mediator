package main

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/convert"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		panic("usage: autocoder <input.dm> ... <input.dm> <output.go>")
	}
	err := convert.ConvertFiles(os.Args[1:len(os.Args)-1], os.Args[len(os.Args)-1])
	if err != nil {
		fmt.Println("generate error:", err)
	}
}
