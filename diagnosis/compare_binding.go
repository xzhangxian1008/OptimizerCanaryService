package diagnosis

import (
	"fmt"
	"strings"
)

func formatExecTime(ns uint64) string {
	return fmt.Sprintf("%d.%09ds", ns/1_000_000_000, ns%1_000_000_000)
}

// currentPlanBinding generates report text only; it never installs a binding.
// Selects inside CTE definitions and subqueries have greater parenthesis depth
// than the main query. At equal depth the first SELECT wins, including UNION.
func currentPlanBinding(sqlText, hint string) (string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return "", nil
	}
	if strings.Contains(hint, "*/") {
		return "", newStackErrorf("PLAN_HINT contains a comment terminator")
	}
	depth, bestDepth, selectEnd := 0, int(^uint(0)>>1), -1
	for i := 0; i < len(sqlText); {
		c := sqlText[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			quote := c
			i++
			closed := false
			for i < len(sqlText) {
				if sqlText[i] == '\\' {
					i += 2
				} else if sqlText[i] == quote {
					i++
					if i < len(sqlText) && sqlText[i] == quote {
						i++
						continue
					}
					closed = true
					break
				} else {
					i++
				}
			}
			if !closed {
				return "", newStackErrorf("unterminated quoted token in binding SQL")
			}
		case c == '#' || (c == '-' && i+2 < len(sqlText) && sqlText[i+1] == '-' && sqlText[i+2] <= ' '):
			for i < len(sqlText) && sqlText[i] != '\n' && sqlText[i] != '\r' {
				i++
			}
		case c == '/' && i+1 < len(sqlText) && sqlText[i+1] == '*':
			end := strings.Index(sqlText[i+2:], "*/")
			if end < 0 {
				return "", newStackErrorf("unterminated comment in binding SQL")
			}
			i += end + 4
		case c == '(':
			depth++
			i++
		case c == ')':
			depth--
			i++
		case sqlWordByte(c):
			start := i
			for i < len(sqlText) && sqlWordByte(sqlText[i]) {
				i++
			}
			if strings.EqualFold(sqlText[start:i], "select") && depth < bestDepth {
				bestDepth, selectEnd = depth, i
			}
		default:
			i++
		}
	}
	if selectEnd < 0 {
		return "", newStackErrorf("no main SELECT found in binding SQL")
	}
	// Replace an existing optimizer hint immediately after the main SELECT,
	// rather than emitting two adjacent hints. All other SQL bytes are retained.
	hintStart := selectEnd
	for hintStart < len(sqlText) && sqlText[hintStart] <= ' ' {
		hintStart++
	}
	if strings.HasPrefix(sqlText[hintStart:], "/*+") {
		end := strings.Index(sqlText[hintStart+3:], "*/")
		if end < 0 {
			return "", newStackErrorf("unterminated optimizer hint in binding SQL")
		}
		return sqlText[:selectEnd] + " /*+ " + hint + " */" + sqlText[hintStart+3+end+2:], nil
	}
	return sqlText[:selectEnd] + " /*+ " + hint + " */" + sqlText[selectEnd:], nil
}

func sqlWordByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
		c >= '0' && c <= '9' || c == '_' || c == '$' || c >= 128
}
