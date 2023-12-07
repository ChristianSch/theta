package outbound

import "github.com/ChristianSch/Theta/domain/models"

type MessageFormatter interface {
	Format(message models.Message) (string, error)
}
