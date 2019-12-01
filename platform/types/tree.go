package types

type TypeTree interface {
	Parent(path TypePath) TypePath
	New(realm *Realm, path TypePath, params ...Value) Value
}
