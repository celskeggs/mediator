package datum

import "github.com/celskeggs/mediator/platform/types"

type ConcreteList struct {
	Contents []*types.Ref
}

var _ ListProvider = &ConcreteList{}

func (c *ConcreteList) Length() int {
	return len(c.Contents)
}

func (c *ConcreteList) Get(i int) *types.Ref {
	return c.Contents[i]
}

func (c *ConcreteList) Set(i int, v *types.Ref) {
	c.Contents[i] = v
}

func (c *ConcreteList) Append(v *types.Ref) {
	c.Contents = append(c.Contents, v)
}

func (c *ConcreteList) RemoveLast() {
	c.Contents = c.Contents[:len(c.Contents)-1]
}

func (c *ConcreteList) RemoveIndex(i int) {
	copy(c.Contents[i:], c.Contents[i+1:])
	c.Contents = c.Contents[:len(c.Contents)-1]
}

func NewList(initial ...*types.Ref) types.Value {
	return List{&ConcreteList{Contents: initial}}
}

func Elements(list types.Value) []types.Value {
	ll := list.(List)
	result := make([]types.Value, ll.Length())
	for i := 0; i < len(result); i++ {
		result[i] = ll.Get(i).Dereference()
	}
	return result
}

func ElementsDatums(list types.Value) []*types.Datum {
	ll := list.(List)
	result := make([]*types.Datum, ll.Length())
	for i := 0; i < len(result); i++ {
		result[i] = ll.Get(i).Dereference().(*types.Datum)
	}
	return result
}
