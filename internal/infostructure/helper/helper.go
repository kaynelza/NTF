package helper

import "time"

type Status int

const (
	StatusUnspecified Status = iota
	StatusPending
	StatusSending
	StatusSent
	StatusCancelled
	StatusDead
)

type Notification struct {
	Id        string
	Recipient string
	Title     string
	Body      string
	LastError string
	SendAt    time.Time
	CreatedAt time.Time
	SentAt    time.Time
	Status    Status
	Attempts  int32
}
