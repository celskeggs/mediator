package atoms

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare MobData /mob /atom/movable
type MobData struct {
	key    string
	client types.Value // not a ref to avoid refcounting cycle
}

func NewMobData(src *types.Datum, _ *MobData, _ ...types.Value) {
	src.SetVar("name", types.String("mob"))
	src.SetVar("layer", types.Int(MobLayer))
	src.SetVar("density", types.Int(1))
}

func (m *MobData) OperatorWrite(src *types.Datum, usr *types.Datum, output types.Value) types.Value {
	client := m.GetClient(src)
	if client != nil {
		client.Invoke(usr, "<<", output)
	}
	return nil
}

func (m *MobData) GetClient(src *types.Datum) types.Value {
	if m.client == nil {
		return nil
	} else if WorldOf(src).PlayerExists(m.client) {
		return m.client
	} else {
		m.client = nil
		return nil
	}
}

func MobDataChunk(v types.Value) (*MobData, bool) {
	impl, ok := types.Unpack(v)
	if !ok {
		return nil, false
	}
	chunk := impl.Chunk("github.com/celskeggs/mediator/platform/atoms.MobData")
	if chunk == nil {
		return nil, false
	}
	return chunk.(*MobData), true
}

func MobSetClient(mobV types.Value, client types.Value) {
	util.FIXME("make this publicly accessible?")
	mob, ismob := MobDataChunk(mobV)
	if !ismob {
		panic("attempt to set client on not-a-mob: " + mobV.String())
	}
	if mob.client != nil {
		panic("client already set!")
	}
	if client != nil {
		mob.key = types.Unstring(client.Var("key"))
	}
	mob.client = client
}

func (m *MobData) GetKey(src *types.Datum) types.Value {
	return types.String(m.key)
}

func (m *MobData) ProcLogin(src *types.Datum, usr *types.Datum) types.Value {
	util.FIXME("make sure that Login gets called when client.mob is changed")
	// algorithm:
	// start at (1,1,1), scan across horizontally, then vertically, then in Z direction
	// pick first location that is *not* dense. move into it. if failed, continue. if none, leave location as null.
	mx, my, mz := WorldOf(src).MaxXYZ()
	for z := uint(1); z <= mz; z++ {
		for y := uint(1); y <= my; y++ {
			for x := uint(1); x <= mx; x++ {
				turf := WorldOf(src).LocateXYZ(x, y, z)
				if turf != nil && !types.AsBool(turf.Var("density")) {
					if types.AsBool(src.Invoke(usr, "Move", turf, common.Direction(0))) {
						return nil
					}
				}
			}
		}
	}
	util.FIXME("change stat object to mob")
	return nil
}
