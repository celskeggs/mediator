package websession

import (
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient"
	"github.com/celskeggs/mediator/webclient/sprite"
	"io/ioutil"
	"path"
	"time"
	"fmt"
)

type worldServer struct {
	World             WorldAPI
	SingleThread      *util.SingleThread
	Subscribers       map[chan struct{}]struct{}
	CoreResourcesDir  string
	ExtraResourcesDir string
}

func (ws worldServer) CoreResourcePath() string {
	return ws.CoreResourcesDir
}

func (ws worldServer) ListResources() (map[string]string, error) {
	contents, err := ioutil.ReadDir(ws.ExtraResourcesDir)
	if err != nil {
		return nil, err
	}
	nameToPath := map[string]string{}
	for _, info := range contents {
		if !info.IsDir() {
			nameToPath[info.Name()] = path.Join(ws.ExtraResourcesDir, info.Name())
		}
	}
	return nameToPath, nil
}

func (ws worldServer) Connect() webclient.ServerSession {
	// TODO: for efficiency, this should probably be bounded and drop messages
	subscription := make(chan struct{})
	session := &worldSession{
		WS:           ws,
		Active:       true,
		Subscription: subscription,
	}
	ws.SingleThread.Run(func() {
		session.Player = ws.World.AddPlayer()
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
	ws.WS.SingleThread.Run(func() {
		ws.removeSubscription()
		ws.Player.Remove()
	})
}

var totalTimeSpent time.Duration = 0
var countTimeSpent = 0

func (e *worldSession) OnMessage(cmd webclient.Command) {
	if e.Active {
		e.WS.SingleThread.Run(func() {
			if !e.Player.IsValid() {
				e.removeSubscription()
			} else {
				start := time.Now()
				e.Player.Command(cmd)
				total := time.Now().Sub(start)
				totalTimeSpent += total
				countTimeSpent += 1
				fmt.Printf("frame: %v\t, avg=%v\n", total, totalTimeSpent/time.Duration(countTimeSpent))
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
		first := true
		for range e.Subscription {
			diff := false
			e.WS.SingleThread.Run(func() {
				sv2 := e.Player.Render()
				if !sv.Equal(sv2) {
					diff = true
					sv = sv2
				}
				lines = e.Player.PullText()
			})
			vup := sprite.ViewUpdate{
				TextLines: lines,
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

func LaunchServer(world WorldAPI, CoreResourcesDir, ExtraResourcesDir string) error {
	// TODO: teardown for SingleThread and our subscriber?
	ws := worldServer{
		World:             world,
		SingleThread:      util.NewSingleThread(),
		Subscribers:       make(map[chan struct{}]struct{}),
		CoreResourcesDir:  CoreResourcesDir,
		ExtraResourcesDir: ExtraResourcesDir,
	}
	updates := world.SubscribeToUpdates()
	if updates == nil {
		panic("update channel cannot be nil")
	}
	go func() {
		for range updates {
			// this way, even if we run slow, it doesn't matter!
			consumeAnyOutstanding(updates)
			ws.SingleThread.Run(func() {
				for subscriber := range ws.Subscribers {
					subscriber <- struct{}{}
				}
			})
		}
		// TODO: maybe it should sometimes?
		panic("update stream should never end")
	}()
	return webclient.LaunchHTTP(ws)
}
