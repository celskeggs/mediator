package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
	"strings"
)

func DefinePath(dt *gen.DefinedTree, path path.TypePath) {
	if dt.Exists(path.String()) {
		return
	}
	switch path.String() {
	case "/world":
		// nothing to do
	default:
		dt.Types = append(dt.Types, gen.DefinedType{
			TypePath: path.String(),
		})
	}
}

func ConstantString(expr parser.DreamMakerExpression) string {
	if expr.Type == parser.ExprTypeStringLiteral {
		return expr.Str
	} else {
		panic("unimplemented: constant string from expr " + expr.String())
	}
}

func ConstantPath(expr parser.DreamMakerExpression) path.TypePath {
	if expr.Type == parser.ExprTypePathLiteral {
		return expr.Path
	} else {
		panic("unimplemented: constant path from expr " + expr.String())
	}
}

func EscapeString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	return "\"" + str + "\""
}

func ExprToGo(expr parser.DreamMakerExpression, targetType string) string {
	switch expr.Type {
	case parser.ExprTypeResourceLiteral:
		if targetType == "icon.Icon" {
			return "icons.LoadOrPanic(" + EscapeString(expr.Str) + ")"
		} else {
			panic("unimplemented: converting expr " + expr.String() + " to type " + targetType)
		}
	case parser.ExprTypeIntegerLiteral:
		if targetType == "bool" {
			return strconv.FormatBool(expr.Integer != 0)
		} else {
			return strconv.FormatInt(expr.Integer, 10)
		}
	default:
		panic("unimplemented: converting expr " + expr.String() + " to type " + targetType)
	}
}

func AssignPath(dt *gen.DefinedTree, path path.TypePath, variable string, expr parser.DreamMakerExpression, loc tokenizer.SourceLocation) error {
	switch path.String() {
	case "/world":
		switch variable {
		case "name":
			dt.WorldName = ConstantString(expr)
		case "mob":
			dt.WorldMob = ConstantPath(expr).String()
			if !dt.Exists(dt.WorldMob) {
				panic("path " + dt.WorldMob + " does not actually exist in the tree")
			}
		default:
			panic("unimplemented: assigning path " + path.String() + " var " + variable)
		}
	default:
		if !dt.Exists(path.String()) {
			panic("unimplemented: assigning path " + path.String() + " var " + variable)
		}
		_, _, goType, found := dt.ResolveField(path.String(), variable)
		if !found {
			return fmt.Errorf("no such field %s on %s at %v", variable, path.String(), loc)
		}
		// CHECK: is this broken by assigning to a pointer grabbed from a slice?
		defType := dt.GetTypeByPath(path.String())
		defType.Inits = append(defType.Inits, gen.DefinedInit{
			ShortName: variable,
			Value:     ExprToGo(expr, goType),
			SourceLoc: loc,
		})
	}
	return nil
}

// Used when injecting new code
func SourceHere() tokenizer.SourceLocation {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		return tokenizer.SourceLocation{file, line, 0}
	}
	return tokenizer.SourceLocation{"", 0, 0}
}

func Convert(dmf *parser.DreamMakerFile) (*gen.DefinedTree, error) {
	dt := &gen.DefinedTree{
		WorldMob:  "/mob",
		WorldName: "World",
	}
	for _, def := range dmf.Definitions {
		if def.Type == parser.DefTypeDefine {
			DefinePath(dt, def.Path)
		}
	}
	for _, def := range dmf.Definitions {
		if def.Type == parser.DefTypeAssign {
			err := AssignPath(dt, def.Path, def.Variable, def.Expression, def.SourceLoc)
			if err != nil {
				return nil, err
			}
		}
	}
	for i, t := range dt.Types {
		if dt.Extends(t.TypePath, "/atom") {
			specifiesName := false
			for _, init := range t.Inits {
				if init.ShortName == "name" {
					specifiesName = true
				}
			}
			if !specifiesName {
				parts := strings.Split(t.TypePath, "/")
				t.Inits = append(t.Inits, gen.DefinedInit{
					ShortName: "name",
					Value:     EscapeString(parts[len(parts)-1]),
					SourceLoc: SourceHere(),
				})
				dt.Types[i] = t
			}
		}
	}
	return dt, nil
}

func ConvertFiles(inputFile string, outputFile string) error {
	dmf, err := parser.ParseFile(inputFile)
	if err != nil {
		return errors.Wrap(err, "while parsing input file")
	}
	tree, err := Convert(dmf)
	if err != nil {
		return errors.Wrap(err, "while building tree")
	}
	err = gen.GenerateTo(tree, outputFile)
	if err != nil {
		return errors.Wrap(err, "while generating output file")
	}
	return nil
}
