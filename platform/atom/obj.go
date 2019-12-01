package atom

import "github.com/celskeggs/mediator/platform/types"

//mediator:declare ObjData /obj /atom/movable
type ObjData struct{}

func NewObjData(_ ...types.Value) ObjData {
	return ObjData{}
}
