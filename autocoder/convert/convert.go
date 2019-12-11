package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/autocoder/pack"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"github.com/pkg/errors"
	"strings"
)

func DefinePath(dt *gen.DefinedTree, path path.TypePath) error {
	if dt.GetTypeByPath(path) != nil {
		return nil
	}
	switch path.String() {
	case "/world":
		// nothing to do
	default:
		if dt.Exists(path) {
			// exists, but not as a locally-defined type: we're trying to override something!
			dt.Types = append(dt.Types, gen.DefinedType{
				TypePath: path,
				BasePath: path,
			})
		} else {
			dt.Types = append(dt.Types, gen.DefinedType{
				TypePath: path,
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

type CodeGenContext struct {
	Tree    *gen.DefinedTree
	SrcType path.TypePath
}

type ResourceType int

const (
	ResourceTypeNone ResourceType = iota
	ResourceTypeIcon
	ResourceTypeAudio
)

func ResourceTypeByName(name string) ResourceType {
	if strings.HasSuffix(name, ".mid") || strings.HasSuffix(name, ".wav") {
		return ResourceTypeAudio
	} else if strings.HasSuffix(name, ".dmi") {
		return ResourceTypeIcon
	} else {
		return ResourceTypeNone
	}
}

const LocalVariablePrefix = "var"

// expressions should always produce a types.Value
func ExprToGo(expr parser.DreamMakerExpression, ctx CodeGenContext) (exprString string, etype dtype.DType, err error) {
	switch expr.Type {
	case parser.ExprTypeResourceLiteral:
		switch ResourceTypeByName(expr.Str) {
		case ResourceTypeIcon:
			util.FIXME("make this load work in more cases")
			ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/atoms")
			return fmt.Sprintf("atoms.WorldOf(src).Icon(%q)", expr.Str), dtype.ConstPath("/icon"), nil
		case ResourceTypeAudio:
			ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/procs")
			return fmt.Sprintf("procs.NewSound(%q)", expr.Str), dtype.ConstPath("/sound"), nil
		default:
			return "", dtype.None(), fmt.Errorf("cannot interpret resource name %q", expr.Str)
		}
	case parser.ExprTypeIntegerLiteral:
		return fmt.Sprintf("types.Int(%d)", expr.Integer), dtype.Integer(), nil
	case parser.ExprTypeStringLiteral:
		return fmt.Sprintf("types.String(%q)", expr.Str), dtype.String(), nil
	case parser.ExprTypeBooleanNot:
		innerString, _, err := ExprToGo(expr.Children[0], ctx)
		if err != nil {
			return "", dtype.None(), err
		}
		return fmt.Sprintf("procs.Invoke(\"!\", %s)", innerString), dtype.Integer(), nil
	case parser.ExprTypeCall:
		target := expr.Children[0]
		args := expr.Children[1:]

		if target.Type != parser.ExprTypeGetNonLocal {
			return "", dtype.None(), fmt.Errorf("calling non-global functions is not yet implemented")
		}

		found := ctx.Tree.GlobalProcedureExists(target.Str)
		if !found {
			return "", dtype.None(), fmt.Errorf("no such global function %s", target.Str)
		}

		kwargs := make(map[string]string)

		var convArgs []string
		for i, arg := range args {
			ce, _, err := ExprToGo(arg, ctx)
			if err != nil {
				return "", dtype.None(), err
			}
			kname := expr.Names[i]
			if kname != "" {
				if _, found := kwargs[kname]; found {
					return "", dtype.None(), fmt.Errorf("duplicate keyword argument %q", kname)
				}
				kwargs[expr.Names[i]] = ce
			} else {
				convArgs = append(convArgs, ", "+ce)
			}
		}

		if len(kwargs) != 0 {
			kwargStr := "map[string]types.Value{"
			first := true
			for name, arg := range kwargs {
				if first {
					first = false
				} else {
					kwargStr += ", "
				}
				kwargStr += fmt.Sprintf("%q: %s", name, arg)
			}
			kwargStr += "}"
			return fmt.Sprintf("procs.KWInvoke(%q, %s%s)", target.Str, kwargStr, strings.Join(convArgs, "")), dtype.Any(), nil
		} else {
			return fmt.Sprintf("procs.Invoke(%q%s)", target.Str, strings.Join(convArgs, "")), dtype.Any(), nil
		}
	case parser.ExprTypeGetNonLocal:
		util.FIXME("resolve more types of nonlocals")
		// look for local fields
		if !ctx.SrcType.IsEmpty() {
			ftype, found := ctx.Tree.ResolveField(ctx.SrcType, expr.Str)
			if found {
				return LocalVariablePrefix + fmt.Sprintf("src.Var(%q)", expr.Str), ftype, nil
			}
		}
		return "", dtype.None(), fmt.Errorf("cannot find nonlocal %s at %v", expr.Str, expr.SourceLoc)
	case parser.ExprTypeGetLocal:
		util.FIXME("types for locals")
		return LocalVariablePrefix + expr.Str, dtype.Any(), nil
	case parser.ExprTypeStringConcat:
		var terms []string
		for _, term := range expr.Children {
			termString, actualType, err := ExprToGo(term, ctx)
			if err != nil {
				return "", dtype.None(), err
			}
			if actualType.IsString() {
				terms = append(terms, "("+termString+")")
			} else {
				util.FIXME("think more carefully here than just assuming it's an atom")
				ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/format")
				terms = append(terms, "format.FormatAtom("+termString+")")
			}
		}
		return strings.Join(terms, " + "), dtype.String(), nil
	default:
		return "", dtype.None(), fmt.Errorf("unimplemented evaluation of expr %v at %v", expr, expr.SourceLoc)
	}
}

func DefineVar(dt *gen.DefinedTree, path path.TypePath, variable string, loc tokenizer.SourceLocation) error {
	if !dt.Exists(path) {
		return fmt.Errorf("no such path %v for declaration of variable %v at %v", path, variable, loc)
	}
	defType := dt.GetTypeByPath(path)
	if defType == nil {
		panic("expected non-nil type " + path.String())
	}

	_, found := dt.ResolveField(path, variable)
	if found {
		return fmt.Errorf("field %s already defined on %v at %v", variable, path, loc)
	}

	defType.Fields = append(defType.Fields, gen.DefinedField{
		Name: variable,
		Type: dtype.Any(),
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
			dt.WorldMob = ConstantPath(expr)
			if !dt.Exists(dt.WorldMob) {
				panic("path " + dt.WorldMob.String() + " does not actually exist in the tree")
			}
		default:
			return fmt.Errorf("no such path %v for assignment of variable %v", path, variable)
		}
	default:
		if !dt.Exists(path) {
			return fmt.Errorf("no such path %v for assignment of variable %v", path, variable)
		}
		util.FIXME("some sort of typechecking for field assignments?")
		_, found := dt.ResolveField(path, variable)
		if !found {
			return fmt.Errorf("no such field %s on %v at %v", variable, path, loc)
		}
		defType := dt.GetTypeByPath(path)
		expr, _, err := ExprToGo(expr, CodeGenContext{
			Tree: dt,
		})
		if err != nil {
			return err
		}
		defType.Inits = append(defType.Inits, gen.DefinedInit{
			Name:      variable,
			Value:     expr,
			SourceLoc: loc,
		})
	}
	return nil
}

func StatementToGo(statement parser.DreamMakerStatement, ctx CodeGenContext) (lines []string, err error) {
	switch statement.Type {
	case parser.StatementTypeIf:
		condition, _, err := ExprToGo(statement.From, ctx)
		if err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("if (types.AsBool(%v)) {", condition))
		for _, bodyStatement := range statement.Body {
			extraLines, err := StatementToGo(bodyStatement, ctx)
			if err != nil {
				return nil, err
			}
			lines = append(lines, extraLines...)
		}
		lines = append(lines, "}")
		return lines, nil
	case parser.StatementTypeWrite:
		target, _, err := ExprToGo(statement.To, ctx)
		if err != nil {
			return nil, err
		}
		value, _, err := ExprToGo(statement.From, ctx)
		if err != nil {
			return nil, err
		}
		return []string{
			fmt.Sprintf("(%s).Invoke(\"<<\", %s)", target, value),
		}, nil
	case parser.StatementTypeReturn:
		util.FIXME("support returning values")
		return []string{
			"return nil",
		}, nil
	}
	return nil, fmt.Errorf("cannot convert statement %v to Go at %v", statement, statement.SourceLoc)
}

func FuncBodyToGo(body []parser.DreamMakerStatement, ctx CodeGenContext) (lines []string, err error) {
	hadReturn := false
	for _, statement := range body {
		if hadReturn {
			return nil, fmt.Errorf("found statement after return: %v", statement.String())
		}
		extraLines, err := StatementToGo(statement, ctx)
		if err != nil {
			return nil, err
		}
		lines = append(lines, extraLines...)
		if statement.Type == parser.StatementTypeReturn {
			hadReturn = true
		}
	}
	if !hadReturn {
		lines = append(lines, "return nil")
	}
	return lines, nil
}

func MergeGoLines(lines []string) string {
	indent := 1
	var parts []string
	for _, line := range lines {
		if strings.HasPrefix(line, "}") {
			indent -= 1
			if indent < 1 {
				panic("internal error: generated code went into negative indentation")
			}
		}
		parts = append(parts, strings.Repeat("\t", indent)+line)
		if strings.HasSuffix(line, "{") {
			indent += 1
		}
	}
	if indent > 1 {
		panic("internal error: generated code had unresolved indentation")
	}
	return strings.Join(parts, "\n")
}

func ImplementFunction(dt *gen.DefinedTree, path path.TypePath, function string, arguments []parser.DreamMakerTypedName, body []parser.DreamMakerStatement, loc tokenizer.SourceLocation) error {
	if !dt.Exists(path) {
		return fmt.Errorf("no such path %v for implementation of function %v at %v", path, function, loc)
	}
	defType := dt.GetTypeByPath(path)
	if defType == nil {
		panic("expected non-nil type " + path.String())
	}

	_, found := dt.ResolveProcedure(path, function)
	if !found {
		return fmt.Errorf("no such function %s to implement on %v at %v", function, path, loc)
	}

	var params []string
	for _, a := range arguments {
		params = append(params, LocalVariablePrefix+a.Name)
	}

	lines, err := FuncBodyToGo(body, CodeGenContext{
		Tree:    dt,
		SrcType: path,
	})
	if err != nil {
		return err
	}

	defType.Funcs = append(defType.Funcs, gen.DefinedFunc{
		Name:   function,
		This:   LocalVariablePrefix + "src",
		Params: params,
		Body:   MergeGoLines(lines),
	})
	return nil
}

func Convert(dmf *parser.DreamMakerFile, packageName string) (*gen.DefinedTree, error) {
	dt := &gen.DefinedTree{
		Package:   packageName,
		WorldMob:  path.ConstTypePath("/mob"),
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
	// implement all functions
	for _, def := range dmf.Definitions {
		if def.Type == parser.DefTypeImplement {
			err := ImplementFunction(dt, def.Path, def.Variable, def.Arguments, def.Body, def.SourceLoc)
			if err != nil {
				return nil, err
			}
		}
	}
	// insert names for everything unnamed
	for i, t := range dt.Types {
		if dt.Extends(t.TypePath, path.ConstTypePath("/atom")) && !t.IsOverride() {
			specifiesName := false
			for _, init := range t.Inits {
				if init.Name == "name" {
					specifiesName = true
				}
			}
			if !specifiesName {
				_, lastComponent, err := t.TypePath.SplitLast()
				if err != nil {
					return nil, err
				}
				t.Inits = append(t.Inits, gen.DefinedInit{
					Name:      "name",
					Value:     fmt.Sprintf("types.String(%q)", lastComponent),
					SourceLoc: tokenizer.SourceHere(),
				})
				dt.Types[i] = t
			}
		}
	}
	return dt, nil
}

func ConvertFiles(inputFiles []string, outputGo string, outputPack string, packageName string) error {
	dmf, err := parser.ParseFiles(inputFiles)
	if err != nil {
		return errors.Wrap(err, "while parsing input files")
	}
	util.FIXME("use search path")
	tree, err := Convert(dmf, packageName)
	if err != nil {
		return errors.Wrap(err, "while building tree")
	}
	err = gen.GenerateTo(tree, outputGo)
	if err != nil {
		return errors.Wrap(err, "while generating output file")
	}
	err = pack.GenerateResourcePack(dmf, outputPack)
	if err != nil {
		return errors.Wrap(err, "while generating resource pack")
	}
	return nil
}
