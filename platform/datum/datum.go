package datum

import (
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare DatumData /datum /
type DatumData struct{}

func NewDatumData(src *types.Datum, _ *DatumData, _ ...types.Value) {
}

func (d *DatumData) ProcNew(src *types.Datum) types.Value {
	util.FIXME("support tag and vars on /datum")
	util.FIXME("support for Del, Read, Topic, Write")
	// nothing to do for plain /datum
	return nil
}
