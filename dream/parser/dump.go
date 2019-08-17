package parser

import (
	"fmt"
	"io"
)

func (dmd DreamMakerDefinition) Dump(output io.Writer) error {
	_, err := fmt.Fprintf(output, "[definition %v]\n", dmd.Type)
	if err != nil {
		return err
	}
	if !dmd.Path.IsEmpty() {
		_, err := fmt.Fprintf(output, "\tpath = %v\n", dmd.Path)
		if err != nil {
			return err
		}
	}
	if dmd.Variable != "" {
		_, err := fmt.Fprintf(output, "\tvariable = %s\n", dmd.Variable)
		if err != nil {
			return err
		}
	}
	if !dmd.Expression.IsNone() {
		_, err := fmt.Fprintf(output, "\texpression = %v\n", dmd.Expression.String())
		if err != nil {
			return err
		}
	}
	return nil
}

func (dmf *DreamMakerFile) Dump(output io.Writer) error {
	_, err := fmt.Fprintln(output, "[beginning of parser dump]")
	if err != nil {
		return err
	}
	for _, def := range dmf.Definitions {
		err := def.Dump(output)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(output, "[end of parser dump]")
	if err != nil {
		return err
	}
	return nil
}

func DumpParsedFile(filename string, output io.Writer) error {
	dmf, err := ParseFile(filename)
	if err != nil {
		return err
	}
	err = dmf.Dump(output)
	if err != nil {
		return err
	}
	return nil
}
