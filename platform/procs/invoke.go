package procs

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
)

func KWInvoke(name string, kwargs map[string]types.Value, args ...types.Value) types.Value {
	switch name {
	case "!":
		return types.FromBool(!types.AsBool(types.Param(args, 0)))
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
	default:
		panic(fmt.Sprintf("unimplemented global function %q", name))
	}
}

func Invoke(name string, args ...types.Value) types.Value {
	return KWInvoke(name, nil, args...)
}
