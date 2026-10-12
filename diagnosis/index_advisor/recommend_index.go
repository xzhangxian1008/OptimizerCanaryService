package index_advisor

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/compare"
	"go.uber.org/zap"
)

type Sample struct {
	Schema string
	SQL    string
}

const recommendIndexQueryTimeout = 15 * time.Second
const recommendIndexMaxResults = 30

// recommendIndexSampleQuery selects the statements with the largest recorded
// execution latency. Candidates are evaluated in this order until enough
// statements with useful recommendations have been collected.
const recommendIndexSampleQuery = `SELECT COALESCE(schema_name, ''), COALESCE(digest, ''), query_sample_text,
       COALESCE(plan, ''), SUM(COALESCE(exec_count, 0)), SUM(COALESCE(sum_latency, 0))
FROM information_schema.cluster_statements_summary
WHERE LOWER(stmt_type) = 'select'
  AND query_sample_text IS NOT NULL AND query_sample_text <> ''
  AND digest IS NOT NULL AND digest <> ''
GROUP BY schema_name, digest, query_sample_text, plan
ORDER BY SUM(COALESCE(sum_latency, 0)) DESC
`

// RecommendIndexResult contains the statement and the non-empty values
// returned by TiDB's RECOMMEND INDEX command for that statement.
type RecommendIndexResult struct {
	Schema          string                `json:"schema,omitempty"`
	Digest          string                `json:"digest"`
	SQL             string                `json:"sql"`
	ExecTime        uint64                `json:"exec_time"`
	ExecCount       uint64                `json:"exec_count"`
	Recommendations []IndexRecommendation `json:"recommendations"`
	OldPlan         string                `json:"old_plan"`
	NewSQL          string                `json:"new_sql"`
	NewPlan         string                `json:"new_plan"`
}

type IndexRecommendation struct {
	Table   string `json:"table"`
	Name    string `json:"index_name"`
	Columns string `json:"index_columns"`
	output  string
}

type RecommendIndexRepository interface {
	RecommendIndex(context.Context) ([]RecommendIndexResult, error)
}

type SQLRecommendIndexRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

type recommendIndexSample struct {
	Sample
	Digest    string
	Plan      string
	ExecTime  uint64
	ExecCount uint64
}

func NewSQLRecommendIndexRepository(db *sql.DB, logger *zap.Logger) *SQLRecommendIndexRepository {
	return &SQLRecommendIndexRepository{db: db, logger: logger}
}

func (r *SQLRecommendIndexRepository) RecommendIndex(ctx context.Context) ([]RecommendIndexResult, error) {
	queryCtx, cancel := context.WithTimeout(ctx, recommendIndexQueryTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(queryCtx, recommendIndexSampleQuery)
	if err != nil {
		return nil, fmt.Errorf("query top SELECT statements: %w", err)
	}
	defer rows.Close()

	samples := make([]recommendIndexSample, 0)
	for rows.Next() {
		var sample recommendIndexSample
		if err := rows.Scan(&sample.Schema, &sample.Digest, &sample.SQL, &sample.Plan, &sample.ExecCount, &sample.ExecTime); err != nil {
			return nil, fmt.Errorf("scan SELECT statement: %w", err)
		}
		// TODO: detect truncated query_sample_text and recover or otherwise
		// handle the complete SQL before sending it to the index advisor.
		if strings.TrimSpace(sample.SQL) != "" {
			samples = append(samples, sample)
			r.logger.Debug("collected SQL from cluster_statements_summary",
				zap.String("schema", sample.Schema),
				zap.String("digest", sample.Digest),
				zap.String("sql", sample.SQL),
				zap.Uint64("exec_count", sample.ExecCount),
				zap.Uint64("exec_time", sample.ExecTime))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read SELECT statements: %w", err)
	}

	results := make([]RecommendIndexResult, 0, recommendIndexMaxResults)
	for _, sample := range samples {
		recommendations, err := r.recommendFor(ctx, sample.Sample)
		if err != nil {
			if isSQLSyntaxError(err) {
				r.logger.Warn("skip SQL with syntax error during index recommendation",
					zap.String("digest", sample.Digest),
					zap.String("query_sample_text", sample.SQL),
					zap.Error(err))
				continue
			}
			if isEmptyQuerySetError(err) {
				r.logger.Debug("skip SQL rejected by index advisor",
					zap.String("schema", sample.Schema),
					zap.String("digest", sample.Digest),
					zap.Error(err))
				continue
			}
			return nil, err
		}
		if len(recommendations) == 0 {
			continue
		}
		hintedSQL, err := buildHypoIndexSQL(sample.SQL, recommendations)
		if err != nil {
			return nil, fmt.Errorf("build hypothetical index SQL: %w", err)
		}
		newPlan, err := r.explain(ctx, sample.Sample, hintedSQL)
		if err != nil {
			if isSQLSyntaxError(err) {
				r.logger.Warn("skip SQL with syntax error during hypothetical index EXPLAIN",
					zap.String("digest", sample.Digest),
					zap.String("query_sample_text", sample.SQL),
					zap.Error(err))
				continue
			}
			return nil, err
		}
		// TODO: compare sample.Plan and newPlan and skip this result when the
		// old and hypothetical-index plans are equal.
		results = append(results, RecommendIndexResult{
			Schema:          sample.Schema,
			Digest:          sample.Digest,
			SQL:             sample.SQL,
			ExecTime:        sample.ExecTime,
			ExecCount:       sample.ExecCount,
			Recommendations: recommendations,
			OldPlan:         sample.Plan,
			NewSQL:          hintedSQL,
			NewPlan:         newPlan,
		})
		if len(results) >= recommendIndexMaxResults {
			break
		}
	}
	return results, nil
}

func isEmptyQuerySetError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "empty query set after filtering invalid queries")
}

func isSQLSyntaxError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "you have an error in your sql syntax") ||
		strings.Contains(message, "syntax error")
}

func (r *SQLRecommendIndexRepository) recommendFor(ctx context.Context, sample Sample) ([]IndexRecommendation, error) {
	recommendCtx, cancel := context.WithTimeout(ctx, recommendIndexQueryTimeout)
	defer cancel()

	conn, err := r.db.Conn(recommendCtx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection for RECOMMEND INDEX: %w", err)
	}
	defer conn.Close()

	if sample.Schema != "" {
		quotedSchema := "`" + strings.ReplaceAll(sample.Schema, "`", "``") + "`"
		if _, err := conn.ExecContext(recommendCtx, "USE "+quotedSchema); err != nil {
			return nil, fmt.Errorf("select schema %q for RECOMMEND INDEX: %w", sample.Schema, err)
		}
	}

	statement := recommendableStatement(sample.SQL)
	recommendSQL := "RECOMMEND INDEX RUN FOR '" + strings.ReplaceAll(statement, "'", "''") + "'"
	r.logger.Info("execute SQL", zap.String("schema", sample.Schema), zap.String("sql", recommendSQL))
	rows, err := conn.QueryContext(recommendCtx, recommendSQL)
	if err != nil {
		return nil, fmt.Errorf("recommend index run for statement: %w", err)
	}
	defer rows.Close()

	columnCount, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read RECOMMEND INDEX columns: %w", err)
	}
	values := make([]IndexRecommendation, 0)
	for rows.Next() {
		cells := make([]any, len(columnCount))
		pointers := make([]any, len(cells))
		for i := range cells {
			pointers[i] = &cells[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("scan RECOMMEND INDEX result: %w", err)
		}
		row := make(map[string]string, len(columnCount))
		for i, name := range columnCount {
			row[strings.ToLower(name)] = recommendCellString(cells[i])
		}
		fields := make([]string, 0, len(columnCount))
		for i, name := range columnCount {
			fields = append(fields, fmt.Sprintf("%s: %s", name, recommendCellString(cells[i])))
		}
		advice := IndexRecommendation{Table: row["table"], Name: row["index_name"], Columns: row["index_columns"], output: strings.Join(fields, "\n")}
		if advice.Table != "" || advice.Name != "" || advice.Columns != "" {
			values = append(values, advice)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read RECOMMEND INDEX result: %w", err)
	}
	return values, nil
}

func recommendCellString(cell any) string {
	if cell == nil {
		return ""
	}
	if value, ok := cell.([]byte); ok {
		return strings.TrimSpace(string(value))
	}
	return strings.TrimSpace(fmt.Sprint(cell))
}

func (r *SQLRecommendIndexRepository) explain(ctx context.Context, sample Sample, statement string) (string, error) {
	explainCtx, cancel := context.WithTimeout(ctx, recommendIndexQueryTimeout)
	defer cancel()
	conn, err := r.db.Conn(explainCtx)
	if err != nil {
		return "", fmt.Errorf("acquire connection for EXPLAIN: %w", err)
	}
	defer conn.Close()
	if sample.Schema != "" {
		quoted := "`" + strings.ReplaceAll(sample.Schema, "`", "``") + "`"
		if _, err := conn.ExecContext(explainCtx, "USE "+quoted); err != nil {
			return "", fmt.Errorf("select schema %q for EXPLAIN: %w", sample.Schema, err)
		}
	}
	explainSQL := "EXPLAIN " + strings.TrimSuffix(strings.TrimSpace(statement), ";")
	rows, err := conn.QueryContext(explainCtx, explainSQL)
	if err != nil {
		return "", fmt.Errorf("explain statement: %w", err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("read EXPLAIN columns: %w", err)
	}
	var out strings.Builder
	out.WriteString(strings.Join(columns, "\t"))
	out.WriteByte('\n')
	for rows.Next() {
		cells := make([]any, len(columns))
		pointers := make([]any, len(cells))
		for i := range cells {
			pointers[i] = &cells[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return "", fmt.Errorf("scan EXPLAIN result: %w", err)
		}
		for i, cell := range cells {
			if i > 0 {
				out.WriteByte('\t')
			}
			out.WriteString(recommendCellString(cell))
		}
		out.WriteByte('\n')
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("read EXPLAIN result: %w", err)
	}
	return strings.TrimRight(out.String(), "\n"), nil
}

func buildHypoIndexSQL(statement string, recommendations []IndexRecommendation) (string, error) {
	hints := make([]string, 0, len(recommendations))
	for _, advice := range recommendations {
		if advice.Table == "" || advice.Name == "" || advice.Columns == "" {
			continue
		}
		hints = append(hints, fmt.Sprintf("hypo_index(%s, %s, %s)", advice.Table, advice.Name, advice.Columns))
	}
	if len(hints) == 0 {
		return recommendableStatement(statement), nil
	}
	return compare.AddOptimizerHint(recommendableStatement(statement), strings.Join(hints, " "))
}

// query_sample_text can contain TiDB's server-side prepared statement suffix,
// for example "SELECT ... [arguments: 1]". That suffix is metadata rather
// than SQL and must be removed before the statement is sent to TiDB again.
func recommendableStatement(statement string) string {
	statement = strings.TrimSpace(statement)
	if markerIndex := strings.LastIndex(statement, "[arguments:"); markerIndex > 0 && strings.TrimSpace(statement[:markerIndex]) != "" {
		if strings.HasSuffix(strings.TrimSpace(statement), "]") {
			statement = strings.TrimSpace(statement[:markerIndex])
		}
	}
	return strings.TrimSuffix(statement, ";")
}

func RenderRecommendIndex(results []RecommendIndexResult) string {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].ExecTime != results[j].ExecTime {
			return results[i].ExecTime > results[j].ExecTime
		}
		return results[i].Digest < results[j].Digest
	})
	var report strings.Builder
	report.WriteString("| SQL | Total Exec Time | Total Exec Count | Index Advice | Old Plan | New Plan |\n")
	report.WriteString("| --- | ---: | ---: | --- | --- | --- |\n")
	for _, result := range results {
		digest := firstEightRecommend(result.Digest)
		anchor := "sql-" + strings.ToLower(result.Digest)
		fmt.Fprintf(&report, "| [%s](#%s) | %s | %d | [detail](#index-recommendation-%s) | [detail](#old-plan-%s) | [detail](#new-plan-%s) |\n",
			markdownRecommendCell(digest), anchor, formatRecommendExecTime(result.ExecTime), result.ExecCount,
			strings.ToLower(result.Digest), strings.ToLower(result.Digest), strings.ToLower(result.Digest))
	}
	for _, result := range results {
		digest := strings.ToLower(result.Digest)
		fmt.Fprintf(&report, "\n<a id=\"sql-%s\"></a>\n\n## SQL: %s\n\n", digest, html.EscapeString(result.Digest))
		report.WriteString("### SQL Statement\n\n")
		writeRecommendCodeBlock(&report, result.SQL)
		fmt.Fprintf(&report, "\n<a id=\"index-recommendation-%s\"></a>\n\n### Index Recommendation\n\n", digest)
		adviceLines := make([]string, 0, len(result.Recommendations))
		for i, advice := range result.Recommendations {
			adviceLines = append(adviceLines, "#### Recommendation "+fmt.Sprint(i+1))
			adviceLines = append(adviceLines, advice.output)
		}
		writeRecommendCodeBlock(&report, strings.Join(adviceLines, "\n"))
		fmt.Fprintf(&report, "\n<a id=\"old-plan-%s\"></a>\n\n### Old Plan\n\n", digest)
		writeRecommendCodeBlock(&report, result.OldPlan)
		fmt.Fprintf(&report, "\n<a id=\"new-plan-%s\"></a>\n\n### New Plan\n\n", digest)
		writeRecommendCodeBlock(&report, result.NewSQL+"\n\n"+result.NewPlan)
	}
	return report.String()
}

func firstEightRecommend(value string) string {
	if len(value) > 8 {
		return value[:8]
	}
	return value
}

func formatRecommendExecTime(ns uint64) string {
	return fmt.Sprintf("%d.%02ds", ns/1_000_000_000, (ns%1_000_000_000)/10_000_000)
}

func markdownRecommendCell(value string) string {
	return strings.NewReplacer("|", "&#124;", "[", "\\[", "]", "\\]").Replace(value)
}

func writeRecommendCodeBlock(out *strings.Builder, content string) {
	fence := "```"
	for strings.Contains(content, fence) {
		fence += "`"
	}
	fmt.Fprintf(out, "%stext\n%s\n%s\n", fence, content, fence)
}
