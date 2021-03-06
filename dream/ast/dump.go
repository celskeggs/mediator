package ast

import (
	"fmt"
	"io"
	"strings"
)

func makeIndent(indent int) string {
	return strings.Repeat("\t", indent)
}

func (dms Statement) Dump(output io.Writer, indent int) error {
	_, err := fmt.Fprintf(output, "%s[statement %v]\n", makeIndent(indent), dms.Type)
	if err != nil {
		return err
	}
	if dms.Name != "" {
		_, err := fmt.Fprintf(output, "%sname = %q\n", makeIndent(indent+1), dms.Name)
		if err != nil {
			return err
		}
	}
	if !dms.VarType.IsNone() {
		_, err := fmt.Fprintf(output, "%spath = %v\n", makeIndent(indent+1), dms.VarType)
		if err != nil {
			return err
		}
	}
	if !dms.From.IsNone() {
		_, err := fmt.Fprintf(output, "%sfrom = %v\n", makeIndent(indent+1), dms.From)
		if err != nil {
			return err
		}
	}
	if !dms.To.IsNone() {
		_, err := fmt.Fprintf(output, "%sto = %v\n", makeIndent(indent+1), dms.To)
		if err != nil {
			return err
		}
	}
	if len(dms.Body) > 0 {
		err := DumpStatementList(output, "body", indent+1, dms.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func DumpStatementList(output io.Writer, header string, indent int, statements []Statement) error {
	_, err := fmt.Fprintf(output, "%s[%s len=%d]\n", makeIndent(indent), header, len(statements))
	if err != nil {
		return err
	}
	for _, s := range statements {
		err := s.Dump(output, indent+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dmd Definition) Dump(output io.Writer) error {
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
	if !dmd.VarType.IsEmpty() {
		_, err := fmt.Fprintf(output, "\tvartype = %v\n", dmd.VarType)
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
		_, err := fmt.Fprintf(output, "\texpression = %v\n", dmd.Expression)
		if err != nil {
			return err
		}
	}
	if len(dmd.Arguments) > 0 {
		_, err := fmt.Fprintf(output, "\t[arguments %d]\n", len(dmd.Arguments))
		if err != nil {
			return err
		}
		for i, argument := range dmd.Arguments {
			_, err := fmt.Fprintf(output, "\t\t[%d] %s: %v\n", i, argument.Name, argument.Type)
			if err != nil {
				return err
			}
		}
	}
	if len(dmd.Body) > 0 {
		err := DumpStatementList(output, "statements", 1, dmd.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *File) Dump(output io.Writer) error {
	_, err := fmt.Fprintln(output, "[beginning of parser dump]")
	if err != nil {
		return err
	}
	for _, dir := range f.SearchPath {
		_, err := fmt.Fprintf(output, "searchpath = %q\n", dir)
		if err != nil {
			return err
		}
	}
	for _, mapName := range f.Maps {
		_, err := fmt.Fprintf(output, "map = %q\n", mapName)
		if err != nil {
			return err
		}
	}
	for _, def := range f.Definitions {
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
