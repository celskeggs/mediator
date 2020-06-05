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

type ProcArgumentAs int

const (
	ProcArgumentNone ProcArgumentAs = iota
	ProcArgumentText
	ProcArgumentMessage
	ProcArgumentNum
	ProcArgumentIcon
	ProcArgumentSound
	ProcArgumentFile
	ProcArgumentKey
	ProcArgumentNull
	ProcArgumentMob
	ProcArgumentObj
	ProcArgumentTurf
	ProcArgumentArea
	ProcArgumentAnything
)

func (t ProcArgumentAs) String() string {
	switch t {
	case ProcArgumentNone:
		return "none"
	case ProcArgumentText:
		return "text"
	case ProcArgumentMessage:
		return "message"
	case ProcArgumentNum:
		return "num"
	case ProcArgumentIcon:
		return "icon"
	case ProcArgumentSound:
		return "sound"
	case ProcArgumentFile:
		return "file"
	case ProcArgumentKey:
		return "key"
	case ProcArgumentNull:
		return "null"
	case ProcArgumentMob:
		return "mob"
	case ProcArgumentObj:
		return "obj"
	case ProcArgumentTurf:
		return "turf"
	case ProcArgumentArea:
		return "area"
	case ProcArgumentAnything:
		return "anything"
	default:
		panic(fmt.Sprintf("unexpected proc argument as-type %d", t))
	}
}

func ProcArgumentFromString(arg string) ProcArgumentAs {
	switch arg {
	case "text":
		return ProcArgumentText
	case "message":
		return ProcArgumentMessage
	case "num":
		return ProcArgumentNum
	case "icon":
		return ProcArgumentIcon
	case "sound":
		return ProcArgumentSound
	case "file":
		return ProcArgumentFile
	case "key":
		return ProcArgumentKey
	case "null":
		return ProcArgumentNull
	case "mob":
		return ProcArgumentMob
	case "obj":
		return ProcArgumentObj
	case "turf":
		return ProcArgumentTurf
	case "area":
		return ProcArgumentArea
	case "anything":
		return ProcArgumentAnything
	default:
		return ProcArgumentNone
	}
}

/*
 * So here's the weird thing with verb types.
 *
 * You can do verb/V(obj/test), and it does the same thing as verb/V(test as obj)
 * But it's ALSO the same as verb/V(obj/cheese/test)!
 *
 * So clearly these are different, orthogonal properties of a verb parameter type.
 */

type ProcArgument struct {
	Type dtype.DType
	Name string
	As   ProcArgumentAs // for verbs
}

type Definition struct {
	Type       DefType
	Path       path.TypePath
	VarType    path.TypePath
	Variable   string
	Expression Expression
	Arguments  []ProcArgument
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

func DefImplement(path path.TypePath, function string, arguments []ProcArgument, body []Statement, location tokenizer.SourceLocation) Definition {
	return Definition{
		Type:      DefTypeImplement,
		Path:      path,
		Variable:  function,
		Arguments: arguments,
		Body:      body,
		SourceLoc: location,
	}
}
