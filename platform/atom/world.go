package atom

import "github.com/celskeggs/mediator/platform/types"

type World interface {
	PlayerExists(client types.Value) bool
	MaxXYZ() (uint, uint, uint)
	LocateXYZ(x uint, y uint, z uint) (turf types.Value)
}

func WorldOf(t *types.Datum) World {
	return t.Realm().WorldRef().(World)
}
