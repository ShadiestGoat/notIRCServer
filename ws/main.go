package ws

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shadiestgoat/log"
)

type Conn interface {
	Ping()

	Write(pm *websocket.PreparedMessage, excludedClient string)

	// Not safe for concurrency
	Close(reason string)

	// Cleans the instance, and returns true if the conn is empty & should be removed
	Clean() bool

	AddConn(c Conn, cIDs ...string)
}

type realConn struct {
	*sync.Mutex

	isPinging atomic.Bool

	isClosed bool
	c        *websocket.Conn
}

// Real conns are no-ops
func (*realConn) AddConn(_ Conn, _ ...string) {
	return
}

func (r *realConn) Clean() bool {
	r.Lock()
	defer r.Unlock()

	return r.isClosed
}

func (c *realConn) Write(pm *websocket.PreparedMessage, _ string) {
	c.Lock()
	defer c.Unlock()

	if c.isClosed {
		return
	}

	c.c.WritePreparedMessage(pm)
}

func (c *realConn) Close(reason string) {
	c.Lock()
	defer c.Unlock()

	c.close(reason)
}

func (c *realConn) close(reason string) {
	if c.isClosed {
		return
	}

	log.Warn("Closing WS: %s", reason)

	c.isClosed = true
	c.c.WriteControl(websocket.CloseMessage, nil, time.Time{})
	c.c.Close()
}

func (c *realConn) Ping() {
	if c.isPinging.Swap(true) {
		return
	}

	c.Lock()
	defer c.Unlock()
	defer c.isPinging.Store(false)

	if c.isClosed {
		return
	}

	m := make(chan bool)
	t := time.NewTimer(7 * time.Second)

	c.c.SetPongHandler(func(appData string) error {
		m <- true
		return nil
	})

	err := c.c.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
	if err != nil {
		c.c.SetPongHandler(nil)
		c.close("Failed to write ping")

		return
	}

	select {
	case <-m:
		return
	case <-t.C:
		c.c.SetPongHandler(nil)
		c.close("No pong")
	}
}
