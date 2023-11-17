package main

import (
	"strings"

	"github.com/shadiestgoat/log"
)

const MAX_MESSAGES = 300

type Message struct {
	// Author names are encoded as Name, and the last 6 bytes are the author's HEX code 
	Author string `json:"author"`
	Content string `json:"content"`
}

var db = NewRing[*Message](MAX_MESSAGES)

var allowedCharacters = []byte{'\t', '\n'}

var allowedCharMap = map[byte]bool{}

func init() {
	for _, b := range allowedCharacters {
		allowedCharMap[b] = true
	}
}

func AddMsg(msg *Message) bool {
	if msg == nil {
		return false
	}
	msg.Content = strings.TrimSpace(msg.Content)
	
	if len(msg.Author) < 7 || len(msg.Content) == 0 {
		return false
	}

	c := msg.Author[len(msg.Author)-6:]

	for _, b := range c {
		if ('0' <= b && b <= '9') || ('A' <= b && b <= 'F') || ('a' <= b && b <= 'f') {
			continue
		}
		return false
	}

	for _, b := range msg.Author {
		if b < 32 {
			return false
		}
	}
	for i := 0; i < len(msg.Content); i++ {
		b := msg.Content[i]

		if b < 32 && !allowedCharMap[b] {
			return false
		}
	}
	
	db.Insert(msg)

	go ws.WriteMsg(msg)
	log.Success("New Message from %s (%s): %s", msg.Author[:len(msg.Author)-6], c, msg.Content)
	go storageFile.WriteString(msg.StorageString())

	return true
}
