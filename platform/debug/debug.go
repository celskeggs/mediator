package debug

import (
	"reflect"
	"strings"
)

type Output struct {
	Sections []string
}

func (o *Output) Println(text string) {
	println(strings.Repeat("  ", len(o.Sections)) + " " + text)
}

func (o *Output) Header(section string) {
	o.Println("<=== " + section + " ===>")
	o.Sections = append(o.Sections, section)
}

func (o *Output) Footer() {
	lastindex := len(o.Sections) - 1
	last := o.Sections[lastindex]
	o.Sections = o.Sections[:lastindex]
	o.Println("++++ " + last + " ++++")
}

func (o *Output) End() {
	if len(o.Sections) > 0 {
		panic("nonzero indent at the end")
	}
}

func DumpFields(rvalue reflect.Value, o *Output) {
	for i := 0; i < rvalue.NumField(); i++ {
		field := rvalue.Field(i)
		name := rvalue.Type().Field(i).Name
		if name == "Impl" {
			continue
		}
		if i == 0 && field.Type().Kind() == reflect.Struct {
			DumpFields(field, o)
			continue
		}
		o.Println("field: " + name)
		if !field.CanInterface() {
			o.Println("HIDDEN")
		} else {
			DumpReflect(field.Interface(), o)
		}
	}
}

func DumpReflect(i interface{}, o *Output) {
	rvalue := reflect.ValueOf(i)
	if rvalue.Kind() == reflect.Ptr {
		rvalue = rvalue.Elem()
	}
	if rvalue.Kind() == reflect.Bool {
		if rvalue.Bool() {
			o.Println("value: true")
		} else {
			o.Println("value: false")
		}
		return
	}
	if rvalue.Kind() != reflect.Struct {
		o.Println("value: " + rvalue.String())
		return
	}
	o.Header("struct " + rvalue.Type().Name())
	DumpFields(rvalue, o)
	o.Footer()
}
