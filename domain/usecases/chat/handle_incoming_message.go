package chat

import (
	"context"
	"errors"

	"github.com/ChristianSch/Theta/domain/models"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/fiber/v2/log"
)

type IncomingMessageHandlerConfig struct {
	// dependencies
	Sender    outbound.SendMessageService
	Formatter outbound.MessageFormatter
	Llm       outbound.LlmService
}

type IncomingMessageHandler struct {
	// dependencies
	cfg IncomingMessageHandlerConfig
}

func NewIncomingMessageHandler(cfg IncomingMessageHandlerConfig) *IncomingMessageHandler {
	return &IncomingMessageHandler{
		cfg: cfg,
	}
}

func (h *IncomingMessageHandler) Handle(message models.Message, connection interface{}) error {
	// check if message is valid
	// get answer ...
	msg, err := h.cfg.Formatter.Format(message)
	if err != nil {
		return err
	}

	// send message for question
	if err := h.cfg.Sender.SendMessage(
		msg,
		connection); err != nil {
		return err
	}

	// complete answer
	bytes := []byte{}

	// this channel keeps track if the answer is finished
	done := make(chan bool)
	defer close(done)

	// get answer
	fn := func(ctx context.Context, chunk []byte) error {
		if len(chunk) == 0 {
			log.Debug("final answer received", outbound.LogField{Key: "component", Value: "handle_incoming_message"})
			done <- true
			return nil
		}

		log.Debug("chunk received",
			outbound.LogField{Key: "component", Value: "handle_incoming_message"},
			outbound.LogField{Key: "length", Value: len(chunk)},
		)

		bytes = append(bytes, chunk...)
		return nil
	}

	go func() {
		err = h.cfg.Llm.SendMessage(message.Text, []string{}, fn)
		if err != nil {
			done <- true
		}
	}()

	// wait for answer to be finished
	<-done

	if err != nil {
		log.Error("error while receiving answer",
			outbound.LogField{Key: "component", Value: "handle_incoming_message"},
			outbound.LogField{Key: "error", Value: err},
		)
		return err
	}

	if len(bytes) == 0 {
		log.Error("no answer received",
			outbound.LogField{Key: "component", Value: "handle_incoming_message"})
		return errors.New("no answer received")
	}

	log.Debug("answer completed",
		outbound.LogField{Key: "component", Value: "handle_incoming_message"},
		outbound.LogField{Key: "length", Value: len(bytes)},
	)

	// send message for answer
	answer, err := h.cfg.Formatter.Format(models.Message{
		Text: string(bytes),
		Type: models.GptMessage,
	})
	if err != nil {
		return err
	}

	return h.cfg.Sender.SendMessage(
		answer,
		connection)
}
