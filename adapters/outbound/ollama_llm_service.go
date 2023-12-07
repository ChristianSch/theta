package outbound

import (
	cont "context"
	"errors"

	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/jmorganca/ollama/api"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type OllamaLlmService struct {
	client *api.Client
	llm    *ollama.LLM
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

	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		return err
	}

	s.llm = llm
	s.log.Debug("set ollama model", outbound.LogField{Key: "model", Value: model})

	return nil
}

func (s *OllamaLlmService) SendMessage(prompt string, context []string, resHandler outbound.ResponseHandler) error {
	if s.llm == nil {
		return errors.New(ModelNotSetError)
	}

	ctx := cont.Background()

	_, err := s.llm.Call(ctx, prompt,
		llms.WithStreamingFunc(resHandler),
	)
	return err
}
