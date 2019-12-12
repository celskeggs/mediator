package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
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
