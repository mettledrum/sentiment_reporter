package websockets

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type Hub struct{}

type conn struct {
	Done chan bool
	ID   string
	WS   *websocket.Conn
}

var (
	connMap map[string]conn
	mtx     sync.RWMutex
)

func init() {
	connMap = map[string]conn{}
}

func (h *Hub) Add(w http.ResponseWriter, r *http.Request) chan bool {
	up := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := up.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket error: %s", err)
		return nil
	}

	c := conn{
		Done: make(chan bool),
		ID:   fmt.Sprintf("%v", uuid.NewV4()),
		WS:   ws,
	}

	mtx.Lock()
	connMap[c.ID] = c
	mtx.Unlock()

	return c.Done
}

func (h *Hub) Publish(b []byte) {
	wg := sync.WaitGroup{}
	mtx.RLock()

	for _, c := range connMap {
		wg.Add(1)
		go writeCleanup(c, b, &wg)
	}

	mtx.RUnlock()
	wg.Wait()
}

func writeCleanup(c conn, b []byte, wg *sync.WaitGroup) {
	log.Println("writing message")

	err := c.WS.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		c.WS.Close()
		c.Done <- true

		mtx.Lock()
		delete(connMap, c.ID)
		mtx.Unlock()

		log.Printf("error writing message to websocket: %s; closing", err)
	}
	wg.Done()
}
