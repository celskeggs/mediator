package dtype

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/util"
)

type DKind int

const (
	KNone DKind = iota
	KAny
	KString
	KInteger
	KPath
)

type DType struct {
	kind DKind
	path path.TypePath
}

func (d DType) IsNone() bool {
	return d.kind == KNone
}

func (d DType) Path() path.TypePath {
	if d.kind != KPath {
		panic("not a path")
	}
	return d.path
}

func (d DType) IsInteger() bool {
	return d.kind == KInteger
}

func (d DType) IsString() bool {
	return d.kind == KString
}

func (d DType) IsAnyPath() bool {
	return d.kind == KPath
}

func (d DType) IsPath(path path.TypePath) bool {
	return d.kind == KPath && d.path.Equals(path)
}

func (d DType) IsPathConst(tp string) bool {
	return d.IsPath(path.ConstTypePath(tp))
}

func (d DType) String() string {
	switch d.kind {
	case KNone:
		return "none"
	case KAny:
		return "any"
	case KString:
		return "string"
	case KInteger:
		return "integer"
	case KPath:
		return "path:" + d.path.String()
	default:
		panic(fmt.Sprintf("unknown dtype kind: %d", d.kind))
	}
}

func Any() DType {
	return DType{
		kind: KAny,
	}
}

func None() DType {
	return DType{
		kind: KNone,
	}
}

func String() DType {
	return DType{
		kind: KString,
	}
}

func Integer() DType {
	return DType{
		kind: KInteger,
	}
}

// assumes that the type path is for a datum, unlike FromPath
func Path(typePath path.TypePath) DType {
	return DType{
		kind: KPath,
		path: typePath,
	}
}

func ConstPath(typePath string) DType {
	return Path(path.ConstTypePath(typePath))
}

// decodes path if it's not a datum, unlike Path
func FromPath(typePath path.TypePath) DType {
	if !typePath.IsAbsolute {
		typePath = path.Root().Join(typePath)
	}
	if len(typePath.Segments) == 1 {
		util.FIXME("figure out what the right set of things here is")
		switch typePath.Segments[0] {
		case "string":
			return String()
		}
	}
	return Path(typePath)
}
