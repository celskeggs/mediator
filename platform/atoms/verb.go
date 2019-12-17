package atoms

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
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

func (v Verb) Invoke(name string, parameters ...types.Value) types.Value {
	panic(fmt.Sprintf("no such proc %q on /verb", name))
}

func (v Verb) String() string {
	return fmt.Sprintf("[verb name=%q proc=%v/%s]", v.VisibleName, v.DefiningType, v.ProcName)
}

func (v Verb) Matches(name string, src *types.Datum, usr *types.Datum) bool {
	util.FIXME("apply src matching")
	return name == v.VisibleName
}

func (v Verb) ResolveArgs(src *types.Datum, usr *types.Datum, args []string) ([]types.Value, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (v Verb) Apply(src *types.Datum, usr *types.Datum, strings []types.Value) {
	panic("unimplemented")
}

func NewVerb(visibleName, defType, procName string) Verb {
	return Verb{
		VisibleName:  visibleName,
		DefiningType: defType,
		ProcName:     procName,
	}
}
