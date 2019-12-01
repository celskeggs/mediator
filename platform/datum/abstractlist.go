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
}

type List struct {
	ListProvider
}

var _ types.Value = List{}

func (l List) Reference() *types.Ref {
	return &types.Ref{l}
}

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

func (l List) Invoke(name string, parameters ...types.Value) types.Value {
	panic("unimplemented: list procs")
}

func (l List) String() string {
	return fmt.Sprintf("[list of length %d]", l.Length())
}
