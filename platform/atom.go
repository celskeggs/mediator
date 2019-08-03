package platform

import (
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
)

// **** atom

type IAtom interface {
	datum.IDatum
	AsAtom() *Atom
	XYZ() (uint, uint, uint)
	Location() IAtom
	SetLocation(atom IAtom)
	Contents() []IAtom
}

var _ IAtom = &Atom{}

type Atom struct {
	datum.Datum
	Appearance Appearance
	Density    bool
	Opacity    bool
	location   *datum.Ref
	contents   map[*Atom]*datum.Ref
}

func (d Atom) RawClone() datum.IDatum {
	return &d
}

func (d *Atom) AsAtom() *Atom {
	return d
}

func (d *Atom) Location() IAtom {
	if d.location == nil {
		return nil
	} else {
		return d.location.Dereference().(IAtom)
	}
}

func (d *Atom) SetLocation(location IAtom) {
	if d.location != nil {
		contents := d.location.Dereference().(IAtom).AsAtom().contents
		if _, found := contents[d]; !found {
			panic("did not find self in location's contents")
		}
		delete(contents, d)
	}
	d.location = nil
	if location != nil {
		d.location = location.Reference()
		if location.AsAtom().contents == nil {
			location.AsAtom().contents = map[*Atom]*datum.Ref{}
		}
		contents := location.AsAtom().contents
		if _, found := contents[d]; found {
			panic("should not have found self in new location's contents")
		}
		contents[d] = d.Reference()
	}
}

func (d *Atom) Contents() (contents []IAtom) {
	for atom := range d.contents {
		contents = append(contents, atom.Impl.(IAtom))
	}
	return contents
}

func (d *Atom) XYZ() (uint, uint, uint) {
	location := d.Location()
	if location == nil {
		return 0, 0, 0
	}
	return location.XYZ()
}

// **** movable atom

type IAtomMovable interface {
	IAtom
	AsAtomMovable() *AtomMovable
}

var _ IAtomMovable = &AtomMovable{}

type AtomMovable struct {
	Atom
}

func (d AtomMovable) RawClone() datum.IDatum {
	return &d
}

func (d *AtomMovable) AsAtomMovable() *AtomMovable {
	return d
}

// **** obj

type IObj interface {
	IAtomMovable
	AsObj() *Obj
}

var _ IObj = &Obj{}

type Obj struct {
	AtomMovable
}

func (d Obj) RawClone() datum.IDatum {
	return &d
}

func (d *Obj) AsObj() *Obj {
	return d
}

// **** turf

type ITurf interface {
	IAtom
	AsTurf() *Turf
	SetXYZ(x uint, y uint, z uint)
}

var _ ITurf = &Turf{}

type Turf struct {
	Atom
	X uint
	Y uint
	Z uint
}

func (d Turf) RawClone() datum.IDatum {
	return &d
}

func (d *Turf) XYZ() (uint, uint, uint) {
	util.FIXME("better XYZ handling for Turfs")
	return d.X, d.Y, d.Z
}

func (d *Turf) SetXYZ(x uint, y uint, z uint) {
	d.X, d.Y, d.Z = x, y, z
}

func (d *Turf) AsTurf() *Turf {
	return d
}

// **** area

type IArea interface {
	IAtom
	AsArea() *Area
	Turfs() []ITurf
}

var _ IArea = &Area{}

type Area struct {
	Atom
}

func (d Area) RawClone() datum.IDatum {
	return &d
}

func (d *Area) AsArea() *Area {
	return d
}

func (d *Area) Turfs() (turfs []ITurf) {
	for atom := range d.contents {
		if turf, isTurf := atom.Impl.(ITurf); isTurf {
			turfs = append(turfs, turf)
		}
	}
	return turfs
}

// **** mob

type IMob interface {
	IAtom
	AsMob() *Mob
}

var _ IMob = &Mob{}

type Mob struct {
	AtomMovable
	Key string
}

func (d Mob) RawClone() datum.IDatum {
	return &d
}

func (d *Mob) AsMob() *Mob {
	return d
}

func NewAtomicTree() *datum.TypeTree {
	tree := datum.NewTypeTree()
	tree.RegisterStruct("/atom", &Atom{})
	tree.RegisterStruct("/atom/movable", &AtomMovable{})

	area := tree.RegisterStruct("/area", &Area{}).(*Area)
	area.Appearance.Layer = AreaLayer

	turf := tree.RegisterStruct("/turf", &Turf{}).(*Turf)
	turf.Appearance.Layer = TurfLayer

	obj := tree.RegisterStruct("/obj", &Obj{}).(*Obj)
	obj.Appearance.Layer = ObjLayer

	mob := tree.RegisterStruct("/mob", &Mob{}).(*Mob)
	mob.Appearance.Layer = MobLayer
	mob.Density = true

	tree.RegisterStruct("/client", &Client{})
	return tree
}
