package main

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		println("Usage: dream <input.dm>")
		os.Exit(1)
	}
	err := tokenizer.DumpTokensFromFile(os.Args[1], os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	err = parser.DumpParsedFile(os.Args[1], os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
