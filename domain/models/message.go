package models

import "time"

type MessageType int

const (
	UserMessage MessageType = iota
	GptMessage
)

type Message struct {
	AuthorId  string
	Text      string
	Timestamp time.Time
	Type      MessageType
}
