package platform

func IsMob(atom IAtom) bool {
	_, ismob := atom.(IMob)
	return ismob
}
