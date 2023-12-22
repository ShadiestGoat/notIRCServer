package db

import (
	"strings"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/notIRCServer/users"
	"github.com/shadiestgoat/notIRCServer/utils"
)

type MsgBase struct {
	Content string `json:"content" yaml:"content"`
	// Is either a '*' or a specific username to indicate a whisper
	To string `json:"to" yaml:"to"`
}

type Msg struct {
	MsgBase

	Author string `json:"from" yaml:"from"`
}

var db *Ring[*Msg]

var allowedCharacters = []byte{'\t', '\n'}

var allowedCharMap = map[byte]bool{}

func init() {
	for _, b := range allowedCharacters {
		allowedCharMap[b] = true
	}
}

func AddMsg(msg *Msg) error {
	msg.Content = strings.TrimSpace(msg.Content)

	for i := 0; i < len(msg.Content); i++ {
		b := msg.Content[i]

		if b < 32 && !allowedCharMap[b] {
			return utils.HTTPErr{
				Msg:    "Illegal chars >:(",
				Status: 400,
			}
		}
	}

	db.Insert(msg)

	log.Success("New Message from %s to %s: %s", msg.Author, msg.To, msg.Content)

	flushLock.Lock()
	curFlush = append(curFlush, *msg)
	flushLock.Unlock()

	return nil
}

func GetMessages() []*Msg {
	return db.Items()
}

func FilterMsgsForPerms(msgs []*Msg, allowedWhispers map[string]bool) []*Msg {
	out := make([]*Msg, 0, len(msgs))

	for _, m := range msgs {
		if m.To == "*" || allowedWhispers[m.To] {
			out = append(out, m)
		}
	}

	return out
}

func FilterForUser(msgs []*Msg, u *users.User) []*Msg {
	if u.AbleToReadAllWhispers {
		return msgs
	}
	return FilterMsgsForPerms(msgs, u.ReadWhispers)
}

func DeleteLast() {
	db.Lock()

	db.deleteLast()

	flushLock.Lock()

	fileNeeded := false

	if len(curFlush) == 0 {
		fileNeeded = true
		fileLock.Lock()
	} else {
		curFlush = curFlush[:len(curFlush)-1]
	}

	db.Unlock()

	flushLock.Unlock()

	if !fileNeeded {
		return
	}

	stat, _ := dbFile.Stat()
	size := stat.Size()

	if size != 0 {
		found := int64(-1)

		for i := int64(2); i <= size; i++ {
			dbFile.Seek(-i, 2)

			bs := make([]byte, 1)
			dbFile.Read(bs)

			if bs[0] == b_RS {
				found = i
				break
			}
		}

		if found == -1 {
			found = size + 1
		}

		dbFile.Truncate(size - found + 1)
	}

	fileLock.Unlock()
}
