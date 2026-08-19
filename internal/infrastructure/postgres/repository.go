package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Storage struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func New(db *pgxpool.Pool, log *zap.Logger) *Storage {
	//pgxpool.New(context.Background(), "host=127.0.0.1 port=5432 user=user db=users password=password_hash")
	return &Storage{
		db:  db,
		log: log,
	}
}
