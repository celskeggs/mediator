package platform

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
	"log"
)

type IClient interface {
	datum.IDatum
	// not intended to be overridden
	AsClient() *Client
	Eye() IAtom
	SetEye(atom IAtom)
	// intended to be overridden
	New(usr IMob) IMob
	Del()
	InvokeVerb(s string)
	RenderViewAsAtoms() (center IAtom, visible []IAtom)
	North() bool
	East() bool
	South() bool
	West() bool
	Move(loc IAtom, dir common.Direction) bool
	OutputString(output string)
	OutputSound(output ISound)
	PullClientRequests() (textDisplay []string, sounds []ISound)
}

var _ IClient = &Client{}

type Client struct {
	datum.IDatum
	Key          string
	World        *World
	mob          *datum.Ref
	eye          *datum.Ref
	virtualEye   *datum.Ref
	ViewDistance uint
	textBuffer   []string
	soundBuffer  []ISound
}

func (d *Client) Mob() IMob {
	return d.mob.Dereference().(IMob)
}

func (d *Client) SetMob(mob IMob) {
	datum.AssertConsistent(mob)

	d.mob = mob.Reference()
	d.SetEye(mob)
	mob.AsMob().setClient(d.Impl().(IClient))
}

func (d *Client) Eye() IAtom {
	if d.eye != nil {
		return d.eye.Dereference().(IAtom)
	} else {
		return nil
	}
}

func (d *Client) SetEye(eye IAtom) {
	datum.AssertConsistent(eye)

	d.eye = eye.Reference()

	if d.World.setVirtualEye {
		_, _, eyeZ := eye.XYZ()
		turf := d.World.LocateXYZ((d.World.MaxX+1)/2, (d.World.MaxY+1)/2, eyeZ)
		if turf == nil {
			util.NiceToHave("figure out if this is the right behavior in this case and maybe adjust it")
			d.virtualEye = nil
		} else {
			d.virtualEye = turf.Reference()
		}
	} else {
		d.virtualEye = d.eye
	}
}

func (d *Client) VirtualEye() IAtom {
	if d.virtualEye != nil {
		return d.virtualEye.Dereference().(IAtom)
	} else {
		return nil
	}
}

func (d *Client) InvokeVerb(verb string) {
	switch verb {
	case ".north":
		d.Impl().(IClient).North()
	case ".south":
		d.Impl().(IClient).South()
	case ".east":
		d.Impl().(IClient).East()
	case ".west":
		d.Impl().(IClient).West()
	default:
		log.Println("got unknown verb:", verb)
	}
}

func (d *Client) RelMove(direction common.Direction) bool {
	var turf ITurf
	mob := d.Mob()
	if mob != nil {
		x, y, z := mob.XYZ()
		dx, dy := direction.XY()
		turf = d.World.LocateXYZ(uint(int(x)+dx), uint(int(y)+dy), z)
	}
	return d.Impl().(IClient).Move(turf, direction)
}

func (d *Client) North() bool {
	return d.RelMove(common.North)
}

func (d *Client) East() bool {
	return d.RelMove(common.East)
}

func (d *Client) South() bool {
	return d.RelMove(common.South)
}

func (d *Client) West() bool {
	return d.RelMove(common.West)
}

func (d *Client) Move(loc IAtom, dir common.Direction) bool {
	datum.AssertConsistent(loc)
	mob := d.Mob()
	util.FIXME("cancel automated movement if necessary")
	if mob != nil {
		return mob.Move(loc, dir)
	}
	return false
}

func (d *Client) OutputString(output string) {
	d.textBuffer = append(d.textBuffer, output)
}

func (d *Client) OutputSound(output ISound) {
	d.soundBuffer = append(d.soundBuffer, output)
}

func (d *Client) PullClientRequests() (textDisplay []string, sounds []ISound) {
	textDisplay, sounds = d.textBuffer, d.soundBuffer
	d.textBuffer = nil
	d.soundBuffer = nil
	return textDisplay, sounds
}

func (d *Client) RenderViewAsAtoms() (center IAtom, atoms []IAtom) {
	util.FIXME("actually do this correctly")
	veye := d.VirtualEye()
	return veye, d.World.ViewX(d.ViewDistance, veye, d.Eye())
}

func (d Client) RawClone() datum.IDatum {
	d.IDatum = d.IDatum.RawClone()
	return &d
}

func (d *Client) AsClient() *Client {
	return d
}

func isTurf(atom IAtom) bool {
	datum.AssertConsistent(atom)
	_, isturf := atom.(ITurf)
	return isturf
}

func (d *Client) constructNewMob() IMob {
	mob := d.World.Tree.New(d.World.Mob).(IMob)
	util.FIXME("initialize name and gender")
	return mob
}

func (d *Client) findExistingMob() IMob {
	if d.Key == "" {
		return nil
	}
	return d.World.FindOne(func(atom IAtom) bool {
		mob, ismob := atom.(IMob)
		return ismob && mob.AsMob().Key() == d.Key
	}).(IMob)
}

func (d *Client) New(usr IMob) IMob {
	datum.AssertConsistent(usr)

	mob := usr
	util.FIXME("add support for 'prototype mobs'")
	if mob == nil {
		mob = d.constructNewMob()
	}
	util.NiceToHave("add support for Topics")
	mob.Login()
	return mob
}

func (d *Client) Del() {
	util.FIXME("call Logout() on mob")
	util.FIXME("should killing the connection go here, maybe in addition to other places?")
}
