package platform

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
)

// **** atom

type IAtom interface {
	datum.IDatum
	// not intended to be overridden
	AsAtom() *Atom
	XYZ() (uint, uint, uint)
	Location() IAtom
	SetLocation(atom IAtom)
	Contents() []IAtom
	// intended to be overridden
	Exit(atom IAtomMovable, newloc IAtom) bool
	Enter(atom IAtomMovable, oldloc IAtom) bool
	Exited(atom IAtomMovable, newloc IAtom)
	Entered(atom IAtomMovable, oldloc IAtom)
}

var _ IAtom = &Atom{}

type Atom struct {
	datum.Datum
	Appearance Appearance
	Density    bool
	Opacity    bool
	Direction  common.Direction
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
	datum.AssertConsistent(location)

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

// you *can* mutate the result of this function
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

func (d *Atom) Exit(atom IAtomMovable, newloc IAtom) bool {
	datum.AssertConsistent(atom, newloc)
	return true
}

func (d *Atom) Enter(atom IAtomMovable, oldloc IAtom) bool {
	datum.AssertConsistent(atom, oldloc)
	return true
}

func (d *Atom) Exited(atom IAtomMovable, newloc IAtom) {
	datum.AssertConsistent(atom, newloc)
	// nothing to do
}

func (d *Atom) Entered(atom IAtomMovable, oldloc IAtom) {
	datum.AssertConsistent(atom, oldloc)
	// nothing to do
}

// **** movable atom

type IAtomMovable interface {
	IAtom
	// not intended to be overridden
	AsAtomMovable() *AtomMovable
	// intended to be overridden
	Move(atom IAtom, direction common.Direction) bool
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

func (d *AtomMovable) Move(newloc IAtom, direction common.Direction) bool {
	datum.AssertConsistent(newloc)
	util.NiceToHave("implement pixel movement/slides")
	oldloc := d.Location()
	impl := d.Impl.(IAtomMovable)
	d.Direction = direction
	if newloc != oldloc && newloc != nil {
		if oldloc != nil {
			if !oldloc.Exit(impl, newloc) {
				return false
			}
			util.NiceToHave("handle Cross and Uncross and Crossed and Uncrossed")
		}
		if !newloc.Enter(impl, oldloc) {
			util.FIXME("bump obstacles")
			return false
		}
		d.SetLocation(newloc)
		if oldloc != nil {
			oldloc.Exited(impl, newloc)
		}
		newloc.Entered(impl, oldloc)
	}
	return true
}

// **** obj

type IObj interface {
	IAtomMovable
	// not intended to be overridden
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
	// not intended to be overridden
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
	util.FIXME("should this actually be public?")
	d.X, d.Y, d.Z = x, y, z
}

func (d *Turf) AsTurf() *Turf {
	return d
}

func (d *Turf) Exit(atom IAtomMovable, newloc IAtom) bool {
	datum.AssertConsistent(atom, newloc)
	util.NiceToHave("call Uncross here")
	return true
}

func (d *Turf) Enter(atom IAtomMovable, oldloc IAtom) bool {
	datum.AssertConsistent(atom, oldloc)
	util.NiceToHave("call Cross here")
	if atom.AsAtom().Density {
		if d.Density {
			return false
		}
		util.NiceToHave("something about only atoms that take up the full tile?")
		for _, existingAtom := range d.Contents() {
			if existingAtom.AsAtom().Density {
				return false
			}
		}
	}
	return true
}

func (d *Turf) Exited(atom IAtomMovable, newloc IAtom) {
	datum.AssertConsistent(atom, newloc)
	util.NiceToHave("call Uncrossed here")
}

func (d *Turf) Entered(atom IAtomMovable, oldloc IAtom) {
	datum.AssertConsistent(atom, oldloc)
	util.NiceToHave("call Crossed here")
}

// **** area

type IArea interface {
	IAtom
	// not intended to be overridden
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
	IAtomMovable
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

	templateAtom := Atom{
		Direction: common.North,
	}

	templateAtomMovable := AtomMovable{
		Atom: templateAtom,
	}

	templateArea := Area{
		Atom: templateAtom,
	}
	templateArea.Appearance.Layer = AreaLayer

	templateTurf := Turf{
		Atom: templateAtom,
	}
	templateTurf.Appearance.Layer = TurfLayer

	templateObj := Obj{
		AtomMovable: templateAtomMovable,
	}
	templateObj.Appearance.Layer = ObjLayer

	templateMob := Mob{
		AtomMovable: templateAtomMovable,
	}
	templateMob.Appearance.Layer = MobLayer
	templateMob.Density = true

	tree.RegisterStruct("/atom", &templateAtom)
	tree.RegisterStruct("/atom/movable", &templateAtomMovable)
	tree.RegisterStruct("/area", &templateArea)
	tree.RegisterStruct("/turf", &templateTurf)
	tree.RegisterStruct("/obj", &templateObj)
	tree.RegisterStruct("/mob", &templateMob)

	tree.RegisterStruct("/client", &Client{})
	return tree
}
