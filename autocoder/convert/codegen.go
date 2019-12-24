package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"strings"
)

func ConstantString(expr ast.Expression) string {
	if expr.Type == ast.ExprTypeStringLiteral {
		return expr.Str
	} else {
		panic("unimplemented: constant string from expr " + expr.String())
	}
}

func ConstantPath(expr ast.Expression) path.TypePath {
	if expr.Type == ast.ExprTypePathLiteral {
		return expr.Path
	} else {
		panic("unimplemented: constant path from expr " + expr.String())
	}
}

type CodeGenContext struct {
	WorldRef string
	Tree     *gen.DefinedTree
	VarTypes map[string]dtype.DType
	Result   string
	ThisProc string
	DefIndex uint
}

func (ctx CodeGenContext) ChunkName() string {
	src, ok := ctx.VarTypes["src"]
	if !ok || ctx.Tree.PackageImport == "" {
		return ""
	}
	structName := ctx.Tree.GetTypeByPath(src.Path()).DataStructName()
	return ctx.Tree.PackageImport + "." + structName
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

func (ctx CodeGenContext) ResolveNonLocal(name string) (get string, set func(to string) string, vtype dtype.DType, ok bool) {
	util.FIXME("resolve more types of nonlocals")
	// look for local fields
	if srctype, ok := ctx.VarTypes["src"]; ok && srctype.IsAnyPath() {
		ftype, found := ctx.Tree.ResolveField(srctype.Path(), name)
		if found {
			return fmt.Sprintf("%ssrc.Var(%q)", LocalVariablePrefix, name),
				func(value string) string {
					return fmt.Sprintf("%ssrc.SetVar(%q, %s)", LocalVariablePrefix, name, value)
				},
				ftype, true
		}
	}
	return "", nil, dtype.None(), false
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
func ExprToGo(expr ast.Expression, ctx CodeGenContext) (exprString string, etype dtype.DType, err error) {
	switch expr.Type {
	case ast.ExprTypeResourceLiteral:
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
	case ast.ExprTypeIntegerLiteral:
		return fmt.Sprintf("types.Int(%d)", expr.Integer), dtype.Integer(), nil
	case ast.ExprTypeStringLiteral:
		return fmt.Sprintf("types.String(%q)", expr.Str), dtype.String(), nil
	case ast.ExprTypeStringMacro:
		innerExpr, _, err := ExprToGo(expr.Children[0], ctx)
		if err != nil {
			return "", dtype.None(), err
		}
		ctx.Tree.AddImport("github.com/celskeggs/mediator/platform/format")
		return fmt.Sprintf("types.String(format.FormatMacro(%q, %s))", expr.Str, innerExpr), dtype.String(), nil
	case ast.ExprTypeBooleanNot:
		innerString, _, err := ExprToGo(expr.Children[0], ctx)
		if err != nil {
			return "", dtype.None(), err
		}
		return fmt.Sprintf("procs.OperatorNot(%s)", innerString), dtype.Integer(), nil
	case ast.ExprTypeCall:
		target := expr.Children[0]
		args := expr.Children[1:]

		var invokeSrc string
		var super bool
		var found bool

		if target.Type == ast.ExprTypeGetNonLocal {
			if target.Str == ".." {
				if ctx.ChunkName() == "" || ctx.ThisProc == "" {
					return "", dtype.None(), fmt.Errorf("cannot use ..() in this non-proc location at %v", target.SourceLoc)
				}
				super = true
				found = true
			} else if ctx.Tree.GlobalProcedureExists(target.Str) {
				found = true
			} else if srctype, ok := ctx.VarTypes["src"]; ok && srctype.IsAnyPath() {
				if _, ok := ctx.Tree.ResolveProcedure(srctype.Path(), target.Str); ok {
					found = true
					invokeSrc = LocalVariablePrefix + "src"
				}
			}
		} else if target.Type == ast.ExprTypeGetField {
			evchild, dt, err := ExprToGo(target.Children[0], ctx)
			if err != nil {
				return "", dtype.None(), err
			}
			if !dt.IsAnyPath() {
				return "", dtype.None(), fmt.Errorf("calling functions on non-datum type %v at %v", dt, expr.SourceLoc)
			}
			invokeSrc = evchild
			_, found = ctx.Tree.ResolveProcedure(dt.Path(), target.Str)
		} else {
			return "", dtype.None(), fmt.Errorf("calling functions like %v is not yet implemented at %v", target, expr.SourceLoc)
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
			if invokeSrc != "" || super {
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
		} else if super {
			util.FIXME("should we be passing all original parameters by default if ..() is used without parameters?")
			if ctx.DefIndex > 0 {
				di := ctx.Tree.LookupIndexedImpl(ctx.VarTypes["src"].Path(), ctx.ThisProc, ctx.DefIndex-1)
				slots := make([]string, len(di.Params))
				if len(convArgs) > len(slots) {
					util.FIXME("handle additional arguments to redecl by evaluating but dropping them")
					return "", dtype.None(), fmt.Errorf("passed more arguments to internal redecl than permitted at %v", expr.SourceLoc)
				}
				for i := range slots {
					if i < len(convArgs) {
						slots[i] = ", " + convArgs[i]
					} else {
						slots[i] = ", nil"
					}
				}
				// we're a subsequent declaration on this particular type; we need to call within our struct
				return fmt.Sprintf("chunk.Shadow%dForProc%s(%s, %s%s)", di.DefIndex, di.Name, LocalVariablePrefix+"src", ctx.UsrRef(), strings.Join(slots, "")), dtype.Any(), nil
			}
			return fmt.Sprintf("varsrc.SuperInvoke(%s, %q, %q%s)", ctx.UsrRef(), ctx.ChunkName(), ctx.ThisProc, strings.Join(convArgs, "")), dtype.Any(), nil
		} else if invokeSrc == "" {
			return fmt.Sprintf("procs.Invoke(%s, %s, %q%s)", ctx.WorldRef, ctx.UsrRef(), target.Str, strings.Join(convArgs, "")), dtype.Any(), nil
		} else {
			return fmt.Sprintf("(%s).Invoke(%s, %q%s)", invokeSrc, ctx.UsrRef(), target.Str, strings.Join(convArgs, "")), dtype.Any(), nil
		}
	case ast.ExprTypeNew:
		for _, name := range expr.Names {
			if name != "" {
				return "", dtype.None(), fmt.Errorf("unhandled: keyword argument in new operation at %v", expr.SourceLoc)
			}
		}
		var argStrs []string
		for _, arg := range expr.Children {
			argStr, _, err := ExprToGo(arg, ctx)
			if err != nil {
				return "", dtype.None(), err
			}
			argStrs = append(argStrs, ", "+argStr)
		}
		return fmt.Sprintf("%s.Realm().New(%q, %s%s)", ctx.WorldRef, expr.Path, ctx.UsrRef(), strings.Join(argStrs, "")), dtype.Path(expr.Path), nil
	case ast.ExprTypeGetNonLocal:
		getExpr, _, ftype, ok := ctx.ResolveNonLocal(expr.Str)
		if ok {
			return getExpr, ftype, nil
		}
		return "", dtype.None(), fmt.Errorf("cannot find nonlocal %s at %v", expr.Str, expr.SourceLoc)
	case ast.ExprTypeGetLocal:
		if expr.Str == "." {
			if ctx.Result == "" {
				return "", dtype.None(), fmt.Errorf("attempt to use . outside of a proc at %v", expr.SourceLoc)
			}
			return ctx.Result, dtype.Any(), nil
		}
		vtype, ok := ctx.VarTypes[expr.Str]
		if !ok {
			return "", dtype.None(), fmt.Errorf("unexpectedly could not find type for var %q at %v ... there may be a ast.bug", expr.Str, expr.SourceLoc)
		}
		return LocalVariablePrefix + expr.Str, vtype, nil
	case ast.ExprTypeGetField:
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
	case ast.ExprTypeStringConcat:
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

func StatementToGo(statement ast.Statement, ctx CodeGenContext) (lines []string, err error) {
	switch statement.Type {
	case ast.StatementTypeIf:
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
	case ast.StatementTypeForList:
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
	case ast.StatementTypeWrite:
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
	case ast.StatementTypeReturn:
		if ctx.Result == "" {
			panic("should never have an empty result name here")
		}
		util.FIXME("support returning values")
		return []string{
			"return " + ctx.Result,
		}, nil
	case ast.StatementTypeEvaluate:
		value, _, err := ExprToGo(statement.To, ctx)
		if err != nil {
			return nil, err
		}
		return []string{
			"_ = " + value,
		}, nil
	case ast.StatementTypeAssign:
		value, _, err := ExprToGo(statement.From, ctx)
		if err != nil {
			return nil, err
		}
		if statement.To.Type == ast.ExprTypeGetNonLocal {
			name := statement.To.Str
			_, setExpr, _, ok := ctx.ResolveNonLocal(name)
			if ok {
				util.FIXME("should any typechecking happen here?")
				return []string{
					setExpr(value),
				}, nil
			}
			return nil, fmt.Errorf("cannot resolve nonlocal %q at %v", name, statement.SourceLoc)
		} else if statement.To.Type == ast.ExprTypeGetLocal {
			assign, _, err := ExprToGo(statement.To, ctx)
			if err != nil {
				return nil, err
			}
			return []string{
				fmt.Sprintf("%s = %s", assign, value),
			}, nil
		} else {
			return nil, fmt.Errorf("not sure how to handle assignment to expression %v at %v", statement.To, statement.SourceLoc)
		}
	case ast.StatementTypeDel:
		value, _, err := ExprToGo(statement.From, ctx)
		if err != nil {
			return nil, err
		}
		return []string{
			fmt.Sprintf("types.Del(%s)", value),
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

func ParseSrcSetting(expr ast.Expression, stype ast.StatementType) (types.SrcSetting, error) {
	var sst types.SrcSettingType
	var dist int
	switch expr.Type {
	case ast.ExprTypeCall:
		for _, name := range expr.Names {
			if name != "" {
				return types.SrcSetting{}, fmt.Errorf("cannot handle keyword arguments in src setting at %v", expr.SourceLoc)
			}
		}
		if expr.Children[0].Type != ast.ExprTypeGetNonLocal || expr.Children[0].Str != "oview" {
			return types.SrcSetting{}, fmt.Errorf("expected call only to oview, not %q, in src setting at %v", expr.Children[0].Str, expr.Children[0].SourceLoc)
		}
		if len(expr.Children) > 2 {
			return types.SrcSetting{}, fmt.Errorf("expected call to have 0-1 arguments when in src setting at %v", expr.SourceLoc)
		}
		sst = types.SrcSettingTypeOView
		if len(expr.Children) == 2 {
			if expr.Children[1].Type != ast.ExprTypeIntegerLiteral {
				return types.SrcSetting{}, fmt.Errorf("expected integer literal in oview parameter at %v", expr.Children[1].SourceLoc)
			}
			dist = int(expr.Children[1].Integer)
			if dist < 0 || int64(dist) != expr.Children[1].Integer {
				return types.SrcSetting{}, fmt.Errorf("integer literal out of range at %v", expr.Children[1].SourceLoc)
			}
		} else {
			dist = types.SrcDistUnspecified
		}
	default:
		return types.SrcSetting{}, fmt.Errorf("unexpected expression %v while parsing src setting at %v", expr, expr.SourceLoc)
	}
	return types.SrcSetting{
		Type: sst,
		Dist: dist,
		In:   stype == ast.StatementTypeSetIn,
	}, nil
}

func ParseSettings(dt *gen.DefinedTree, typePath path.TypePath, body []ast.Statement) (types.ProcSettings, []ast.Statement, error) {
	settings := types.ProcSettings{}
	settings.Src = DefaultSrcSetting(dt, typePath)
	setSrc := false
	for len(body) > 0 && (body[0].Type == ast.StatementTypeSetIn || body[0].Type == ast.StatementTypeSetTo) {
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

func FuncBodyToGo(body []ast.Statement, ctx CodeGenContext) (lines []string, err error) {
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
		if statement.Type == ast.StatementTypeReturn {
			hadReturn = true
		}
	}
	if !hadReturn {
		if ctx.Result == "" {
			panic("result should not be nil here")
		}
		lines = append(lines, "return "+ctx.Result)
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
