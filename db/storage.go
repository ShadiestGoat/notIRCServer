package db

import (
	"errors"
	"os"
	"sync"

	"github.com/shadiestgoat/log"
)

const (
	b_RS byte = 30
	b_US byte = 31
)

func (msg Msg) StorageString() string {
	return msg.Author + string(b_US) + msg.To + string(b_US) + msg.Content + string(b_RS)
}

func loadOldDB(fileName string, maxMsgs int) []*Msg {
	stat, err := os.Stat(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*Msg{}
		}

		log.Fatal("Can't stat msg db")
	}

	if stat.IsDir() {
		log.Fatal("Msg db is folder, quitting...")
	}

	size := stat.Size()

	f, err := os.OpenFile(fileName, os.O_RDONLY, 0755)
	log.FatalIfErr(err, "Opening old msg db")

	messages := []*Msg{}

	var i int64 = 2 // last byte is always EOF, and the one before is RS

	curMsg := &Msg{}
	curF := []byte{}
	// 0 - content, 1 - to, 2 - author
	fID := 0

	setStr := func(str *string) {
		for i, j := 0, len(curF)-1; i < j; i, j = i+1, j-1 {
			curF[i], curF[j] = curF[j], curF[i]
		}
		*str = string(curF)
		curF = []byte{}
	}

	for i <= size && (maxMsgs == -1 || len(messages) < maxMsgs) {
		f.Seek(-i, 2)
		bs := make([]byte, 1)
		f.Read(bs)
		b := bs[0]

		switch bs[0] {
		case b_RS:
			setStr(&curMsg.Author)
			messages = append(messages, curMsg)
			curMsg = &Msg{}
			fID = 0
		case b_US:
			if fID == 0 {
				setStr(&curMsg.Content)
			} else {
				setStr(&curMsg.To)
			}
			fID++
		default:
			curF = append(curF, b)
		}

		i++
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages
}

var (
	dbFile   *os.File
	fileLock = &sync.Mutex{}

	curFlush  []Msg
	flushLock = &sync.Mutex{}
)

func flush() {
	log.Debug("Flushing...")

	flushLock.Lock()

	if len(curFlush) == 0 {
		flushLock.Unlock()
		log.Debug("Nothing to flush")
		return
	}

	str := ""

	for _, v := range curFlush {
		str += v.StorageString()
	}

	curFlush = []Msg{}

	flushLock.Unlock()

	fileLock.Lock()
	dbFile.WriteString(str)
	fileLock.Unlock()

	log.Debug("Flush complete")
}
