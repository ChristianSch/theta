package repo

import "github.com/ChristianSch/Theta/domain/models"

type ConversationRepo interface {
	// CreateConversation creates a new conversation with the given model
	CreateConversation(model string) (models.Conversation, error)
	// GetConversation returns the conversation with the given id
	GetConversation(id string) (models.Conversation, error)
	AddMessage(id string, message models.Message) (models.Conversation, error)
}
