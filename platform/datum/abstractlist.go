package datum

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
)

// remember: indices start from 1!

type ListProvider interface {
	Length() int
	Get(i int) *types.Ref
	Set(i int, v *types.Ref)
	Append(v *types.Ref)
	RemoveLast()
	RemoveIndex(i int)
}

type List struct {
	ListProvider
}

var _ types.Value = List{}

func (l List) Var(name string) types.Value {
	switch name {
	case "len":
		return types.Int(l.ListProvider.Length())
	default:
		panic("no such field " + name + " on list")
	}
}

func (l List) SetVar(name string, value types.Value) {
	switch name {
	case "len":
		target := types.Unint(value)
		for target > l.ListProvider.Length() {
			l.ListProvider.Append(nil)
		}
		for target < l.ListProvider.Length() {
			l.ListProvider.RemoveIndex(l.ListProvider.Length())
		}
	default:
		panic("no such field " + name + " on list")
	}
}

func (l List) Invoke(usr *types.Datum, name string, parameters ...types.Value) types.Value {
	switch name {
	case "+":
		value := types.Param(parameters, 0)
		if otherList, ok := value.(List); ok {
			// concatenate
			refsA := ElementsAsRefs(l)
			refsB := ElementsAsRefs(otherList)
			return NewListFromRefs(append(refsA, refsB...)...)
		} else {
			// append
			refsA := ElementsAsRefs(l)
			return NewListFromRefs(append(refsA, types.Reference(value))...)
		}
	case "<<":
		for _, element := range Elements(l) {
			if types.IsType(element, "/mob") {
				element.Invoke(usr, "<<", parameters...)
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("unimplemented: list proc %q", name))
	}
}

func (l List) String() string {
	return fmt.Sprintf("[list of length %d]", l.Length())
}
