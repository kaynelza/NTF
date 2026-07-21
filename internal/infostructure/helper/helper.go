package helper

import "time"

type Status string

const (
	STATUS_UNSPECIFIED Status = "UNSPECIFIED"
	STATUS_PENDING     Status = "PENDING"
	STATUS_SENDING     Status = "SENDING"
	STATUS_SENT        Status = "SENT"
	STATUS_CANCELLED   Status = "CANCELLED"
	STATUS_DEAD        Status = "DEAD"
)

type Notification struct {
	id         string
	recipient  string
	title      string
	body       string
	send_at    time.Time
	status     Status
	attempts   int32
	last_error string
	created_at time.Time
	sent_at    time.Time
}
