package platform

import (
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
)

type IClient interface {
	datum.IDatum
	AsClient() *Client
	New(usr IMob) IMob
	Del()
	InvokeVerb(s string)
	RenderViewAsAtoms() []IAtom
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
	d.mob = mob.Reference()
}

func (d *Client) InvokeVerb(s string) {
	panic("implement me")
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
	_, isturf := atom.(ITurf)
	return isturf
}

func (d *Client) constructNewMob() IMob {
	mob := d.World.Tree.New(d.World.Mob).(IMob)
	util.FIXME("initialize name and gender")
	mob.SetLocation(d.World.FindOne(isTurf))
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
	mob := usr
	util.FIXME("add support for 'prototype mobs'")
	if mob == nil {
		mob = d.constructNewMob()
	}
	util.FIXME("add support for Topics")
	util.FIXME("call Login() on mob")
	return mob
}

func (d *Client) Del() {
	util.FIXME("call Logout() on mob")
	util.FIXME("should killing the connection go here, maybe in addition to other places?")
}
