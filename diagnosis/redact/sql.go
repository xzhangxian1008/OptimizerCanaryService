// Package redact produces report-safe representations of diagnostic data.
package redact

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// Omitted is used when a field cannot be safely redacted.
const Omitted = "[content omitted: unable to redact]"

var errUnsafeSQL = errors.New("SQL cannot be safely redacted")

// SQL redacts SQL literals for display, not execution. An error never includes
// input text, and callers must not fall back to displaying the original SQL.
func SQL(text string) (result string, err error) {
	defer func() {
		if recover() != nil {
			result, err = "", errUnsafeSQL
		}
	}()
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errUnsafeSQL
	}
	// TiDB can append prepared-statement arguments after the SQL. Parse the
	// whole input first, then try removing the annotation only if needed.
	result, err = redactSQLModes(text)
	if err == nil {
		return result, nil
	}
	trimmed := strings.TrimSpace(strings.TrimSuffix(text, ";"))
	marker := strings.LastIndex(trimmed, "[arguments:")
	if marker <= 0 || !strings.HasSuffix(trimmed, "]") || !strings.ContainsAny(trimmed[marker-1:marker], " \t\r\n") {
		return "", errUnsafeSQL
	}
	return redactSQLModes(strings.TrimSpace(trimmed[:marker]))
}

func redactSQLModes(text string) (string, error) {
	// The source session's SQL mode is unavailable. Only emit output when
	// parsing with all combinations of these quote/escape modes agrees.
	modes := []mysql.SQLMode{0, mysql.ModeANSIQuotes, mysql.ModeNoBackslashEscapes, mysql.ModeANSIQuotes | mysql.ModeNoBackslashEscapes}
	var result string
	for _, mode := range modes {
		p := parser.New()
		p.SetSQLMode(mode)
		stmt, err := p.ParseOneStmt(text, "", "")
		if err != nil {
			return "", errUnsafeSQL
		}
		switch stmt.(type) {
		case *ast.SelectStmt, *ast.SetOprStmt, *ast.CreateIndexStmt:
		default:
			return "", errUnsafeSQL
		}
		node, ok := stmt.Accept(sqlRedactor{})
		if !ok {
			return "", errUnsafeSQL
		}
		var out strings.Builder
		if err := node.Restore(format.NewRestoreCtx(format.DefaultRestoreFlags, &out)); err != nil {
			return "", errUnsafeSQL
		}
		// Restoring first produces a canonical SQL mode. The normalizer also
		// masks literals stored outside ordinary ValueExpr nodes.
		safe := parser.NormalizeKeepHint(out.String())
		// The normalizer collapses lists of literals to "..."; the report
		// uses a single placeholder for those values instead.
		safe = replaceListMarkers(safe)
		if safe == "" || (result != "" && result != safe) {
			return "", errUnsafeSQL
		}
		result = safe
	}
	return result, nil
}

func replaceListMarkers(text string) string {
	var out strings.Builder
	quoted := false
	for i := 0; i < len(text); {
		if text[i] == '`' {
			out.WriteByte(text[i])
			i++
			if quoted && i < len(text) && text[i] == '`' {
				out.WriteByte(text[i])
				i++
				continue
			}
			quoted = !quoted
			continue
		}
		if !quoted && strings.HasPrefix(text[i:], " ... ") {
			out.WriteString(" ? ")
			i += len(" ... ")
			continue
		}
		out.WriteByte(text[i])
		i++
	}
	return out.String()
}

type sqlRedactor struct{}

func (sqlRedactor) Enter(node ast.Node) (ast.Node, bool) {
	if selectStmt, ok := node.(*ast.SelectStmt); ok {
		// Hints containing literals or unknown payloads are omitted. Retain
		// explicitly supported hints that contain only object identifiers.
		var safeHints []*ast.TableOptimizerHint
		for _, hint := range selectStmt.TableHints {
			if hint.HintData == nil && safeHint(hint.HintName.L) {
				safeHints = append(safeHints, hint)
			}
		}
		selectStmt.TableHints = safeHints
	}
	if _, ok := node.(ast.ValueExpr); ok {
		return ast.NewParamMarkerExpr(0), true
	}
	return node, false
}

func (sqlRedactor) Leave(node ast.Node) (ast.Node, bool) { return node, true }

func safeHint(name string) bool {
	switch strings.ToLower(name) {
	case "use_index", "ignore_index", "use_index_merge", "hash_join", "merge_join", "inl_join", "inl_hash_join", "inl_merge_join", "hash_agg", "stream_agg":
		return true
	default:
		return false
	}
}

// Binding redacts the original and hinted statements separately.
func Binding(original, hinted string) (string, error) {
	originalSafe, err := SQL(original)
	if err != nil {
		return "", errUnsafeSQL
	}
	hintedSafe, err := SQL(hinted)
	if err != nil {
		return "", errUnsafeSQL
	}
	return fmt.Sprintf("CREATE GLOBAL BINDING FOR %s USING %s;", originalSafe, hintedSafe), nil
}
