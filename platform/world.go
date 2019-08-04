package platform

import (
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/debug"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/websession"
)

type World struct {
	Name    string
	Mob     datum.TypePath
	Client  datum.TypePath
	Tree    *datum.TypeTree
	clients map[IClient]*datum.Ref

	// true if this instance has an API associated with it
	// we never provide more than one API so that we avoid double-threading
	claimed bool
}

func (w *World) FindAll(predicate func(IAtom) bool) []IAtom {
	var atoms []IAtom
	newPredicate := func(d datum.IDatum) bool {
		atom, isatom := d.(IAtom)
		return isatom && (predicate == nil || predicate(atom))
	}
	for _, d := range w.Tree.Realm().FindAll(newPredicate) {
		atoms = append(atoms, d.(IAtom))
	}
	return atoms
}

func (w *World) FindOne(predicate func(IAtom) bool) IAtom {
	newPredicate := func(d datum.IDatum) bool {
		atom, isatom := d.(IAtom)
		return isatom && predicate(atom)
	}
	atom := w.Tree.Realm().FindOne(newPredicate)
	if atom == nil {
		return nil
	} else {
		return atom.(IAtom)
	}
}

func (w *World) Dump(o *debug.Output) {
	o.Header("world")
	o.Println("name: " + w.Name)
	o.Println("mob: " + string(w.Mob))
	o.Footer()
}

func (w *World) CreateNewPlayer(key string) IClient {
	client := w.Tree.New(w.Client).(IClient)

	if _, found := w.clients[client]; found {
		panic("duplicate client objects should not exist")
	}
	w.clients[client] = client.Reference()

	c := client.AsClient()
	c.World = w
	c.Key = key
	c.SetMob(client.New(c.findExistingMob()))

	return client
}

func (w *World) RemovePlayer(client IClient) {
	delete(w.clients, client)
	client.Del()
}

func (w *World) PlayerExists(client IClient) bool {
	_, found := w.clients[client]
	return found
}

func (w *World) ServerAPI() websession.WorldAPI {
	if w.claimed {
		panic("second ServerAPI() call unexpected")
	}
	w.claimed = true
	return &worldAPI{
		World:   w,
		updates: make(chan struct{}, 1),
	}
}

func (w *World) LocateXYZ(x, y, z uint) ITurf {
	util.FIXME("this can definitely be more efficient")
	turf := w.FindOne(func(atom IAtom) bool {
		turf, isturf := atom.(ITurf)
		if isturf {
			tx, ty, tz := turf.XYZ()
			if tx == x && ty == y && tz == z {
				return true
			}
		}
		return false
	})
	if turf == nil {
		return nil
	} else {
		return turf.(ITurf)
	}
}

func NewWorld(tree *datum.TypeTree, name string, defaultMob datum.TypePath, defaultClient datum.TypePath) *World {
	return &World{
		Name:    name,
		Mob:     defaultMob,
		Client:  defaultClient,
		Tree:    tree,
		clients: map[IClient]*datum.Ref{},
		claimed: false,
	}
}
