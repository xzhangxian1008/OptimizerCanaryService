package redact

import (
	"errors"
	"regexp"
	"strings"
)

var (
	errUnsafePlan = errors.New("plan cannot be safely redacted")
	planID        = regexp.MustCompile(`^[ │├└─]*[A-Za-z][A-Za-z0-9_]*_[0-9]+(?:\((?:Build|Probe)\))?$`)
	planNumber    = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?|N/A)$`)
	planTask      = regexp.MustCompile(`^(?:root|cop\[(?:tikv|tiflash|tidb)\]|mpp\[tiflash\])$`)
)

// Plan redacts plan details without using SQL context to identify columns.
func Plan(text string) (string, error) {
	return plan(text, nil)
}

// PlanWithSQL uses known identifiers from the SQL and schema to retain column
// references in recognized expressions. Unrecognized content is still hidden.
func PlanWithSQL(text, sqlText, schema string) (string, error) {
	return plan(text, planIdentifiers(sqlText, schema))
}

func plan(text string, identifiers map[string]bool) (string, error) {
	lines := strings.Split(strings.Trim(text, "\r\n"), "\n")
	if len(lines) < 2 || !strings.Contains(lines[0], "\t") {
		return "", errUnsafePlan
	}
	// Some statement-summary plans prefix every line with a formatting tab.
	if strings.HasPrefix(lines[0], "\t") {
		for i, line := range lines {
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "\t") {
				return "", errUnsafePlan
			}
			lines[i] = strings.TrimPrefix(line, "\t")
		}
	}
	headers := strings.Split(lines[0], "\t")
	kinds := make([]string, len(headers))
	hasID := false
	for i, header := range headers {
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "id":
			headers[i], kinds[i], hasID = "id", "id", true
		case "task":
			headers[i], kinds[i] = "task", "task"
		case "estrows":
			headers[i], kinds[i] = "estRows", "number"
		case "estcost":
			headers[i], kinds[i] = "estCost", "number"
		case "access object":
			headers[i], kinds[i] = "access object", "access"
		case "operator info":
			headers[i], kinds[i] = "operator info", "info"
		default:
			headers[i] = "?"
		}
	}
	if !hasID {
		return "", errUnsafePlan
	}
	var out strings.Builder
	out.WriteString(strings.Join(headers, "\t"))
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		cells := strings.Split(line, "\t")
		if len(cells) != len(headers) {
			return "", errUnsafePlan
		}
		for i, cell := range cells {
			value := strings.TrimSpace(cell)
			switch {
			case kinds[i] == "id" && planID.MatchString(cell):
				cells[i] = cell
			case kinds[i] == "task" && planTask.MatchString(value):
				cells[i] = value
			case kinds[i] == "number" && planNumber.MatchString(value):
				cells[i] = value
			case kinds[i] == "access":
				cells[i] = redactAccessObject(value)
			case kinds[i] == "info":
				cells[i] = redactOperatorInfo(value, identifiers)
			default:
				cells[i] = "?"
			}
		}
		out.WriteByte('\n')
		out.WriteString(strings.Join(cells, "\t"))
	}
	return out.String(), nil
}
