package procs

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/types"
)

func KWInvoke(w atoms.World, usr *types.Datum, name string, kwargs map[string]types.Value, args ...types.Value) types.Value {
	switch name {
	case "ismob":
		return types.FromBool(types.IsType(types.Param(args, 0), "/mob"))
	case "sound":
		return NewSoundFull(
			types.KWParam(args, 0, kwargs, "file"),
			types.KWParam(args, 1, kwargs, "repeat"),
			types.KWParam(args, 2, kwargs, "wait"),
			types.KWParam(args, 3, kwargs, "channel"),
			types.KWParam(args, 4, kwargs, "volume"),
		)
	case "oview":
		if usr == nil {
			panic("usr was nil when calling oview")
		}
		return datum.NewList(w.View1(usr, atoms.ViewExclusive)...)
	default:
		panic(fmt.Sprintf("unimplemented global function %q", name))
	}
}

func Invoke(w atoms.World, usr *types.Datum, name string, args ...types.Value) types.Value {
	return KWInvoke(w, usr, name, nil, args...)
}

func OperatorNot(x types.Value) types.Value {
	return types.FromBool(!types.AsBool(x))
}
