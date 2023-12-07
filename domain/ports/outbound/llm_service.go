package outbound

import "context"

// ResponseHandler is a function that is called for every data chunk that is received. EOF is indicated by an empty chunk.
type ResponseHandler func(ctx context.Context, chunk []byte) error

type LlmService interface {
	ListModels() ([]string, error)
	SetModel(model string) error
	SendMessage(prompt string, context []string, resHandler ResponseHandler) error
}
