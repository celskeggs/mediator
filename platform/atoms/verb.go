package atoms

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"strconv"
	"strings"
)

// NOTE: although verbs have a "defining type", which can be inspected, all that ultimately matters is the procedure
// name, because that is what actually gets called when the verb is invoked.
type Verb struct {
	VisibleName  string
	DefiningType string
	ProcName     string
}

var _ types.Value = Verb{}

func (v Verb) Var(name string) types.Value {
	panic(fmt.Sprintf("no such variable %q on /verb", name))
}

func (v Verb) SetVar(name string, value types.Value) {
	panic(fmt.Sprintf("no such variable %q on /verb", name))
}

func (v Verb) Invoke(usr *types.Datum, name string, parameters ...types.Value) types.Value {
	panic(fmt.Sprintf("no such proc %q on /verb", name))
}

func (v Verb) String() string {
	return fmt.Sprintf("[verb name=%q proc=%v/%s]", v.VisibleName, v.DefiningType, v.ProcName)
}

func (v Verb) Matches(name string, src *types.Datum, usr *types.Datum, args []string) bool {
	if name != v.VisibleName {
		return false
	}
	settings, ok := types.UnpackDatum(src).ProcSettings(v.VisibleName)
	if !ok {
		util.FIXME("make sure this never actually happens")
		panic(fmt.Sprintf("attempt to use procedure %s on %v as a verb, when it has no metadata", name, src))
	}
	if settings.Src.In && len(args) >= 1 {
		// then we expect the first argument to refer to us; if it doesn't, we don't match.
		if strings.HasPrefix(args[0], "#") {
			// reference by UID
			uid, err := strconv.ParseUint(args[0][1:], 10, 64)
			if err != nil {
				return false
			}
			if uid != src.UID() {
				return false
			}
		} else {
			if args[0] != types.Unstring(src.Var("name")) {
				return false
			}
		}
	}
	switch settings.Src.Type {
	case types.SrcSettingTypeUsr:
		if settings.Src.In {
			return src.Var("loc") == usr
		} else {
			// src = usr
			return src == usr
		}
	case types.SrcSettingTypeOView:
		util.FIXME("come up with a more efficient way to do this")
		if settings.Src.In {
			// src in oview(N)
			var objects []types.Value
			if settings.Src.Dist == types.SrcDistUnspecified {
				objects = WorldOf(src).View1(usr, ViewExclusive)
			} else {
				objects = WorldOf(src).View(uint(settings.Src.Dist), usr, ViewExclusive)
			}
			for _, obj := range objects {
				if obj == src {
					return true
				}
			}
			return false
		} else {
			panic("support not implemented for proc src setting 'src = oview(...)'")
		}
	default:
		panic("support not implemented for proc src setting " + settings.Src.Type.String())
	}
}

func (v Verb) ResolveArgs(src *types.Datum, usr *types.Datum, args []string) ([]types.Value, error) {
	settings, ok := types.UnpackDatum(src).ProcSettings(v.VisibleName)
	if !ok {
		panic("should not have gotten here; verb should not have matched")
	}
	if settings.Src.In && len(args) >= 1 {
		args = args[1:]
	}
	results := make([]types.Value, len(args))
	for _, _ = range args {
		return nil, fmt.Errorf("unimplemented")
	}
	return results, nil
}

func (v Verb) Apply(src *types.Datum, usr *types.Datum, args []types.Value) {
	src.Invoke(usr, v.ProcName, args...)
}

func NewVerb(visibleName, defType, procName string) Verb {
	return Verb{
		VisibleName:  visibleName,
		DefiningType: defType,
		ProcName:     procName,
	}
}
