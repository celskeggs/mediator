package world

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/types"
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

var _ websession.WorldAPI = &worldAPI{}

func (w *worldAPI) AddPlayer() websession.PlayerAPI {
	util.FIXME("get a key for this")
	client := w.World.CreateNewPlayer("")
	if !types.IsType(client.Var("mob"), "/mob") {
		panic("nonexistent or invalid mob")
	}
	w.Update()
	return playerAPI{
		Client: client,
		API:    w,
	}
}

func (w *worldAPI) Tick() {
	// update stat panels
	for _, player := range w.World.clients {
		p := player.Dereference()
		_, client := ClientDataChunk(p)
		mob := p.Var("mob")
		md, ok := atoms.MobDataChunk(mob)
		if ok {
			md.StartStatContext()
			util.FIXME("handle Stat sleeping correctly")
			p.Invoke(mob.(*types.Datum), "Stat")
			client.statDisplay = md.EndStatContext()
		} else {
			// cannot run stat for this client; return empty display
			client.statDisplay = sprite.StatDisplay{}
		}
	}
	w.Update()
}

// should be called by functions in session.go primarily.
func (w *worldAPI) Update() {
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
	Client *types.Datum
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
		InvokeVerb(p.Client, cmd.Verb)
		p.API.Update()
	}
}

const SpriteSize = 32

func (p playerAPI) Render() sprite.SpriteView {
	center, viewAtoms, stats := p.API.World.RenderClientView(p.Client)

	util.FIXME("don't use hardcoded tile sizes here")
	util.FIXME("add adjacent cell movement animations")

	viewDist := types.Unuint(p.Client.Var("view"))
	sizeInCells := (viewDist * 2) + 1
	viewportSize := sizeInCells * SpriteSize

	var view sprite.SpriteView
	view.WindowTitle = p.API.World.Name
	view.ViewPortWidth = viewportSize
	view.ViewPortHeight = viewportSize
	view.Stats = stats

	if center != nil {
		cX, cY := XY(center)
		shiftX, shiftY := (cX-viewDist)*SpriteSize, (cY-viewDist)*SpriteSize

		layers := map[int][]sprite.GameSprite{}
		for _, visibleAtom := range viewAtoms {
			x, y := XY(visibleAtom)
			found, layer, s := visibleAtom.Var("appearance").(atoms.Appearance).ToSprite(x*SpriteSize-shiftX, y*SpriteSize-shiftY, visibleAtom.Var("dir").(common.Direction))
			if found {
				layers[layer] = append(layers[layer], s)
			}
		}
		var layerOrder []int
		for layer := range layers {
			layerOrder = append(layerOrder, layer)
		}
		sort.Ints(layerOrder)
		for _, layer := range layerOrder {
			view.Sprites = append(view.Sprites, layers[layer]...)
		}
	}

	return view
}

func (p playerAPI) PullRequests() (lines []string, sounds []sprite.Sound) {
	return PullClientRequests(p.Client)
}
