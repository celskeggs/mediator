package main

import (
	"github.com/celskeggs/mediator/web"
	"log"
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

func (e ExampleSession) BeginSend(sender chan<- interface{}) {
	go func() {
		log.Println("sending first message for: ", e.ID)
		sender <- struct{
			Test string
		}{
			Test: "Hello JSON World",
		}
		log.Println("sent first message for: ", e.ID)
	}()
}

func main() {
	es := ExampleServer{}
	err := web.LaunchHTTP(es)
	panic(err)
}