package main

import (
	"encoding/json"
	"fmt"
	"github.com/celskeggs/mediator/dmi"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: dmitool <icon.dmi> [<icon.dmi> ...]")
		os.Exit(1)
	}
	for _, arg := range os.Args[1:] {
		png, err := ioutil.ReadFile(arg)
		if err != nil {
			panic(err)
		}
		dmiInfo, err := dmi.ParseDMI(png)
		if err != nil {
			panic(err)
		}
		println(arg)
		result, err := json.MarshalIndent(dmiInfo, "", "  ")
		if err != nil {
			panic(err)
		}
		println(string(result))
	}
}
