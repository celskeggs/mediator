package gotype

type Kind uint8

const (
	KindNone Kind = iota
	KindInterfaceAny
	KindExternal
	KindBool
	KindInt
	KindString
	KindFunc
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
	default:
		panic("unrecognized kind with ID=" + string(k))
	}
}

type GoType struct {
	Kind
	Param string
}

func (g GoType) String() string {
	switch g.Kind {
	case KindNone:
		return "<none>"
	case KindInterfaceAny:
		return "interface{}"
	case KindExternal:
		return g.Param
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindString:
		return "string"
	case KindFunc:
		return "func"
	default:
		panic("cannot stringify Go type with unrecognized kind " + g.Kind.String())
	}
}

func (g GoType) Equals(other GoType) bool {
	return g == other
}

func (g GoType) IsInterfaceAny() bool {
	return g.Kind == KindInterfaceAny
}

func (g GoType) IsExternal(other string) bool {
	return g.Kind == KindExternal && g.Param == other
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
		Kind:  KindExternal,
		Param: ref,
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

func Func() GoType {
	return GoType{
		Kind: KindFunc,
	}
}
