package atom

import "github.com/celskeggs/mediator/platform/types"

type World interface {
	PlayerExists(client types.Value) bool
	MaxXYZ() (int, int, int)
	LocateXYZ(x int, y int, z int) (turf types.Value)
	Realm() *types.Realm
}

func WorldOf(t *types.Datum) World {
	return t.Realm().WorldRef().(World)
}
