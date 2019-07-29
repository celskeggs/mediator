package platform

// **** atom

type Atom struct {
	Datum
	Icon    string
	Density bool
	Opacity bool
}

func (d Atom) RawClone() (IDatum) {
	return &d
}

func (d *Atom) AsAtom() *Atom {
	return d
}

// **** movable atom

type AtomMovable struct {
	Atom
}

func (d AtomMovable) RawClone() (IDatum) {
	return &d
}

// **** obj

type Obj struct {
	AtomMovable
}

func (d Obj) RawClone() (IDatum) {
	return &d
}

// **** turf

type Turf struct {
	Atom
}

func (d Turf) RawClone() (IDatum) {
	return &d
}

// **** area

type Area struct {
	Atom
}

func (d Area) RawClone() (IDatum) {
	return &d
}

// **** mob

type Mob struct {
	AtomMovable
}

func (d Mob) RawClone() (IDatum) {
	return &d
}
