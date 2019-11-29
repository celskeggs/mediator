package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"strings"
)

type ExprType uint8

const (
	ExprTypeNone ExprType = iota
	ExprTypeResourceLiteral
	ExprTypePathLiteral
	ExprTypeIntegerLiteral
	ExprTypeStringLiteral
	ExprTypeStringConcat
	ExprTypeGetLocal
	ExprTypeGetNonLocal
	ExprTypeBooleanNot
	ExprTypeCall
)

func (et ExprType) String() string {
	switch et {
	case ExprTypeNone:
		return "None"
	case ExprTypeResourceLiteral:
		return "ResourceLiteral"
	case ExprTypePathLiteral:
		return "PathLiteral"
	case ExprTypeIntegerLiteral:
		return "IntegerLiteral"
	case ExprTypeStringLiteral:
		return "StringLiteral"
	case ExprTypeStringConcat:
		return "StringConcat"
	case ExprTypeGetLocal:
		return "GetLocal"
	case ExprTypeGetNonLocal:
		return "GetNonLocal"
	case ExprTypeBooleanNot:
		return "BooleanNot"
	case ExprTypeCall:
		return "Call"
	default:
		panic(fmt.Sprintf("unrecognized expression type: %d", et))
	}
}

type DreamMakerExpression struct {
	Type      ExprType
	Str       string
	Integer   int64
	Names     []string
	Children  []DreamMakerExpression
	Path      path.TypePath
	SourceLoc tokenizer.SourceLocation
}

func ExprNone() DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypeNone,
	}
}

func ExprResourceLiteral(literal string, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypeResourceLiteral,
		Str:  literal,
		SourceLoc: loc,
	}
}

func ExprIntegerLiteral(literal int64, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:    ExprTypeIntegerLiteral,
		Integer: literal,
		SourceLoc: loc,
	}
}

func ExprStringLiteral(literal string, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypeStringLiteral,
		Str:  literal,
		SourceLoc: loc,
	}
}

func ExprPathLiteral(path path.TypePath, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypePathLiteral,
		Path: path,
		SourceLoc: loc,
	}
}

func ExprStringConcat(exprs []DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:     ExprTypeStringConcat,
		Children: exprs,
		SourceLoc: loc,
	}
}

func ExprGetLocal(name string, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeGetLocal,
		Str:       name,
		SourceLoc: loc,
	}
}

func ExprGetNonLocal(name string, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeGetNonLocal,
		Str:       name,
		SourceLoc: loc,
	}
}

func ExprBooleanNot(expr DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeBooleanNot,
		Children:  []DreamMakerExpression{expr},
		SourceLoc: loc,
	}
}

func ExprCall(expr DreamMakerExpression, keywords []string, arguments []DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerExpression {
	children := []DreamMakerExpression{expr}
	children = append(children, arguments...)
	return DreamMakerExpression{
		Type:      ExprTypeCall,
		Names:     keywords,
		Children:  children,
		SourceLoc: loc,
	}
}

func (dme DreamMakerExpression) IsNone() bool {
	return dme.Type == ExprTypeNone
}

func (dme DreamMakerExpression) String() string {
	var params []string
	if dme.Integer != 0 || dme.Type == ExprTypeIntegerLiteral {
		params = append(params, fmt.Sprintf("integer=%d", dme.Integer))
	}
	if dme.Str != "" || (dme.Type == ExprTypeStringLiteral || dme.Type == ExprTypeResourceLiteral) {
		params = append(params, fmt.Sprintf("string=%q", dme.Str))
	}
	if !dme.Path.IsEmpty() {
		params = append(params, fmt.Sprintf("path=%v", dme.Path))
	}
	for _, name := range dme.Names {
		params = append(params, fmt.Sprintf("name=%q", name))
	}
	for _, child := range dme.Children {
		params = append(params, child.String())
	}
	return fmt.Sprintf("%v(%s)", dme.Type, strings.Join(params, ", "))
}

type StatementType uint8

const (
	StatementTypeNone StatementType = iota
	StatementTypeWrite
	StatementTypeIf
	StatementTypeReturn
)

func (et StatementType) String() string {
	switch et {
	case StatementTypeNone:
		return "None"
	case StatementTypeWrite:
		return "Write"
	case StatementTypeIf:
		return "If"
	case StatementTypeReturn:
		return "Return"
	default:
		panic(fmt.Sprintf("unrecognized statement type: %d", et))
	}
}

type DreamMakerStatement struct {
	Type      StatementType
	From      DreamMakerExpression
	To        DreamMakerExpression
	Body      []DreamMakerStatement
	SourceLoc tokenizer.SourceLocation
}

func StatementNone() DreamMakerStatement {
	return DreamMakerStatement{
		Type: StatementTypeNone,
	}
}

func StatementWrite(destination DreamMakerExpression, value DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeWrite,
		From:      value,
		To:        destination,
		SourceLoc: loc,
	}
}

func StatementIf(condition DreamMakerExpression, body []DreamMakerStatement, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeIf,
		From:      condition,
		Body:      body,
		SourceLoc: loc,
	}
}

func StatementReturn(loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeReturn,
		SourceLoc: loc,
	}
}

func (dms DreamMakerStatement) IsNone() bool {
	return dms.Type == StatementTypeNone
}

func (dms DreamMakerStatement) String() string {
	var params []string
	if !dms.From.IsNone() {
		params = append(params, fmt.Sprintf("from=%v", dms.From))
	}
	if !dms.To.IsNone() {
		params = append(params, fmt.Sprintf("to=%v", dms.To))
	}
	if len(dms.Body) > 0 {
		for _, statement := range dms.Body {
			params = append(params, statement.String())
		}
	}
	return fmt.Sprintf("%v(%s)", dms.Type, strings.Join(params, ", "))
}

type DreamMakerTypedName struct {
	Type path.TypePath // always absolute; root path for "no type"
	Name string
}

type DreamMakerDefType int

const (
	DefTypeNone DreamMakerDefType = iota
	DefTypeDefine
	DefTypeAssign
	DefTypeVarDef
	DefTypeImplement
)

func (t DreamMakerDefType) String() string {
	switch t {
	case DefTypeNone:
		return "None"
	case DefTypeDefine:
		return "Define"
	case DefTypeAssign:
		return "Assign"
	case DefTypeVarDef:
		return "VarDef"
	case DefTypeImplement:
		return "Implement"
	default:
		panic(fmt.Sprintf("unexpected definition type %d", t))
	}
}

type DreamMakerDefinition struct {
	Type       DreamMakerDefType
	Path       path.TypePath
	Variable   string
	Expression DreamMakerExpression
	Arguments  []DreamMakerTypedName
	Body       []DreamMakerStatement
	SourceLoc  tokenizer.SourceLocation
}

func DefDefine(path path.TypePath, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:      DefTypeDefine,
		Path:      path,
		SourceLoc: location,
	}
}

func DefAssign(path path.TypePath, variable string, value DreamMakerExpression, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:       DefTypeAssign,
		Path:       path,
		Variable:   variable,
		Expression: value,
		SourceLoc:  location,
	}
}

func DefVarDef(path path.TypePath, variable string, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:      DefTypeVarDef,
		Path:      path,
		Variable:  variable,
		SourceLoc: location,
	}
}

func DefImplement(path path.TypePath, function string, arguments []DreamMakerTypedName, body []DreamMakerStatement, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:      DefTypeImplement,
		Path:      path,
		Variable:  function,
		Arguments: arguments,
		Body:      body,
		SourceLoc: location,
	}
}

type DreamMakerFile struct {
	Definitions []DreamMakerDefinition
}

func (dmf *DreamMakerFile) Extend(file *DreamMakerFile) {
	dmf.Definitions = append(dmf.Definitions, file.Definitions...)
}
