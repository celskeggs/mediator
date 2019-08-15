package datum

import "sync"

var TRACE = false

type Realm struct {
	busylock         sync.Mutex
	busy             bool
	datums           map[*Datum]struct{}
	deferredRemovals []*Datum
	worldRef         interface{}
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
			delete(r.datums, dr)
		}
		r.deferredRemovals = nil
	}
}

func (r *Realm) add(d *Datum) {
	if _, found := r.datums[d]; found {
		panic("datum already found in realm")
	}
	// no busy check here; it's only really for garbage collection, which means remove
	r.datums[d] = struct{}{}
	if TRACE {
		println("added datum", d, "of type", d.Type, "to realm")
	}
}

func (r *Realm) remove(d *Datum) {
	r.busylock.Lock()
	defer r.busylock.Unlock()
	if _, found := r.datums[d]; !found {
		panic("datum not found in realm")
	}
	if r.busy {
		r.deferredRemovals = append(r.deferredRemovals)
	} else {
		delete(r.datums, d)
	}
	if TRACE {
		println("removed datum", d, "of type", d.Type, "from realm")
	}
}

func (r *Realm) FindAll(predicate func(IDatum) bool) (out []IDatum) {
	r.setBusy(true)
	defer r.setBusy(false)
	for datum := range r.datums {
		if predicate(datum.impl) {
			out = append(out, datum.impl)
		}
	}
	return out
}

// returns nil if not found
func (r *Realm) FindOne(predicate func(IDatum) bool) IDatum {
	r.setBusy(true)
	defer r.setBusy(false)
	for datum := range r.datums {
		if predicate(datum.impl) {
			return datum.impl
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
