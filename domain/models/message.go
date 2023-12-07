package models

import (
	"time"
)

type MessageType int

const (
	UserMessage MessageType = iota
	GptMessage
)

type Message struct {
	Text      string
	Timestamp time.Time
	Type      MessageType
	Id        string
}
