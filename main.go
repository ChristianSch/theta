package main

import (
	"encoding/json"
	"errors"
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

	// init llms first to see if we have any models available
	ollama, err := outbound.NewOllamaLlmService(log)
	if err != nil {
		panic(err)
	}

	ollamaModels, err := ollama.ListModels()
	if err != nil {
		panic(err)
	}

	log.Debug("available ollama models", outboundPorts.LogField{Key: "models", Value: ollamaModels})

	// TODO: init openai
	if len(ollamaModels) == 0 {
		panic(errors.New("no models available"))
	}

	// convenience check: if we have more than one model, use the first one
	ollama.SetModel(ollamaModels[0])

	web := inbound.NewFiberWebServer(inbound.FiberWebServerConfig{
		Port:                8080,
		TemplatesPath:       "./infrastructure/views",
		TemplatesExtension:  ".gohtml",
		StaticResourcesPath: "./infrastructure/static",
	}, inbound.FiberWebServerAdapters{Log: log})

	msgSender := outbound.NewSendFiberWebsocketMessage(outbound.SendFiberWebsocketMessageConfig{Log: log})
	msgFormatter := outbound.NewFiberMessageFormatter(outbound.FiberMessageFormatterConfig{
		MessageTemplatePath: "./infrastructure/views/components/message.gohtml",
	})
	msgHandler := chat.NewIncomingMessageHandler(chat.IncomingMessageHandlerConfig{
		Sender:    msgSender,
		Formatter: msgFormatter,
		Llm:       ollama,
	})

	web.AddRoute("GET", "/", func(ctx interface{}) error {
		log.Debug("handling request", outboundPorts.LogField{Key: "path", Value: "/"})
		fiberCtx := ctx.(*fiber.Ctx)
		return fiberCtx.Render("chat", fiber.Map{
			"Title": "",
			"Model": ollamaModels[0],
		}, "layouts/main")
	})

	web.AddWebsocketRoute("/ws/chat", func(conn interface{}) error {
		log.Debug("handling websocket request", outboundPorts.LogField{Key: "path", Value: "/ws/chat"})
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
