package atoms

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
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

func NewVerb(visibleName, defType, procName string) Verb {
	return Verb{
		VisibleName:  visibleName,
		DefiningType: defType,
		ProcName:     procName,
	}
}
