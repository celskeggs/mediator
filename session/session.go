package session

import (
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/web"
)

type worldServer struct {
	World        WorldAPI
	SingleThread *util.SingleThread
	Subscribers  map[chan struct{}]struct{}
}

func (ws worldServer) Connect() web.ServerSession {
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

func (ws *worldSession) Close() {
	if !ws.Active {
		panic("session already closed")
	}
	ws.Active = false
	ws.WS.SingleThread.Run(func() {
		ws.Player.Remove()
		delete(ws.WS.Subscribers, ws.Subscription)
	})
}

func (e *worldSession) NewMessageHolder() interface{} {
	return &Command{}
}

func (e *worldSession) OnMessage(message interface{}) {
	cmd := *message.(*Command)
	if e.Active {
		e.WS.SingleThread.Run(func() {
			e.Player.Command(cmd)
		})
	}
}

func (e *worldSession) BeginSend(send func(interface{}) error) {
	go func() {
		defer func() {
			_ = send(nil)
		}()
		var sv SpriteView
		for range e.Subscription {
			diff := false
			e.WS.SingleThread.Run(func() {
				sv2 := e.Player.Render()
				if !sv.Equal(sv2) {
					diff = true
					sv = sv2
				}
			})
			if diff {
				if send(sv) != nil {
					break
				}
			}
		}
		panic("subscription should not end")
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

func LaunchServer(world WorldAPI) error {
	// TODO: teardown for SingleThread and our subscriber?
	ws := worldServer{
		World:        world,
		SingleThread: util.NewSingleThread(),
		Subscribers:  make(map[chan struct{}]struct{}),
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
	return web.LaunchHTTP(ws)
}
