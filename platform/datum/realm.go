package datum

var TRACE = false

type Realm struct {
	datums map[*Datum]struct{}
}

func (r *Realm) add(d *Datum) {
	if _, found := r.datums[d]; found {
		panic("datum already found in realm")
	}
	r.datums[d] = struct{}{}
	if TRACE {
		println("added datum", d, "of type", d.Type, "to realm")
	}
}

func (r *Realm) remove(d *Datum) {
	if _, found := r.datums[d]; !found {
		panic("datum not found in realm")
	}
	delete(r.datums, d)
	if TRACE {
		println("removed datum", d, "of type", d.Type, "from realm")
	}
}

func (r *Realm) FindAll(predicate func(IDatum) bool) (out []IDatum) {
	for datum := range r.datums {
		if predicate(datum.impl) {
			out = append(out, datum.impl)
		}
	}
	return out
}

// returns nil if not found
func (r *Realm) FindOne(predicate func(IDatum) bool) IDatum {
	for datum := range r.datums {
		if predicate(datum.impl) {
			return datum.impl
		}
	}
	return nil
}
