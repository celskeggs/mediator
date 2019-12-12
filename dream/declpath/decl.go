package declpath

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/path"
)

type DeclType int

const (
	DeclInvalid DeclType = iota
	DeclPlain
	DeclVar
	DeclProc
	DeclVerb
)

func (t DeclType) String() string {
	switch t {
	case DeclInvalid:
		return "<invalid>"
	case DeclPlain:
		return "plain"
	case DeclVar:
		return "var"
	case DeclProc:
		return "proc"
	case DeclVerb:
		return "verb"
	default:
		panic("unrecognized decl type " + string(t))
	}
}

type DeclPath struct {
	Type   DeclType
	Prefix path.TypePath
	Suffix string // only valid when Type != DeclPlain
}

func Empty() DeclPath {
	return DeclPath{
		Type:   DeclPlain,
		Prefix: path.Empty(),
	}
}

func Root() DeclPath {
	return DeclPath{
		Type:   DeclPlain,
		Prefix: path.Root(),
	}
}

func (d DeclPath) CanAdd() bool {
	return d.Type == DeclPlain || d.Suffix == ""
}

func (d DeclPath) Add(segment string) DeclPath {
	if d.Type == DeclPlain {
		d.Prefix = d.Prefix.Add(segment)
	} else if d.Suffix == "" {
		d.Suffix = segment
	} else {
		panic("attempt to Add when all segments were already populated!")
	}
	return d
}

func (d DeclPath) CanAddDecl() bool {
	return d.Type == DeclPlain
}

func (d DeclPath) AddDecl(t DeclType) DeclPath {
	if t == DeclPlain || t == DeclInvalid {
		panic("invalid decl type to add")
	}
	if d.Type == DeclPlain {
		d.Type = t
	} else {
		panic("attempt to AddDecl when decl was already populated!")
	}
	return d
}

func (d DeclPath) IsEmpty() bool {
	return d.Type == DeclPlain && d.Prefix.IsEmpty()
}

func (d DeclPath) IsAbsolute() bool {
	return d.Prefix.IsAbsolute
}

func (d DeclPath) IsPlain() bool {
	return d.Type == DeclPlain
}

func (d DeclPath) Unwrap() path.TypePath {
	if !d.IsPlain() {
		panic("attempt to unwrap non-plain type")
	}
	if d.Suffix != "" {
		panic("suffix should never be populated on plain type")
	}
	return d.Prefix
}

func (d DeclPath) Join(o DeclPath) (DeclPath, bool) {
	if o.IsAbsolute() {
		return o, true
	} else if d.IsPlain() {
		o.Prefix = d.Unwrap().Join(o.Prefix)
		return o, true
	} else if !o.IsPlain() {
		return Empty(), false
	} else if o.IsEmpty() {
		return d, true
	} else if len(o.Prefix.Segments) != 1 || d.Suffix != "" {
		return Empty(), false
	} else {
		d.Suffix = o.Prefix.Segments[0]
		return d, true
	}
}

func (t DeclPath) IsVarDef() bool {
	return t.Type == DeclVar && t.Suffix != ""
}

func (t DeclPath) IsProcDef() bool {
	return t.Type == DeclProc && t.Suffix != ""
}

func (t DeclPath) IsVerbDef() bool {
	return t.Type == DeclVerb && t.Suffix != ""
}

func (t DeclPath) SplitDef() (path.TypePath, string) {
	if t.IsPlain() || t.Suffix == "" {
		panic("not a declaration in SplitDef")
	}
	return t.Prefix, t.Suffix
}

func (d DeclPath) String() string {
	if d.IsPlain() {
		return d.Prefix.String()
	} else if len(d.Prefix.Segments) == 0 {
		if d.Suffix == "" {
			return fmt.Sprintf("%v%v", d.Prefix, d.Type)
		} else {
			return fmt.Sprintf("%v%v/%v", d.Prefix, d.Type, d.Suffix)
		}
	} else {
		if d.Suffix == "" {
			return fmt.Sprintf("%v/%v", d.Prefix, d.Type)
		} else {
			return fmt.Sprintf("%v/%v/%s", d.Prefix, d.Type, d.Suffix)
		}
	}
}
