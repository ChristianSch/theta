package outbound

import (
	"errors"

	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/websocket/v2"
)

// adapter for sending messages to the user via websocket

type SendFiberWebsocketMessageConfig struct {
	// dependencies
	Log outbound.Log
}

type SendFiberWebsocketMessage struct {
	cfg SendFiberWebsocketMessageConfig
}

func NewSendFiberWebsocketMessage(cfg SendFiberWebsocketMessageConfig) *SendFiberWebsocketMessage {
	return &SendFiberWebsocketMessage{
		cfg: cfg,
	}
}

func (s *SendFiberWebsocketMessage) SendMessage(message string, connection interface{}) error {
	fiberConn, ok := connection.(*websocket.Conn)
	if !ok {
		s.cfg.Log.Error("invalid connection type", outbound.LogField{Key: "connection", Value: connection})
		return errors.New("invalid connection type")
	}

	err := fiberConn.WriteMessage(websocket.TextMessage, []byte(message))
	s.cfg.Log.Debug("sent message", outbound.LogField{Key: "message", Value: message})
	return err
}
