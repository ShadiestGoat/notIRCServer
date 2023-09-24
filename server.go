package main

import (
	"strings"

	"github.com/shadiestgoat/log"
)

type Message struct {
	// Author names are encoded as Name, and the last 6 bytes are the author's HEX code 
	Author string `json:"author"`
	Content string `json:"content"`
}

var db = NewRing[*Message](400)

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
	
	db.Insert(msg)
	
	go ws.WriteMsg(msg)
	go log.Success("New Message from %s (%s): %s", msg.Author[:len(msg.Author)-6], c, msg.Content)
	return true
}
