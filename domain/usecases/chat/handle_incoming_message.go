package chat

import (
	"time"

	"github.com/ChristianSch/Theta/domain/models"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
)

type IncomingMessageHandlerConfig struct {
	// dependencies
	Sender outbound.SendMessageService
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

	answer := "<div hx-swap-oob=\"beforeend:#content\"><div class=\"user-message message\">" +
		message.Timestamp.Format(time.DateTime) + " " + message.Text + "</div>" +
		"<div class=\"gpt-message message\">pong!</div></div>"

	return h.cfg.Sender.SendMessage(
		answer,
		connection)
}
