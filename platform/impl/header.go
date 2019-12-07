package impl

//go:generate go run github.com/celskeggs/mediator/boilerplate
import (
	_ "github.com/celskeggs/mediator/platform/atoms"
	_ "github.com/celskeggs/mediator/platform/datum"
	_ "github.com/celskeggs/mediator/platform/world"
)
