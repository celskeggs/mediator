package parser

import (
	"github.com/celskeggs/mediator/dream/path"
	"fmt"
	"strings"
)

type ExprType uint8

const (
	ExprTypeNone            ExprType = iota
	ExprTypeResourceLiteral
	ExprTypePathLiteral
	ExprTypeIntegerLiteral
	ExprTypeStringLiteral
	ExprTypeStringConcat
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
	default:
		panic(fmt.Sprintf("unrecognized expression type: %d", et))
	}
}

type DreamMakerExpression struct {
	Type     ExprType
	Str      string
	Integer  int64
	Children []DreamMakerExpression
	Path     path.TypePath
}

func ExprNone() DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypeNone,
	}
}

func ExprResourceLiteral(literal string) DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypeResourceLiteral,
		Str:  literal,
	}
}

func ExprIntegerLiteral(literal int64) DreamMakerExpression {
	return DreamMakerExpression{
		Type:    ExprTypeIntegerLiteral,
		Integer: literal,
	}
}

func ExprStringLiteral(literal string) DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypeStringLiteral,
		Str:  literal,
	}
}

func ExprPathLiteral(path path.TypePath) DreamMakerExpression {
	return DreamMakerExpression{
		Type: ExprTypePathLiteral,
		Path: path,
	}
}

func ExprStringConcat(exprs []DreamMakerExpression) DreamMakerExpression {
	return DreamMakerExpression{
		Type:     ExprTypeStringConcat,
		Children: exprs,
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
		params = append(params, fmt.Sprintf("string='%s'", dme.Str))
	}
	if !dme.Path.IsEmpty() {
		params = append(params, fmt.Sprintf("path=%v", dme.Path))
	}
	for _, child := range dme.Children {
		params = append(params, child.String())
	}
	return fmt.Sprintf("%v(%s)", dme.Type, strings.Join(params, ", "))
}

type DreamMakerDefType int

const (
	DefTypeNone   DreamMakerDefType = iota
	DefTypeDefine
	DefTypeAssign
)

func (t DreamMakerDefType) String() string {
	switch t {
	case DefTypeNone:
		return "None"
	case DefTypeDefine:
		return "Define"
	case DefTypeAssign:
		return "Assign"
	default:
		panic(fmt.Sprintf("unexpected definition type %d", t))
	}
}

type DreamMakerDefinition struct {
	Type       DreamMakerDefType
	Path       path.TypePath
	Variable   string
	Expression DreamMakerExpression
}

func DefDefine(path path.TypePath) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type: DefTypeDefine,
		Path: path,
	}
}

func DefAssign(path path.TypePath, variable string, value DreamMakerExpression) DreamMakerDefinition {
	return DreamMakerDefinition{
		Type:       DefTypeAssign,
		Path:       path,
		Variable:   variable,
		Expression: value,
	}
}

type DreamMakerFile struct {
	Definitions []DreamMakerDefinition
}
