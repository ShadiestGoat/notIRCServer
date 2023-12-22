package ws

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/notIRCServer/db"
	"github.com/shadiestgoat/notIRCServer/users"
	"github.com/shadiestgoat/stopper"
)

var serv = &WSRoom{
	Mutex: &sync.Mutex{},
	conns: map[string]Conn{},
}

const (
	INTERNAL_PREFIX = "$"
	ALL_ROOM        = INTERNAL_PREFIX + "*"
)

func AddConn(u *users.User, cID string, conn *websocket.Conn) {
	c := &realConn{
		Mutex:    &sync.Mutex{},
		isClosed: false,
		c:        conn,
	}

	serv.AddConn(c, ALL_ROOM)

	for uName, allowed := range u.ReadWhispers {
		if allowed {
			serv.AddConn(c, INTERNAL_PREFIX+uName, cID)
		}
	}
}

func WriteMsg(msg *db.Msg, cID string) {
	if strings.HasPrefix(cID, INTERNAL_PREFIX) {
		cID = ""
	}

	b, err := json.Marshal(msg)

	if log.ErrorIfErr(err, "marshal of ws msg") {
		return
	}

	pm, err := websocket.NewPreparedMessage(websocket.TextMessage, b)
	if log.ErrorIfErr(err, "creating pm") {
		return
	}

	serv.WriteToRoom(pm, cID, INTERNAL_PREFIX+msg.To)
}

var stop = make(chan bool)

func Init(stop *stopper.Receiver) {
	go func() {
		t := time.NewTicker(15 * time.Second)
		i := 0

		for {
			select {
			case <-t.C:
				if i == 3 {
					serv.Clean()
					i = 0
				} else {
					serv.Ping()
					i++
				}
			case <-stop.C:
				serv.Clean()
				stop.Done()
				return
			}
		}
	}()
}

func Stop() {
	serv.Close("server closing")
}
