package types

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/debug"
	"github.com/celskeggs/mediator/util"
	"runtime"
)

// long-lived pointers to Values need to be Refs
type Ref struct {
	v Value
}

func (d *Ref) Dereference() Value {
	if d == nil {
		return nil
	}
	if d.v == nil {
		panic("nil value during dereference")
	}
	return d.v
}

func finalizeRef(r *Ref) {
	d := r.v.(*Datum)
	if d == nil {
		panic("finalizeRef should only be set up for Datum Refs")
	}
	d.decrementRefs()
	r.v = nil
}

// a datum or a primitive. a nil Value means null.
type Value interface {
	Reference() *Ref
	Var(name string) Value
	SetVar(name string, value Value)
	Invoke(name string, parameters ...Value) Value
}

type Datum struct {
	impl DatumImpl

	// refcount is the number of Refs to this Datum. the datum only counts as being in the realm when this is nonzero.
	refCount uint
	realm    *Realm

	// TODO: readd singleton support
	//	// used for areas; causes there to only exist a single instance of each type
	//	singleton bool
}

var _ Value = &Datum{}

type SetResult int

const (
	SetResultOk SetResult = iota
	SetResultNonexistent
	SetResultReadOnly
)

type DatumImpl interface {
	Type() TypePath
	Var(src *Datum, name string) (Value, bool)
	SetVar(src *Datum, name string, value Value) SetResult
	Proc(src *Datum, name string, params ...Value) (Value, bool)
	Chunk(ref string) interface{}
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

func (d *Datum) Type() TypePath {
	return d.impl.Type()
}

func (d *Datum) Reference() *Ref {
	if d == nil {
		return nil
	}
	if d.refCount == 0 {
		d.realm.add(d)
	}
	// we're ignoring the possibility of overflow
	d.refCount += 1
	ref := &Ref{v: d}
	runtime.SetFinalizer(ref, finalizeRef)
	return ref
}

func (d *Datum) Var(name string) Value {
	v, ok := d.impl.Var(d, name)
	if !ok {
		panic(fmt.Sprintf("no such variable %s found on type %v during read", name, d.Type()))
	}
	return v
}

func (d *Datum) SetVar(name string, value Value) {
	switch d.impl.SetVar(d, name, value) {
	case SetResultOk:
	case SetResultNonexistent:
		panic(fmt.Sprintf("no such variable %s found on type %v during write", name, d.Type()))
	case SetResultReadOnly:
		panic(fmt.Sprintf("variable %s on type %v is read-only", name, d.Type()))
	default:
		panic("invalid result type")
	}
}

func (d *Datum) Invoke(name string, params ...Value) Value {
	result, ok := d.impl.Proc(d, name, params...)
	if !ok {
		panic(fmt.Sprintf("no such procedure %s found on type %v", name, d.Type()))
	}
	return result
}

func (d *Datum) Dump(o *debug.Output) {
	util.FIXME("replace dump with code for new model")
	debug.DumpReflect(d.Reference, o)
}

func (d *Datum) Realm() *Realm {
	return d.realm
}

func Unpack(v Value) (DatumImpl, bool) {
	d, ok := v.(*Datum)
	if ok {
		return d.impl, true
	}
	return nil, false
}

func Chunk(v Value, ref string) interface{} {
	d, ok := v.(*Datum)
	if ok {
		return d.impl.Chunk(ref)
	}
	return nil
}

func Param(params []Value, i int) Value {
	if i >= len(params) {
		return nil
	}
	return params[i]
}
