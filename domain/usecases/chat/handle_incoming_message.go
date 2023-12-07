package chat

import (
	"github.com/ChristianSch/Theta/domain/models"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
)

type IncomingMessageHandlerConfig struct {
	// dependencies
	Sender    outbound.SendMessageService
	Formatter outbound.MessageFormatter
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
	answer, err := h.cfg.Formatter.Format(message)
	if err != nil {
		return err
	}

	return h.cfg.Sender.SendMessage(
		answer,
		connection)
}
