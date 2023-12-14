package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"time"

	"github.com/ChristianSch/Theta/adapters/inbound"
	"github.com/ChristianSch/Theta/adapters/outbound"
	"github.com/ChristianSch/Theta/adapters/outbound/repo"
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

	// all models
	llmModels := ollamaModels

	// TODO: init openai
	if len(ollamaModels) == 0 {
		panic(errors.New("no models available"))
	}

	// convenience check: if we have more than one model, use the first one
	ollama.SetModel(ollamaModels[0])

	web := inbound.NewFiberWebServer(inbound.FiberWebServerConfig{
		Port:                5467,
		TemplatesPath:       "./infrastructure/views",
		TemplatesExtension:  ".gohtml",
		StaticResourcesPath: "./infrastructure/static",
	}, inbound.FiberWebServerAdapters{Log: log})

	// markdown 2 html post processor
	mdToHtmlPostProcessor := outbound.NewMdToHtmlLlmPostProcessor()

	// conversation repo
	convRepo := repo.NewInMemoryConversationRepo()

	msgSender := outbound.NewSendFiberWebsocketMessage(outbound.SendFiberWebsocketMessageConfig{Log: log})
	msgFormatter := outbound.NewFiberMessageFormatter(outbound.FiberMessageFormatterConfig{
		MessageTemplatePath: "./infrastructure/views/components/message.gohtml",
	})
	msgHandler := chat.NewIncomingMessageHandler(chat.IncomingMessageHandlerConfig{
		Sender:    msgSender,
		Formatter: msgFormatter,
		Llm:       ollama,
		PostProcessors: []outboundPorts.PostProcessor{
			{
				Processor: mdToHtmlPostProcessor,
				Order:     0, // first one
				Name:      mdToHtmlPostProcessor.GetName(),
			},
		},
		ConversationRepo: convRepo,
	})

	web.AddRoute("GET", "/", func(ctx interface{}) error {
		fiberCtx := ctx.(*fiber.Ctx)
		return fiberCtx.Render("new_chat", fiber.Map{
			"Title":  "",
			"Models": llmModels,
		}, "layouts/main")
	})

	web.AddRoute("GET", "/chat", func(ctx interface{}) error {
		fiberCtx := ctx.(*fiber.Ctx)
		return fiberCtx.Redirect("/")
	})

	// create new conversation
	web.AddRoute("POST", "/chat", func(ctx interface{}) error {
		fiberCtx := ctx.(*fiber.Ctx)

		// get model from form
		model := fiberCtx.FormValue("model")
		if model == "" {
			log.Error("no model specified", outboundPorts.LogField{Key: "error", Value: "no model specified"})
			return fiberCtx.Redirect("/")
		}

		// get message from form
		message := fiberCtx.FormValue("message")
		if message == "" {
			log.Error("no message specified", outboundPorts.LogField{Key: "error", Value: "no message specified"})
			return fiberCtx.Redirect("/")
		}

		// get conversation
		conv, err := convRepo.CreateConversation(model)
		if err != nil {
			log.Error("error while creating conversation", outboundPorts.LogField{Key: "error", Value: err})
			return err
		}

		fiberCtx.Append("HX-Replace-Url", "/chat/"+conv.Id)

		return fiberCtx.Render("chat", fiber.Map{
			"Title":          "",
			"Models":         llmModels,
			"ConversationId": conv.Id,
			"UserMessage":    message,
		}, "layouts/empty")
	})

	// open existing conversation
	web.AddRoute("GET", "/chat/:id", func(ctx interface{}) error {
		fiberCtx := ctx.(*fiber.Ctx)
		convId := fiberCtx.Params("id")

		// get conversation
		conv, err := convRepo.GetConversation(convId)
		if err != nil {
			log.Error("error while getting conversation", outboundPorts.LogField{Key: "error", Value: err})
			return fiberCtx.Redirect("/")
		}

		var renderedMessages []template.HTML

		for _, msg := range conv.Messages {
			renderedMsg, err := msgFormatter.Format(msg)
			if err != nil {
				log.Error("error while formatting message", outboundPorts.LogField{Key: "error", Value: err})
				return err
			}

			// note that you shouldn't do this under no circumstances, this circumvents the XSS protection
			renderedMessages = append(renderedMessages, template.HTML(renderedMsg))
		}

		return fiberCtx.Render("chat", fiber.Map{
			"Title":          "",
			"Model":          conv.Model,
			"ConversationId": conv.Id,
			"Messages":       renderedMessages,
		}, "layouts/main")
	})

	web.AddWebsocketRoute("/ws/chat/:id", func(conn interface{}) error {
		fiberConn := conn.(*websocket.Conn)
		convId := fiberConn.Params("id")
		log.Debug("handling websocket request",
			outboundPorts.LogField{Key: "path", Value: "/ws/chat/:id"},
			outboundPorts.LogField{Key: "id", Value: convId})

		// get conversation
		conv, err := convRepo.GetConversation(convId)
		if err != nil {
			log.Error("error while getting conversation", outboundPorts.LogField{Key: "error", Value: err})
			return err
		}

		log.Debug("conversation received message", outboundPorts.LogField{Key: "conversation", Value: conv})

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

			if len(wsMsg.Message) > 0 {
				msg := models.Message{
					Text:      wsMsg.Message,
					Timestamp: time.Now(),
					Type:      models.UserMessage,
				}

				// add message to conversation!
				if err := msgHandler.Handle(msg, conv, fiberConn); err != nil {
					log.Error("error while writing message", outboundPorts.LogField{Key: "error", Value: err})
					break
				}
			}
		}

		return nil
	})

	if err := web.Start(); err != nil {
		panic(err)
	}
}
