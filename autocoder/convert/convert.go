package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
	"strings"
)

func DefinePath(dt *gen.DefinedTree, path path.TypePath) error {
	if dt.GetTypeByPath(path.String()) != nil {
		return nil
	}
	switch path.String() {
	case "/world":
		// nothing to do
	default:
		if dt.Exists(path.String()) {
			// exists, but not as a locally-defined type: we're trying to override something!
			dt.Types = append(dt.Types, gen.DefinedType{
				TypePath: path.String(),
				BasePath: path.String(),
			})
		} else {
			dt.Types = append(dt.Types, gen.DefinedType{
				TypePath: path.String(),
			})
		}
	}
	return nil
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

type ResourceType int
const (
	ResourceTypeNone ResourceType = iota
	ResourceTypeIcon
	ResourceTypeAudio
)

func ResourceTypeByName(name string) ResourceType {
	if strings.HasSuffix(name, ".mid") {
		return ResourceTypeAudio
	} else if strings.HasSuffix(name, ".dmi") {
		return ResourceTypeIcon
	} else {
		return ResourceTypeNone
	}
}

func ExprToGo(expr parser.DreamMakerExpression, targetType string) (string, error) {
	switch expr.Type {
	case parser.ExprTypeResourceLiteral:
		switch ResourceTypeByName(expr.Str) {
		case ResourceTypeIcon:
			if targetType == "icon.Icon" || targetType == "interface{}" {
				return "icons.LoadOrPanic(" + EscapeString(expr.Str) + ")", nil
			}
		case ResourceTypeAudio:
			util.FIXME("implement audio support")
		}
	case parser.ExprTypeIntegerLiteral:
		if targetType == "bool" {
			return strconv.FormatBool(expr.Integer != 0), nil
		} else if targetType == "int" || targetType == "interface{}" {
			return strconv.FormatInt(expr.Integer, 10), nil
		}
	case parser.ExprTypeStringLiteral:
		if targetType == "string" || targetType == "interface{}" {
			return fmt.Sprintf("%q", expr.Str), nil
		}
	}
	return "", fmt.Errorf("cannot convert expr %v to type %v at %v", expr, targetType, expr.SourceLoc)
}

func DefineVar(dt *gen.DefinedTree, path path.TypePath, variable string, loc tokenizer.SourceLocation) error {
	if !dt.Exists(path.String()) {
		return fmt.Errorf("no such path %v for declaration of variable %v at %v", path, variable, loc)
	}
	defType := dt.GetTypeByPath(path.String())
	if defType == nil {
		panic("expected non-nil type " + path.String())
	}

	_, _, _, found := dt.ResolveField(path.String(), variable)
	if found {
		return fmt.Errorf("field %s already defined on %v at %v", variable, path, loc)
	}

	defType.Fields = append(defType.Fields, gen.DefinedField{
		Name: variable,
		Type: "interface{}",
	})
	return nil
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
			return fmt.Errorf("no such path %v for assignment of variable %v", path, variable)
		}
	default:
		if !dt.Exists(path.String()) {
			return fmt.Errorf("no such path %v for assignment of variable %v", path, variable)
		}
		_, _, goType, found := dt.ResolveField(path.String(), variable)
		if !found {
			return fmt.Errorf("no such field %s on %s at %v", variable, path.String(), loc)
		}
		// CHECK: is this broken by assigning to a pointer grabbed from a slice?
		defType := dt.GetTypeByPath(path.String())
		expr, err := ExprToGo(expr, goType)
		if err != nil {
			return err
		}
		defType.Inits = append(defType.Inits, gen.DefinedInit{
			ShortName: variable,
			Value:     expr,
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
	// define all types
	for _, def := range dmf.Definitions {
		if def.Type == parser.DefTypeDefine {
			err := DefinePath(dt, def.Path)
			if err != nil {
				return nil, err
			}
		}
	}
	// declare all variables
	for _, def := range dmf.Definitions {
		if def.Type == parser.DefTypeVarDef {
			err := DefineVar(dt, def.Path, def.Variable, def.SourceLoc)
			if err != nil {
				return nil, err
			}
		}
	}
	// assign all values
	for _, def := range dmf.Definitions {
		if def.Type == parser.DefTypeAssign {
			err := AssignPath(dt, def.Path, def.Variable, def.Expression, def.SourceLoc)
			if err != nil {
				return nil, err
			}
		}
	}
	// insert names for everything unnamed
	for i, t := range dt.Types {
		if dt.Extends(t.TypePath, "/atom") && !t.IsOverride() {
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

func ConvertFiles(inputFiles []string, outputFile string) error {
	dmf, err := parser.ParseFiles(inputFiles)
	if err != nil {
		return errors.Wrap(err, "while parsing input files")
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
