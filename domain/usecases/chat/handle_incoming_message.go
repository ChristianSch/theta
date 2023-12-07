package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/ChristianSch/Theta/domain/models"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
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
	msgId := fmt.Sprintf("msg-%s", strings.Split(uuid.New().String(), "-")[0])
	log.Debug("starting processing of message", outbound.LogField{Key: "messageId", Value: msgId})

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

	// send initial (empty) answer that we'll update as we get more chunks
	answer, err := h.cfg.Formatter.Format(models.Message{
		Text: "",
		Type: models.GptMessage,
		Id:   msgId,
	})
	if err != nil {
		return err
	}

	if err = h.cfg.Sender.SendMessage(answer, connection); err != nil {
		return err
	}

	// this channel keeps track if the answer is finished
	done := make(chan bool)
	defer close(done)

	// read answer chunks and update the message chunk by chunk
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

		// update sent message
		if err := h.cfg.Sender.SendMessage(
			fmt.Sprintf("<div hx-swap-oob=\"beforeend:#%s\">%s</div>", msgId, string(chunk)),
			connection); err != nil {
			return err
		}

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

	return nil
}
