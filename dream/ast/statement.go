package ast

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"strings"
)

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
