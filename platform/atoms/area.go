package atoms

import (
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare AreaData /area /atom
type AreaData struct{}

func NewAreaData(src *types.Datum, _ ...types.Value) AreaData {
	util.FIXME("handle area.X, .Y, .Z correctly")
	src.SetVar("name", types.String("area"))
	src.SetVar("layer", types.Int(AreaLayer))
	return AreaData{}
}

func TurfsInArea(area types.Value) (turfs []types.Value) {
	util.FIXME("is this method actually needed?")
	if !types.IsType(area, "/area") {
		panic("not an /area")
	}
	for _, atom := range datum.Elements(area.Var("contents")) {
		if types.IsType(atom, "/turf") {
			turfs = append(turfs, atom)
		}
	}
	return turfs
}
