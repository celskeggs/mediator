package declpath

import (
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
	Suffix path.TypePath // only valid when Type != DeclPlain; may never be absolute
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
	return d.Type == DeclPlain || d.Type == DeclVar || d.Suffix.IsEmpty()
}

func (d DeclPath) Add(segment string) DeclPath {
	if d.Type == DeclPlain {
		d.Prefix = d.Prefix.Add(segment)
	} else if d.Type == DeclVar || d.Suffix.IsEmpty() {
		// vars can have multiple segments for the sake of specifying a type
		d.Suffix = d.Suffix.Add(segment)
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
	if !d.Suffix.IsEmpty() {
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
	} else {
		for _, segment := range o.Prefix.Segments {
			if !d.CanAdd() {
				return Empty(), false
			}
			d = d.Add(segment)
		}
		return d, true
	}
}

func (t DeclPath) IsVarDef() bool {
	return t.Type == DeclVar && !t.Suffix.IsEmpty()
}

func (t DeclPath) IsProcDef() bool {
	return t.Type == DeclProc && !t.Suffix.IsEmpty()
}

func (t DeclPath) IsVerbDef() bool {
	return t.Type == DeclVerb && !t.Suffix.IsEmpty()
}

func (t DeclPath) SplitDef() (target path.TypePath, typePath path.TypePath, name string) {
	if t.IsPlain() || t.Suffix.IsEmpty() {
		panic("not a declaration in SplitDef")
	}
	typePath, name, err := t.Suffix.SplitLast()
	if err != nil {
		panic("unexpected internal error: " + err.Error())
	}
	if !typePath.IsEmpty() && t.Type != DeclVar {
		panic("unexpected state mismatch: non-var must not have non-empty type path")
	}
	return t.Prefix, typePath, name
}

func (d DeclPath) String() string {
	tmp := d.Prefix
	if d.Type != DeclPlain {
		tmp = tmp.Add(d.Type.String())
		tmp = tmp.Join(d.Suffix)
	}
	return tmp.String()
}
