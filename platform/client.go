package platform

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
	"log"
)

type IClient interface {
	datum.IDatum
	AsClient() *Client
	New(usr IMob) IMob
	Del()
	InvokeVerb(s string)
	RenderViewAsAtoms() []IAtom
	North() bool
	East() bool
	South() bool
	West() bool
	Move(loc IAtom, dir common.Direction) bool
}

var _ IClient = &Client{}

type Client struct {
	datum.Datum
	Key   string
	mob   *datum.Ref
	World *World
}

func (d *Client) Mob() IMob {
	return d.mob.Dereference().(IMob)
}

func (d *Client) SetMob(mob IMob) {
	datum.AssertConsistent(mob)

	d.mob = mob.Reference()
}

func (d *Client) InvokeVerb(verb string) {
	switch verb {
	case ".north":
		d.Impl.(IClient).North()
	case ".south":
		d.Impl.(IClient).South()
	case ".east":
		d.Impl.(IClient).East()
	case ".west":
		d.Impl.(IClient).West()
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
	return d.Impl.(IClient).Move(turf, direction)
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

func (d *Client) RenderViewAsAtoms() []IAtom {
	util.FIXME("actually do this correctly")
	return d.World.FindAll(nil)
}

func (d Client) RawClone() datum.IDatum {
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
	util.FIXME("don't randomly select entry turf; follow the rules. especially don't start in a wall")
	mob.SetLocation(d.World.FindOne(isTurf))
	util.FIXME("use Enter to join world, not SetLocation")
	return mob
}

func (d *Client) findExistingMob() IMob {
	if d.Key == "" {
		return nil
	}
	return d.World.FindOne(func(atom IAtom) bool {
		mob, ismob := atom.(IMob)
		return ismob && mob.AsMob().Key == d.Key
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
	util.FIXME("call Login() on mob")
	return mob
}

func (d *Client) Del() {
	util.FIXME("call Logout() on mob")
	util.FIXME("should killing the connection go here, maybe in addition to other places?")
}
