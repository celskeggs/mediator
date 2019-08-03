package main

import (
	"encoding/json"
	"github.com/celskeggs/mediator/dmi"
	"io/ioutil"
	"os"
)

func main() {
	png, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	dmiInfo, err := dmi.ParseDMI(png)
	if err != nil {
		panic(err)
	}
	result, err := json.MarshalIndent(dmiInfo, "", "  ")
	if err != nil {
		panic(err)
	}
	println(string(result))
}
