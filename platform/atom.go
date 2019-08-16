package platform

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient/sprite"
)

// **** atom

type IAtom interface {
	datum.IDatum
	// not intended to be overridden
	AsAtom() *Atom
	XYZ() (uint, uint, uint)
	ContainingArea() IArea
	Location() IAtom
	SetLocation(atom IAtom)
	Contents() []IAtom
	World() *World
	// intended to be overridden
	Exit(atom IAtomMovable, newloc IAtom) bool
	Enter(atom IAtomMovable, oldloc IAtom) bool
	Exited(atom IAtomMovable, newloc IAtom)
	Entered(atom IAtomMovable, oldloc IAtom)
}

var _ IAtom = &Atom{}

type Atom struct {
	datum.IDatum
	Appearance Appearance
	Density    bool
	Opacity    bool
	Direction  common.Direction
	location   *datum.Ref
	contents   map[*Atom]*datum.Ref
}

func (d Atom) RawClone() datum.IDatum {
	d.IDatum = d.IDatum.RawClone()
	return &d
}

func (d *Atom) AsAtom() *Atom {
	return d
}

func (d *Atom) World() *World {
	return d.Realm().WorldRef().(*World)
}

func (d *Atom) Location() IAtom {
	if d.location == nil {
		return nil
	} else {
		return d.location.Dereference().(IAtom)
	}
}

func (d *Atom) ContainingArea() IArea {
	loc := d.Location()
	if loc == nil {
		return nil
	} else {
		return loc.ContainingArea()
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
		contents = append(contents, atom.Impl().(IAtom))
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
	Bump(atom IAtom)
}

var _ IAtomMovable = &AtomMovable{}

type AtomMovable struct {
	IAtom
}

func (d AtomMovable) RawClone() datum.IDatum {
	d.IAtom = d.IAtom.RawClone().(IAtom)
	return &d
}

func (d *AtomMovable) AsAtomMovable() *AtomMovable {
	return d
}

func (d *AtomMovable) Move(newloc IAtom, direction common.Direction) bool {
	datum.AssertConsistent(newloc)

	util.NiceToHave("implement pixel movement/slides")

	oldloc := d.Location()
	oldarea := d.ContainingArea()
	impl := d.Impl().(IAtomMovable)
	if direction != 0 {
		d.AsAtom().Direction = direction
	}
	if newloc != oldloc && newloc != nil {
		newarea := newloc.ContainingArea()
		if oldloc != nil {
			if !oldloc.Exit(impl, newloc) {
				return false
			}
			util.NiceToHave("handle Cross and Uncross and Crossed and Uncrossed")
		}
		if newarea != oldarea && oldarea != nil {
			if !oldarea.Exit(impl, newarea) {
				return false
			}
		}
		if !newloc.Enter(impl, oldloc) {
			util.FIXME("bump obstacles")
			return false
		}
		if newarea != oldarea && newarea != nil {
			if !newarea.Enter(impl, oldarea) {
				return false
			}
		}
		d.SetLocation(newloc)
		if oldloc != nil {
			oldloc.Exited(impl, newloc)
		}
		if newarea != oldarea && oldarea != nil {
			oldarea.Exited(impl, newarea)
		}
		newloc.Entered(impl, oldloc)
		if newarea != oldarea && newarea != nil {
			newarea.Entered(impl, oldarea)
		}
	}
	return true
}

func (d *AtomMovable) Bump(obstacle IAtom) {
	datum.AssertConsistent(obstacle)
	// nothing to do in the general case
	util.NiceToHave("group support for mob bumping")
}

// **** obj

type IObj interface {
	IAtomMovable
	// not intended to be overridden
	AsObj() *Obj
}

var _ IObj = &Obj{}

type Obj struct {
	IAtomMovable
}

func (d Obj) RawClone() datum.IDatum {
	d.IAtomMovable = d.IAtomMovable.RawClone().(IAtomMovable)
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
}

var _ ITurf = &Turf{}

type Turf struct {
	IAtom
	X uint
	Y uint
	Z uint
}

func (d Turf) RawClone() datum.IDatum {
	d.IAtom = d.IAtom.RawClone().(IAtom)
	return &d
}

func (d *Turf) XYZ() (uint, uint, uint) {
	util.FIXME("better XYZ handling for Turfs")
	return d.X, d.Y, d.Z
}

func (d *Turf) SetXYZ(x uint, y uint, z uint) {
	util.NiceToHave("see if this can be made private, and if so, if it should be")
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
		if d.AsAtom().Density {
			atom.Bump(d.Impl().(IAtom))
			return false
		}
		util.NiceToHave("something about only atoms that take up the full tile?")
		for _, existingAtom := range d.Contents() {
			if existingAtom.AsAtom().Density {
				atom.Bump(existingAtom)
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
	IAtom
}

func (d Area) RawClone() datum.IDatum {
	d.IAtom = d.IAtom.RawClone().(IAtom)
	return &d
}

func (d *Area) AsArea() *Area {
	return d
}

func (d *Area) ContainingArea() IArea {
	return d.Impl().(IArea)
}

func (d *Area) Turfs() (turfs []ITurf) {
	for atom := range d.AsAtom().contents {
		if turf, isTurf := atom.Impl().(ITurf); isTurf {
			turfs = append(turfs, turf)
		}
	}
	return turfs
}

// **** mob

type IMob interface {
	IAtomMovable
	AsMob() *Mob
	OutputString(output string)
	OutputSound(output ISound)
	Client() IClient
	Key() string
	Login()
}

var _ IMob = &Mob{}

type Mob struct {
	IAtomMovable
	key    string
	client IClient
}

func (d Mob) RawClone() datum.IDatum {
	d.IAtomMovable = d.IAtomMovable.RawClone().(IAtomMovable)
	return &d
}

func (d *Mob) Key() string {
	return d.key
}

func (d *Mob) Client() IClient {
	if d.client == nil {
		return nil
	} else if d.World().PlayerExists(d.client) {
		return d.client
	} else {
		d.client = nil
		return nil
	}
}

func (d *Mob) setClient(client IClient) {
	datum.AssertConsistent(client)
	util.FIXME("make this publicly accessible")
	if d.Client() != nil {
		panic("client already set!")
	}
	if client != nil {
		d.key = client.AsClient().Key
	}
	d.client = client
}

func (d *Mob) AsMob() *Mob {
	return d
}

func (d *Mob) OutputString(output string) {
	client := d.Client()
	if client != nil {
		client.OutputString(output)
	}
}

func (d *Mob) OutputSound(output ISound) {
	client := d.Client()
	if client != nil {
		client.OutputSound(output)
	}
}

func (d *Mob) Login() {
	// algorithm:
	// start at (1,1,1), scan across horizontally, then vertically, then in Z direction
	// pick first location that is *not* dense. move into it. if failed, continue. if none, leave location as null.
	mx, my, mz := d.World().MaxX, d.World().MaxY, d.World().MaxZ
	for z := uint(1); z <= mz; z++ {
		for y := uint(1); y <= my; y++ {
			for x := uint(1); x <= mx; x++ {
				turf := d.World().LocateXYZ(x, y, z)
				if turf != nil && !turf.AsAtom().Density {
					if d.Impl().(IAtomMovable).Move(turf, 0) {
						return
					}
				}
			}
		}
	}
	util.FIXME("change stat object to mob")
}

// **** tree definition

type TreeDefiner interface {
	AtomTemplate(parent datum.IDatum) IAtom
	AtomMovableTemplate(parent IAtom) IAtomMovable
	AreaTemplate(parent IAtom) IArea
	TurfTemplate(parent IAtom) ITurf
	ObjTemplate(parent IAtomMovable) IObj
	MobTemplate(parent IAtomMovable) IMob
	ClientTemplate(parent datum.IDatum) IClient
	SoundTemplate(parent datum.IDatum) ISound
}

type BaseTreeDefiner struct{}

var _ TreeDefiner = BaseTreeDefiner{}

func (td BaseTreeDefiner) AtomTemplate(parent datum.IDatum) IAtom {
	return &Atom{
		IDatum:    parent,
		Direction: common.South,
		Appearance: Appearance{
			Name: "atom",
		},
	}
}

func (td BaseTreeDefiner) AtomMovableTemplate(parent IAtom) IAtomMovable {
	parent.AsAtom().Appearance.Name = "movable"
	return &AtomMovable{
		IAtom: parent,
	}
}

func (td BaseTreeDefiner) AreaTemplate(parent IAtom) IArea {
	parent.AsAtom().Appearance.Layer = AreaLayer
	parent.AsAtom().Appearance.Name = "area"
	return &Area{
		IAtom: parent,
	}
}

func (td BaseTreeDefiner) TurfTemplate(parent IAtom) ITurf {
	parent.AsAtom().Appearance.Layer = TurfLayer
	parent.AsAtom().Appearance.Name = "turf"
	return &Turf{
		IAtom: parent,
	}
}

func (td BaseTreeDefiner) ObjTemplate(parent IAtomMovable) IObj {
	parent.AsAtom().Appearance.Layer = ObjLayer
	parent.AsAtom().Appearance.Name = "obj"
	return &Obj{
		IAtomMovable: parent,
	}
}

func (td BaseTreeDefiner) MobTemplate(parent IAtomMovable) IMob {
	parent.AsAtom().Appearance.Layer = MobLayer
	parent.AsAtom().Appearance.Name = "mob"
	parent.AsAtom().Density = true
	return &Mob{
		IAtomMovable: parent,
	}
}

func (td BaseTreeDefiner) ClientTemplate(parent datum.IDatum) IClient {
	return &Client{
		IDatum: parent,
	}
}

func (td BaseTreeDefiner) SoundTemplate(parent datum.IDatum) ISound {
	return &Sound{
		IDatum: parent,
		Sound: sprite.Sound{
			Volume: 100,
		},
	}
}

func NewAtomicTree(td TreeDefiner) *datum.TypeTree {
	tree := datum.NewTypeTree()

	tree.RegisterStruct("/atom",
		td.AtomTemplate(
			tree.DeriveNew("/datum")))

	tree.RegisterStruct("/atom/movable",
		td.AtomMovableTemplate(
			tree.DeriveNew("/atom").(IAtom)))

	tree.RegisterStruct("/area",
		td.AreaTemplate(
			tree.DeriveNew("/atom").(IAtom)))
	tree.SetSingleton("/area")

	tree.RegisterStruct("/turf",
		td.TurfTemplate(
			tree.DeriveNew("/atom").(IAtom)))

	tree.RegisterStruct("/obj",
		td.ObjTemplate(
			tree.DeriveNew("/atom/movable").(IAtomMovable)))

	tree.RegisterStruct("/mob",
		td.MobTemplate(
			tree.DeriveNew("/atom/movable").(IAtomMovable)))

	tree.RegisterStruct("/client",
		td.ClientTemplate(
			tree.DeriveNew("/datum")))

	tree.RegisterStruct("/sound",
		td.SoundTemplate(
			tree.DeriveNew("/datum")))

	return tree
}
