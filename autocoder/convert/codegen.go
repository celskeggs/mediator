package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"strings"
)

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
	WorldRef string
	Tree     *gen.DefinedTree
	VarTypes map[string]dtype.DType
}

func (ctx CodeGenContext) UsrRef() string {
	if _, hasusr := ctx.VarTypes["usr"]; hasusr {
		return LocalVariablePrefix + "usr"
	} else {
		return "nil"
	}
}

func (ctx CodeGenContext) WithVar(name string, varType dtype.DType) CodeGenContext {
	if _, found := ctx.VarTypes[name]; found {
		util.FIXME("should probably not panic here")
		panic("duplicate variable " + name)
	}
	vt := map[string]dtype.DType{}
	for k, v := range ctx.VarTypes {
		vt[k] = v
	}
	vt[name] = varType
	ctx.VarTypes = vt
	return ctx
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

func unstring(expr string) string {
	if strings.HasPrefix(expr, "types.String(") && strings.HasSuffix(expr, ")") {
		return expr[len("types.String(") : len(expr)-1]
	} else {
		return "types.Unstring(" + expr + ")"
	}
}

// expressions should always produce a types.Value
func ExprToGo(expr parser.DreamMakerExpression, ctx CodeGenContext) (exprString string, etype dtype.DType, err error) {
	switch expr.Type {
	case parser.ExprTypeResourceLiteral:
		switch ResourceTypeByName(expr.Str) {
		case ResourceTypeIcon:
			ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/atoms")
			return fmt.Sprintf("%s.Icon(%q)", ctx.WorldRef, expr.Str), dtype.ConstPath("/icon"), nil
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
	case parser.ExprTypeStringMacro:
		innerExpr, _, err := ExprToGo(expr.Children[0], ctx)
		if err != nil {
			return "", dtype.None(), err
		}
		ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/format")
		return fmt.Sprintf("types.String(format.FormatMacro(%q, %s))", expr.Str, innerExpr), dtype.String(), nil
	case parser.ExprTypeBooleanNot:
		innerString, _, err := ExprToGo(expr.Children[0], ctx)
		if err != nil {
			return "", dtype.None(), err
		}
		return fmt.Sprintf("procs.OperatorNot(%s)", innerString), dtype.Integer(), nil
	case parser.ExprTypeCall:
		target := expr.Children[0]
		args := expr.Children[1:]

		if target.Type != parser.ExprTypeGetNonLocal {
			return "", dtype.None(), fmt.Errorf("calling non-global functions is not yet implemented")
		}

		var invokeSrc string
		var found bool
		if ctx.Tree.GlobalProcedureExists(target.Str) {
			found = true
		} else if srctype, ok := ctx.VarTypes["src"]; ok && srctype.IsAnyPath() {
			if _, ok := ctx.Tree.ResolveProcedure(srctype.Path(), target.Str); ok {
				found = true
				invokeSrc = LocalVariablePrefix + "src"
			}
		}
		if !found {
			return "", dtype.None(), fmt.Errorf("no such function %s at %v", target.Str, target.SourceLoc)
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
			if invokeSrc != "" {
				return "", dtype.None(), fmt.Errorf("no support for keyword arguments in datum procedure invocations at %v", expr.SourceLoc)
			}
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
			return fmt.Sprintf("procs.KWInvoke(%s, %s, %q, %s%s)", ctx.WorldRef, ctx.UsrRef(), target.Str, kwargStr, strings.Join(convArgs, "")), dtype.Any(), nil
		} else if invokeSrc == "" {
			return fmt.Sprintf("procs.Invoke(%s, %s, %q%s)", ctx.WorldRef, ctx.UsrRef(), target.Str, strings.Join(convArgs, "")), dtype.Any(), nil
		} else {
			return fmt.Sprintf("(%s).Invoke(%s, %q%s)", invokeSrc, ctx.UsrRef(), target.Str, strings.Join(convArgs, "")), dtype.Any(), nil
		}
	case parser.ExprTypeGetNonLocal:
		util.FIXME("resolve more types of nonlocals")
		// look for local fields
		if srctype, ok := ctx.VarTypes["src"]; ok && srctype.IsAnyPath() {
			ftype, found := ctx.Tree.ResolveField(srctype.Path(), expr.Str)
			if found {
				return LocalVariablePrefix + fmt.Sprintf("src.Var(%q)", expr.Str), ftype, nil
			}
		}
		return "", dtype.None(), fmt.Errorf("cannot find nonlocal %s at %v", expr.Str, expr.SourceLoc)
	case parser.ExprTypeGetLocal:
		vtype, ok := ctx.VarTypes[expr.Str]
		if !ok {
			return "", dtype.None(), fmt.Errorf("unexpectedly could not find type for var %q at %v ... there may be a parser bug", expr.Str, expr.SourceLoc)
		}
		return LocalVariablePrefix + expr.Str, vtype, nil
	case parser.ExprTypeGetField:
		exprStr, exprType, err := ExprToGo(expr.Children[0], ctx)
		if err != nil {
			return "", dtype.None(), err
		}
		if !exprType.IsAnyPath() {
			return "", dtype.None(), fmt.Errorf("attempt to find field %q on non-datum type %v at %v", expr.Str, exprType, expr.SourceLoc)
		}
		fieldType, ok := ctx.Tree.ResolveField(exprType.Path(), expr.Str)
		if !ok {
			return "", dtype.None(), fmt.Errorf("cannot find field %q on datum type %v at %v", expr.Str, exprType, expr.SourceLoc)
		}
		return fmt.Sprintf("(%s).Var(%q)", exprStr, expr.Str), fieldType, nil
	case parser.ExprTypeStringConcat:
		var terms []string
		for _, term := range expr.Children {
			termString, actualType, err := ExprToGo(term, ctx)
			if err != nil {
				return "", dtype.None(), err
			}
			if !actualType.IsString() {
				return "", dtype.None(), fmt.Errorf("expected string concat to have string term, not %v at %v", actualType, expr.SourceLoc)
			}
			ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/types")
			terms = append(terms, unstring(termString))
		}
		return fmt.Sprintf("types.String(%s)", strings.Join(terms, " + ")), dtype.String(), nil
	default:
		return "", dtype.None(), fmt.Errorf("unimplemented evaluation of expr %v at %v", expr, expr.SourceLoc)
	}
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
	case parser.StatementTypeForList:
		list, _, err := ExprToGo(statement.From, ctx)
		if err != nil {
			return nil, err
		}
		ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/datum")
		lines = append(lines, fmt.Sprintf("for _, %s := range datum.Elements(%s) {", LocalVariablePrefix+statement.Name, list))
		if !statement.VarType.IsNone() {
			lines = append(lines, fmt.Sprintf("if !types.IsType(%s, %q) {", LocalVariablePrefix+statement.Name, statement.VarType.Path()))
			lines = append(lines, "continue")
			lines = append(lines, "}")
		}
		subctx := ctx.WithVar(statement.Name, statement.VarType)
		for _, bodyStatement := range statement.Body {
			extraLines, err := StatementToGo(bodyStatement, subctx)
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
			fmt.Sprintf("(%s).Invoke(%s, \"<<\", %s)", target, ctx.UsrRef(), value),
		}, nil
	case parser.StatementTypeReturn:
		util.FIXME("support returning values")
		return []string{
			"return nil",
		}, nil
	case parser.StatementTypeEvaluate:
		value, _, err := ExprToGo(statement.To, ctx)
		if err != nil {
			return nil, err
		}
		return []string{
			"_ = " + value,
		}, nil
	}
	return nil, fmt.Errorf("cannot convert statement %v to Go at %v", statement, statement.SourceLoc)
}

func DefaultSrcSetting(tree *gen.DefinedTree, typePath path.TypePath) types.SrcSetting {
	if tree.Extends(typePath, path.ConstTypePath("/mob")) {
		return types.SrcSetting{
			Type: types.SrcSettingTypeUsr,
			In:   false,
		}
	} else if tree.Extends(typePath, path.ConstTypePath("/obj")) {
		return types.SrcSetting{
			Type: types.SrcSettingTypeUsr,
			In:   true,
		}
	} else if tree.Extends(typePath, path.ConstTypePath("/turf")) {
		return types.SrcSetting{
			Type: types.SrcSettingTypeView,
			Dist: 0,
			In:   false,
		}
	} else if tree.Extends(typePath, path.ConstTypePath("/area")) {
		return types.SrcSetting{
			Type: types.SrcSettingTypeView,
			Dist: 0,
			In:   false,
		}
	} else {
		util.FIXME("figure out what the correct default should be for other cases")
		return types.SrcSetting{}
	}
}

func ParseSrcSetting(expr parser.DreamMakerExpression, stype parser.StatementType) (types.SrcSetting, error) {
	var sst types.SrcSettingType
	var dist uint
	switch expr.Type {
	case parser.ExprTypeCall:
		for _, name := range expr.Names {
			if name != "" {
				return types.SrcSetting{}, fmt.Errorf("cannot handle keyword arguments in src setting at %v", expr.SourceLoc)
			}
		}
		if expr.Children[0].Type != parser.ExprTypeGetNonLocal || expr.Children[0].Str != "oview" {
			return types.SrcSetting{}, fmt.Errorf("expected call only to oview, not %q, in src setting at %v", expr.Children[0].Str, expr.Children[0].SourceLoc)
		}
		if len(expr.Children) > 2 {
			return types.SrcSetting{}, fmt.Errorf("expected call to have 0-1 arguments when in src setting at %v", expr.SourceLoc)
		}
		if expr.Children[1].Type != parser.ExprTypeIntegerLiteral {
			return types.SrcSetting{}, fmt.Errorf("expected integer literal in oview parameter at %v", expr.Children[1].SourceLoc)
		}
		sst = types.SrcSettingTypeOView
		dist = uint(expr.Children[1].Integer)
		if int64(dist) != expr.Children[1].Integer {
			return types.SrcSetting{}, fmt.Errorf("integer literal out of range at %v", expr.Children[1].SourceLoc)
		}
	default:
		return types.SrcSetting{}, fmt.Errorf("unexpected expression %v while parsing src setting at %v", expr, expr.SourceLoc)
	}
	return types.SrcSetting{
		Type: sst,
		Dist: dist,
		In:   stype == parser.StatementTypeSetIn,
	}, nil
}

func ParseSettings(dt *gen.DefinedTree, typePath path.TypePath, body []parser.DreamMakerStatement) (types.ProcSettings, []parser.DreamMakerStatement, error) {
	settings := types.ProcSettings{}
	settings.Src = DefaultSrcSetting(dt, typePath)
	setSrc := false
	for len(body) > 0 && (body[0].Type == parser.StatementTypeSetIn || body[0].Type == parser.StatementTypeSetTo) {
		if body[0].Name == "src" {
			if setSrc {
				return types.ProcSettings{}, nil, fmt.Errorf("duplicate setting for src at %v", body[0].SourceLoc)
			}
			var err error
			settings.Src, err = ParseSrcSetting(body[0].To, body[0].Type)
			if err != nil {
				return types.ProcSettings{}, nil, err
			}
			setSrc = true
		}
		body = body[1:]
	}
	return settings, body, nil
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
