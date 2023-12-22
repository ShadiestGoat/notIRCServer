package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Internal conn IDs must start with $
type WSRoom struct {
	*sync.Mutex

	conns map[string]Conn
}

// A more safe way to write a specific room
func (r *WSRoom) WriteToRoom(pm *websocket.PreparedMessage, cID, rID string) {
	r.Lock()
	defer r.Unlock()

	if r.conns[rID] == nil {
		return
	} else {
		r.conns[rID].Write(pm, cID)
	}
}

// AddConn implements Conn.
func (r *WSRoom) AddConn(c Conn, cID ...string) {
	if len(cID) == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()

	if len(cID) == 1 {
		r.conns[cID[0]] = c
	} else {
		if r.conns[cID[0]] == nil {
			r.conns[cID[0]] = &WSRoom{
				Mutex: &sync.Mutex{},
				conns: map[string]Conn{},
			}
		}

		r.conns[cID[0]].AddConn(c, cID[1:]...)
	}
}

func (r *WSRoom) massAction(f func(cID string)) {
	wg := &sync.WaitGroup{}

	wg.Add(len(r.conns))

	for cID := range r.conns {
		go func(cID string) {
			f(cID)
			wg.Done()
		}(cID)
	}

	wg.Wait()
}

func (r *WSRoom) Clean() bool {
	r.Lock()
	defer r.Unlock()

	newConns := map[string]Conn{}

	l := sync.Mutex{}

	r.massAction(func(cID string) {
		c := r.conns[cID]
		if !c.Clean() {
			l.Lock()
			newConns[cID] = c
			l.Unlock()
		}
	})

	r.conns = newConns

	return len(r.conns) == 0
}

func (r *WSRoom) Close(reason string) {
	r.Lock()
	defer r.Unlock()

	r.massAction(func(cID string) {
		r.conns[cID].Close(reason)
	})
}

func (r *WSRoom) Ping() {
	r.Lock()
	defer r.Unlock()

	r.massAction(func(cID string) {
		r.conns[cID].Ping()
	})
}

func (r *WSRoom) Write(pm *websocket.PreparedMessage, excludedClient string) {
	r.Lock()
	defer r.Unlock()

	r.massAction(func(cID string) {
		if cID == excludedClient {
			return
		}

		r.conns[cID].Write(pm, excludedClient)
	})
}
