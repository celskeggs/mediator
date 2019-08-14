package platform

import (
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient"
	"github.com/celskeggs/mediator/webclient/sprite"
	"github.com/celskeggs/mediator/websession"
	"sort"
)

type worldAPI struct {
	World   *World
	updates chan struct{}
}

func (w *worldAPI) AddPlayer() websession.PlayerAPI {
	util.FIXME("get a key for this")
	client := w.World.CreateNewPlayer("")
	if client.AsClient().mob == nil {
		panic("no mob??")
	}
	w.Update()
	return playerAPI{
		Client: client,
		API:    w,
	}
}

// should be called by functions in session.go primarily.
func (w *worldAPI) Update() {
	util.FIXME("figure out timed updates and how they work with both THIS update system and the SingleThread thing")
	select {
	case w.updates <- struct{}{}:
	default:
		// discard the update.
		// the other end will already know something changed when it gets the next token;
		// we don't need to tell them again.
	}
}

func (w *worldAPI) SubscribeToUpdates() <-chan struct{} {
	return w.updates
}

type playerAPI struct {
	API    *worldAPI
	Client IClient
}

func (p playerAPI) Remove() {
	p.API.World.RemovePlayer(p.Client)
	p.API.Update()
}

func (p playerAPI) IsValid() bool {
	return p.API.World.PlayerExists(p.Client)
}

func (p playerAPI) Command(cmd webclient.Command) {
	if cmd.Verb != "" {
		p.Client.InvokeVerb(cmd.Verb)
		p.API.Update()
	}
}

const SpriteSize = 32

func (p playerAPI) Render() sprite.SpriteView {
	center, atoms := p.Client.RenderViewAsAtoms()

	util.FIXME("don't use hardcoded tile sizes here")
	util.FIXME("add adjacent cell movement animations")

	viewDist := p.Client.AsClient().ViewDistance
	sizeInCells := (viewDist * 2) + 1
	viewportSize := sizeInCells * SpriteSize

	cX, cY, _ := center.XYZ()
	shiftX, shiftY := (cX-viewDist)*SpriteSize, (cY-viewDist)*SpriteSize

	layers := map[int][]sprite.GameSprite{}
	for _, atom := range atoms {
		x, y, _ := atom.XYZ()
		found, layer, s := atom.AsAtom().Appearance.ToSprite(x*SpriteSize-shiftX, y*SpriteSize-shiftY, atom.AsAtom().Direction)
		if found {
			layers[layer] = append(layers[layer], s)
		}
	}
	var layerOrder []int
	for layer := range layers {
		layerOrder = append(layerOrder, layer)
	}
	sort.Ints(layerOrder)

	var view sprite.SpriteView
	view.WindowTitle = p.API.World.Name
	view.ViewPortWidth = viewportSize
	view.ViewPortHeight = viewportSize
	for _, layer := range layerOrder {
		view.Sprites = append(view.Sprites, layers[layer]...)
	}
	return view
}

func (p playerAPI) PullText() (lines []string) {
	return p.Client.PullClientRequests()
}
