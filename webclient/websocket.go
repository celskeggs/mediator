package webclient

import (
	"github.com/celskeggs/mediator/webclient/sprite"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type ErrorStop struct{}

func (e ErrorStop) Error() string {
	return "Cannot send: connection closed"
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketClient struct {
	api ServerSession
}

type WebSocketServer struct {
	api ServerAPI
}

func setConnectionTimeout(conn *websocket.Conn, timeout time.Duration) error {
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}
	conn.SetPongHandler(func(string) error {
		// TODO: figure out if this error handling is correct
		return conn.SetReadDeadline(time.Now().Add(timeout))
	})
	return nil
}

func (wss *WebSocketServer) handleSessionTransmit(session ServerSession, conn *websocket.Conn) {
	defer func() {
		session.Close()
		err := conn.Close()
		if err != nil {
			log.Printf("error during session close: %v", err)
		}
	}()
	conn.SetReadLimit(1024)
	err := setConnectionTimeout(conn, time.Minute)
	if err != nil {
		log.Printf("error during session setup: %v", err)
		return
	}
	for {
		var command Command
		err := conn.ReadJSON(&command)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		session.OnMessage(command)
	}
}

func (wss *WebSocketServer) handleSessionReceive(session ServerSession, conn *websocket.Conn) {
	ticker := time.NewTicker(40 * time.Second)
	sendChannel := make(chan *sprite.SpriteView)
	terminated := false
	defer func() {
		terminated = true
		ticker.Stop()
		err := conn.Close()
		if err != nil {
			log.Printf("error during session close: %v", err)
		}
		for range sendChannel {
			log.Printf("still receiving messages after close")
		}
	}()
	session.BeginSend(func(message *sprite.SpriteView) error {
		if message == nil {
			close(sendChannel)
			return nil
		} else if terminated {
			return ErrorStop{}
		} else {
			sendChannel <- message
			return nil
		}
	})
	for {
		err := conn.SetWriteDeadline(time.Now().Add(time.Second * 60))
		if err != nil {
			log.Printf("error setting write deadline: %v", err)
			return
		}
		select {
		case message, ok := <-sendChannel:
			if !ok {
				err := conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("error sending close message: %v", err)
					return
				}
				return
			}

			err = conn.WriteJSON(message)
			if err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (wss *WebSocketServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("websocket establishment error: %v", err)
		http.Error(writer, "Failed", 400)
		return
	}
	session := wss.api.Connect()
	go wss.handleSessionTransmit(session, conn)
	go wss.handleSessionReceive(session, conn)
}

func NewWebSocketServer(api ServerAPI) *WebSocketServer {
	return &WebSocketServer{
		api: api,
	}
}
