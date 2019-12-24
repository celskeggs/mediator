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
	case "view":
		if usr == nil {
			panic("usr was nil when calling view")
		}
		return datum.NewList(w.View1(usr, atoms.ViewInclusive)...)
	case "stat":
		if usr == nil {
			panic("usr is nil during attempt to use stat()")
		}
		mob, ok := atoms.MobDataChunk(usr)
		if !ok {
			panic("usr is not a /mob during attempt to use stat()")
		}
		context := mob.StatContext()
		if context == nil {
			panic("attempt to use stat() when not presently within a Stat() invocation")
		}
		if len(args) >= 2 {
			context.Stat(types.Unstring(types.Param(args, 0)), types.Param(args, 1))
		} else {
			context.Stat("", types.Param(args, 0))
		}
		return nil
	case "statpanel":
		if usr == nil {
			panic("usr is nil during attempt to use statpanel()")
		}
		mob, ok := atoms.MobDataChunk(usr)
		if !ok {
			panic("usr is not a /mob during attempt to use statpanel()")
		}
		context := mob.StatContext()
		if context == nil {
			panic("attempt to use statpanel() when not presently within a Stat() invocation")
		}
		visible := context.StatPanel(types.Unstring(types.Param(args, 0)))
		if len(args) >= 2 {
			context.Stat("", types.Param(args, 1))
			return nil
		} else if len(args) >= 3 {
			context.Stat(types.Unstring(types.Param(args, 1)), types.Param(args, 2))
			return nil
		} else {
			return types.FromBool(visible)
		}
	case "walk_to":
		panic("unimplemented: walk_to")
	case "get_dir":
		panic("unimplemented: get_dir")
	case "flick":
		panic("unimplemented: flick")
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
