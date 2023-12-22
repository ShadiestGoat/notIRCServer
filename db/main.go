package db

import (
	"os"
	"sync"
	"time"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/stopper"
)

func Init(fileName string, maxMsgs int, stop *stopper.Receiver) {
	var err error

	dbFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	log.FatalIfErr(err, "Opening db for wr")

	db = &Ring[*Msg]{
		RWMutex: sync.RWMutex{},
		items:   loadOldDB(fileName, maxMsgs),
		maxLen:  maxMsgs,
	}

	go func() {
		t := time.NewTicker(5 * time.Minute)

		for {
			select {
			case <-t.C:
				flush()
			case <-stop.C:
				log.Debug("Gotta stop flushing")
				flush()
				stop.Done()
				return
			}
		}
	}()
}

func Stop() {
	fileLock.Lock()
	dbFile.Close()
	fileLock.Unlock()
}
