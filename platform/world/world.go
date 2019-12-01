package world

import (
	"github.com/celskeggs/mediator/platform/atom"
	"github.com/celskeggs/mediator/platform/debug"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/websession"
)

type World struct {
	Name         string
	Mob          types.TypePath
	Client       types.TypePath
	ViewDistance uint

	defaultLazyEye uint

	MaxX, MaxY, MaxZ uint

	Tree *types.TypeTree

	clients map[world.IClient]*types.Ref

	// true if this instance has an API associated with it
	// we never provide more than one API so that we avoid double-threading
	claimed bool

	// true if the virtual eye should be set to the middle of the may
	setVirtualEye bool
}

var _ atom.World = &World{}

func (w World) PlayerExists(client types.Value) bool {
	panic("implement me")
}

func (w World) MaxXYZ() (int, int, int) {
	panic("implement me")
}

func (w World) LocateXYZ(x int, y int, z int) (turf types.Value) {
	panic("implement me")
}

func (w *World) ViewX(value types.Value, value2 types.Value, value3 types.Value) []types.Value {
	panic("implement me")
}

func (w *World) Realm() *types.Realm {
	panic("implement me")
}

func (w World) FindOne(predicate func(atom types.Value) bool) types.Value {
	panic("implement me")
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

func (w *World) CreateNewPlayer(key string) world.IClient {
	client := w.Tree.New(w.Client).(world.IClient)

	if _, found := w.clients[client]; found {
		panic("duplicate client objects should not exist")
	}
	w.clients[client] = client.Reference()

	c := client.AsClient()
	c.World = w
	c.Key = key
	if c.ViewDistance == 0 {
		util.NiceToHave("handle the /client/view = 0 situation correctly")
		c.ViewDistance = w.ViewDistance
	}
	client.New(c.findExistingMob())

	return client
}

func (w *World) RemovePlayer(client world.IClient) {
	datum.AssertConsistent(client)
	delete(w.clients, client)
	client.Del()
}

func (w *World) PlayerExists(client world.IClient) bool {
	datum.AssertConsistent(client)
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

func (w *World) UpdateDefaultViewDistance() {
	// if the map is <= 21x21, adjust view to fit the whole thing
	if w.MaxX > 0 && w.MaxY > 0 && w.MaxX <= 21 && w.MaxY <= 21 {
		w.ViewDistance = MaxUint(w.MaxX, w.MaxY) / 2
		// note: the documentation SAYS that we should turn on lazy_eye, but it actually doesn't get turned on.
		w.setVirtualEye = true
	}
}

func NewWorld(tree *types.TypeTree) *World {
	world := &World{
		Name:         "Untitled",
		Mob:          "/mob",
		Client:       "/client",
		ViewDistance: 5,
		Tree:         tree,
		clients:      map[world.IClient]*types.Ref{},
		claimed:      false,
	}
	tree.Realm().SetWorldRef(world)
	return world
}
