package atoms

import "github.com/celskeggs/mediator/platform/types"

//mediator:declare ObjData /obj /atom/movable
type ObjData struct{}

func NewObjData(src *types.Datum, _ *ObjData, _ ...types.Value) {
	src.SetVar("name", types.String("obj"))
	src.SetVar("layer", types.Int(ObjLayer))
}
