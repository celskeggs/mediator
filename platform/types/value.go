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

func Reference(v Value) *Ref {
	if v == nil {
		return nil
	}
	if d, ok := v.(*Datum); ok {
		if d.impl == nil {
			panic("attempt to Reference deleted datum")
		}
		if d.refCount == 0 {
			d.realm.add(d)
		}
		// we're ignoring the possibility of overflow
		d.refCount += 1
		ref := &Ref{v: d}
		runtime.SetFinalizer(ref, finalizeRef)
		return ref
	} else {
		return &Ref{v}
	}
}

func (r *Ref) Dereference() Value {
	if r == nil {
		return nil
	}
	if r.v == nil {
		panic("nil value during dereference")
	}
	if datum, ok := r.v.(*Datum); ok {
		// datum has been deleted; return nil now
		if datum.impl == nil {
			return nil
		}
	}
	return r.v
}

func Del(v Value) {
	datum, ok := v.(*Datum)
	if !ok {
		panic("cannot delete non-datum value " + v.String())
	}
	datum.delete()
}

func finalizeRef(r *Ref) {
	d := r.v.(*Datum)
	if d == nil {
		panic("finalizeRef should only be set up for Datum Refs")
	}
	if d.impl != nil {
		d.decrementRefs()
	}
	r.v = nil
}

// a datum or a primitive. a nil Value means null.
type Value interface {
	Var(name string) Value
	SetVar(name string, value Value)
	Invoke(usr *Datum, name string, parameters ...Value) Value
	String() string
}

type Datum struct {
	impl DatumImpl

	// refcount is the number of Refs to this Datum. the datum only counts as being in the realm when this is nonzero.
	refCount uint
	realm    *Realm
	uid      uint64
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
	Proc(src *Datum, usr *Datum, name string, params ...Value) (Value, bool)
	SuperProc(src *Datum, usr *Datum, chunk string, name string, params ...Value) (Value, bool)
	ProcSettings(name string) (ProcSettings, bool)
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

func (d *Datum) delete() {
	if d.impl == nil {
		panic("attempt to delete deleted datum")
	}
	// if you attempt to use a deleted datum for anything, it will panic; this is intentional.
	// you should never be able to get your hands on a deleted datum, anyway -- if it COULD become deleted, then it's
	// a long-lived reference, and therefore should be stored as a *Ref, which will handle the ability for it to become
	// nil through deletion.
	d.impl = nil
	if d.refCount > 0 {
		d.realm.remove(d)
		d.refCount = 0
	}
}

func (d *Datum) Type() TypePath {
	if d.impl == nil {
		panic("attempt to get type of deleted datum")
	}
	return d.impl.Type()
}

func (d *Datum) UID() uint64 {
	return d.uid
}

func (d *Datum) Var(name string) Value {
	if d == nil {
		panic(fmt.Sprintf("attempt to access variable %s on null value", name))
	}
	if d.impl == nil {
		panic("attempt to access variable on deleted datum")
	}
	v, ok := d.impl.Var(d, name)
	if !ok {
		panic(fmt.Sprintf("no such variable %s found on type %v during read", name, d.Type()))
	}
	return v
}

func (d *Datum) SetVar(name string, value Value) {
	if d.impl == nil {
		panic("attempt to set variable on deleted datum")
	}
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

func (d *Datum) Invoke(usr *Datum, name string, params ...Value) Value {
	if d.impl == nil {
		panic("attempt to invoke proc on deleted datum")
	}
	result, ok := d.impl.Proc(d, usr, name, params...)
	if !ok {
		panic(fmt.Sprintf("no such procedure %s found on type %v", name, d.Type()))
	}
	return result
}

func (d *Datum) SuperInvoke(usr *Datum, chunk string, name string, params ...Value) Value {
	if d.impl == nil {
		panic("attempt to invoke super proc on deleted datum")
	}
	result, ok := d.impl.SuperProc(d, usr, chunk, name, params...)
	if !ok {
		panic(fmt.Sprintf("no such super procedure %s.%s found on type %v", chunk, name, d.Type()))
	}
	return result
}

func (d *Datum) Dump(o *debug.Output) {
	util.FIXME("replace dump with code for new model")
	debug.DumpReflect(d, o)
}

func (d *Datum) Realm() *Realm {
	return d.realm
}

func (d *Datum) String() string {
	if d.impl == nil {
		// see comment in delete(), above, about why you should never be able to access this
		return "[deleted datum that you should never be able to have a pointer to]"
	}
	return fmt.Sprintf("[datum of type %s]", d.Type())
}

func Unpack(v Value) (DatumImpl, bool) {
	d, ok := v.(*Datum)
	if ok {
		return UnpackDatum(d), true
	}
	return nil, false
}

func UnpackDatum(d *Datum) DatumImpl {
	if d.impl == nil {
		panic("attempt to unpack deleted datum")
	}
	return d.impl
}

func Chunk(v Value, ref string) interface{} {
	d, ok := v.(*Datum)
	if ok {
		if d.impl == nil {
			panic("attempt to get chunk from deleted datum")
		}
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

func KWParam(params []Value, i int, kwparams map[string]Value, kw string) Value {
	kwv, haskw := kwparams[kw]
	if i >= len(params) {
		if haskw {
			return kwv
		}
		return nil
	}
	if haskw {
		panic("parameter " + kw + " specified both as positional and named argument")
	}
	return params[i]
}

func IsType(v Value, path TypePath) bool {
	if datum, ok := v.(*Datum); ok {
		if datum == nil {
			panic("found half-nil datum")
		}
		if datum.impl == nil {
			panic("attempt to check type of deleted datum")
		}
		return datum.Realm().IsSubType(datum.Type(), path)
	}
	return false
}

func AssertType(v Value, path TypePath) {
	if !IsType(v, path) {
		panic("unexpected value " + v.String() + " when datum of type " + path.String() + " expected")
	}
}
