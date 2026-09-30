package compare

import (
	"database/sql"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/util"
)

const (
	queryTimeout   = 15 * time.Second
	explainTimeout = 15 * time.Second
)

const compareSqlInfoQuery = `SELECT COALESCE(schema_name, ''), digest,
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

const explainPlanDigestQuery = `SELECT COALESCE(query_sample_text, ''), COALESCE(plan_digest, ''), COALESCE(plan, '')
FROM information_schema.cluster_statements_summary
WHERE LOWER(stmt_type) IN ('explainsql', 'explain')
  AND (? = '' OR COALESCE(schema_name, '') = ?)
  AND plan_digest IS NOT NULL AND plan_digest <> ''
  AND plan IS NOT NULL AND plan <> ''
ORDER BY summary_end_time DESC`

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
		return "", nil, util.NewStackErrorf("arguments annotation is not closed")
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
				return nil, util.NewStackErrorf("unexpected ')' in arguments")
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
		return nil, util.NewStackErrorf("unterminated quoted argument")
	}
	if depth != 0 {
		return nil, util.NewStackErrorf("unclosed parenthesized argument")
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
		return nil, util.NewStackErrorf("empty argument")
	}
	if strings.EqualFold(text, "NULL") {
		return nil, nil
	}
	if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
		if text[0] == '"' {
			value, err := strconv.Unquote(text)
			if err != nil {
				return nil, util.NewStackErrorf("decode string argument %q: %w", text, err)
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

func readStringRows(rows *sql.Rows) ([]string, [][]string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, util.NewStackErrorf("read result columns: %w", err)
	}
	var result [][]string
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		destinations := make([]any, len(columns))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return nil, nil, util.NewStackErrorf("scan result row: %w", err)
		}
		row := make([]string, len(columns))
		for i := range values {
			row[i] = values[i].String
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, util.NewStackErrorf("read result rows: %w", err)
	}
	return columns, result, nil
}
