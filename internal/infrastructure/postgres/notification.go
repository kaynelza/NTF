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

func (s *Storage) GetNotificationByID(ctx context.Context, etx entity.Transaction, id string) (notification entity.Notification, err error) {
	tx, ok := etx.(pgx.Tx)
	if !ok {
		return entity.Notification{}, errors.New("failed to convert types")
	}

	q := `
		select id,recipient,title,body,send_at,status,attempts,last_error,created_at,sent_at
		from ntf.notification
		where id=$1`

	row, err := tx.Query(ctx, q, id)
	if err != nil {
		return entity.Notification{}, errors.Wrap(entity.ErrInternal, "failed to list notification row")
	}

	notification, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Notification])
	if err != nil {
		return entity.Notification{}, errors.Wrap(entity.ErrInternal, "failed to form notification ")
	}

	return notification, nil
}

func (s *Storage) ListAllNotifications(ctx context.Context, etx entity.Transaction, recipient string, status entity.Status, limit, offset int) (list []entity.Notification, total int, err error) {
	tx, ok := etx.(pgx.Tx)
	if !ok {
		return nil, 0, errors.New("failed to convert types")
	}

	q := `
		select id,recipient,title,body,send_at,status,attempts,last_error,created_at,sent_at
		from ntf.notification
		where recipient=$1 and status = $2
		limit $3
		offset $4`

	rows, err := tx.Query(ctx, q, recipient, status, limit, offset)
	if err != nil {
		return nil, 0, errors.Wrap(entity.ErrInternal, "failed to list all notifications")
	}

	list, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Notification])
	if err != nil {
		return nil, 0, errors.Wrap(entity.ErrInternal, "failed to form list of all notifications")
	}

	return list, len(list), nil
}

func (s *Storage) CancelNotification(ctx context.Context, etx entity.Transaction, id string) error {
	tx, ok := etx.(pgx.Tx)
	if !ok {
		return errors.New("failed to convert types")
	}

	q := `
		update ntf.notification
		set status = $1
		where id = $2`

	_, err := tx.Exec(ctx, q, entity.StatusCancelled, id)
	if err != nil {
		return errors.Wrap(entity.ErrInternal, "failed to cancel notification")
	}

	return nil
}

func (s *Storage) LockNotificationForUpdate(ctx context.Context, etx entity.Transaction, id string) (notification entity.Notification, err error) {
	tx, ok := etx.(pgx.Tx)
	if !ok {
		return entity.Notification{}, errors.New("failed to convert types")
	}

	q := `
		select id, recipient, title, body, send_at, status, attempts, last_error, created_at, sent_at
		from ntf.notification 
		where id = $1
		for update skip locked`

	row, err := tx.Query(ctx, q, id)
	if err != nil {
		return entity.Notification{}, errors.Wrap(entity.ErrInternal, "failed to lock notification")
	}
	notification, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Notification])

	return notification, nil
}
