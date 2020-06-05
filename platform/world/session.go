package world

import (
	"fmt"
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient"
	"github.com/celskeggs/mediator/webclient/sprite"
	"github.com/celskeggs/mediator/websession"
	"math/rand"
	"sort"
)

type worldAPI struct {
	World   *World
	updates chan struct{}
}

var _ websession.WorldAPI = &worldAPI{}

func (w *worldAPI) AddPlayer(key string) websession.PlayerAPI {
	if key == "" {
		key = fmt.Sprintf("Guest-%v", rand.Uint64())
	}
	client := w.World.CreateNewPlayer(key)
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
			md.StartStatContext(mob.(*types.Datum))
			util.FIXME("handle Stat sleeping correctly")
			p.Invoke(mob.(*types.Datum), "Stat")
			client.statDisplay = md.EndStatContext()
		} else {
			// cannot run stat for this client; return empty display
			client.statDisplay = sprite.StatDisplay{}
		}
	}
	// update walk
	for _, movable := range w.World.FindAllType("/atom/movable") {
		UpdateWalk(movable)
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
	center, viewAtoms, stats, verbs, verbsOn := p.API.World.RenderClientView(p.Client)

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
	view.Verbs = verbs

	if center != nil {
		cX, cY := XY(center)
		shiftX, shiftY := (cX-viewDist)*SpriteSize, (cY-viewDist)*SpriteSize

		layers := map[int][]sprite.GameSprite{}
		for _, visibleAtom := range viewAtoms {
			x, y := XY(visibleAtom)
			found, layer, s := visibleAtom.Var("appearance").(atoms.Appearance).ToSprite(x*SpriteSize-shiftX, y*SpriteSize-shiftY, visibleAtom.Var("dir").(common.Direction))
			if found {
				s.Name = types.Unstring(visibleAtom.Var("name"))
				s.Verbs = verbsOn[visibleAtom.(*types.Datum)]
				s.UID = visibleAtom.(*types.Datum).UID()
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

func (p playerAPI) PullRequests() (lines []string, sounds []sprite.Sound, flicks []sprite.Flick) {
	return PullClientRequests(p.Client)
}
