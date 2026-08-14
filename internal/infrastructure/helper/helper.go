package helper

import (
	"fmt"
	"math"
	"math/rand"
	"net/textproto"
	"time"

	"github.com/go-faster/errors"
	v1 "github.com/kaynelza/NTF/pkg/grpc/notifyd/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Status int

const (
	MaxAttempts              = 6
	StatusUnspecified Status = iota
	StatusPending
	StatusSending
	StatusSent
	StatusCancelled
	StatusDead
)

var ErrPermanent = errors.New("error as is not ok")

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

func NotificationFromPB(n *v1.Notification) Notification {
	status := StatusFromPB(n.Status)
	return Notification{
		Id:        n.GetId(),
		Recipient: n.GetRecipient(),
		Title:     n.GetTitle(),
		Body:      n.GetBody(),
		SendAt:    n.GetSendAt().AsTime(),
		Status:    status,
		Attempts:  int(n.GetAttempts()),
		LastError: n.LastError,
		CreatedAt: n.CreatedAt.AsTime(),
		SentAt:    n.SentAt.AsTime(),
	}
}

func NotificationToPB(n Notification) *v1.Notification {
	return &v1.Notification{
		Id:        n.Id,
		Recipient: n.Recipient,
		Title:     n.Title,
		Body:      n.Body,
		SendAt:    timestamppb.New(n.SendAt),
		Status:    StatusToPB(n.Status),
		Attempts:  int32(n.Attempts),
		LastError: n.LastError,
		CreatedAt: timestamppb.New(n.CreatedAt),
		SentAt:    timestamppb.New(n.SentAt),
	}
}

func StatusFromPB(status v1.Status) Status {
	switch status {
	case v1.Status_STATUS_PENDING:
		return StatusPending

	case v1.Status_STATUS_SENDING:
		return StatusSending

	case v1.Status_STATUS_SENT:
		return StatusSent

	case v1.Status_STATUS_CANCELLED:
		return StatusCancelled

	case v1.Status_STATUS_DEAD:
		return StatusDead

	default:
		return StatusUnspecified
	}
}

func StatusToPB(status Status) v1.Status {
	switch status {
	case StatusPending:
		return v1.Status_STATUS_PENDING

	case StatusSending:
		return v1.Status_STATUS_SENDING

	case StatusSent:
		return v1.Status_STATUS_SENT

	case StatusCancelled:
		return v1.Status_STATUS_CANCELLED

	case StatusDead:
		return v1.Status_STATUS_DEAD

	default:
		return v1.Status_STATUS_UNSPECIFIED
	}
}

func CanCancel(notification Notification) bool {
	if notification.Status == StatusPending && notification.Attempts != MaxAttempts {
		return true
	}
	return false
}

func NextAttempt(notification *Notification) {
	if notification.Attempts == MaxAttempts {
		notification.Status = StatusDead
		return
	}
	jitter := time.Duration(rand.Intn(10)) * time.Second
	timer := time.Second*time.Duration(int(math.Pow(2, float64(notification.Attempts)))*30) + jitter
	notification.SendAt = time.Now().Add(timer)
}

func IsErrorRetryable(err error) (bool, error) {
	var protoError *textproto.Error
	if errors.As(err, &protoError) {
		code := protoError.Code
		if code < 500 {
			return false, nil
		}
		return true, nil
	}
	return false, fmt.Errorf("%w:%w", err, ErrPermanent)
}
