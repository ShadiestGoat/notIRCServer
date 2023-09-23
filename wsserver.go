package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shadiestgoat/log"
)

var ws = &WSServer{
	Mutex:        &sync.Mutex{},
	closeCleaner: make(chan bool),
	closePinger:  make(chan bool),
	Clients:      []*Conn{},
}

func init() {
	go ws.Cleaner()
	go ws.Pinger()
}

type WSServer struct {
	*sync.Mutex

	closeCleaner chan bool
	closePinger chan bool

	Clients []*Conn
}

func (s *WSServer) CloseConns(reason string) {
	s.Lock()
	defer s.Unlock()

	close(s.closeCleaner)
	close(s.closePinger)

	wg := &sync.WaitGroup{}

	for _, c := range s.Clients {
		wg.Add(1)

		go func(c *Conn) {
			c.Lock()
			c.Close(reason)
			c.Unlock()

			wg.Done()
		}(c)
	}
	
	wg.Wait()
}

func (s *WSServer) AddConn(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()

	conn := &Conn{
		Mutex:    &sync.Mutex{},
		isClosed: false,
		c:        c,
	}

	go func() {
		for {
			_, _, err := c.ReadMessage()

			if err != nil {
				conn.Lock()
				conn.Close("Read Err: " + err.Error())
				conn.Unlock()
				
				break
			}
		}
	}()

	s.Clients = append(s.Clients, conn)
}

func (s *WSServer) Pinger() {
	t := time.NewTicker(15 * time.Second)

	for {
		select {
		case <- s.closePinger:
			return
		case <- t.C:
			log.Debug("Ping Loop")

			s.Lock()
			for _, c := range s.Clients {
				go c.Ping()
			}
			s.Unlock()
		}
	}
}

func (s *WSServer) Cleaner() {
	t := time.NewTicker(30 * time.Second)

	for {
		select {
		case <- s.closeCleaner:
			return
		case <- t.C:
			log.Debug("Close Loop")

			s.Lock()
			newL := &sync.Mutex{}
			newConns := []*Conn{}
			wg := &sync.WaitGroup{}
			
			for _, c := range s.Clients {
				wg.Add(1)

				go func (c *Conn)  {
					defer wg.Done()
					if c.IsClosing() {
						return
					}
					
					newL.Lock()
					newConns = append(newConns, c)
					newL.Unlock()
				}(c)
			}

			wg.Wait()

			s.Clients = newConns
			s.Unlock()
		}
	}
}

func (s *WSServer) WriteMsg(rawMsg any) {
	b, err := json.Marshal(rawMsg)
	if log.ErrorIfErr(err, "marshal for rawMsg") {
		return
	}

	pm, err := websocket.NewPreparedMessage(websocket.TextMessage, b)
	s.Lock()
	for _, c := range s.Clients {
		go c.WriteMsg(pm)
	}
	s.Unlock()
}

type Conn struct {
	*sync.Mutex

	isClosed bool
	c *websocket.Conn
}

func (c *Conn) Ping() {
	c.Lock()
	defer c.Unlock()

	if c.isClosed {
		return
	}

	m := make(chan bool)
	t := time.NewTimer(7 * time.Second)

	c.c.SetPongHandler(func(appData string) error {
		m <- true
		return nil
	})

	err := c.c.WriteControl(websocket.PingMessage, nil, time.Now().Add(5 * time.Second))
	if err != nil {
		c.c.SetPongHandler(nil)
		c.Close("Failed to write ping")

		return
	}

	select {
	case <- m:
		return
	case <- t.C:
		c.c.SetPongHandler(nil)
		c.Close("No pong")
	}
}

func (c *Conn) WriteMsg(pm *websocket.PreparedMessage) {
	c.Lock()
	defer c.Unlock()

	if c.isClosed {
		return
	}

	c.c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := c.c.WritePreparedMessage(pm)
	if err != nil {
		c.Close("Failed to write msg: " + err.Error())
	}
}

func (c *Conn) IsClosing() bool {
	c.Lock()
	defer c.Unlock()
	return c.isClosed
}

// Not safe for concurrency
func (c *Conn) Close(reason string) {
	if c.isClosed {
		return
	}

	log.Warn("Closing WS: %s", reason)

	c.isClosed = true
	c.c.WriteControl(websocket.CloseMessage, nil, time.Time{})
	c.c.Close()
}
