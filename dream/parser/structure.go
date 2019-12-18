package parser

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
		Type:      ExprTypeResourceLiteral,
		Str:       literal,
		SourceLoc: loc,
	}
}

func ExprIntegerLiteral(literal int64, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeIntegerLiteral,
		Integer:   literal,
		SourceLoc: loc,
	}
}

func ExprStringLiteral(literal string, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeStringLiteral,
		Str:       literal,
		SourceLoc: loc,
	}
}

func ExprPathLiteral(path path.TypePath, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypePathLiteral,
		Path:      path,
		SourceLoc: loc,
	}
}

func ExprStringConcat(exprs []DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeStringConcat,
		Children:  exprs,
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

func ExprGetField(expr DreamMakerExpression, field string, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeGetField,
		Str:       field,
		Children:  []DreamMakerExpression{expr},
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

func ExprNew(typepath path.TypePath, keywords []string, arguments []DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerExpression {
	return DreamMakerExpression{
		Type:      ExprTypeNew,
		Names:     keywords,
		Path:      typepath,
		Children:  arguments,
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

type DreamMakerStatement struct {
	Type      StatementType
	VarType   dtype.DType
	Name      string
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

func StatementSetTo(field string, expr DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeSetTo,
		Name:      field,
		To:        expr,
		SourceLoc: loc,
	}
}

func StatementSetIn(field string, expr DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeSetIn,
		Name:      field,
		To:        expr,
		SourceLoc: loc,
	}
}

func StatementEvaluate(expr DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeEvaluate,
		To:        expr,
		SourceLoc: loc,
	}
}

func StatementAssign(destination DreamMakerExpression, value DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeAssign,
		From:      value,
		To:        destination,
		SourceLoc: loc,
	}
}

func StatementDel(expr DreamMakerExpression, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeDel,
		From:      expr,
		SourceLoc: loc,
	}
}

func StatementForList(vartype dtype.DType, varname string, inExpr DreamMakerExpression, body []DreamMakerStatement, loc tokenizer.SourceLocation) DreamMakerStatement {
	return DreamMakerStatement{
		Type:      StatementTypeForList,
		VarType:   vartype,
		Name:      varname,
		From:      inExpr,
		Body:      body,
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

type DreamMakerTypedName struct {
	Type dtype.DType
	Name string
}

type DreamMakerDefType int

const (
	DefTypeNone DreamMakerDefType = iota
	DefTypeDefine
	DefTypeAssign
	DefTypeVarDef
	DefTypeProcDecl
	DefTypeVerbDecl
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

type DreamMakerDefinition struct {
	Type       DreamMakerDefType
	Path       path.TypePath
	VarType    path.TypePath
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

func DefVarDef(path path.TypePath, varType path.TypePath, variable string, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:      DefTypeVarDef,
		Path:      path,
		VarType:   varType,
		Variable:  variable,
		SourceLoc: location,
	}
}

func DefProcDecl(path path.TypePath, variable string, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:      DefTypeProcDecl,
		Path:      path,
		Variable:  variable,
		SourceLoc: location,
	}
}

func DefVerbDecl(path path.TypePath, variable string, location tokenizer.SourceLocation) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:      DefTypeVerbDecl,
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
	SearchPath  []string
	Maps        []string
}

func (dmf *DreamMakerFile) Extend(file *DreamMakerFile) {
	dmf.Definitions = append(dmf.Definitions, file.Definitions...)
	dmf.SearchPath = append(dmf.SearchPath, file.SearchPath...)
	dmf.Maps = append(dmf.Maps, file.Maps...)
}
