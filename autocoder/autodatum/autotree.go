package autodatum

import "github.com/celskeggs/mediator/autocoder/iface"

type autoTree struct {
	G   iface.Gen
	Out iface.OutSource
}

type AutocodeTree func(iface.Gen, OutTree)
type OutTree interface {
	ExtendBuiltin(path string)
	// WORKING HERE...
}

func AutoTree(g iface.Gen, out iface.OutSource) {
	platform := g.Import(PlatformImportPath)

	out.Struct("DerivedWorld",
		func(gen iface.Gen, out iface.OutStruct) {
			out.Include(platform.Type("BaseTreeDefiner"))
		})
}
