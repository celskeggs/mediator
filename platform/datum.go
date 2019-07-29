package platform

type IDatum interface {
	RawClone() IDatum
	Clone() IDatum
	AsDatum() *Datum
	AsAtom() *Atom
	Dump(*DebugOutput)
}

// invariant: Reference should point back at the top level of the Datum
type Datum struct {
	Reference IDatum
	Type      TypePath
}

func (d Datum) RawClone() (IDatum) {
	return &d
}

func (d *Datum) AsDatum() *Datum {
	return d
}

func (d *Datum) Clone() IDatum {
	if d.Reference == nil {
		panic("reference is nil when cloning")
	}
	cloned := d.Reference.RawClone()
	cloned.AsDatum().Reference = cloned
	return cloned
}

func (d *Datum) Dump(o *DebugOutput) {
	DumpReflect(d.Reference, o)
}

func setReference(datum IDatum) IDatum {
	datum.AsDatum().Reference = datum
	return datum
}

func (d *Datum) AsAtom() *Atom {
	panic("not an atom type: " + d.Type)
}
