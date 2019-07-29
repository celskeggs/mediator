package platform

import (
	"strings"
)

type TypePath string
type Type func() IDatum

type World struct {
	Name string
	Mob  TypePath
}

func (w *World) Dump(o *DebugOutput) {
	o.Header("world")
	o.Println("name: " + w.Name)
	o.Println("mob: " + string(w.Mob))
	o.Footer()
}

type Tree struct {
	World World
	types map[TypePath]IDatum
}

func NewBaseTree() (*Tree) {
	tree := &Tree{
		types: map[TypePath]IDatum{},
	}

	tree.set("/datum", setReference(&Datum{}))
	tree.set("/atom", setReference(&Atom{}))
	tree.set("/atom/movable", setReference(&AtomMovable{}))
	tree.set("/obj", setReference(&Obj{}))
	tree.set("/turf", setReference(&Turf{}))
	tree.set("/area", setReference(&Area{}))
	mob := setReference(&Mob{}).(*Mob)
	mob.Density = true
	tree.set("/mob", mob)
	return tree
}

func (path TypePath) IsValid() bool {
	sp := string(path)
	return strings.Count(sp, "/") < len(sp) &&
		strings.Count(sp, "//") == 0 &&
		sp[0] == '/' &&
		sp[len(sp)-1] != '/'
}

func (path TypePath) Validate() {
	if !path.IsValid() {
		panic("path is not valid: " + path)
	}
}

func (t *Tree) New(path TypePath) IDatum {
	return t.Get(path).Clone()
}

func (t *Tree) Exists(path TypePath) bool {
	if !path.IsValid() {
		return false
	}
	_, found := t.types[path]
	return found
}

func (t *Tree) Get(path TypePath) IDatum {
	path.Validate()
	datum, found := t.types[path]
	if !found {
		panic("missing type: " + string(path))
	}
	return datum
}

// also populates Type field
func (t *Tree) set(path TypePath, datum IDatum) {
	path.Validate()
	if datum.AsDatum().Type != "" && datum == t.Get(datum.AsDatum().Type) {
		panic("attempt to re-insert previously inserted datum")
	}
	datum.AsDatum().Type = path
	_, found := t.types[path]
	if found {
		panic("duplicate type: " + path)
	}
	t.types[path] = datum
}

func (t *Tree) Derive(parent TypePath, child TypePath) IDatum {
	childDatum := t.New(parent)
	t.set(child, childDatum)
	return childDatum
}

func (t *Tree) Dump(o *DebugOutput) {
	o.Header("tree")
	t.World.Dump(o)
	for path, datum := range t.types {
		o.Header("type: " + string(path))
		datum.Dump(o)
		o.Footer()
	}
	o.Footer()
}
