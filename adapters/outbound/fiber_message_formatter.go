package outbound

import (
	"bytes"
	"html/template"

	"github.com/ChristianSch/Theta/domain/models"
)

type FiberMessageFormatterConfig struct {
	MessageTemplatePath string
}

type FiberMessageFormatter struct {
	cfg FiberMessageFormatterConfig
}

func NewFiberMessageFormatter(cfg FiberMessageFormatterConfig) *FiberMessageFormatter {
	return &FiberMessageFormatter{
		cfg: cfg,
	}
}

type MessageData struct {
	Author    string
	Text      string
	MessageId string
	IsGpt     bool
}

func (f *FiberMessageFormatter) Format(message models.Message) (string, error) {
	tmpl, err := template.ParseFiles(f.cfg.MessageTemplatePath)
	if err != nil {
		return "", err
	}

	// author
	var author string
	if message.Type == models.UserMessage {
		author = "You"
	} else {
		author = "GPT"
	}

	// buffer to write the template to
	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, MessageData{
		Author:    author,
		Text:      message.Text,
		MessageId: message.Id,
		IsGpt:     message.Type == models.GptMessage,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
