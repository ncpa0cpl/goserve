package utils

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WsController struct {
	connections []*websocket.Conn
	mutex       *sync.Mutex
}

func CreateWsController() *WsController {
	return &WsController{
		connections: make([]*websocket.Conn, 0),
		mutex:       &sync.Mutex{},
	}
}

func (controller *WsController) AddConnection(conn *websocket.Conn) {
	controller.mutex.Lock()
	controller.connections = append(controller.connections, conn)
	controller.mutex.Unlock()

	go func() {
		defer conn.Close()
		defer controller.RemoveConnection(conn)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func (controller *WsController) RemoveConnection(conn *websocket.Conn) {
	controller.mutex.Lock()
	for i, c := range controller.connections {
		if c == conn {
			controller.connections = append(
				controller.connections[:i],
				controller.connections[i+1:]...,
			)
			break
		}
	}
	controller.mutex.Unlock()

}

func (controller *WsController) SendToAll(strMsg string) {
	msg := []byte(strMsg)
	controller.mutex.Lock()
	for _, conn := range controller.connections {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
	controller.mutex.Unlock()
}
