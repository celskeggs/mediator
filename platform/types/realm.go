package types

import (
	"sync"
)

var TRACE = false

type Realm struct {
	busylock         sync.Mutex
	busy             bool
	datums           map[*Datum]struct{}
	deferredRemovals []*Datum

	datumsByUID map[uint64]*Datum // locked under busy
	nextUID     uint64            // locked under uidlock
	uidlock     sync.Mutex

	worldRef         interface{}
	typeTree         TypeTree
	TreePrivateState interface{} // populated by the TypeTree
}

func NewRealm(tree TypeTree) *Realm {
	realm := &Realm{
		datumsByUID: map[uint64]*Datum{},
		nextUID:     1000,

		datums:   map[*Datum]struct{}{},
		typeTree: tree,
	}
	tree.PopulateRealm(realm)
	return realm
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
			if _, found := r.datumsByUID[dr.uid]; !found {
				panic("deferred datum removal: UID not found in realm")
			}
			delete(r.datumsByUID, dr.uid)
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
	if _, found := r.datumsByUID[d.uid]; found {
		panic("UID already found in realm")
	}
	// no busy check here; it's only really for garbage collection, which means remove
	r.datums[d] = struct{}{}
	r.datumsByUID[d.uid] = d
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
		if _, found := r.datumsByUID[d.uid]; !found {
			panic("UID not found in realm")
		}
		delete(r.datumsByUID, d.uid)
	}
	if TRACE {
		println("removed datum", d, "of type", d.Type(), "from realm")
	}
}

func (r *Realm) FindAll(predicate func(*Datum) bool) (out []Value) {
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
func (r *Realm) FindOne(predicate func(*Datum) bool) Value {
	r.setBusy(true)
	defer r.setBusy(false)
	for datum := range r.datums {
		if predicate(datum) {
			return datum
		}
	}
	return nil
}

func (r *Realm) Lookup(uid uint64) Value {
	if d, ok := r.datumsByUID[uid]; ok {
		return d
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

func (realm *Realm) NewPlain(path TypePath, params ...Value) *Datum {
	return realm.typeTree.New(realm, path, params...)
}

func (realm *Realm) New(path TypePath, usr *Datum, params ...Value) *Datum {
	datum := realm.typeTree.New(realm, path, params...)
	datum.Invoke(usr, "New", params...)
	return datum
}

func (realm *Realm) NewDatum(impl DatumImpl) *Datum {
	if impl == nil {
		panic("datum impl should never start as nil")
	}
	return &Datum{
		impl:     impl,
		refCount: 0,
		realm:    realm,
		uid:      realm.getNextUID(),
	}
}

func (realm *Realm) getNextUID() (id uint64) {
	realm.uidlock.Lock()
	defer realm.uidlock.Unlock()
	id = realm.nextUID
	realm.nextUID += 1
	return id
}

func (realm *Realm) IsSubType(path TypePath, of TypePath) bool {
	for path != "" {
		if path == of {
			return true
		}
		path = realm.typeTree.Parent(path)
	}
	return false
}
