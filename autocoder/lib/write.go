package lib

import (
	"io"
	"github.com/hashicorp/go-multierror"
	"fmt"
	"unicode"
	"strings"
)

type genWriter struct {
	Output      io.Writer
	Errors      []error
	IndentCount int
	WasNewline  bool
}

func (w *genWriter) HasError() bool {
	return len(w.Errors) > 0
}

func (w *genWriter) Error() error {
	if len(w.Errors) == 1 {
		return w.Errors[0]
	} else if len(w.Errors) > 0 {
		return multierror.Append(nil, w.Errors...)
	}
	return nil
}

func (w *genWriter) AddError(err error) {
	if err != nil {
		w.Errors = append(w.Errors, err)
	}
}

func (w *genWriter) Indent() {
	w.IndentCount += 1
}

func (w *genWriter) Unindent() {
	if w.IndentCount <= 0 {
		panic("indent depth going negative")
	}
	w.IndentCount -= 1
}

func (w *genWriter) Write(format string, params ...interface{}) {
	var prefix string
	if w.WasNewline {
		prefix = strings.Repeat("\t", w.IndentCount)
		w.WasNewline = false
	}
	_, err := fmt.Fprintf(w.Output, prefix + format, params...)
	if err != nil {
		w.AddError(err)
	}
}

func (w *genWriter) WriteLn(format string, params ...interface{}) {
	w.Write(format, params...)
	w.Newline()
}

func (w *genWriter) Newline() {
	w.Write("\n")
	w.WasNewline = true
}

func IsValidIdentifier(s string) bool {
	for _, r := range s {
		if !(unicode.IsLetter(r) || r == '_' || unicode.IsNumber(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func (w *genWriter) Identifier(symbol string) string {
	if !IsValidIdentifier(symbol) {
		w.AddError(fmt.Errorf("invalid symbol: %s", symbol))
	}
	return symbol
}

func (w *genWriter) String(str string) string {
	chunks := []string{"\""}
	for _, r := range str {
		chunk := string([]rune{r})
		if r == '"' {
			chunk = "\\\""
		} else if r == '\\' {
			chunk = "\\\\"
		} else if r == '\n' {
			chunk = "\\n"
		} else if !unicode.IsPrint(r) {
			panic(fmt.Sprintf("unimplemented: stringification for rune %d", r))
		}
		chunks = append(chunks, chunk)
	}
	chunks = append(chunks, "\"")
	return strings.Join(chunks, "")
}

type codeContext struct {
	Self string
}

func (w *genWriter) StringTypes(goTypes []Type) string {
	var parts []string
	for _, goType := range goTypes {
		parts = append(parts, goType.String())
	}
	return strings.Join(parts, ", ")
}

func (w *genWriter) WriteElement(elem Element) {
	switch elem.Type {
	case ElemTypeInclude:
		w.WriteLn("%s", elem.GoType.String())
	default:
		panic(fmt.Sprintf("unrecognized element type: %d", elem.Type))
	}
}

func (w *genWriter) WriteExpression(expr Expression, ctx codeContext) {
	switch expr.Type {
	case ExprTypeLiteralStruct:
		goType := expr.GoType
		if goType.IsPtr() {
			w.Write("&")
			goType = expr.GoType.UnwrapPtr()
		}
		w.Write("%s {\n", goType.String())
		w.Indent()
		for _, kv := range expr.KVs {
			w.Write("%s: ", kv.Key)
			w.WriteExpression(kv.Value, ctx)
		}
		w.Unindent()
		w.Write("}")
	case ExprTypeField:
		w.WriteExpression(expr.Exprs[0], ctx)
		w.Write(".%s", w.Identifier(expr.FieldName))
	case ExprTypeInvoke:
		w.WriteExpression(expr.Exprs[0], ctx)
		w.Write(".%s(", w.Identifier(expr.FieldName))
		for i, expr := range expr.Exprs[1:] {
			if i > 1 {
				w.Write(", ")
			}
			w.WriteExpression(expr, ctx)
		}
		w.Write(")")
	case ExprTypeCast:
		w.WriteExpression(expr.Exprs[0], ctx)
		w.Write(".(%s)", expr.GoType.String())
	case ExprTypeSelf:
		w.Write("%s", w.Identifier(ctx.Self))
	case ExprTypeSelfRef:
		w.Write("&%s", w.Identifier(ctx.Self))
	default:
		panic(fmt.Sprintf("unrecognized expression type: %d", expr.Type))
	}
}

func (w *genWriter) WriteStatement(statement Statement, ctx codeContext) {
	switch statement.Type {
	case StatementTypeAssign:
		w.WriteExpression(statement.Lvalue, ctx)
		w.Write(" = ")
		w.WriteExpression(statement.Rvalue, ctx)
		w.Newline()
	case StatementTypeReturn:
		w.Write("return ")
		w.WriteExpression(statement.Rvalue, ctx)
		w.Newline()
	default:
		panic(fmt.Sprintf("unrecognized statement type: %d", statement.Type))
	}
}

func (w *genWriter) WriteDef(def Definition) {
	switch def.Type {
	case DefTypeStruct:
		w.WriteLn("type %s struct {", w.Identifier(def.Name))
		w.Indent()
		for _, element := range def.Elements {
			w.WriteElement(element)
		}
		w.Unindent()
		w.WriteLn("}")
	case DefTypeInterface:
		w.WriteLn("type %s interface {", w.Identifier(def.Name))
		w.Indent()
		for _, element := range def.Elements {
			w.WriteElement(element)
		}
		w.Unindent()
		w.WriteLn("}")
	case DefTypeGlobal:
		w.Write("var %s %s = ", w.Identifier(def.Name), def.GoType.String())
		w.WriteExpression(def.Initializer, codeContext{})
		w.Newline()
	case DefTypeFunctionOn:
		selfName := w.Identifier(strings.ToLower(def.GoType.Name()[0:1]))
		w.WriteLn("func (%s %s) %s(%s) (%s) {",
			selfName, def.GoType.String(), def.Name, w.StringTypes(def.Params), w.StringTypes(def.Results))
		w.Indent()
		for _, statement := range def.Code {
			w.WriteStatement(statement, codeContext{
				Self: selfName,
			})
		}
		w.Unindent()
		w.WriteLn("}")
	default:
		panic(fmt.Sprintf("unrecognized definition type %d", def.Type))
	}
}

func (w *genWriter) WriteImports(g *Generator) {
	w.Newline()
	w.WriteLn("import (")
	w.Indent()

	for importPath, _ := range g.Imports {
		w.WriteLn("%s", w.String(importPath))
	}

	w.Unindent()
	w.WriteLn(")")

	for _, definition := range g.Defs {
		w.Newline()
		w.WriteDef(definition)
	}
	w.Newline()
}

func (w *genWriter) WriteGenerator(g *Generator) {
	w.WriteLn("package %s", w.Identifier(g.Package))

	w.WriteImports(g)
}

func (g *Generator) WriteTo(output io.Writer) error {
	gw := &genWriter{
		Output: output,
		Errors: nil,
	}
	gw.WriteGenerator(g)
	return gw.Error()
}
