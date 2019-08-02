package main

import (
	"github.com/celskeggs/mediator/web"
	"log"
	"time"
)

type ExampleServer struct {
	LastID uint
}

func (e ExampleServer) Connect() web.ServerSession {
	e.LastID += 1
	log.Println("opened session: ", e.LastID)
	return &ExampleSession{
		ID: e.LastID,
	}
}

type ExampleSession struct {
	ID uint
}

func (e ExampleSession) Close() {
	log.Println("closed session: ", e.ID)
}

type MessageHolder struct {
	Test string
}

func (e ExampleSession) NewMessageHolder() interface{} {
	return &MessageHolder{}
}

func (e ExampleSession) OnMessage(message interface{}) {
	mh := message.(*MessageHolder)
	log.Println("got message holder for: ", e.ID, " with message ", mh.Test)
}

type GameSprite struct {
	Icon string `json:"icon"`
	SourceX uint `json:"sx"`
	SourceY uint `json:"sy"`
	SourceWidth uint `json:"sw"`
	SourceHeight uint `json:"sh"`
	X uint `json:"x"`
	Y uint `json:"y"`
	Width uint `json:"w"`
	Height uint `json:"h"`
}

type GameSpriteMessage struct {
	Sprites map[string]GameSprite `json:"sprites"`
}

func (e ExampleSession) BeginSend(send func(interface{}) error) {
	go func() {
		defer func() {
			_ = send(nil)
		}()
		for {
			log.Println("sending first message for: ", e.ID)
			if send(GameSpriteMessage{
				Sprites: map[string]GameSprite{
					"test": {
						Icon: "cheese.dmi",
						X:    128,
						Y:    128,
					},
				},
			}) != nil {
				break
			}
			log.Println("sent message for: ", e.ID)
			time.Sleep(time.Second / 2)
			if send(GameSpriteMessage{
				Sprites: map[string]GameSprite{
					"test": {
						Icon: "cheese.dmi",
						X:    128,
						Y:    192,
					},
				},
			}) != nil {
				break
			}
			log.Println("sent another message for: ", e.ID)
			time.Sleep(time.Second / 2)
		}
	}()
}

func main() {
	es := ExampleServer{}
	err := web.LaunchHTTP(es)
	panic(err)
}