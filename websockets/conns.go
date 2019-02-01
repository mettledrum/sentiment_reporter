package websockets

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type hub struct{}

type conn struct {
	Done chan bool
	ID   string
	WS   *websocket.Conn
}

var (
	badConns     map[string]bool
	badConnsLock sync.RWMutex
	connMap      map[string]conn
	connMapLock  sync.RWMutex

	// Hub is the exported manager of websocket connections
	Hub hub
)

func init() {
	connMap = map[string]conn{}
	badConns = map[string]bool{}

	Hub = hub{}
}

// Add creates a new ws connection and returns a done chan
func (h *hub) Add(w http.ResponseWriter, r *http.Request) chan bool {
	up := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := up.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket error: %s\n", err)
		return nil
	}

	c := conn{
		Done: make(chan bool),
		ID:   getID(),
		WS:   ws,
	}

	connMapLock.Lock()
	connMap[c.ID] = c
	connMapLock.Unlock()

	log.Printf("created websocket: %s\n", c.ID)

	return c.Done
}

// Publish writes bytes to all connections
// closes, removes connections, and calls done chan if error
func (h *hub) Publish(b []byte) {
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

func write(c conn, b []byte, wg *sync.WaitGroup) {
	log.Printf("writing message to websocket: %s\n", c.ID)

	err := c.WS.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Printf("error writing message to websocket: %s; %s\n", c.ID, err)

		// slate bad connections for removal
		badConnsLock.Lock()
		badConns[c.ID] = true
		badConnsLock.Unlock()
	}

	wg.Done()
}

func cleanup() {
	badConnsLock.Lock()

	for id := range badConns {

		connMapLock.Lock()

		c := connMap[id]
		fmt.Printf("cleaning up websocket: %s\n", c.ID)

		err := c.WS.Close() // close connection
		if err != nil {
			fmt.Printf("error closing websocket: %s; %s\n", c.ID, err)
		}
		c.Done <- true      // notify done with connection
		delete(connMap, id) // rm from connections map

		connMapLock.Unlock()
	}

	badConns = map[string]bool{} // reset bad connections map
	badConnsLock.Unlock()
}

func getID() string {
	return fmt.Sprintf("%v", uuid.NewV4())[0:5]
}
