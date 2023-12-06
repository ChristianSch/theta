package main

import (
	"encoding/json"
	"time"

	"github.com/ChristianSch/Theta/adapters/inbound"
	"github.com/ChristianSch/Theta/adapters/outbound"
	"github.com/ChristianSch/Theta/domain/models"
	outboundPorts "github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/ChristianSch/Theta/domain/usecases/chat"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type WebsocketMsg struct {
	Message string `json:"message,omitempty"`
	Headers struct {
	} `json:"HEADERS,omitempty"`
}

func main() {
	log := outbound.NewZapLogger(outbound.ZapLoggerConfig{Debug: true})
	web := inbound.NewFiberWebServer(inbound.FiberWebServerConfig{
		Port:                8080,
		TemplatesPath:       "./infrastructure/views",
		TemplatesExtension:  ".gohtml",
		StaticResourcesPath: "./infrastructure/static",
	}, inbound.FiberWebServerAdapters{Log: log})

	msgSender := outbound.NewSendFiberWebsocketMessage(outbound.SendFiberWebsocketMessageConfig{Log: log})
	msgHandler := chat.NewIncomingMessageHandler(chat.IncomingMessageHandlerConfig{Sender: msgSender})

	web.AddRoute("GET", "/", func(ctx interface{}) error {
		log.Debug("handling request", outboundPorts.LogField{Key: "path", Value: "/"})
		fiberCtx := ctx.(*fiber.Ctx)
		return fiberCtx.Render("chat", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/main")
	})

	web.AddWebsocketRoute("/ws", func(conn interface{}) error {
		log.Debug("handling websocket request", outboundPorts.LogField{Key: "path", Value: "/ws"})
		fiberConn := conn.(*websocket.Conn)

		for {
			messageType, message, err := fiberConn.ReadMessage()
			if err != nil {
				log.Error("error while reading message", outboundPorts.LogField{Key: "error", Value: err})
				break
			}

			// message is json, marshall it to WebsocketMsg
			var wsMsg WebsocketMsg
			if err := json.Unmarshal([]byte(message), &wsMsg); err != nil {
				log.Error("error while unmarshalling message", outboundPorts.LogField{Key: "error", Value: err})
				break
			}

			log.Debug("received message",
				outboundPorts.LogField{Key: "message", Value: wsMsg.Message},
				outboundPorts.LogField{Key: "messageType", Value: messageType},
			)

			msg := models.Message{
				AuthorId:  "", // FIXME:
				Text:      wsMsg.Message,
				Timestamp: time.Now(),
				Type:      models.UserMessage,
			}

			if err := msgHandler.Handle(msg, fiberConn); err != nil {
				log.Error("error while writing message", outboundPorts.LogField{Key: "error", Value: err})
				break
			}
		}

		return nil
	})

	if err := web.Start(); err != nil {
		panic(err)
	}
}
