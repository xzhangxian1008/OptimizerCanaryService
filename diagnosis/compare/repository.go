package compare

import (
	"database/sql"

	"go.uber.org/zap"
)

type SQLRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSQLRepository(db *sql.DB, logger *zap.Logger) *SQLRepository {
	return &SQLRepository{db: db, logger: logger}
}
