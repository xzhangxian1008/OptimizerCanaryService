package index_advisor

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/compare"
	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/redact"
	"go.uber.org/zap"
)

type Sample struct {
	Schema string
	SQL    string
}

const recommendIndexQueryTimeout = 15 * time.Second
const recommendIndexMaxResults = 20
const recommendIndexMaxWarningLogCnt = 3000

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
	SizeBytes       uint64
	Table           string `json:"table"`
	Name            string `json:"index_name"`
	Columns         string `json:"index_columns"`
	output          string
	createStatement string
	Database        string
	newSQL          string
	newPlan         string
	improvement     *float64
}

type RecommendIndexRepository interface {
	RecommendIndex(context.Context) ([]RecommendIndexResult, error)
}

// ConfigurableRecommendIndexRepository supports selecting the SQL column used
// by the cluster slow-query source for a single request.
type ConfigurableRecommendIndexRepository interface {
	RecommendIndexWithSQLColumn(context.Context, string) ([]RecommendIndexResult, error)
}

type SQLRecommendIndexRepository struct {
	db                 *sql.DB
	logger             *zap.Logger
	slowQuerySQLColumn string
}

type recommendIndexSample struct {
	Sample
	Source             string
	Digest             string
	ExecTime           uint64
	ExecCount          uint64
	AvgMem             float64
	AvgRequestUnitRead float64
	Instance           string
	WindowStart        string
	LastSeen           string
}

func NewSQLRecommendIndexRepository(db *sql.DB, logger *zap.Logger, slowQuerySQLColumn ...string) *SQLRecommendIndexRepository {
	column := DefaultSlowQuerySQLColumn
	if len(slowQuerySQLColumn) > 0 {
		column = NormalizeSlowQuerySQLColumn(slowQuerySQLColumn[0])
	}
	return &SQLRecommendIndexRepository{db: db, logger: logger, slowQuerySQLColumn: column}
}

func (r *SQLRecommendIndexRepository) RecommendIndex(ctx context.Context) ([]RecommendIndexResult, error) {
	return r.recommendIndex(ctx)
}

func (r *SQLRecommendIndexRepository) RecommendIndexWithSQLColumn(ctx context.Context, column string) ([]RecommendIndexResult, error) {
	copy := *r
	copy.slowQuerySQLColumn = NormalizeSlowQuerySQLColumn(column)
	return copy.recommendIndex(ctx)
}

func (r *SQLRecommendIndexRepository) recommendIndex(ctx context.Context) ([]RecommendIndexResult, error) {
	current, err := r.readSamples(ctx, currentSummaryQuery)
	if err != nil {
		return nil, fmt.Errorf("read current statement summary: %w", err)
	}
	// for _, sample := range current {
	// 	r.logger.Info(fmt.Sprintf("xzxdebug: %s", sample.SQL))
	// }
	markSampleSource(current, "cluster_statements_summary")

	history, err := r.readSamples(ctx, historySummaryQuery)
	if err != nil {
		return nil, fmt.Errorf("read statement summary history: %w", err)
	}
	// for _, sample := range history {
	// 	r.logger.Info(fmt.Sprintf("xzxdebug: %s", sample.SQL))
	// }
	markSampleSource(history, "cluster_statements_summary_history")

	slowQueries, err := r.readSlowQuerySamples(ctx)
	if err != nil {
		// CLUSTER_SLOW_QUERY is not available on some TiDB Cloud tiers and
		// older TiDB versions. Keep statement-summary recommendations usable
		// when that optional source is unavailable.
		if isStaleObjectError(err) {
			r.logger.Warn("skip cluster slow query source", zap.Error(err))
			slowQueries = nil
		} else {
			return nil, fmt.Errorf("read cluster slow query: %w", err)
		}
	}
	// for _, sample := range slowQueries {
	// 	r.logger.Info(fmt.Sprintf("xzxdebug: %s", sample.SQL))
	// }
	markSampleSource(slowQueries, "cluster_slow_query")

	samples := aggregateAndSortSamples(append(mergeSummaryWindows(current, history), slowQueries...))
	return recommendSamples(ctx, samples, r, r.logger)
}

type indexRecommendationRunner interface {
	recommendFor(context.Context, Sample) ([]IndexRecommendation, error)
	getSQLPlan(context.Context, Sample, string) (string, error)
}

func recommendSamples(ctx context.Context, samples []recommendIndexSample, runner indexRecommendationRunner, logger *zap.Logger) ([]RecommendIndexResult, error) {
	results := make([]RecommendIndexResult, 0, recommendIndexMaxResults)
	warningCount := 0
	warnSQL := func(message string, sample recommendIndexSample, err error) {
		if warningCount >= recommendIndexMaxWarningLogCnt {
			return
		}
		logger.Warn(message,
			zap.String("schema", sample.Schema),
			zap.String("digest", sample.Digest),
			zap.String("source", sample.Source),
			zap.String("sql", sample.SQL),
			zap.Error(err))
		warningCount++
	}

	for _, sample := range samples {
		if isSampleInBlacklist(sample) {
			continue
		}
		recommendations, err := runner.recommendFor(ctx, sample.Sample)
		if err != nil {
			if isStaleObjectError(err) {
				warnSQL("skip SQL referencing a missing database object", sample, err)
				continue
			}
			if isSQLSyntaxError(err) {
				warnSQL("skip SQL with syntax error during index recommendation", sample, err)
				continue
			}

			if isEmptyQuerySetError(err) {
				warnSQL("skip SQL rejected by index advisor", sample, err)
				continue
			}
			warnSQL("skip SQL after index recommendation error", sample, err)
			continue
		}

		if len(recommendations) == 0 {
			continue
		}

		oldPlan, err := runner.getSQLPlan(ctx, sample.Sample, convertToRecommendableStatement(sample.SQL))
		if err != nil {
			if isStaleObjectError(err) {
				warnSQL("skip SQL referencing a missing database object", sample, err)
				continue
			}
			if isSQLSyntaxError(err) {
				warnSQL("skip SQL with syntax error during original EXPLAIN", sample, err)
				continue
			}
			warnSQL("skip SQL after original EXPLAIN error", sample, err)
			continue
		}

		var completed []IndexRecommendation
		for _, advice := range recommendations {
			hintedSQL, err := buildHypoIndexSQL(sample.SQL, []IndexRecommendation{advice})
			if err != nil {
				warnSQL("skip index advice after building hypothetical SQL error", sample, err)
				continue
			}

			newPlan, err := runner.getSQLPlan(ctx, sample.Sample, hintedSQL)
			if err != nil {
				if isStaleObjectError(err) {
					warnSQL("skip advice referencing a missing database object", sample, err)
					continue
				}
				if isSQLSyntaxError(err) {
					warnSQL("skip advice with syntax error during hypothetical index EXPLAIN", sample, err)
					continue
				}
				warnSQL("skip advice after hypothetical index EXPLAIN error", sample, err)
				continue
			}

			// TODO: compare oldPlan and newPlan; skip advice if the plans are equal.
			advice.newSQL, advice.newPlan = hintedSQL, newPlan
			completed = append(completed, advice)
		}

		if len(completed) == 0 {
			continue
		}

		results = append(results, RecommendIndexResult{
			Schema:             sample.Schema,
			Digest:             sample.Digest,
			SQL:                sample.SQL,
			ExecTime:           sample.ExecTime,
			ExecCount:          sample.ExecCount,
			Recommendations:    completed,
			OldPlan:            oldPlan,
			AvgMem:             sample.AvgMem,
			AvgRequestUnitRead: sample.AvgRequestUnitRead,
		})
		if len(results) >= recommendIndexMaxResults {
			break
		}
	}
	return results, nil
}

func isSampleInBlacklist(sample recommendIndexSample) bool {
	switch strings.ToLower(sample.Schema) {
	case "information_schema", "mysql":
		return true
	}
	sql := strings.ToLower(sample.SQL)
	for _, fragment := range []string{"from mysql.", "from information_schema.", "cluster_slow_query"} {
		if strings.Contains(sql, fragment) {
			return true
		}
	}
	return false
}

type recommendIndexSampleKey struct {
	schema string
	digest string
}

func aggregateAndSortSamples(rawSamples []recommendIndexSample) []recommendIndexSample {
	byKey := make(map[recommendIndexSampleKey]recommendIndexSample, len(rawSamples))
	for _, sample := range rawSamples {
		key := recommendIndexSampleKey{schema: strings.ToLower(sample.Schema), digest: strings.ToLower(sample.Digest)}
		if existing, ok := byKey[key]; ok {
			totalCount := existing.ExecCount + sample.ExecCount
			if totalCount > 0 {
				existing.AvgMem = (existing.AvgMem*float64(existing.ExecCount) + sample.AvgMem*float64(sample.ExecCount)) / float64(totalCount)
				existing.AvgRequestUnitRead = (existing.AvgRequestUnitRead*float64(existing.ExecCount) + sample.AvgRequestUnitRead*float64(sample.ExecCount)) / float64(totalCount)
			}

			existing.ExecCount = totalCount
			existing.ExecTime += sample.ExecTime
			if sample.LastSeen > existing.LastSeen || (sample.LastSeen == existing.LastSeen && sample.SQL < existing.SQL) {
				existing.SQL, existing.LastSeen, existing.Source = sample.SQL, sample.LastSeen, sample.Source
			}
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
		if samples[i].Digest != samples[j].Digest {
			return samples[i].Digest < samples[j].Digest
		}
		return samples[i].Schema < samples[j].Schema
	})
	return samples
}

func isEmptyQuerySetError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "empty query set after filtering invalid queries")
}

func isStaleObjectError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}
	switch mysqlErr.Number {
	case 1049, 1054, 1146: // Unknown database, unknown column, or missing table.
		return true
	default:
		return false
	}
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

	statement := convertToRecommendableStatement(sample.SQL)
	recommendSQL := "RECOMMEND INDEX RETURN RESULT FOR'" + strings.ReplaceAll(statement, "'", "''") + "'"
	r.logger.Info("execute SQL", zap.String("schema", sample.Schema), zap.String("sql", recommendSQL))
	rows, err := conn.QueryContext(recommendCtx, recommendSQL)
	if err != nil {
		return nil, fmt.Errorf("recommend index return result statement: %w", err)
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
		advice.createStatement = row["create_index_statement"]
		advice.SizeBytes, err = strconv.ParseUint(row["est_index_size"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse estimated index size %q: %w", row["est_index_size"], err)
		}
		advice.Database = row["database"]
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

func (r *SQLRecommendIndexRepository) getSQLPlan(ctx context.Context, sample Sample, statement string) (string, error) {
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
		return convertToRecommendableStatement(statement), nil
	}
	return compare.AddOptimizerHint(convertToRecommendableStatement(statement), strings.Join(hints, " "))
}

// query_sample_text can contain TiDB's server-side prepared statement suffix,
// for example "SELECT ... [arguments: 1]". That suffix is metadata rather
// than SQL and must be removed before the statement is sent to TiDB again.
func convertToRecommendableStatement(statement string) string {
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
	report.WriteString("# Index Recommendation Overview\n\n")
	report.WriteString("SQL and index statements are redacted display templates; hints and plan details may be omitted.\n\n")
	report.WriteString("<table>\n<thead><tr><th>SQL</th><th>Total Time</th><th>Count</th><th>Database</th><th>Table</th><th>Est Cost Change</th><th>Est Improvement</th><th>Advice</th></tr></thead>\n<tbody>\n")
	for _, result := range results {
		digest := recommendAnchor(result)
		for i, advice := range result.Recommendations {
			report.WriteString("<tr>")
			if i == 0 {
				fmt.Fprintf(&report, "<td rowspan=\"%d\"><a href=\"#sql-%s\">%s</a></td><td rowspan=\"%d\">%s</td><td rowspan=\"%d\">%d</td>", len(result.Recommendations), html.EscapeString(digest), html.EscapeString(firstEightRecommend(result.Digest)), len(result.Recommendations), formatRecommendExecTime(result.ExecTime), len(result.Recommendations), result.ExecCount)
			}
			database := advice.Database
			if database == "" {
				database = result.Schema
			}
			fmt.Fprintf(&report, "<td>%s</td><td>%s</td><td>%s</td><td>%s</td><td><a href=\"#index-advice-%s-%d\">detail</a></td></tr>\n", html.EscapeString(database), html.EscapeString(advice.Table), html.EscapeString(topPlanCost(redact.DisplayPlan(result.OldPlan))+" -> "+topPlanCost(redact.DisplayPlan(advice.newPlan))), formatEstImprovement([]IndexRecommendation{advice}), html.EscapeString(digest), i+1)
		}
	}
	report.WriteString("</tbody>\n</table>\n")
	report.WriteString("\n# Recommendation List\n")
	for _, result := range results {
		digest := recommendAnchor(result)
		fmt.Fprintf(&report, "\n<a id=\"sql-%s\"></a>\n\n## SQL: %s\n\n", digest, html.EscapeString(firstEightRecommend(result.Digest)))
		writeMetadataBlock(&report, fmt.Sprintf("SQL Digest: %s", result.Digest))
		report.WriteString("\n")
		report.WriteString("### SQL Statement\n\n")
		writeRecommendSQLBlock(&report, redact.DisplaySQL(result.SQL))
		for i, advice := range result.Recommendations {
			fmt.Fprintf(&report, "\n<a id=\"index-advice-%s-%d\"></a>\n\n### Advice\n\n", digest, i+1)
			report.WriteString("#### Create Index Statement\n\n")
			writeRecommendSQLBlock(&report, redact.DisplaySQL(advice.createStatement))
			report.WriteString("\n#### Old Plan\n\n")
			writePlanBlock(&report, alignPlanColumns(redact.DisplayPlanWithSQL(compare.WithoutPlanColumns(result.OldPlan, "actRows", "execution info", "memory", "disk"), result.SQL, result.Schema)))
			report.WriteString("\n#### New Plan\n\n")
			writeRecommendSQLBlock(&report, redact.DisplaySQL(advice.newSQL))
			report.WriteString("\n")
			writePlanBlock(&report, alignPlanColumns(redact.DisplayPlanWithSQL(advice.newPlan, result.SQL, result.Schema)))
		}

	}
	return report.String()
}

func recommendAnchor(result RecommendIndexResult) string {
	digest := strings.ToLower(result.Digest)
	if result.Schema == "" {
		return digest
	}
	return digest + "-" + hex.EncodeToString([]byte(strings.ToLower(result.Schema)))
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

func topPlanCost(plan string) string {
	lines := strings.Split(plan, "\n")
	if len(lines) < 2 {
		return "N/A"
	}
	header := strings.Split(lines[0], "\t")
	for i, name := range header {
		if strings.EqualFold(strings.TrimSpace(name), "estCost") {
			for _, line := range lines[1:] {
				if strings.TrimSpace(line) == "" {
					continue
				}
				cells := strings.Split(line, "\t")
				if i < len(cells) && strings.TrimSpace(cells[i]) != "" {
					return strings.TrimSpace(cells[i])
				}
				return "N/A"
			}
		}
	}
	return "N/A"
}
