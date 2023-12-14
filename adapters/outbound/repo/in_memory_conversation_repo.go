package repo

import (
	"errors"
	"sync"

	"github.com/ChristianSch/Theta/domain/models"
	"github.com/google/uuid"
)

const (
	ErrConversationNotFound = "conversation not found"
)

type InMemoryConversationRepo struct {
	conversations map[string]models.Conversation
	mu            sync.Mutex
}

func NewInMemoryConversationRepo() *InMemoryConversationRepo {
	return &InMemoryConversationRepo{
		conversations: make(map[string]models.Conversation),
	}
}

func nextId() string {
	return uuid.New().String()
}

func (r *InMemoryConversationRepo) CreateConversation(model string) (models.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv := models.Conversation{
		Id:    nextId(),
		Model: model,
	}
	r.conversations[conv.Id] = conv

	return conv, nil
}

func (r *InMemoryConversationRepo) GetConversation(id string) (models.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, ok := r.conversations[id]
	if !ok {
		return models.Conversation{}, errors.New(ErrConversationNotFound)
	}

	return conv, nil
}

func (r *InMemoryConversationRepo) AddMessage(id string, message models.Message) (models.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, ok := r.conversations[id]
	if !ok {
		return models.Conversation{}, errors.New(ErrConversationNotFound)
	}

	conv.Messages = append(conv.Messages, message)
	r.conversations[id] = conv

	return conv, nil
}
