package atom

import "github.com/celskeggs/mediator/platform/types"

//mediator:declare ObjData /obj /atom/movable
type ObjData struct{}

func NewObjData(src *types.Datum, _ ...types.Value) ObjData {
	src.SetVar("name", types.String("obj"))
	src.SetVar("layer", types.Int(ObjLayer))
	return ObjData{}
}
