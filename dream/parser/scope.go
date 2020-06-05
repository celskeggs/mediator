package parser

type marker struct{}

// This doesn't actually track types... because the parser doesn't care about them!
type Scope struct {
	vars map[string]marker
}

func NewScope() *Scope {
	return &Scope{
		vars: map[string]marker{},
	}
}

func (vs *Scope) HasVar(name string) bool {
	_, ok := vs.vars[name]
	return ok
}

func (vs *Scope) AddVar(name string) {
	if vs.HasVar(name) {
		panic("attempt to override existing variable " + name)
	}
	vs.vars[name] = marker{}
}

func (vs *Scope) RemoveVar(name string) {
	if !vs.HasVar(name) {
		panic("attempt to remove nonexistent variable " + name)
	}
	delete(vs.vars, name)
}
