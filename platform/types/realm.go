package types

import "sync"

var TRACE = false

type Realm struct {
	busylock         sync.Mutex
	busy             bool
	datums           map[*Datum]struct{}
	deferredRemovals []*Datum
	worldRef         interface{}
	typeTree         TypeTree
}

func NewRealm(tree TypeTree) *Realm {
	return &Realm{
		datums:   map[*Datum]struct{}{},
		typeTree: nil,
	}
}

func (r *Realm) setBusy(busy bool) {
	r.busylock.Lock()
	defer r.busylock.Unlock()

	if r.busy && busy {
		panic("already busy!")
	}
	if !r.busy && !busy {
		panic("already not busy!")
	}
	r.busy = busy
	if !busy && r.deferredRemovals != nil {
		for _, dr := range r.deferredRemovals {
			if _, found := r.datums[dr]; !found {
				panic("deferred datum removal: datum not found in realm")
			}
			delete(r.datums, dr)
		}
		r.deferredRemovals = nil
	}
}

func (r *Realm) add(d *Datum) {
	r.busylock.Lock()
	defer r.busylock.Unlock()
	if _, found := r.datums[d]; found {
		panic("datum already found in realm")
	}
	// no busy check here; it's only really for garbage collection, which means remove
	r.datums[d] = struct{}{}
	if TRACE {
		println("added datum", d, "of type", d.Type(), "to realm")
	}
}

func (r *Realm) remove(d *Datum) {
	r.busylock.Lock()
	defer r.busylock.Unlock()
	if r.busy {
		r.deferredRemovals = append(r.deferredRemovals)
	} else {
		if _, found := r.datums[d]; !found {
			panic("datum not found in realm")
		}
		delete(r.datums, d)
	}
	if TRACE {
		println("removed datum", d, "of type", d.Type(), "from realm")
	}
}

func (r *Realm) FindAll(predicate func(*Datum) bool) (out []*Datum) {
	r.setBusy(true)
	defer r.setBusy(false)
	for datum := range r.datums {
		if predicate(datum) {
			out = append(out, datum)
		}
	}
	return out
}

// returns nil if not found
func (r *Realm) FindOne(predicate func(*Datum) bool) *Datum {
	r.setBusy(true)
	defer r.setBusy(false)
	for datum := range r.datums {
		if predicate(datum) {
			return datum
		}
	}
	return nil
}

func (realm *Realm) SetWorldRef(worldRef interface{}) {
	if worldRef == nil {
		panic("worldref cannot be nil")
	}
	if realm.worldRef != nil {
		panic("worldref already set")
	}
	realm.worldRef = worldRef
}

func (realm *Realm) WorldRef() interface{} {
	if realm.worldRef == nil {
		panic("no worldref registered")
	}
	return realm.worldRef
}

func (realm *Realm) NewDatum(impl DatumImpl) *Datum {
	return &Datum{
		impl:     impl,
		refCount: 0,
		realm:    realm,
	}
}

func (realm *Realm) IsSubType(path TypePath, of TypePath) bool {
	for path != "" {
		if path == of {
			return true
		}
		path = realm.typeTree.Parent(path)
	}
}
