package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/shadiestgoat/log"
)

var wsUp = &websocket.Upgrader{
	HandshakeTimeout: 5 * time.Second,
	ReadBufferSize:   2048,
	WriteBufferSize:  2048,
	Error: nil,
	CheckOrigin: nil,
	EnableCompression: true,
}

var PORT = "3000"

func init() {
	godotenv.Load()
	if v, ok := os.LookupEnv("PORT"); ok {
		PORT = v
	}
}

func init() {
	log.Init(log.NewLoggerPrint(), log.NewLoggerFile("logs/log"))
}

func init() {
	db.items = LoadOldStorage()
	if len(db.items) != 0 {
		log.Success("Loaded %d messages", len(db.items))
	}
}

func init() {
	InitStorage()
}

func main() {
	log.Success("Starting on port %s", PORT)

	r := chi.NewRouter()

	r.Get("/messages", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(db.Items())

		if log.ErrorIfErr(err, "messages get marshal") {
			w.WriteHeader(500)
			w.Write(nil)
			return
		}
		
		w.Write(b)
	})

	r.Post(`/messages`, func(w http.ResponseWriter, r *http.Request) {
		msg := &Message{}
		defer w.Write(nil)

		err := json.NewDecoder(r.Body).Decode(msg)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		
		out := AddMsg(msg)

		if !out {
			w.WriteHeader(400)
			return
		}
	})

	r.HandleFunc(`/ws`, func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUp.Upgrade(w, r, nil)
		if log.ErrorIfErr(err, "upgrading ws") {
			return
		}

		ws.AddConn(conn)
	})

	s := &http.Server{
		Addr: ":" + PORT,
		Handler: r,
	}

	var errClose = make(chan bool)

	go func ()  {
		err := s.ListenAndServe()
		if log.ErrorIfErr(err, "loading listen") {
			errClose <- true
		}
	}()

	endApp := make(chan os.Signal, 4)

	signal.Notify(endApp, os.Interrupt)

	select {
	case <- endApp:
		ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		err := s.Shutdown(ctx)
		log.ErrorIfErr(err, "closing http server")
		cancel()
	case <- errClose:
		log.Error("Bad stuff - server dead")
	}

	storageFile.Close()
	log.PrintWarn("Closing...")
	ws.CloseConns("Server Shutdown")
	log.Success("Closed!")
}