package gotype

import (
	"fmt"
	"strings"
)

type Kind uint8

const (
	KindNone Kind = iota
	KindInterfaceAny
	KindExternal
	KindBool
	KindInt
	KindString
	KindFunc
	KindFuncAbstractParams
)

func (k Kind) String() string {
	switch k {
	case KindNone:
		return "KindNone"
	case KindInterfaceAny:
		return "KindInterfaceAny"
	case KindExternal:
		return "KindExternal"
	case KindBool:
		return "KindBool"
	case KindInt:
		return "KindInt"
	case KindString:
		return "KindString"
	case KindFunc:
		return "KindFunc"
	case KindFuncAbstractParams:
		return "KindFuncAbstractParams"
	default:
		panic("unrecognized kind with ID=" + string(k))
	}
}

type GoType struct {
	Kind
	Ref string
	Params []GoType
	Results []GoType
}

func typesToString(types []GoType) string {
	var parts []string
	for _, t := range types {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, ", ")
}

func (g GoType) String() string {
	switch g.Kind {
	case KindNone:
		return "<none>"
	case KindInterfaceAny:
		return "interface{}"
	case KindExternal:
		return g.Ref
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindString:
		return "string"
	case KindFunc:
		return fmt.Sprintf("func(%s)(%s)", typesToString(g.Params), typesToString(g.Results))
	case KindFuncAbstractParams:
		return fmt.Sprintf("func(...)(%s)", typesToString(g.Results))
	default:
		panic("cannot stringify Go type with unrecognized kind " + g.Kind.String())
	}
}

func (g GoType) Equals(other GoType) bool {
	if g.Kind != other.Kind {
		return false
	}
	switch g.Kind {
	case KindExternal:
		return g.Ref == other.Ref
	case KindFunc:
		if len(g.Params) != len(other.Params) || len(g.Results) != len(other.Results) {
			return false
		}
		for i, param := range g.Params {
			if !param.Equals(other.Params[i]) {
				return false
			}
		}
		for i, result := range g.Results {
			if !result.Equals(other.Results[i]) {
				return false
			}
		}
		return true
	case KindFuncAbstractParams:
		if len(g.Results) != len(other.Results) {
			return false
		}
		for i, result := range g.Results {
			if !result.Equals(other.Results[i]) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

func (g GoType) IsInterfaceAny() bool {
	return g.Kind == KindInterfaceAny
}

func (g GoType) IsExternal(other string) bool {
	return g.Kind == KindExternal && g.Ref == other
}

func (g GoType) IsBool() bool {
	return g.Kind == KindBool
}

func (g GoType) IsInt() bool {
	return g.Kind == KindInt
}

func (g GoType) IsString() bool {
	return g.Kind == KindString
}

func (g GoType) IsFunc() bool {
	return g.Kind == KindFunc || g.Kind == KindFuncAbstractParams
}

func (g GoType) IsFuncConcrete() bool {
	return g.Kind == KindFunc
}

func None() GoType {
	return GoType{
		Kind: KindNone,
	}
}

func InterfaceAny() GoType {
	return GoType{
		Kind: KindInterfaceAny,
	}
}

func External(ref string) GoType {
	return GoType{
		Kind: KindExternal,
		Ref:  ref,
	}
}

func Bool() GoType {
	return GoType{
		Kind: KindBool,
	}
}

func Int() GoType {
	return GoType{
		Kind: KindInt,
	}
}

func String() GoType {
	return GoType{
		Kind: KindString,
	}
}

func Func(params []GoType, results []GoType) GoType {
	return GoType{
		Kind:    KindFunc,
		Params:  params,
		Results: results,
	}
}

func FuncAbstractParams(results []GoType) GoType {
	return GoType{
		Kind:    KindFuncAbstractParams,
		Results: results,
	}
}
