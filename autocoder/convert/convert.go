package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"github.com/pkg/errors"
	"strconv"
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

func EscapeString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	return "\"" + str + "\""
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

// Typechecking is done on input; all ExprToGo paths should check that they can provide targetType. Callers do not need to validate goType.
func ExprToGo(expr parser.DreamMakerExpression, targetType gotype.GoType, ctx CodeGenContext) (exprString string, goType gotype.GoType, err error) {
	switch expr.Type {
	case parser.ExprTypeResourceLiteral:
		switch ResourceTypeByName(expr.Str) {
		case ResourceTypeIcon:
			if targetType.IsExternal("*icon.Icon") || targetType.IsInterfaceAny() {
				return "icons.LoadOrPanic(" + EscapeString(expr.Str) + ")", gotype.External("*icons.Icon"), nil
			}
		case ResourceTypeAudio:
			if targetType.IsExternal("sprite.Sound") || targetType.IsInterfaceAny() {
				return "platform.NewSound(" + EscapeString(expr.Str) + ")", gotype.External("sprite.Sound"), nil
			}
		}
	case parser.ExprTypeIntegerLiteral:
		if targetType.IsBool() {
			return strconv.FormatBool(expr.Integer != 0), gotype.Bool(), nil
		} else if targetType.IsInteger() || targetType.IsInterfaceAny() {
			return strconv.FormatInt(expr.Integer, 10), gotype.Int(), nil
		}
	case parser.ExprTypeStringLiteral:
		if targetType.IsString() || targetType.IsInterfaceAny() {
			return fmt.Sprintf("%q", expr.Str), gotype.String(), nil
		}
	case parser.ExprTypeBooleanNot:
		if targetType.IsBool() || targetType.IsInterfaceAny() {
			innerString, _, err := ExprToGo(expr.Children[0], gotype.Bool(), ctx)
			if err != nil {
				return "", gotype.None(), err
			}
			return fmt.Sprintf("!(%s)", innerString), gotype.Bool(), nil
		}
	case parser.ExprTypeCall:
		target := expr.Children[0]
		args := expr.Children[1:]
		targetString, concrete, err := ExprToGo(target, gotype.FuncAbstractParams([]gotype.GoType{targetType}), ctx)
		if err != nil {
			return "", gotype.None(), err
		}
		if !concrete.IsFuncConcrete() {
			return "", gotype.None(), fmt.Errorf("concrete func required, but found %v", concrete)
		}
		if len(concrete.Results) != 1 {
			return "", gotype.None(), fmt.Errorf("can only handle funcs with exactly 1 result, but found %v", concrete)
		}
		filledArgs := make([]string, len(concrete.Params))

		// fill keyword args first
		for ni, name := range expr.Names {
			if name != "" {
				found := false
				for ai, ak := range concrete.KeywordNames {
					if name == ak {
						if filledArgs[ai] != "" {
							return "", gotype.None(), fmt.Errorf("duplicate keyword argument: %s", name)
						}
						argString, _, err := ExprToGo(args[ni], concrete.Params[ai], ctx)
						if err != nil {
							return "", gotype.None(), err
						}
						filledArgs[ai] = argString
						found = true
						break
					}
				}
				if !found {
					return "", gotype.None(), fmt.Errorf("no such keyword argument: %s", name)
				}
			}
		}
		// fill positional arguments second
		for ai, arg := range args {
			if expr.Names[ai] == "" {
				found := false
				for fi, fill := range filledArgs {
					if fill == "" {
						argString, _, err := ExprToGo(arg, concrete.Params[fi], ctx)
						if err != nil {
							return "", gotype.None(), err
						}
						filledArgs[fi] = argString
						found = true
						break
					}
				}
				if !found {
					return "", gotype.None(), fmt.Errorf("too many positional arguments when trying to use argument %d", ai)
				}
			}
		}
		// fill default arguments third
		for i, expr := range concrete.DefaultExprs {
			if filledArgs[i] == "" {
				if expr != "" {
					filledArgs[i] = expr
				} else {
					return "", gotype.None(), fmt.Errorf("too few parameters; no expression for argument %d", i)
				}
			}
		}

		return fmt.Sprintf("(%s)(%s)", targetString, strings.Join(filledArgs, ", ")), concrete.Results[0], nil
	case parser.ExprTypeGetNonLocal:
		util.FIXME("resolve more types of nonlocals")
		// look for local fields
		if !ctx.SrcType.IsEmpty() {
			definingStruct, longName, goType, found := ctx.Tree.ResolveField(ctx.SrcType, expr.Str)
			if found {
				if !targetType.IsInterfaceAny() && !targetType.Equals(goType) {
					return "", gotype.None(), fmt.Errorf("type mismatch: needed %v but src.%s was %v at %v", targetType, expr.Str, goType, expr.SourceLoc)
				}
				if definingStruct == ctx.Tree.GetTypeByPath(ctx.SrcType).StructName() {
					return LocalVariablePrefix + "src." + longName, goType, nil
				} else {
					return LocalVariablePrefix + "src.As" + definingStruct + "()." + longName, goType, nil
				}
			}
		}
		// look for global procedures
		if targetType.IsFunc() || targetType.IsInterfaceAny() {
			record, found := ctx.Tree.ResolveGlobalProcedure(expr.Str)
			if found {
				if !record.GoType.IsFuncConcrete() {
					return "", gotype.None(), fmt.Errorf("discovered global func for %s must be of a concrete type, not %v", expr.Str, record.GoType)
				}
				return record.GoRef, record.GoType, nil
			}
		}
		return "", gotype.None(), fmt.Errorf("cannot find nonlocal %s at %v", expr.Str, expr.SourceLoc)
	case parser.ExprTypeGetLocal:
		util.FIXME("type checking for locals")
		return LocalVariablePrefix + expr.Str, gotype.InterfaceAny(), nil
	case parser.ExprTypeStringConcat:
		var terms []string
		for _, term := range expr.Children {
			termString, actualType, err := ExprToGo(term, gotype.InterfaceAny(), ctx)
			if err != nil {
				return "", gotype.None(), err
			}
			if actualType.IsString() {
				terms = append(terms, "("+termString+")")
			} else {
				util.FIXME("think more carefully here than just assuming it's an atom")
				ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/format")
				terms = append(terms, "format.FormatAtom("+termString+")")
			}
		}
		return strings.Join(terms, " + "), gotype.String(), nil
	}
	return "", gotype.None(), fmt.Errorf("cannot convert expr %v to type %v at %v", expr, targetType, expr.SourceLoc)
}

func DefineVar(dt *gen.DefinedTree, path path.TypePath, variable string, loc tokenizer.SourceLocation) error {
	if !dt.Exists(path) {
		return fmt.Errorf("no such path %v for declaration of variable %v at %v", path, variable, loc)
	}
	defType := dt.GetTypeByPath(path)
	if defType == nil {
		panic("expected non-nil type " + path.String())
	}

	_, _, _, found := dt.ResolveField(path, variable)
	if found {
		return fmt.Errorf("field %s already defined on %v at %v", variable, path, loc)
	}

	defType.Fields = append(defType.Fields, gen.DefinedField{
		Name: variable,
		Type: gotype.InterfaceAny(),
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
		_, _, goType, found := dt.ResolveField(path, variable)
		if !found {
			return fmt.Errorf("no such field %s on %v at %v", variable, path, loc)
		}
		// CHECK: is this broken by assigning to a pointer grabbed from a slice?
		defType := dt.GetTypeByPath(path)
		expr, _, err := ExprToGo(expr, goType, CodeGenContext{
			Tree: dt,
		})
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

func StatementToGo(statement parser.DreamMakerStatement, ctx CodeGenContext) (lines []string, err error) {
	switch statement.Type {
	case parser.StatementTypeIf:
		condition, _, err := ExprToGo(statement.From, gotype.Bool(), ctx)
		if err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("if (%v) {", condition))
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
		target, _, err := ExprToGo(statement.To, gotype.External("platform.IMob"), ctx)
		if err != nil {
			return nil, err
		}
		value, valueType, err := ExprToGo(statement.From, gotype.InterfaceAny(), ctx)
		if err != nil {
			return nil, err
		}
		if valueType.IsString() {
			return []string{
				fmt.Sprintf("(%s).OutputString(%s)", target, value),
			}, nil
		} else if valueType.IsExternal("sprite.Sound") {
			return []string{
				fmt.Sprintf("(%s).OutputSound(%s)", target, value),
			}, nil
		} else {
			return nil, fmt.Errorf("write operator with unknown variant not implemented at %v", statement.SourceLoc)
		}
	case parser.StatementTypeReturn:
		util.FIXME("support returning values")
		return []string{
			"return",
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
		// not strictly necessary YET, but it will be at some point
		lines = append(lines, "return")
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

	proc, found := dt.ResolveProcedure(path, function)
	if !found {
		return fmt.Errorf("no such function %s to implement on %v at %v", function, path, loc)
	}
	if !proc.GoType.IsFuncConcrete() {
		panic("resolved procedures should always have a concrete type")
	}
	if len(proc.GoType.Results) > 0 {
		return fmt.Errorf("unsupported: function with results at %v", tokenizer.SourceHere())
	}

	var params []gen.DefinedParam
	for _, a := range arguments {
		util.FIXME("check argument type compatibility")
		params = append(params,
			gen.DefinedParam{
				Name: LocalVariablePrefix + a.Name,
				Type: dt.Ref(a.Type, true),
			},
		)
	}
	if len(params) > len(proc.GoType.Params) {
		return fmt.Errorf("attempt to define function with more parameters (%v) than overridden function (%v)", len(params), len(proc.GoType.Params))
	}
	for i := len(params); i < len(proc.GoType.Params); i++ {
		params = append(params,
			gen.DefinedParam{
				Name: "_",
				Type: proc.GoType.Params[i].String(),
			},
		)
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

func Convert(dmf *parser.DreamMakerFile) (*gen.DefinedTree, error) {
	dt := &gen.DefinedTree{
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
				if init.ShortName == "name" {
					specifiesName = true
				}
			}
			if !specifiesName {
				_, lastComponent, err := t.TypePath.SplitLast()
				if err != nil {
					return nil, err
				}
				t.Inits = append(t.Inits, gen.DefinedInit{
					ShortName: "name",
					Value:     EscapeString(lastComponent),
					SourceLoc: tokenizer.SourceHere(),
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
