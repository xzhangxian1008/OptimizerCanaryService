package diagnosis

import (
	"context"
	"errors"
	"strings"
	"sync"

	"database/sql"
	"go.uber.org/zap"
)

var ErrNoTiDBConnection = errors.New("TiDB connection has not been configured")

// ConnectionStatus describes the result of a connection attempt. It is also
// returned by the test-only HTTP endpoint.
type ConnectionStatus struct {
	DSN       string
	Connected bool
	Err       error
}

// ConnectionManager owns the currently active TiDB connection and implements
// Repository by forwarding each operation to the current database. A new
// connection can be installed while the HTTP server is running.
type ConnectionManager struct {
	mu     sync.RWMutex
	db     *sql.DB
	dsn    string
	logger *zap.Logger
}

func NewConnectionManager(logger *zap.Logger) *ConnectionManager {
	return &ConnectionManager{logger: logger}
}

// Replace verifies a new DSN and installs it after a successful ping. The
// previous connection is closed after the replacement. If the new connection
// fails, the existing connection is retained.
func (m *ConnectionManager) Replace(ctx context.Context, dsn string) ConnectionStatus {
	dsn = strings.TrimSpace(dsn)
	status := ConnectionStatus{DSN: dsn}
	if dsn == "" {
		status.Err = errors.New("dsn must not be empty")
		return status
	}

	db, err := OpenDB(dsn)
	if err != nil {
		status.Err = err
		return status
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		status.Err = err
		return status
	}

	m.mu.Lock()
	oldDB := m.db
	m.db = db
	m.dsn = dsn
	m.mu.Unlock()
	if oldDB != nil {
		_ = oldDB.Close()
	}
	status.Connected = true
	m.logger.Info("TiDB connection replaced", zap.String("dsn", dsn))
	return status
}

func (m *ConnectionManager) Close() {
	m.mu.Lock()
	db := m.db
	m.db = nil
	m.dsn = ""
	m.mu.Unlock()
	if db != nil {
		_ = db.Close()
	}
}

func (m *ConnectionManager) current() (*sql.DB, error) {
	m.mu.RLock()
	db := m.db
	m.mu.RUnlock()
	if db == nil {
		return nil, ErrNoTiDBConnection
	}
	return db, nil
}

func (m *ConnectionManager) Sample(ctx context.Context, source Source, limit int) ([]Sample, error) {
	db, err := m.current()
	if err != nil {
		return nil, err
	}
	return NewSQLRepository(db, m.logger).Sample(ctx, source, limit)
}

func (m *ConnectionManager) Explain(ctx context.Context, sample Sample) error {
	db, err := m.current()
	if err != nil {
		return err
	}
	return NewSQLRepository(db, m.logger).Explain(ctx, sample)
}
