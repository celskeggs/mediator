package ast

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
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
	ExprTypeStringMacro
	ExprTypeStringConcat
	ExprTypeGetLocal
	ExprTypeGetNonLocal
	ExprTypeGetField
	ExprTypeBooleanNot
	ExprTypeCall
	ExprTypeNew
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
	case ExprTypeStringMacro:
		return "StringMacro"
	case ExprTypeStringConcat:
		return "StringConcat"
	case ExprTypeGetLocal:
		return "GetLocal"
	case ExprTypeGetNonLocal:
		return "GetNonLocal"
	case ExprTypeGetField:
		return "GetField"
	case ExprTypeBooleanNot:
		return "BooleanNot"
	case ExprTypeCall:
		return "Call"
	case ExprTypeNew:
		return "New"
	default:
		panic(fmt.Sprintf("unrecognized expression type: %d", et))
	}
}

type Expression struct {
	Type      ExprType
	Str       string
	Integer   int64
	Names     []string
	Children  []Expression
	Path      path.TypePath
	SourceLoc tokenizer.SourceLocation
}

func ExprNone() Expression {
	return Expression{
		Type: ExprTypeNone,
	}
}

func ExprResourceLiteral(literal string, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeResourceLiteral,
		Str:       literal,
		SourceLoc: loc,
	}
}

func ExprIntegerLiteral(literal int64, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeIntegerLiteral,
		Integer:   literal,
		SourceLoc: loc,
	}
}

func ExprStringLiteral(literal string, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeStringLiteral,
		Str:       literal,
		SourceLoc: loc,
	}
}

func ExprPathLiteral(path path.TypePath, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypePathLiteral,
		Path:      path,
		SourceLoc: loc,
	}
}

func ExprStringMacro(macro string, expr Expression, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeStringMacro,
		Str:       macro,
		Children:  []Expression{expr},
		SourceLoc: loc,
	}
}

func ExprStringConcat(exprs []Expression, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeStringConcat,
		Children:  exprs,
		SourceLoc: loc,
	}
}

func ExprGetLocal(name string, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeGetLocal,
		Str:       name,
		SourceLoc: loc,
	}
}

func ExprGetNonLocal(name string, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeGetNonLocal,
		Str:       name,
		SourceLoc: loc,
	}
}

func ExprGetField(expr Expression, field string, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeGetField,
		Str:       field,
		Children:  []Expression{expr},
		SourceLoc: loc,
	}
}

func ExprBooleanNot(expr Expression, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeBooleanNot,
		Children:  []Expression{expr},
		SourceLoc: loc,
	}
}

func ExprCall(expr Expression, keywords []string, arguments []Expression, loc tokenizer.SourceLocation) Expression {
	children := []Expression{expr}
	children = append(children, arguments...)
	return Expression{
		Type:      ExprTypeCall,
		Names:     keywords,
		Children:  children,
		SourceLoc: loc,
	}
}

func ExprNew(typepath path.TypePath, keywords []string, arguments []Expression, loc tokenizer.SourceLocation) Expression {
	return Expression{
		Type:      ExprTypeNew,
		Names:     keywords,
		Path:      typepath,
		Children:  arguments,
		SourceLoc: loc,
	}
}

func (dme Expression) IsNone() bool {
	return dme.Type == ExprTypeNone
}

func (dme Expression) String() string {
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
	StatementTypeSetIn
	StatementTypeSetTo
	StatementTypeEvaluate
	StatementTypeAssign
	StatementTypeDel
	StatementTypeForList
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
	case StatementTypeSetIn:
		return "SetIn"
	case StatementTypeSetTo:
		return "SetTo"
	case StatementTypeEvaluate:
		return "Evaluate"
	case StatementTypeAssign:
		return "Assign"
	case StatementTypeDel:
		return "Del"
	case StatementTypeForList:
		return "ForList"
	default:
		panic(fmt.Sprintf("unrecognized statement type: %d", et))
	}
}

type Statement struct {
	Type      StatementType
	VarType   dtype.DType
	Name      string
	From      Expression
	To        Expression
	Body      []Statement
	SourceLoc tokenizer.SourceLocation
}

func StatementNone() Statement {
	return Statement{
		Type: StatementTypeNone,
	}
}

func StatementWrite(destination Expression, value Expression, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeWrite,
		From:      value,
		To:        destination,
		SourceLoc: loc,
	}
}

func StatementIf(condition Expression, body []Statement, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeIf,
		From:      condition,
		Body:      body,
		SourceLoc: loc,
	}
}

func StatementReturn(loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeReturn,
		SourceLoc: loc,
	}
}

func StatementSetTo(field string, expr Expression, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeSetTo,
		Name:      field,
		To:        expr,
		SourceLoc: loc,
	}
}

func StatementSetIn(field string, expr Expression, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeSetIn,
		Name:      field,
		To:        expr,
		SourceLoc: loc,
	}
}

func StatementEvaluate(expr Expression, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeEvaluate,
		To:        expr,
		SourceLoc: loc,
	}
}

func StatementAssign(destination Expression, value Expression, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeAssign,
		From:      value,
		To:        destination,
		SourceLoc: loc,
	}
}

func StatementDel(expr Expression, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeDel,
		From:      expr,
		SourceLoc: loc,
	}
}

func StatementForList(vartype dtype.DType, varname string, inExpr Expression, body []Statement, loc tokenizer.SourceLocation) Statement {
	return Statement{
		Type:      StatementTypeForList,
		VarType:   vartype,
		Name:      varname,
		From:      inExpr,
		Body:      body,
		SourceLoc: loc,
	}
}

func (dms Statement) IsNone() bool {
	return dms.Type == StatementTypeNone
}

func (dms Statement) String() string {
	var params []string
	if !dms.From.IsNone() {
		params = append(params, fmt.Sprintf("from=%v", dms.From))
	}
	if !dms.To.IsNone() {
		params = append(params, fmt.Sprintf("to=%v", dms.To))
	}
	if !dms.VarType.IsNone() {
		params = append(params, fmt.Sprintf("vartype=%v", dms.VarType))
	}
	if dms.Name != "" {
		params = append(params, fmt.Sprintf("name=%q", dms.Name))
	}
	if len(dms.Body) > 0 {
		for _, statement := range dms.Body {
			params = append(params, statement.String())
		}
	}
	return fmt.Sprintf("%v(%s)", dms.Type, strings.Join(params, ", "))
}

type TypedName struct {
	Type dtype.DType
	Name string
}

type DefType int

const (
	DefTypeNone DefType = iota
	DefTypeDefine
	DefTypeAssign
	DefTypeVarDef
	DefTypeProcDecl
	DefTypeVerbDecl
	DefTypeImplement
)

func (t DefType) String() string {
	switch t {
	case DefTypeNone:
		return "None"
	case DefTypeDefine:
		return "Define"
	case DefTypeAssign:
		return "Assign"
	case DefTypeVarDef:
		return "VarDef"
	case DefTypeProcDecl:
		return "ProcDecl"
	case DefTypeVerbDecl:
		return "VerbDecl"
	case DefTypeImplement:
		return "Implement"
	default:
		panic(fmt.Sprintf("unexpected definition type %d", t))
	}
}

type Definition struct {
	Type       DefType
	Path       path.TypePath
	VarType    path.TypePath
	Variable   string
	Expression Expression
	Arguments  []TypedName
	Body       []Statement
	SourceLoc  tokenizer.SourceLocation
}

func DefDefine(path path.TypePath, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:      DefTypeDefine,
		Path:      path,
		SourceLoc: location,
	}
}

func DefAssign(path path.TypePath, variable string, value Expression, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:       DefTypeAssign,
		Path:       path,
		Variable:   variable,
		Expression: value,
		SourceLoc:  location,
	}
}

func DefVarDef(path path.TypePath, varType path.TypePath, variable string, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:      DefTypeVarDef,
		Path:      path,
		VarType:   varType,
		Variable:  variable,
		SourceLoc: location,
	}
}

func DefProcDecl(path path.TypePath, variable string, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:      DefTypeProcDecl,
		Path:      path,
		Variable:  variable,
		SourceLoc: location,
	}
}

func DefVerbDecl(path path.TypePath, variable string, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:      DefTypeVerbDecl,
		Path:      path,
		Variable:  variable,
		SourceLoc: location,
	}
}

func DefImplement(path path.TypePath, function string, arguments []TypedName, body []Statement, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:      DefTypeImplement,
		Path:      path,
		Variable:  function,
		Arguments: arguments,
		Body:      body,
		SourceLoc: location,
	}
}

type File struct {
	Definitions []Definition
	SearchPath  []string
	Maps        []string
}

func (f *File) Extend(file *File) {
	f.Definitions = append(f.Definitions, file.Definitions...)
	f.SearchPath = append(f.SearchPath, file.SearchPath...)
	f.Maps = append(f.Maps, file.Maps...)
}
