package diagnosis

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"go.uber.org/zap"
)

const compareStatementsQuery = `SELECT COALESCE(schema_name, ''), digest,
       query_sample_text, exec_count, COALESCE(plan_digest, ''), COALESCE(plan, ''),
       COALESCE(sum_latency, 0), COALESCE(plan_hint, '')
FROM information_schema.cluster_statements_summary
WHERE LOWER(stmt_type) = 'select'
  AND query_sample_text IS NOT NULL AND query_sample_text <> ''
  AND digest IS NOT NULL AND digest <> ''
  AND plan_digest IS NOT NULL AND plan_digest <> ''
  AND plan IS NOT NULL AND plan <> ''
  AND LOWER(query_sample_text) NOT LIKE '%information_schema.cluster_statements_summary%'
  AND LOWER(query_sample_text) NOT LIKE '%information_schema.tiflash_replica%'
  AND LOWER(query_sample_text) NOT LIKE '%mysql.bind_info%'
  AND LOWER(query_sample_text) NOT LIKE 'select @@version_comment%'
ORDER BY digest, schema_name, plan_digest, instance`

const compareBindingsQuery = `SELECT original_sql, bind_sql, default_db, status,
       create_time, update_time, charset, collation, source, sql_digest, plan_digest
FROM mysql.bind_info
WHERE sql_digest = ? AND status <> 'deleted'
  AND (default_db = ? OR default_db = '')
ORDER BY update_time, bind_sql`

func (r *SQLRepository) StatementPlans(ctx context.Context) ([]StatementPlan, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	r.logger.Info("execute SQL", zap.String("sql", compareStatementsQuery))
	rows, err := r.db.QueryContext(queryCtx, compareStatementsQuery)
	if err != nil {
		return nil, newStackErrorf("query cluster statement plans: %w", err)
	}
	defer rows.Close()
	var statements []StatementPlan
	for rows.Next() {
		var statement StatementPlan
		if err := rows.Scan(&statement.Schema, &statement.SQLDigest, &statement.SQL, &statement.ExecCount, &statement.PlanDigest, &statement.Plan, &statement.ExecTime, &statement.PlanHint); err != nil {
			return nil, newStackErrorf("scan cluster statement plan: %w", err)
		}
		if statement.SQLDigest == "" || statement.PlanDigest == "" || strings.TrimSpace(statement.SQL) == "" {
			return nil, newStackErrorf("cluster statement plan is missing SQL or digest (SQL digest %q)", statement.SQLDigest)
		}
		statements = append(statements, statement)
	}
	if err := rows.Err(); err != nil {
		return nil, newStackErrorf("read cluster statement plans: %w", err)
	}
	return statements, nil
}

func (r *SQLRepository) ExplainPlan(ctx context.Context, sample Sample) (string, error) {
	plan, _, err := r.ExplainPlanWithDigest(ctx, sample)
	return plan, err
}

func (r *SQLRepository) ExplainPlanWithDigest(ctx context.Context, sample Sample) (string, string, error) {
	explainCtx, cancel := context.WithTimeout(ctx, explainTimeout)
	defer cancel()
	conn, err := r.db.Conn(explainCtx)
	if err != nil {
		return "", "", newStackErrorf("acquire connection: %w", err)
	}
	defer conn.Close()
	if sample.Schema != "" {
		useSQL := "USE `" + strings.ReplaceAll(sample.Schema, "`", "``") + "`"
		r.logger.Info("execute SQL", zap.String("sql", useSQL))
		if _, err := conn.ExecContext(explainCtx, useSQL); err != nil {
			return "", "", newStackErrorf("select schema %q: %w", sample.Schema, err)
		}
	}
	statement, arguments, err := explainableSample(sample.SQL)
	if err != nil {
		return "", "", fmt.Errorf("parse prepared statement sample: %w", err)
	}
	if statement == "" {
		return "", "", newStackErrorf("cannot explain an empty statement")
	}
	explainSQL := "EXPLAIN " + statement
	r.logger.Info("execute SQL", zap.String("schema", sample.Schema), zap.String("sql", explainSQL))
	rows, err := conn.QueryContext(explainCtx, explainSQL, arguments...)
	if err != nil {
		return "", "", newStackErrorf("explain statement: %w", err)
	}
	defer rows.Close()
	columns, values, err := readStringRows(rows)
	if err != nil {
		return "", "", fmt.Errorf("read EXPLAIN result: %w", err)
	}
	var plan strings.Builder
	plan.WriteString(strings.Join(columns, "\t"))
	for _, row := range values {
		plan.WriteByte('\n')
		plan.WriteString(strings.Join(row, "\t"))
	}
	planText := plan.String()
	planDigest, err := r.explainPlanDigest(explainCtx, conn, sample.Schema, explainSQL, planText)
	if err != nil {
		return "", "", err
	}
	return planText, planDigest, nil
}

const explainPlanDigestQuery = `SELECT COALESCE(query_sample_text, ''), COALESCE(plan_digest, ''), COALESCE(plan, '')
FROM information_schema.cluster_statements_summary
WHERE LOWER(stmt_type) IN ('explainsql', 'explain')
  AND (? = '' OR COALESCE(schema_name, '') = ?)
  AND plan_digest IS NOT NULL AND plan_digest <> ''
  AND plan IS NOT NULL AND plan <> ''
ORDER BY summary_end_time DESC`

func (r *SQLRepository) explainPlanDigest(ctx context.Context, conn *sql.Conn, schema, explainSQL, newPlan string) (string, error) {

	rows, err := conn.QueryContext(ctx, explainPlanDigestQuery, schema, schema)
	if err != nil {
		return "", newStackErrorf("query EXPLAIN plan digest: %w", err)
	}
	defer rows.Close()
	newOperators, err := planOperators(newPlan)
	if err != nil {
		return "", fmt.Errorf("read EXPLAIN plan for digest lookup: %w", err)
	}
	newDigestOperators := digestLookupOperators(newOperators)
	normalizedExplainSQL := normalizeExplainSQL(explainSQL)
	for rows.Next() {
		var querySample, digest, plan string
		if err := rows.Scan(&querySample, &digest, &plan); err != nil {
			return "", newStackErrorf("scan EXPLAIN plan digest: %w", err)
		}
		if normalizeExplainSQL(querySample) != normalizedExplainSQL {
			continue
		}
		operators, err := planOperators(plan)
		if err == nil && slices.Equal(newDigestOperators, digestLookupOperators(operators)) {
			return digest, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", newStackErrorf("read EXPLAIN plan digests: %w", err)
	}
	return "", newStackErrorf("no plan digest found for EXPLAIN of SQL in schema %q", schema)
}

func normalizeExplainSQL(sqlText string) string {
	var normalized strings.Builder
	spacePending := false
	for i := 0; i < len(sqlText); {
		char := sqlText[i]
		if char == '\'' || (char == '"' && isDoubleQuotedStringStart(sqlText, i)) {
			if spacePending && normalized.Len() > 0 {
				normalized.WriteByte(' ')
				spacePending = false
			}
			end := quotedSQLTokenEnd(sqlText, i, char)
			if char == '"' {
				normalized.WriteByte('\'')
				normalized.WriteString(sqlText[i+1 : end-1])
				normalized.WriteByte('\'')
			} else {
				normalized.WriteString(sqlText[i:end])
			}
			i = end
			continue
		}
		if char == '"' {
			if spacePending && normalized.Len() > 0 {
				normalized.WriteByte(' ')
				spacePending = false
			}
			end := quotedSQLTokenEnd(sqlText, i, char)
			normalized.WriteString(sqlText[i:end])
			i = end
			continue
		}
		if unicode.IsSpace(rune(char)) {
			spacePending = true
			i++
			continue
		}
		if spacePending && normalized.Len() > 0 {
			normalized.WriteByte(' ')
		}
		spacePending = false
		normalized.WriteRune(unicode.ToLower(rune(char)))
		i++
	}
	return normalized.String()
}

func quotedSQLTokenEnd(sqlText string, start int, quote byte) int {
	for i := start + 1; i < len(sqlText); i++ {
		if sqlText[i] != quote {
			continue
		}
		if i+1 < len(sqlText) && sqlText[i+1] == quote {
			i++
			continue
		}
		return i + 1
	}
	return len(sqlText)
}

func isDoubleQuotedStringStart(sqlText string, index int) bool {
	for index > 0 && sqlText[index-1] == ' ' {
		index--
	}
	if index == 0 {
		return true
	}
	switch sqlText[index-1] {
	case '=', '<', '>', '!', '(', ',', '+', '-', '*', '/', '%':
		return true
	default:
		return false
	}
}

// digestLookupOperators accounts for the task name difference between direct
// EXPLAIN output (mpp[...]) and the plan text stored in statement summary
// (cop[...]). This normalization is only used to locate the new plan digest;
// old/new plan comparison still uses the original id/task values.
func digestLookupOperators(operators []planOperator) []planOperator {
	normalized := make([]planOperator, len(operators))
	copy(normalized, operators)
	for i := range normalized {
		if strings.HasPrefix(normalized[i].task, "mpp[") {
			normalized[i].task = "cop" + normalized[i].task[len("mpp"):]
		}
	}
	return normalized
}

// explainableSample removes the annotation TiDB appends to a server-side
// prepared statement in QUERY_SAMPLE_TEXT and turns its values into database
// parameters. For example:
//
//	SELECT * FROM t WHERE id = ? [arguments: 42]
//
// becomes EXPLAIN SELECT * FROM t WHERE id = ? with one bound argument. Sending
// the values as database parameters preserves their types and avoids quoting
// mistakes when a value contains SQL punctuation.
func explainableSample(sample string) (string, []any, error) {
	sample = strings.TrimSpace(sample)
	marker := "[arguments:"
	markerIndex := strings.LastIndex(sample, marker)
	if markerIndex <= 0 || !unicode.IsSpace(rune(sample[markerIndex-1])) {
		return strings.TrimSuffix(sample, ";"), nil, nil
	}

	statement := strings.TrimSpace(sample[:markerIndex])
	argumentText := strings.TrimSpace(sample[markerIndex+len(marker):])
	argumentText = strings.TrimSpace(strings.TrimSuffix(argumentText, ";"))
	if !strings.HasSuffix(argumentText, "]") {
		return "", nil, newStackErrorf("arguments annotation is not closed")
	}
	argumentText = strings.TrimSpace(strings.TrimSuffix(argumentText, "]"))
	arguments, err := parseTiDBArguments(argumentText)
	if err != nil {
		return "", nil, err
	}
	return strings.TrimSuffix(strings.TrimSpace(statement), ";"), arguments, nil
}

func parseTiDBArguments(text string) ([]any, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	if strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") {
		text = strings.TrimSpace(text[1 : len(text)-1])
	}
	if text == "" {
		return nil, nil
	}

	parts := make([]string, 0, 2)
	start, depth := 0, 0
	var quote rune
	escaped := false
	for i, char := range text {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == quote {
				quote = 0
			}
			continue
		}
		switch char {
		case '\'', '"', '`':
			quote = char
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return nil, newStackErrorf("unexpected ')' in arguments")
			}
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(text[start:i]))
				start = i + 1
			}
		}
	}
	if quote != 0 {
		return nil, newStackErrorf("unterminated quoted argument")
	}
	if depth != 0 {
		return nil, newStackErrorf("unclosed parenthesized argument")
	}
	parts = append(parts, strings.TrimSpace(text[start:]))

	arguments := make([]any, 0, len(parts))
	for _, part := range parts {
		argument, err := parseTiDBArgument(part)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
	}
	return arguments, nil
}

func parseTiDBArgument(text string) (any, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, newStackErrorf("empty argument")
	}
	if strings.EqualFold(text, "NULL") {
		return nil, nil
	}
	if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
		if text[0] == '"' {
			value, err := strconv.Unquote(text)
			if err != nil {
				return nil, newStackErrorf("decode string argument %q: %w", text, err)
			}
			return value, nil
		}
		return text[1 : len(text)-1], nil
	}
	if strings.HasPrefix(strings.ToLower(text), "0x") && len(text)%2 == 0 {
		value, err := hex.DecodeString(text[2:])
		if err == nil {
			return value, nil
		}
	}
	if value, err := strconv.ParseInt(text, 10, 64); err == nil {
		return value, nil
	}
	if value, err := strconv.ParseUint(text, 10, 64); err == nil {
		return value, nil
	}
	if value, err := strconv.ParseFloat(text, 64); err == nil {
		return value, nil
	}
	return text, nil
}

func (r *SQLRepository) Bindings(ctx context.Context, digest, schema string) (string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	r.logger.Info("execute SQL", zap.String("sql", compareBindingsQuery), zap.String("digest", digest), zap.String("schema", schema))
	rows, err := r.db.QueryContext(queryCtx, compareBindingsQuery, digest, schema)
	if err != nil {
		return "", newStackErrorf("query bindings: %w", err)
	}
	defer rows.Close()
	columns, values, err := readStringRows(rows)
	if err != nil {
		return "", fmt.Errorf("read bindings: %w", err)
	}
	var binding strings.Builder
	for i, row := range values {
		if i > 0 {
			binding.WriteByte('\n')
		}
		for j, value := range row {
			fmt.Fprintf(&binding, "%s: %s\n", columns[j], value)
		}
	}
	return binding.String(), nil
}

func readStringRows(rows *sql.Rows) ([]string, [][]string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, newStackErrorf("read result columns: %w", err)
	}
	var result [][]string
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		destinations := make([]any, len(columns))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return nil, nil, newStackErrorf("scan result row: %w", err)
		}
		row := make([]string, len(columns))
		for i := range values {
			row[i] = values[i].String
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, newStackErrorf("read result rows: %w", err)
	}
	return columns, result, nil
}

func (m *ConnectionManager) Compare(ctx context.Context) (string, error) {
	db, err := m.current()
	if err != nil {
		return "", err
	}
	return NewComparer(NewSQLRepository(db, m.logger)).Compare(ctx)
}

var _ CompareRepository = (*SQLRepository)(nil)
