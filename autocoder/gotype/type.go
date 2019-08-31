package gotype

import "fmt"

type TypeType int

const (
	TypeTypeNone      TypeType = iota
	TypeTypeImported
	TypeTypeLocal
	TypeTypePtr
	TypeTypePrimitive
)

type Type struct {
	Type    TypeType
	Package string
	RawName string
	Inner   *Type
}

func (t Type) Ptr() Type {
	return Type{
		Type:  TypeTypePtr,
		Inner: &t,
	}
}

func (t Type) Name() string {
	switch t.Type {
	case TypeTypeImported:
		return t.RawName
	case TypeTypeLocal:
		return t.RawName
	case TypeTypePtr:
		return t.Inner.Name()
	case TypeTypePrimitive:
		return t.RawName
	default:
		panic(fmt.Sprintf("unrecognized type type %d", t.Type))
	}
}

func (t Type) String() string {
	switch t.Type {
	case TypeTypeImported:
		return t.Package + "." + t.RawName
	case TypeTypeLocal:
		return t.RawName
	case TypeTypePtr:
		return "*" + t.Inner.Name()
	case TypeTypePrimitive:
		return t.RawName
	default:
		panic(fmt.Sprintf("unrecognized type type %d", t.Type))
	}
}

func (t Type) IsPtr() bool {
	return t.Type == TypeTypePtr
}

func (t Type) UnwrapPtr() Type {
	if !t.IsPtr() {
		panic("not a pointer in UnwrapPtr")
	}
	return *t.Inner
}

func TypeBool() Type {
	return Type{
		Type:    TypeTypePrimitive,
		RawName: "bool",
	}
}

func TypeString() Type {
	return Type{
		Type:    TypeTypePrimitive,
		RawName: "string",
	}
}

func PackageType(pkg string, typeName string) Type {
	return Type{
		Type:    TypeTypeImported,
		Package: pkg,
		RawName: typeName,
	}
}

func LocalType(typeName string) Type {
	return Type{
		Type:    TypeTypeLocal,
		RawName: typeName,
	}
}
