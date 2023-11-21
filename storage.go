package main

import (
	"errors"
	"os"

	"github.com/shadiestgoat/log"
)

const STORAGE_FILE_NAME = "messages.log"

const (
	B_RS byte = 30
	B_US byte = 31
)

func (msg Message) StorageString() string {
	return msg.Author + string(B_US) + msg.Content + string(B_RS)
}

func LoadOldStorage() []*Message {
	stat, err := os.Stat(STORAGE_FILE_NAME)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*Message{}
		}

		log.Fatal("Can't stat old messages file")
	}
	if stat.IsDir() {
		log.Fatal("Messages file is folder, quitting...")
	}
	size := stat.Size()

	f, err := os.OpenFile(STORAGE_FILE_NAME, os.O_RDONLY, 0755)
	log.FatalIfErr(err, "Opening old storage file")

	messages := []*Message{}
	
	var i int64 = 2 // last byte is always EOF, and the one before is RS 

	curMsg := &Message{}
	curF := []byte{}

	setStr := func (str *string) {
		for i, j := 0, len(curF)-1; i < j; i, j = i+1, j-1 {
			curF[i], curF[j] = curF[j], curF[i]
		}
		*str = string(curF)
		curF = []byte{}
	}

	for i <= size && len(messages) < MAX_MESSAGES {
		f.Seek(-i, 2)
		bs := make([]byte, 1)
		f.Read(bs)
		b := bs[0]

		switch bs[0] {
		case B_RS:
			setStr(&curMsg.Author)
			messages = append(messages, curMsg)
			curMsg = &Message{}
		case B_US:
			setStr(&curMsg.Content)
		default:
			curF = append(curF, b)
		}

		i++
	}

	if len(curF) != 0 {
		setStr(&curMsg.Author)
		messages = append(messages, curMsg)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages
}

var storageFile *os.File

func InitStorage() {
	var err error

	storageFile, err = os.OpenFile(STORAGE_FILE_NAME, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0755)
	log.FatalIfErr(err, "Opening storage file for wr")
}
