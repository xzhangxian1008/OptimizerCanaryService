package diagnosis

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryTimeout   = 15 * time.Second
	explainTimeout = 15 * time.Second

	slowQueryQuery = `SELECT COALESCE(db, ''), query
FROM information_schema.slow_query
WHERE is_internal = FALSE AND query IS NOT NULL AND query <> ''
ORDER BY RAND()
LIMIT ?`

	topSQLQuery = `SELECT schema_name, query_sample_text
FROM (
    SELECT COALESCE(schema_name, '') AS schema_name, query_sample_text
    FROM information_schema.statements_summary
    WHERE query_sample_text IS NOT NULL AND query_sample_text <> ''
    ORDER BY sum_latency DESC
    LIMIT 100
) AS top_statements
ORDER BY RAND()
LIMIT ?`

	statementSummaryQuery = `SELECT COALESCE(schema_name, ''), query_sample_text
FROM information_schema.statements_summary
WHERE query_sample_text IS NOT NULL AND query_sample_text <> ''
ORDER BY RAND()
LIMIT ?`
)

var sourceQueries = map[Source]string{
	SourceSlowQuery:        slowQueryQuery,
	SourceTopSQL:           topSQLQuery,
	SourceStatementSummary: statementSummaryQuery,
}

type Repository interface {
	Sample(ctx context.Context, source Source, limit int) ([]Sample, error)
	Explain(ctx context.Context, sample Sample) error
}

type SQLRepository struct {
	db *sql.DB
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return db, nil
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Sample(ctx context.Context, source Source, limit int) ([]Sample, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	query, ok := sourceQueries[source]
	if !ok {
		return nil, fmt.Errorf("unsupported source %q", source)
	}
	rows, err := r.db.QueryContext(queryCtx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", source, err)
	}
	defer rows.Close()

	samples := make([]Sample, 0, limit)
	for rows.Next() {
		var sample Sample
		if err := rows.Scan(&sample.Schema, &sample.SQL); err != nil {
			return nil, fmt.Errorf("scan %s sample: %w", source, err)
		}
		if strings.TrimSpace(sample.SQL) != "" {
			samples = append(samples, sample)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read %s samples: %w", source, err)
	}
	return samples, nil
}

func (r *SQLRepository) Explain(ctx context.Context, sample Sample) error {
	explainCtx, cancel := context.WithTimeout(ctx, explainTimeout)
	defer cancel()

	conn, err := r.db.Conn(explainCtx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Close()

	if sample.Schema != "" {
		quotedSchema := "`" + strings.ReplaceAll(sample.Schema, "`", "``") + "`"
		if _, err := conn.ExecContext(explainCtx, "USE "+quotedSchema); err != nil {
			return fmt.Errorf("select schema %q: %w", sample.Schema, err)
		}
	}

	statement := strings.TrimSpace(sample.SQL)
	statement = strings.TrimSuffix(statement, ";")
	rows, err := conn.QueryContext(explainCtx, "EXPLAIN "+statement)
	if err != nil {
		return fmt.Errorf("explain statement: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		// Draining the result ensures TiDB completes the EXPLAIN request.
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read EXPLAIN result: %w", err)
	}
	return nil
}
