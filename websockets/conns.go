package websockets

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct{}

var (
	conns []*websocket.Conn
	mtx   sync.Mutex
)

func init() {
	conns = []*websocket.Conn{}
}

func (h *Hub) Add(c *websocket.Conn) {
	mtx.Lock()
	conns = append(conns, c)
	mtx.Unlock()
}

func (h *Hub) Publish(b []byte) {
	mtx.Lock()
	defer mtx.Unlock()

	goodConns := []*websocket.Conn{}

	for _, c := range conns {
		err := c.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			log.Printf("error writing to websocket; closing: %s", err)
			c.Close()
		} else {
			goodConns = append(goodConns, c)
		}
	}

	conns = goodConns
}
