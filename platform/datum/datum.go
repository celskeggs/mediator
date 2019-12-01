package datum

import (
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare DatumData /datum /
type DatumData struct {
}

func NewDatumData(_ ...types.Value) DatumData {
	return DatumData{}
}

func (d *DatumData) ProcNew(src *types.Datum) types.Value {
	util.FIXME("support tag and vars on /datum")
	util.FIXME("support for Del, Read, Topic, Write")
	// nothing to do for plain /datum
	return nil
}
