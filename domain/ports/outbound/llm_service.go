package outbound

import "context"

type LlmResponse struct {
	Answer  string
	Context *context.Context
}

type LlmService interface {
	ListModels() ([]string, error)
	SendMessage(messages []string) (LlmResponse, error)
}
