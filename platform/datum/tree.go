package datum

import "github.com/celskeggs/mediator/platform/debug"

// Types are just Datums used as prototypes
type TypeTree struct {
	types map[TypePath]IDatum
	realm Realm
}

func NewTypeTree() *TypeTree {
	tree := &TypeTree{
		types: make(map[TypePath]IDatum),
		realm: Realm{
			datums: map[*Datum]struct{}{},
		},
	}
	tree.RegisterStruct("/datum", &Datum{})
	return tree
}

func (t *TypeTree) Realm() *Realm {
	return &t.realm
}

func (t *TypeTree) New(path TypePath) IDatum {
	return t.Get(path).Clone()
}

func (t *TypeTree) DeriveNew(path TypePath) IDatum {
	return CloneForce(t.Get(path))
}

func (t *TypeTree) Exists(path TypePath) bool {
	if !path.IsValid() {
		return false
	}
	_, found := t.types[path]
	return found
}

func (t *TypeTree) Get(path TypePath) IDatum {
	path.Validate()
	datum, found := t.types[path]
	if !found {
		panic("missing type: " + string(path))
	}
	return datum
}

// also populates Type field
func (t *TypeTree) set(path TypePath, datum IDatum) {
	path.Validate()
	AssertConsistent(datum)
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

func (t *TypeTree) RegisterStruct(path TypePath, datum IDatum) (prototype IDatum) {
	datum.AsDatum().realm = t.Realm()
	setImpl(datum)
	t.set(path, datum)
	return datum
}

func (t *TypeTree) SetSingleton(path TypePath) {
	t.Get(path).AsDatum().singleton = true
}

func (t *TypeTree) Derive(parent TypePath, child TypePath) (prototype IDatum) {
	childDatum := t.DeriveNew(parent)
	t.set(child, childDatum)
	return childDatum
}

func (t *TypeTree) Dump(o *debug.Output) {
	o.Header("tree")
	for path, datum := range t.types {
		o.Header("type: " + string(path))
		datum.Dump(o)
		o.Footer()
	}
	o.Footer()
}
