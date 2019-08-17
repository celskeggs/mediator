package lib

import (
	"fmt"
	"errors"
	"strings"
)

type Generator struct {
	Package string
	Imports map[string]*Import
	Defs    []Definition
	Errors  []error
}

func StartGeneration(pkg string) *Generator {
	return &Generator{
		Package: pkg,
		Imports: make(map[string]*Import),
	}
}

func (g *Generator) AddError(err error) {
	if err != nil {
		g.Errors = append(g.Errors, err)
	}
}

func (g *Generator) Import(path string) *Import {
	if _, ok := g.Imports[path]; !ok {
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			g.AddError(fmt.Errorf("invalid package path: %s", path))
		}
		g.Imports[path] = &Import{
			Path:    path,
			Package: parts[len(parts)-1],
		}
	}
	return g.Imports[path]
}

func (g *Generator) elementsDefinition(defType DefType, typeType TypeType, name string, elements []Element) Type {
	def := Definition{
		Type: defType,
		Name: name,
	}
	for _, element := range elements {
		g.AddError(element.ValidateIn(def))
		def.Elements = append(def.Elements, element)
	}
	g.Defs = append(g.Defs, def)
	return Type{
		Type:    typeType,
		RawName: name,
	}
}

func (g *Generator) Struct(name string, elements ...Element) Type {
	return g.elementsDefinition(DefTypeStruct, TypeTypeStruct, name, elements)
}

func (g *Generator) Interface(name string, elements ...Element) Type {
	return g.elementsDefinition(DefTypeInterface, TypeTypeInterface, name, elements)
}

func (g *Generator) Global(name string, varType Type, value Expression) {
	def := Definition{
		Type:        DefTypeGlobal,
		Name:        name,
		GoType:      varType,
		Initializer: value,
	}
	g.Defs = append(g.Defs, def)
}

func (g *Generator) FuncOn(structType Type, name string, params []Type, results []Type, block ...Statement) {
	def := Definition{
		Type:    DefTypeFunctionOn,
		GoType:  structType,
		Name:    name,
		Params:  params,
		Results: results,
		Code:    block,
	}
	g.Defs = append(g.Defs, def)
}

func (g *Generator) Include(includeType Type) Element {
	return Element{
		Type:   ElemTypeInclude,
		Name:   includeType.Name(),
		GoType: includeType,
	}
}

func (g *Generator) LiteralStructPtr(structType Type, kvs ...KeyValue) Expression {
	return Expression{
		Type:   ExprTypeLiteralStruct,
		GoType: structType.Ptr(),
		KVs:    kvs,
	}
}

func (g *Generator) Self() Expression {
	return Expression{
		Type: ExprTypeSelf,
	}
}

func (g *Generator) SelfRef() Expression {
	return Expression{
		Type: ExprTypeSelfRef,
	}
}

func (g *Generator) Assign(lvalue Expression, rvalue Expression) Statement {
	if !lvalue.IsLvalue() {
		g.AddError(errors.New("lvalue in assignment is not a valid lvalue"))
	}
	return Statement{
		Type:   StatementTypeAssign,
		Lvalue: lvalue,
		Rvalue: rvalue,
	}
}

func (g *Generator) Return(value Expression) Statement {
	return Statement{
		Type:   StatementTypeReturn,
		Rvalue: value,
	}
}

type TypeType int

const (
	TypeTypeNone      TypeType = iota
	TypeTypeStruct
	TypeTypeInterface
	TypeTypePtr
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
	case TypeTypeStruct:
		return t.RawName
	case TypeTypeInterface:
		return t.RawName
	case TypeTypePtr:
		return t.Inner.Name()
	default:
		panic(fmt.Sprintf("unrecognized type type %d", t.Type))
	}
}

func (t Type) String() string {
	switch t.Type {
	case TypeTypeStruct:
		fallthrough
	case TypeTypeInterface:
		if t.Package == "" {
			return t.RawName
		} else {
			return t.Package + "." + t.RawName
		}
	case TypeTypePtr:
		return "*" + t.Inner.Name()
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

type Import struct {
	Path    string
	Package string
}

func (i *Import) GetStruct(typeName string) Type {
	return Type{
		Type:    TypeTypeStruct,
		Package: i.Package,
		RawName: typeName,
	}
}

func (i *Import) GetInterface(typeName string) Type {
	return Type{
		Type:    TypeTypeInterface,
		Package: i.Package,
		RawName: typeName,
	}
}

type DefType int

const (
	DefTypeNone       DefType = iota
	DefTypeStruct
	DefTypeInterface
	DefTypeGlobal
	DefTypeFunctionOn
)

type Definition struct {
	Type        DefType
	Name        string
	GoType      Type
	Elements    []Element
	Initializer Expression
	Params      []Type
	Results     []Type
	Code        []Statement
}

func (def Definition) HasField(name string) bool {
	for _, elem := range def.Elements {
		if elem.Name == name {
			return true
		}
	}
	return false
}

type ElemType int

const (
	ElemTypeNone    ElemType = iota
	ElemTypeInclude
)

type Element struct {
	Type   ElemType
	Name   string
	GoType Type
}

func (e Element) ValidateIn(def Definition) error {
	switch e.Type {
	case ElemTypeInclude:
		if def.HasField(e.Name) {
			return fmt.Errorf("duplicate field: '%s'", e.Name)
		}
		return nil
	default:
		panic(fmt.Sprintf("unrecognized element type: %d", e.Type))
	}
}

type StatementType int

const (
	StatementTypeNone   StatementType = iota
	StatementTypeAssign
	StatementTypeReturn
)

type Statement struct {
	Type     StatementType
	Lvalue   Expression
	Rvalue   Expression
}

type ExprType int

const (
	ExprTypeNone          ExprType = iota
	ExprTypeLiteralStruct
	ExprTypeSelf
	ExprTypeSelfRef
	ExprTypeField
	ExprTypeInvoke
	ExprTypeCast
)

type Expression struct {
	Type      ExprType
	GoType    Type
	Exprs     []Expression
	FieldName string
	KVs       []KeyValue
}

func (e Expression) Field(field string) Expression {
	return Expression{
		Type:      ExprTypeField,
		Exprs:     []Expression{e},
		FieldName: field,
	}
}

func (e Expression) Invoke(name string, args ...Expression) Expression {
	exprs := append([]Expression{e}, args...)
	return Expression{
		Type:      ExprTypeInvoke,
		FieldName: name,
		Exprs:     exprs,
	}
}

func (e Expression) Cast(castType Type) Expression {
	return Expression{
		Type:   ExprTypeCast,
		GoType: castType,
		Exprs:  []Expression{e},
	}
}

func (e Expression) IsLvalue() bool {
	switch e.Type {
	case ExprTypeField:
		return true
	default:
		return false
	}
}

type KeyValue struct {
	Key   string
	Value Expression
}
