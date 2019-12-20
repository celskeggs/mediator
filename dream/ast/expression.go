package ast

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
