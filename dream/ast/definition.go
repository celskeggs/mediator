package ast

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
)

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

type TypedName struct {
	Type dtype.DType
	Name string
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
