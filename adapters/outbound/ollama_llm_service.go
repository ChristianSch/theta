package outbound

import (
	cont "context"
	"errors"

	"github.com/ChristianSch/Theta/domain/models"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/jmorganca/ollama/api"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
)

type OllamaLlmService struct {
	client *api.Client
	llm    *ollama.Chat
	model  *string
	log    outbound.Log
}

const (
	ModelNotSetError = "model not set"
)

func NewOllamaLlmService(log outbound.Log) (*OllamaLlmService, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &OllamaLlmService{
		client: client,
		log:    log,
	}, nil
}

func (s *OllamaLlmService) ListModels() ([]string, error) {
	ctx := cont.Background()
	models, err := s.client.List(ctx)
	if err != nil {
		return []string{}, err
	}

	var modelNames []string

	for _, model := range models.Models {
		modelNames = append(modelNames, model.Name)
	}

	return modelNames, nil
}

func (s *OllamaLlmService) SetModel(model string) error {
	s.model = &model

	llm, err := ollama.NewChat(ollama.WithLLMOptions(ollama.WithModel(model)))
	if err != nil {
		return err
	}

	s.llm = llm
	s.log.Debug("set ollama model", outbound.LogField{Key: "model", Value: model})

	return nil
}

func (s *OllamaLlmService) SendMessage(prompt string, context []models.Message, resHandler outbound.ResponseHandler) error {
	if s.llm == nil {
		return errors.New(ModelNotSetError)
	}

	var messages []schema.ChatMessage

	for _, msg := range context {
		if msg.Type == models.UserMessage {
			messages = append(messages, schema.HumanChatMessage{
				Content: msg.Text,
			})
		} else {
			messages = append(messages, schema.AIChatMessage{
				Content: msg.Text,
			})
		}
	}

	messages = append(messages, schema.HumanChatMessage{
		Content: prompt,
	})

	_, err := s.llm.Call(cont.Background(), messages, llms.WithStreamingFunc(resHandler))
	return err
}
