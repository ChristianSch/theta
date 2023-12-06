package outbound

type SendMessageService interface {
	SendMessage(message string, connection interface{}) error
}
