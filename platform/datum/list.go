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

func NewList(initial ...*types.Ref) types.Value {
	return List{&ConcreteList{Contents: initial}}
}
