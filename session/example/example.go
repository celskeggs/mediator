package main

import (
	"github.com/celskeggs/mediator/session"
	"time"
)

type ExampleAPI struct {
	Cheese session.GameSprite
	Updates chan struct{}
}

func (e *ExampleAPI) AddPlayer() session.PlayerAPI {
	return &ExamplePlayer{
		API: e,
	}
}

func (e *ExampleAPI) SubscribeToUpdates() chan struct{} {
	return e.Updates
}

func (e *ExampleAPI) MoveTheCheese() {
	for {
		time.Sleep(time.Second / 10)
		e.Cheese.X += 10
		if e.Cheese.X >= 600 {
			e.Cheese.X = 40
		}
		e.Updates <- struct{}{}
	}
}

type ExamplePlayer struct {
	API *ExampleAPI
}

func (e ExamplePlayer) Remove() {
	// nothing to do
}

func (e ExamplePlayer) Command(cmd session.Command) {
	// nothing to do
}

func (e ExamplePlayer) Render() session.SpriteView {
	return session.SpriteView{
		Sprites: map[string]session.GameSprite{
			"cheese": e.API.Cheese,
		},
	}
}

func main() {
	api := &ExampleAPI{
		Cheese: session.GameSprite{
			Icon: "cheese.dmi",
			X:    300,
			Y:    100,
		},
		Updates: make(chan struct{}),
	}
	go api.MoveTheCheese()
	err := session.LaunchServer(api)
	// should not get here
	panic(err)
}
