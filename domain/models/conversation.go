package models

import "time"

type Conversation struct {
	Id                string
	ConversationStart time.Time
	Model             string
	Messages          []Message
	// Active means that this conversation can be continued (needs the model to be available in the main context)
	Active bool
}
