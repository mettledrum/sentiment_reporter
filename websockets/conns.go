package websockets

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

// Hub manages websocket connections
type Hub struct{}

type conn struct {
	Done chan bool
	ID   string
	WS   *websocket.Conn
}

var (
	badConns     []string
	badConnsLock sync.RWMutex
	connMap      map[string]conn
	connMapLock  sync.RWMutex
)

func init() {
	connMap = map[string]conn{}
	badConns = []string{}
}

// Add creates a new ws connection and returns a done chan
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

	connMapLock.Lock()
	connMap[c.ID] = c
	connMapLock.Unlock()

	return c.Done
}

// Publish writes bytes to all connections
// it removes connections and calls done chan if error
func (h *Hub) Publish(b []byte) {
	wg := sync.WaitGroup{}

	connMapLock.RLock()
	for _, c := range connMap {
		wg.Add(1)
		go write(c, b, &wg)
	}
	connMapLock.RUnlock()

	wg.Wait()

	cleanup()
}

func cleanup() {
	badConnsLock.RLock()

	for _, id := range badConns {
		connMapLock.RLock()
		c := connMap[id]
		connMapLock.RUnlock()

		c.Done <- true
		c.WS.Close()

		connMapLock.Lock()
		delete(connMap, id)
		connMapLock.Unlock()
	}

	badConnsLock.RUnlock()

	badConnsLock.Lock()
	badConns = []string{}
	badConnsLock.Unlock()
}

func write(c conn, b []byte, wg *sync.WaitGroup) {
	log.Println("writing message")

	err := c.WS.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Printf("error writing message to websocket: %s", err)

		badConnsLock.Lock()
		badConns = append(badConns, c.ID)
		badConnsLock.Unlock()
	}
	wg.Done()
}
