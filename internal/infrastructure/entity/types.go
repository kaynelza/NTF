package entity

import (
	"time"

	"github.com/kaynelza/NTF/internal/infrastructure/helper"
)

type Notification struct {
	Id        string
	Recipient string
	Title     string
	Body      string
	SendAt    time.Time
	Status    helper.Status
	Attempts  int32
	LastError string
	CreatedAt time.Time
	SentAt    time.Time
}
