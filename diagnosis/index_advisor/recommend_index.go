package index_advisor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/compare"
	"go.uber.org/zap"
)

type Sample struct {
	Schema string
	SQL    string
}

const recommendIndexQueryTimeout = 15 * time.Second
const recommendIndexMaxResults = 30

// TODO: switch back to recommendIndexAggregatedQuery after the TiDB bug is
// fixed: https://github.com/pingcap/tidb/issues/71913.
// The aggregation and ordering in that query currently produce incomplete
// results on affected TiDB versions. Keep both queries aligned: the active query
// omits ORDER BY and aggregate functions. When restoring this query, remove
// the local aggregation and sorting workaround.
/*
const recommendIndexAggregatedQuery = `SELECT COALESCE(schema_name, ''), COALESCE(digest, ''), query_sample_text,
       COALESCE(plan, ''), SUM(COALESCE(exec_count, 0)), SUM(COALESCE(sum_latency, 0)),
       SUM(COALESCE(avg_mem, 0) * exec_count) / NULLIF(SUM(exec_count), 0),
       SUM(COALESCE(avg_request_unit_read, 0) * exec_count) / NULLIF(SUM(exec_count), 0)
FROM information_schema.cluster_statements_summary
WHERE LOWER(stmt_type) = 'select'
  AND query_sample_text IS NOT NULL AND query_sample_text <> ''
  AND digest IS NOT NULL AND digest <> ''
GROUP BY schema_name, digest, query_sample_text, plan
ORDER BY SUM(COALESCE(sum_latency, 0)) DESC
`
*/

// recommendIndexSampleQuery deliberately avoids aggregate functions and
// ordering. Rows are aggregated and sorted in Go because of TiDB issue 71913.
const recommendIndexSampleQuery = `SELECT COALESCE(schema_name, ''), COALESCE(digest, ''), query_sample_text,
       COALESCE(plan, ''), COALESCE(exec_count, 0), COALESCE(sum_latency, 0),
       COALESCE(avg_mem, 0), COALESCE(avg_request_unit_read, 0)
FROM information_schema.cluster_statements_summary
WHERE LOWER(stmt_type) = 'select'
  AND query_sample_text IS NOT NULL AND query_sample_text <> ''
  AND digest IS NOT NULL AND digest <> ''
GROUP BY schema_name, digest, query_sample_text, plan, exec_count, sum_latency, avg_mem, avg_request_unit_read, instance`

// RecommendIndexResult contains the statement and the non-empty values
// returned by TiDB's RECOMMEND INDEX command for that statement.
type RecommendIndexResult struct {
	Schema             string                `json:"schema,omitempty"`
	Digest             string                `json:"digest"`
	SQL                string                `json:"sql"`
	ExecTime           uint64                `json:"exec_time"`
	AvgMem             float64               `json:"avg_mem"`
	AvgRequestUnitRead float64               `json:"avg_request_unit_read"`
	ExecCount          uint64                `json:"exec_count"`
	Recommendations    []IndexRecommendation `json:"recommendations"`
	OldPlan            string                `json:"old_plan"`
	NewSQL             string                `json:"new_sql"`
	NewPlan            string                `json:"new_plan"`
}

type IndexRecommendation struct {
	Table       string `json:"table"`
	Name        string `json:"index_name"`
	Columns     string `json:"index_columns"`
	output      string
	improvement *float64
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
	Digest             string
	Plan               string
	ExecTime           uint64
	ExecCount          uint64
	AvgMem             float64
	AvgRequestUnitRead float64
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

	var rawSamples []recommendIndexSample
	for rows.Next() {
		var sample recommendIndexSample
		if err := rows.Scan(&sample.Schema, &sample.Digest, &sample.SQL, &sample.Plan, &sample.ExecCount, &sample.ExecTime, &sample.AvgMem, &sample.AvgRequestUnitRead); err != nil {
			return nil, fmt.Errorf("scan SELECT statement: %w", err)
		}
		// TODO: detect truncated query_sample_text and recover or otherwise
		// handle the complete SQL before sending it to the index advisor.
		if strings.TrimSpace(sample.SQL) != "" {
			rawSamples = append(rawSamples, sample)
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
	samples := aggregateAndSortSamples(rawSamples)

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
		oldPlan, err := r.explain(ctx, sample.Sample, recommendableStatement(sample.SQL))
		if err != nil {
			if isSQLSyntaxError(err) {
				r.logger.Warn("skip SQL with syntax error during original EXPLAIN", zap.String("digest", sample.Digest), zap.String("query_sample_text", sample.SQL), zap.Error(err))
				continue
			}
			return nil, err
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
		// TODO: compare oldPlan and newPlan and skip this result when the
		// old and hypothetical-index plans are equal.
		results = append(results, RecommendIndexResult{
			Schema:             sample.Schema,
			Digest:             sample.Digest,
			SQL:                sample.SQL,
			ExecTime:           sample.ExecTime,
			ExecCount:          sample.ExecCount,
			Recommendations:    recommendations,
			OldPlan:            oldPlan,
			AvgMem:             sample.AvgMem,
			AvgRequestUnitRead: sample.AvgRequestUnitRead,
			NewSQL:             hintedSQL,
			NewPlan:            newPlan,
		})
		if len(results) >= recommendIndexMaxResults {
			break
		}
	}
	return results, nil
}

type recommendIndexSampleKey struct {
	schema string
	digest string
	sql    string
	plan   string
}

func aggregateAndSortSamples(rawSamples []recommendIndexSample) []recommendIndexSample {
	byKey := make(map[recommendIndexSampleKey]recommendIndexSample, len(rawSamples))
	for _, sample := range rawSamples {
		key := recommendIndexSampleKey{schema: sample.Schema, digest: sample.Digest, sql: sample.SQL, plan: sample.Plan}
		if existing, ok := byKey[key]; ok {
			totalCount := existing.ExecCount + sample.ExecCount
			if totalCount > 0 {
				existing.AvgMem = (existing.AvgMem*float64(existing.ExecCount) + sample.AvgMem*float64(sample.ExecCount)) / float64(totalCount)
				existing.AvgRequestUnitRead = (existing.AvgRequestUnitRead*float64(existing.ExecCount) + sample.AvgRequestUnitRead*float64(sample.ExecCount)) / float64(totalCount)
			}
			existing.ExecCount = totalCount
			existing.ExecTime += sample.ExecTime
			byKey[key] = existing
		} else {
			byKey[key] = sample
		}
	}
	samples := make([]recommendIndexSample, 0, len(byKey))
	for _, sample := range byKey {
		samples = append(samples, sample)
	}
	sort.SliceStable(samples, func(i, j int) bool {
		if samples[i].ExecTime != samples[j].ExecTime {
			return samples[i].ExecTime > samples[j].ExecTime
		}
		return samples[i].Digest < samples[j].Digest
	})
	return samples
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
		var impacted []struct {
			Improvement *float64 `json:"Improvement"`
		}
		if err := json.Unmarshal([]byte(row["top_impacted_query"]), &impacted); err == nil && len(impacted) > 0 {
			advice.improvement = impacted[0].Improvement
		}
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
	explainSQL := "EXPLAIN FORMAT='verbose' " + strings.TrimSuffix(strings.TrimSpace(statement), ";")
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
			out.WriteString(planCellString(cell))
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
	report.WriteString("# Overview\n\n")
	report.WriteString("| SQL | Total Exec Time | Exec Count | AVG_MEM | AVG_REQ_UNIT_READ | Est Improvement | Index Advice |\n")
	report.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, result := range results {
		digest := firstEightRecommend(result.Digest)
		anchor := "sql-" + strings.ToLower(result.Digest)
		fmt.Fprintf(&report, "| [%s](#%s) | %s | %d | %.2f | %.4f | %s | [detail](#index-recommendation-%s) |\n",
			markdownRecommendCell(digest), anchor, formatRecommendExecTime(result.ExecTime), result.ExecCount, result.AvgMem, result.AvgRequestUnitRead,
			formatEstImprovement(result.Recommendations), strings.ToLower(result.Digest))
	}
	report.WriteString("\n# Recommendation List\n")
	for _, result := range results {
		digest := strings.ToLower(result.Digest)
		fmt.Fprintf(&report, "\n<a id=\"sql-%s\"></a>\n\n## SQL: %s\n\n", digest, html.EscapeString(firstEightRecommend(result.Digest)))
		writeMetadataBlock(&report, fmt.Sprintf("Schema:     %s\nSQL Digest: %s", result.Schema, result.Digest))
		report.WriteString("\n")
		report.WriteString("### SQL Statement\n\n")
		writeRecommendSQLBlock(&report, result.SQL)
		fmt.Fprintf(&report, "\n<a id=\"index-recommendation-%s\"></a>\n\n### Index Recommendation\n\n", digest)
		for i, advice := range result.Recommendations {
			fmt.Fprintf(&report, "#### Recommendation %d\n\n", i+1)
			var metadata []string
			for _, field := range strings.Split(advice.output, "\n") {
				name, value, _ := strings.Cut(field, ":")
				switch strings.ToLower(strings.TrimSpace(name)) {
				case "database", "table", "index_name", "index_columns", "est_index_size":
					metadata = append(metadata, fmt.Sprintf("%-15s %s", strings.TrimSpace(name)+":", strings.TrimSpace(value)))
				}
			}
			writeMetadataBlock(&report, strings.Join(metadata, "\n"))
			report.WriteString("\n")
			for _, field := range strings.Split(advice.output, "\n") {
				name, value, _ := strings.Cut(field, ":")
				value = strings.TrimSpace(value)
				switch strings.ToLower(strings.TrimSpace(name)) {
				case "database", "table", "index_name", "index_columns", "est_index_size":
					continue
				case "reason":
					fmt.Fprintf(&report, "reason: %s\n\n", formatRecommendationReason(value))
				case "top_impacted_query":
					writeImpactedQueries(&report, value)
				case "create_index_statement":
					report.WriteString("**Create Index Statement**\n\n")
					writeRecommendSQLBlock(&report, value)
					report.WriteString("\n")
				default:
					fmt.Fprintf(&report, "%s\n\n", field)
				}
			}
		}
		fmt.Fprintf(&report, "\n<a id=\"old-plan-%s\"></a>\n\n### Old Plan\n\n", digest)
		writePlanBlock(&report, alignPlanColumns(compare.WithoutPlanColumns(result.OldPlan, "actRows", "execution info", "memory", "disk")))
		fmt.Fprintf(&report, "\n<a id=\"new-plan-%s\"></a>\n\n### New Plan\n\n", digest)
		writeRecommendSQLBlock(&report, "-- Sql with hypo_index hint\n"+result.NewSQL)
		report.WriteString("\n")
		writePlanBlock(&report, alignPlanColumns(result.NewPlan))
	}
	return report.String()
}

func formatRecommendationReason(reason string) string {
	if index := strings.Index(strings.ToLower(reason), "in query:"); index >= 0 {
		return strings.TrimSpace(reason[:index])
	}
	return reason
}

// Metadata uses a transparent monospace block to keep values aligned without
// the background and border normally applied to fenced code blocks.
func writeMetadataBlock(out *strings.Builder, content string) {
	out.WriteString("<pre style=\"background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;\">")
	out.WriteString(html.EscapeString(content))
	out.WriteString("</pre>\n")
}

// Keep wide plan tables on one line per row and allow horizontal scrolling
// in Markdown viewers that support inline HTML styles.
func writePlanBlock(out *strings.Builder, plan string) {
	out.WriteString("<pre style=\"white-space: pre; overflow-x: auto; max-width: 100%;\"><code style=\"white-space: pre; overflow-wrap: normal; word-break: normal;\">")
	out.WriteString(html.EscapeString(plan))
	out.WriteString("</code></pre>\n")
}

// alignPlanColumns replaces tab separators with padding based on the widest
// value in each column. Leading spaces in operator IDs retain the plan tree.
func alignPlanColumns(plan string) string {
	lines := strings.Split(plan, "\n")
	rows := make([][]string, len(lines))
	var widths []int
	for i, line := range lines {
		rows[i] = strings.Split(line, "\t")
		for j, value := range rows[i] {
			value = strings.TrimRight(value, " ")
			rows[i][j] = value
			if j >= len(widths) {
				widths = append(widths, 0)
			}
			widths[j] = max(widths[j], utf8.RuneCountInString(value))
		}
	}
	var out strings.Builder
	for i, row := range rows {
		if i > 0 {
			out.WriteByte('\n')
		}
		for j, value := range row {
			out.WriteString(value)
			if j < len(row)-1 {
				out.WriteString(strings.Repeat(" ", widths[j]-utf8.RuneCountInString(value)+2))
			}
		}
	}
	return out.String()
}

// Each advisor invocation analyzes one SQL. For multiple recommendations,
// display the best individual estimate, without adding their improvements.
func formatEstImprovement(recommendations []IndexRecommendation) string {
	var best *float64
	for _, advice := range recommendations {
		if advice.improvement != nil && (best == nil || *advice.improvement > *best) {
			best = advice.improvement
		}
	}
	if best == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.4f%%", *best*100)
}

func writeImpactedQueries(out *strings.Builder, value string) {
	var queries []struct {
		Query       string   `json:"Query"`
		Improvement *float64 `json:"Improvement"`
	}
	out.WriteString("**Top Impacted Queries**\n\n")
	if err := json.Unmarshal([]byte(value), &queries); err != nil || len(queries) == 0 {
		writeRecommendCodeBlock(out, value)
		out.WriteString("\n")
		return
	}
	for i, query := range queries {
		fmt.Fprintf(out, "**Query %d**\n\n", i+1)
		writeRecommendSQLBlock(out, query.Query)
		if query.Improvement != nil {
			fmt.Fprintf(out, "\nEstimated Improvement: %.4f%%\n\n", *query.Improvement*100)
		} else {
			out.WriteString("\nEstimated Improvement: N/A\n\n")
		}
	}
}

func writeRecommendSQLBlock(out *strings.Builder, content string) {
	fence := "```"
	for strings.Contains(content, fence) {
		fence += "`"
	}
	fmt.Fprintf(out, "%ssql\n%s\n%s\n", fence, content, fence)
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

// Preserve leading spaces and tree branches in EXPLAIN output.
func planCellString(cell any) string {
	if cell == nil {
		return ""
	}
	if value, ok := cell.([]byte); ok {
		return string(value)
	}
	return fmt.Sprint(cell)
}
