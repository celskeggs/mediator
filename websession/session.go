package websession

import (
	"fmt"
	"github.com/celskeggs/mediator/resourcepack"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient"
	"github.com/celskeggs/mediator/webclient/sprite"
	"time"
)

type worldServer struct {
	World              WorldAPI
	SingleThread       *util.SingleThread
	Subscribers        map[chan struct{}]struct{}
	LoadedResourcePack *resourcepack.ResourcePack
}

func (ws worldServer) ResourcePack() *resourcepack.ResourcePack {
	return ws.LoadedResourcePack
}

func (ws worldServer) Connect() webclient.ServerSession {
	subscription := make(chan struct{}, 1)
	session := &worldSession{
		WS:           ws,
		Active:       true,
		Subscription: subscription,
	}
	ws.SingleThread.Run("AddPlayer()", func() {
		util.FIXME("actually populate the key, rather than always using a Guest")
		session.Player = ws.World.AddPlayer("")
		ws.Subscribers[subscription] = struct{}{}
	})
	return session
}

type worldSession struct {
	WS           worldServer
	Player       PlayerAPI
	Active       bool
	Subscription chan struct{}
}

// MUST be called from SingleThread context
func (ws *worldSession) removeSubscription() {
	_, exists := ws.WS.Subscribers[ws.Subscription]
	if exists {
		delete(ws.WS.Subscribers, ws.Subscription)
		close(ws.Subscription)
	}
}

func (ws *worldSession) Close() {
	if !ws.Active {
		panic("session already closed")
	}
	ws.Active = false
	ws.WS.SingleThread.Run("Close()", func() {
		ws.removeSubscription()
		ws.Player.Remove()
	})
}

var totalTimeSpent time.Duration = 0
var countTimeSpent = 0

func (e *worldSession) OnMessage(cmd webclient.Command) {
	if e.Active {
		e.WS.SingleThread.Run("OnMessage()", func() {
			if !e.Player.IsValid() {
				e.removeSubscription()
			} else {
				start := time.Now()
				e.Player.Command(cmd)
				total := time.Now().Sub(start)
				totalTimeSpent += total
				countTimeSpent += 1
				fmt.Printf("update: %v\t, avg=%v\n", total, totalTimeSpent/time.Duration(countTimeSpent))
			}
		})
	}
}

func (e *worldSession) BeginSend(send func(update *sprite.ViewUpdate) error) {
	go func() {
		defer func() {
			_ = send(nil)
		}()
		var sv sprite.SpriteView
		var lines []string
		var sounds []sprite.Sound
		var flicks []sprite.Flick
		first := true
		for range e.Subscription {
			diff := false
			e.WS.SingleThread.Run("Render()", func() {
				sv2 := e.Player.Render()
				if !sv.Equal(sv2) {
					diff = true
					sv = sv2
				}
				lines, sounds, flicks = e.Player.PullRequests()
			})
			vup := sprite.ViewUpdate{
				TextLines: lines,
				Sounds:    sounds,
				Flicks:    flicks,
			}
			if diff || first {
				vup.NewState = &sv
			}
			if vup.TextLines != nil || vup.NewState != nil {
				if send(&vup) != nil {
					break
				}
			}
			first = false
		}
	}()
}

func consumeAnyOutstanding(c <-chan struct{}) {
	foundMessage := true
	for foundMessage {
		select {
		case <-c:
		default:
			foundMessage = false
		}
	}
}

func (ws *worldServer) Ticker(fps int) {
	period := time.Second / time.Duration(fps)
	go func() {
		next := time.Now().Add(period)
		for {
			here := time.Now()
			remaining := next.Sub(here)
			if remaining > 0 {
				time.Sleep(remaining)
			}
			next = here.Add(period)
			ws.SingleThread.Run("Tick()", func() {
				ws.World.Tick()
			})
		}
	}()
}

func LaunchServer(world WorldAPI, pack *resourcepack.ResourcePack) error {
	// TODO: teardown for SingleThread and our subscriber?
	ws := worldServer{
		World:              world,
		SingleThread:       util.NewSingleThread(),
		Subscribers:        make(map[chan struct{}]struct{}),
		LoadedResourcePack: pack,
	}
	updates := world.SubscribeToUpdates()
	if updates == nil {
		panic("update channel cannot be nil")
	}
	go func() {
		for range updates {
			// this way, even if we run slow, it doesn't matter!
			consumeAnyOutstanding(updates)
			ws.SingleThread.Run("Update()", func() {
				for subscriber := range ws.Subscribers {
					select {
					case subscriber <- struct{}{}:
						// notify if we can
					default:
						// if we can't... well, it already has a notification pending, so we're fine
					}
				}
			})
		}
		// TODO: maybe it should sometimes?
		panic("update stream should never end")
	}()
	ws.Ticker(10)
	return webclient.LaunchHTTP(ws)
}
