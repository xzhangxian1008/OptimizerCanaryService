package prepare

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

const (
	DefaultSchema      = "optimizer_canary_prepare"
	DefaultConcurrency = 4
	DefaultDuration    = time.Minute
	connectionTimeout  = 5 * time.Second
)

type Config struct {
	DSN         string
	Schema      string
	Concurrency int
	Duration    time.Duration
}

func (c Config) validate() error {
	if strings.TrimSpace(c.DSN) == "" {
		return errors.New("dsn must not be empty")
	}
	if strings.TrimSpace(c.Schema) == "" {
		return errors.New("prepare schema must not be empty")
	}
	if c.Concurrency <= 0 {
		return errors.New("prepare concurrency must be greater than zero")
	}
	if c.Duration <= 0 {
		return errors.New("prepare duration must be greater than zero")
	}
	return nil
}

// Run connects to the requested TiDB node and prepares a compact TPC-C workload
// whose recorded plans intentionally differ from the optimizer's natural plans.
func Run(ctx context.Context, config Config, logger *zap.Logger) error {
	if err := config.validate(); err != nil {
		return err
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	db, err := sql.Open("mysql", strings.TrimSpace(config.DSN))
	if err != nil {
		return fmt.Errorf("open TiDB connection: %w", err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("connect to TiDB: %w", err)
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire TiDB connection: %w", err)
	}
	defer conn.Close()

	logger.Info("prepare mode connected to TiDB", zap.String("schema", config.Schema))
	preparer := &preparer{
		db:          db,
		conn:        conn,
		schema:      strings.TrimSpace(config.Schema),
		logger:      logger,
		concurrency: config.Concurrency,
		duration:    config.Duration,
	}
	return preparer.prepare(ctx)
}

type preparer struct {
	db          *sql.DB
	conn        *sql.Conn
	schema      string
	logger      *zap.Logger
	concurrency int
	duration    time.Duration
}

func (p *preparer) prepare(ctx context.Context) error {
	if err := p.prepareTableSchema(ctx); err != nil {
		return fmt.Errorf("prepare table schema: %w", err)
	}
	if err := p.prepareStatements(ctx); err != nil {
		return fmt.Errorf("prepare statements: %w", err)
	}
	p.logger.Info("prepare mode completed", zap.String("schema", p.schema))
	return nil
}

func (p *preparer) prepareTableSchema(ctx context.Context) error {
	return p.prepareTPCCSchema(ctx)
}

func (p *preparer) prepareStatements(ctx context.Context) error {
	return p.prepareTPCCStatements(ctx)
}
