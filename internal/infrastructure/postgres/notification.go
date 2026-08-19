package postgres

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/jackc/pgx/v5"
	"github.com/kaynelza/NTF/internal/infrastructure/entity"
)

func (s *Storage) CreateNotification(ctx context.Context, notification entity.Notification) (id string, sendAt time.Time, err error) {
	q := `
		insert into ntf.notification(recipient,title,body,send_at,status) 
		values ($1,$2,$3,$4,$5)
		returning id,send_at`

	err = s.db.QueryRow(ctx, q, notification.Recipient, notification.Title, notification.Body, notification.SendAt, entity.StatusPending).Scan(&id, &sendAt)
	if err != nil {
		return "", time.Time{}, errors.Wrap(entity.ErrInternal, "failed to form new notification")
	}

	return id, sendAt, nil
}

func (s *Storage) GetNotificationByID(ctx context.Context, id string) (notification entity.Notification, err error) {
	q := `
		select id,recipient,title,body,send_at,status,attempts,last_error,created_at,sent_at
		from ntf.notification
		where id=$1`

	err = s.db.QueryRow(ctx, q, id).Scan(&notification)
	if err != nil {
		return entity.Notification{}, errors.Wrap(entity.ErrInternal, "failed to get notification by id")
	}

	return notification, nil
}

func (s *Storage) ListAllNotifications(ctx context.Context, recipient string, status entity.Status, limit, offset int) (list []entity.Notification, total int, err error) {
	q := `
		select id,recipient,title,body,send_at,status,attempts,last_error,created_at,sent_at
		from ntf.notification
		where recipient=$1 and status = $2
		limit $3
		offset $4`

	rows, err := s.db.Query(ctx, q, recipient, status, limit, offset)
	if err != nil {
		return nil, 0, errors.Wrap(entity.ErrInternal, "failed to list all notifications")
	}

	list, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Notification])
	if err != nil {
		return nil, 0, errors.Wrap(entity.ErrInternal, "failed to form list of all notifications")
	}

	return list, len(list), nil
}

func (s *Storage) CancelNotification(ctx context.Context, id string) error {
	q := `
		update ntf.notification
		set status = $1
		where id = $2`

	_, err := s.db.Exec(ctx, q, entity.StatusCancelled, id)
	if err != nil {
		return errors.Wrap(entity.ErrInternal, "failed to cancel notification")
	}

	return nil
}
