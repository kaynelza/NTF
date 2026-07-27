package helper

import (
	"errors"
	"math"
	"math/rand"
	"net/textproto"
	"time"
)

type Status int

const (
	NoMoreAttempts           = 6
	StatusUnspecified Status = iota
	StatusPending
	StatusSending
	StatusSent
	StatusCancelled
	StatusDead
)

var ErrAsIsNotOK = errors.New("error as is not ok")

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
	Attempts  int
}

func CanCancel(notification Notification) bool {
	if notification.Status == StatusPending && notification.Attempts != NoMoreAttempts {
		return true
	}
	return false
}

func NextAttempt(notification *Notification) {
	if notification.Attempts == NoMoreAttempts {
		notification.Status = StatusDead
		return
	}
	jitter := time.Duration(rand.Intn(10)) * time.Second
	timer := time.Second*time.Duration(int(math.Pow(2, float64(notification.Attempts)))*30) + jitter
	notification.SendAt = time.Now().Add(timer)
}

func IsRequestDead(err error) (bool, error) {
	var protoError *textproto.Error
	if errors.As(err, &protoError) {
		code := protoError.Code
		if code < 500 {
			return false, nil
		}
		return true, nil
	}
	return false, ErrAsIsNotOK
}
