package postgres

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kaynelza/NTF/internal/infrastructure/entity"
	"go.uber.org/zap"
)

type Storage struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func New(db *pgxpool.Pool, log *zap.Logger) *Storage {
	return &Storage{
		db:  db,
		log: log,
	}
}

func (s *Storage) BeginTx(ctx context.Context) (entity.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	return tx, nil
}
