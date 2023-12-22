package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/notIRCServer/db"
	"github.com/shadiestgoat/notIRCServer/users"
	"github.com/shadiestgoat/notIRCServer/ws"
	"github.com/shadiestgoat/stopper"
)

var wsUp = &websocket.Upgrader{
	HandshakeTimeout:  5 * time.Second,
	ReadBufferSize:    2048,
	WriteBufferSize:   2048,
	Error:             nil,
	CheckOrigin:       nil,
	EnableCompression: true,
}

var (
	PORT      = "3000"
	RING_SIZE = 700
)

func init() {
	godotenv.Load()

	log.Init(log.NewLoggerPrint(), log.NewLoggerFile("logs/log"))

	if v := os.Getenv("PORT"); v != "" {
		PORT = v
	}
	maxMsgs := os.Getenv("RING_SIZE")

	if maxMsgs != "" {
		size, err := strconv.ParseUint(maxMsgs, 10, 64)
		log.FatalIfErr(err, "parsing RING_SIZE")

		if size == 0 {
			log.Warn("Parsed RING_SIZE, but it's 0. Using default value...")
		} else {
			RING_SIZE = int(size)
		}
	}
}

var cmds []string

func init() {
	for _, c := range os.Args[1:] {
		if c[0] == '-' {
			continue
		}

		cmds = append(cmds, c)
	}

	if len(cmds) == 0 {
		cmds = []string{"exec"}
	}
}

func main() {
	flag.Parse()

	users.Init(*userEnv, cmds[0] != "exec" || os.Getenv("I_KNOW_WHAT_IM_DOING") == "NO_TOKENS")

	stop := stopper.NewSender(func(closerName string) {
		log.Debug("Service '%v' stopped", closerName)
	}, true)

	dbStop := stop.Register("db")

	switch cmds[0] {
	case "export":
		db.Init(*dbFile, -1, dbStop)
		cmdExport()

		stop.Stop()
		db.Stop()
		return
	case "import":
		log.Fatal("Sorry - not implemented right now. Wait for future release...")
		return
	}

	db.Init(*dbFile, RING_SIZE, dbStop)

	log.Success("Starting on port %s", PORT)

	s := &http.Server{
		Addr:    ":" + PORT,
		Handler: Router(),
	}

	var errClose = make(chan bool)

	go func() {
		err := s.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server-ending err: %v", err)
			errClose <- true
		}
	}()

	endApp := make(chan os.Signal, 4)

	signal.Notify(endApp, os.Interrupt)

	select {
	case <-endApp:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := s.Shutdown(ctx)
		log.ErrorIfErr(err, "closing http server")
		cancel()
	case <-errClose:
		log.Error("Bad stuff - server dead")
	}

	log.Warn("Closing...")
	ws.Stop()
	log.Debug("Closed ws conns")
	stop.Stop()
	log.Debug("finished routine stops")
	db.Stop()
	log.Debug("finished closing db")

	log.Success("Closed!")
}
