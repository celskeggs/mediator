package datum

import (
	"github.com/celskeggs/mediator/platform/debug"
	"runtime"
)

// long-lived pointers to Datums need to be Refs
type Ref struct {
	datum *Datum
}

func (d *Ref) Dereference() IDatum {
	if d == nil {
		return nil
	}
	impl := d.datum.Impl
	if impl == nil {
		panic("nil Impl")
	}
	return impl
}

func finalizeRef(r *Ref) {
	r.datum.decrementRefs()
	r.datum = nil
}

type IDatum interface {
	RawClone() IDatum
	Clone() IDatum
	AsDatum() *Datum
	Dump(*debug.Output)
	Reference() *Ref
	Realm() *Realm
}

var _ IDatum = &Datum{}

// invariant: Impl should point back at the top level of the Datum's container struct
type Datum struct {
	Impl IDatum
	Type TypePath

	// refcount is the number of Refs to this Datum. the datum only counts as being in the realm when this is nonzero.
	refCount uint
	realm    *Realm
}

func AssertConsistent(data ...IDatum) {
	for _, datum := range data {
		if datum != nil && datum.AsDatum().Impl != datum {
			panic("inconsistent datum")
		}
	}
}

func (d *Datum) decrementRefs() {
	if d.refCount == 0 {
		panic("refcount should not already have been zero")
	}
	d.refCount -= 1
	if d.refCount == 0 {
		d.realm.remove(d)
	}
}

func (d *Datum) Reference() *Ref {
	if d.refCount == 0 {
		d.realm.add(d)
	}
	// we're ignoring the possibility of overflow
	d.refCount += 1
	ref := &Ref{
		datum: d,
	}
	runtime.SetFinalizer(ref, finalizeRef)
	return ref
}

func (d Datum) RawClone() IDatum {
	return &d
}

func (d *Datum) AsDatum() *Datum {
	return d
}

func (d *Datum) Clone() IDatum {
	if d.Impl == nil {
		panic("reference is nil when cloning")
	}
	if d.realm == nil {
		panic("realm is nil when cloning")
	}
	cloned := d.Impl.RawClone()
	setImpl(cloned)
	cloned.AsDatum().refCount = 0
	return cloned
}

func (d *Datum) Dump(o *debug.Output) {
	debug.DumpReflect(d.Reference, o)
}

func (d *Datum) Realm() *Realm {
	return d.realm
}

func setImpl(datum IDatum) {
	if datum.AsDatum().realm == nil {
		panic("no realm found on datum")
	}
	datum.AsDatum().Impl = datum
}
