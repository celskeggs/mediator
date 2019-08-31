package main

import (
	"os"
	"github.com/celskeggs/mediator/autocoder/convert"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: autocoder <input.dm> <output.go>")
	}

	err := convert.ConvertFiles(os.Args[1], os.Args[2], convert.ConvertConfig{
		DefaultCoreResourcesDir: "../resources",
		DefaultIconsDir:         "resources",
		DefaultMap:              "map.dmm",
	})
	if err != nil {
		panic("generate error: " + err.Error())
	}
}
