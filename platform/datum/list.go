package datum

import (
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"reflect"
)

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

func NewListFromRefs(initial ...*types.Ref) types.Value {
	return List{&ConcreteList{Contents: initial}}
}

func NewList(initial ...types.Value) types.Value {
	refs := make([]*types.Ref, len(initial))
	for i, init := range initial {
		refs[i] = types.Reference(init)
	}
	return NewListFromRefs(refs...)
}

// converts from []<value> to []types.Value
func ToValueSlice(initial interface{}) []types.Value {
	util.FIXME("avoid needing reflection in this codebase")
	val := reflect.ValueOf(initial)
	if val.Kind() != reflect.Slice {
		panic("attempt to run ToValueSlice on something that's not a slice: " + val.String())
	} else {
		elements := make([]types.Value, val.Len())
		for i := 0; i < len(elements); i++ {
			elements[i] = val.Index(i).Interface().(types.Value)
		}
		return elements
	}
}

func NewListFromSlice(initial interface{}) types.Value {
	return NewList(ToValueSlice(initial)...)
}

func Elements(list types.Value) []types.Value {
	ll := list.(List)
	result := make([]types.Value, ll.Length())
	for i := 0; i < len(result); i++ {
		result[i] = ll.Get(i).Dereference()
	}
	return result
}

func ElementsAsRefs(list types.Value) []*types.Ref {
	ll := list.(List)
	result := make([]*types.Ref, ll.Length())
	for i := 0; i < len(result); i++ {
		result[i] = ll.Get(i)
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

// slice should be an empty slice of the type we want for the output.
func ElementsAsType(slice interface{}, list types.Value) interface{} {
	elements := Elements(list)
	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice || sliceVal.Len() > 0 {
		panic("attempt to use non-empty slice as first argument to ElementsAsType")
	}
	build := reflect.MakeSlice(sliceVal.Type(), len(elements), len(elements))
	for i, elem := range elements {
		build.Index(i).Set(reflect.ValueOf(elem))
	}
	return build.Interface()
}
