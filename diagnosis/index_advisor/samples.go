package index_advisor

import (
	"context"
	"fmt"
	"strings"
)

const DefaultSlowQuerySQLColumn = "in_memory_original_query"

// NormalizeSlowQuerySQLColumn returns a supported slow-query SQL column.
// Unknown and empty values intentionally fall back to the historical default.
func NormalizeSlowQuerySQLColumn(column string) string {
	switch strings.ToLower(strings.TrimSpace(column)) {
	case "query":
		return "query"
	case DefaultSlowQuerySQLColumn:
		return DefaultSlowQuerySQLColumn
	default:
		return DefaultSlowQuerySQLColumn
	}
}

func IsSupportedSlowQuerySQLColumn(column string) bool {
	column = strings.ToLower(strings.TrimSpace(column))
	return column == "" || column == "query" || column == DefaultSlowQuerySQLColumn
}

func markSampleSource(samples []recommendIndexSample, source string) {
	for i := range samples {
		samples[i].Source = source
	}
}

// Fetch individual summary rows and aggregate in Go. Aggregation in the
// statement summary query can return incomplete results on affected versions:
// https://github.com/pingcap/tidb/issues/71913.
const summaryColumns = `SELECT COALESCE(schema_name, ''), digest, query_sample_text,
       COALESCE(exec_count, 0), COALESCE(sum_latency, 0),
       COALESCE(avg_mem, 0), COALESCE(avg_request_unit_read, 0),
       COALESCE(instance, ''), COALESCE(CAST(summary_begin_time AS CHAR), ''),
       COALESCE(CAST(last_seen AS CHAR), ''), COALESCE(table_names, '')
FROM `

const summaryFilter = `
WHERE LOWER(stmt_type) = 'select'
  AND query_sample_text IS NOT NULL AND query_sample_text <> ''
  AND digest IS NOT NULL AND digest <> ''
  AND sample_user IS NOT NULL AND sample_user <> ''`

const currentSummaryQuery = summaryColumns + `information_schema.cluster_statements_summary` + summaryFilter
const historySummaryQuery = summaryColumns + `information_schema.cluster_statements_summary_history` + summaryFilter + `
  AND summary_end_time > DATE_SUB(NOW(), INTERVAL 24 HOUR)`

// Slow query rows contain an unredacted statement in the configured SQL
// column. Unlike statement summaries, they do not expose TABLE_NAMES, so the
// index advisor itself is responsible for rejecting any statement it cannot
// explain or recommend an index for.
func slowQueryQuery(sqlColumn string) string {
	column := NormalizeSlowQuerySQLColumn(sqlColumn)
	return fmt.Sprintf(`SELECT COALESCE(db, ''), COALESCE(digest, ''),
       COALESCE(%s, ''), COALESCE(query_time, 0),
       COALESCE(CAST(time AS CHAR), '')
FROM information_schema.cluster_slow_query
WHERE is_internal = FALSE
  AND %s IS NOT NULL
  AND %s <> ''`, column, column, column)
}

func (r *SQLRecommendIndexRepository) readSamples(ctx context.Context, query string) ([]recommendIndexSample, error) {
	queryCtx, cancel := context.WithTimeout(ctx, recommendIndexQueryTimeout)
	defer cancel()
	rows, err := r.db.QueryContext(queryCtx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []recommendIndexSample
	for rows.Next() {
		var sample recommendIndexSample
		var tableNames string
		if err := rows.Scan(&sample.Schema, &sample.Digest, &sample.SQL, &sample.ExecCount,
			&sample.ExecTime, &sample.AvgMem, &sample.AvgRequestUnitRead, &sample.Instance,
			&sample.WindowStart, &sample.LastSeen, &tableNames); err != nil {
			return nil, fmt.Errorf("scan SELECT statement: %w", err)
		}
		if strings.TrimSpace(sample.SQL) == "" || !hasOnlyUserTables(tableNames) {
			continue
		}
		// QUERY_SAMPLE_TEXT may be truncated; TiDB does not expose a reliable
		// truncation marker here. Syntax errors are handled by the advisor.
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read SELECT statements: %w", err)
	}
	return samples, nil
}

func (r *SQLRecommendIndexRepository) readSlowQuerySamples(ctx context.Context) ([]recommendIndexSample, error) {
	queryCtx, cancel := context.WithTimeout(ctx, recommendIndexQueryTimeout)
	defer cancel()
	rows, err := r.db.QueryContext(queryCtx, slowQueryQuery(r.slowQuerySQLColumn))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []recommendIndexSample
	for rows.Next() {
		var sample recommendIndexSample
		var queryTime float64
		if err := rows.Scan(&sample.Schema, &sample.Digest, &sample.SQL, &queryTime, &sample.LastSeen); err != nil {
			return nil, fmt.Errorf("scan slow query statement: %w", err)
		}
		// QUERY_TIME is reported in seconds while statement-summary
		// SUM_LATENCY is represented in nanoseconds by the MySQL driver.
		if queryTime > 0 {
			sample.ExecTime = uint64(queryTime * 1e9)
		}
		if strings.TrimSpace(sample.SQL) == "" {
			continue
		}
		// Each row in CLUSTER_SLOW_QUERY is one slow-log execution.
		sample.ExecCount = 1
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read slow query statements: %w", err)
	}
	return samples, nil
}

// TABLE_NAMES lists fully qualified tables, including tables in nested queries.
// A query without a table cannot benefit from an index recommendation.
func hasOnlyUserTables(tableNames string) bool {
	if strings.TrimSpace(tableNames) == "" {
		return false
	}
	for _, table := range strings.Split(tableNames, ",") {
		schema, name, ok := strings.Cut(strings.TrimSpace(table), ".")
		if !ok || strings.TrimSpace(schema) == "" || strings.TrimSpace(name) == "" {
			return false
		}
		switch strings.ToLower(strings.TrimSpace(schema)) {
		case "information_schema", "performance_schema", "metrics_schema", "inspection_schema", "mysql", "sys":
			return false
		}
	}
	return true
}

type summaryWindow struct {
	instance string
	start    string
}

// The current window can move into history between the two queries. If it
// does, use the completed historical window rather than the earlier snapshot.
func mergeSummaryWindows(current, history []recommendIndexSample) []recommendIndexSample {
	historicalWindows := make(map[summaryWindow]struct{}, len(history))
	for _, sample := range history {
		if sample.Instance != "" && sample.WindowStart != "" {
			historicalWindows[summaryWindow{sample.Instance, sample.WindowStart}] = struct{}{}
		}
	}
	merged := make([]recommendIndexSample, 0, len(current)+len(history))
	merged = append(merged, history...)
	for _, sample := range current {
		if sample.Instance != "" && sample.WindowStart != "" {
			if _, exists := historicalWindows[summaryWindow{sample.Instance, sample.WindowStart}]; exists {
				continue
			}
		}
		merged = append(merged, sample)
	}
	return merged
}
