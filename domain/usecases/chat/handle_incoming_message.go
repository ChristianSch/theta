package chat

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ChristianSch/Theta/domain/models"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/ChristianSch/Theta/domain/ports/outbound/repo"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type IncomingMessageHandlerConfig struct {
	// dependencies
	Sender           outbound.SendMessageService
	Formatter        outbound.MessageFormatter
	Llm              outbound.LlmService
	PostProcessors   []outbound.PostProcessor
	ConversationRepo repo.ConversationRepo
}

type IncomingMessageHandler struct {
	// dependencies
	cfg IncomingMessageHandlerConfig
}

func NewIncomingMessageHandler(cfg IncomingMessageHandlerConfig) *IncomingMessageHandler {
	// sort post processors by order
	sort.SliceStable(cfg.PostProcessors, func(i, j int) bool {
		return cfg.PostProcessors[i].Order < cfg.PostProcessors[j].Order
	})

	for _, p := range cfg.PostProcessors {
		log.Debug("post processor registered",
			outbound.LogField{Key: "name", Value: p.Name},
			outbound.LogField{Key: "order", Value: p.Order},
		)
	}

	return &IncomingMessageHandler{
		cfg: cfg,
	}
}

func (h *IncomingMessageHandler) Handle(message models.Message, conversation models.Conversation, connection interface{}) error {
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

	// chunks contains *all* received chunks (= the whole answer)
	var chunks []byte

	// read answer chunks and update the message chunk by chunk
	fn := func(ctx context.Context, chunk []byte) error {
		if len(chunk) == 0 {
			log.Debug("final answer received", outbound.LogField{Key: "component", Value: "handle_incoming_message"})
			done <- true
			return nil
		}

		chunks = append(chunks, chunk...)

		log.Debug("chunk received",
			outbound.LogField{Key: "component", Value: "handle_incoming_message"},
			outbound.LogField{Key: "length", Value: len(chunk)},
		)

		// post process the message
		// initially, only the original message
		res := chunks

		for _, p := range h.cfg.PostProcessors {
			// each post processor gets the result of the previous one
			res, err = p.Processor.PostProcess(res)
		}

		// update sent message by swapping it entirely (allows for properly render html,
		// which would not work with a simple addition of deltas)
		if err := h.cfg.Sender.SendMessage(
			fmt.Sprintf("<div hx-swap-oob=\"innerHTML:#%s\">%s</div>", msgId, string(res)),
			connection); err != nil {
			return err
		}

		return nil
	}

	// send message to llm via a goroutine so we can wait for the answer
	go func() {
		err = h.cfg.Llm.SendMessage(message.Text, conversation.Messages, fn)
		if err != nil {
			done <- true
		}
	}()

	// wait for answer to be finished
	<-done

	// add message and answer to conversation
	h.cfg.ConversationRepo.AddMessage(conversation.Id, message)
	h.cfg.ConversationRepo.AddMessage(conversation.Id, models.Message{
		Text: string(chunks),
		Type: models.GptMessage,
	})

	if err != nil {
		log.Error("error while receiving answer",
			outbound.LogField{Key: "component", Value: "handle_incoming_message"},
			outbound.LogField{Key: "error", Value: err},
		)
		return err
	}

	return nil
}
