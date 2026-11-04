package redact

import (
	"regexp"
	"strings"
)

const planIdentifier = `[\p{L}_][\p{L}\p{N}_$]*`

var (
	simplePlanName = regexp.MustCompile(`^` + planIdentifier + `$`)
	qualifiedName  = regexp.MustCompile(`^` + planIdentifier + `(?:\.` + planIdentifier + `)*$`)
	accessEntry    = regexp.MustCompile(`(?i)^(table|index|clustered index|partition):\s*(` + planIdentifier + `(?:\.` + planIdentifier + `)*)(?:\((` + planIdentifier + `(?:\s*,\s*` + planIdentifier + `)*)\))?$`)
	functionCall   = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\((.*)\)$`)
)

// Identifiers are taken only from already-redacted SQL and the known schema;
// arbitrary words from operator info might instead be unquoted user values.
func planIdentifiers(sqlText, schema string) map[string]bool {
	identifiers := make(map[string]bool)
	if simplePlanName.MatchString(schema) {
		identifiers[strings.ToLower(schema)] = true
	}
	safeSQL, err := SQL(sqlText)
	if err != nil {
		return identifiers
	}
	for pos := 0; pos < len(safeSQL); {
		start := strings.IndexByte(safeSQL[pos:], '`')
		if start < 0 {
			break
		}
		start += pos + 1
		end := strings.IndexByte(safeSQL[start:], '`')
		if end < 0 {
			break
		}
		end += start
		name := safeSQL[start:end]
		if simplePlanName.MatchString(name) {
			identifiers[strings.ToLower(name)] = true
		}
		pos = end + 1
	}
	return identifiers
}

func redactAccessObject(value string) string {
	if value == "" {
		return ""
	}
	parts, ok := splitPlanParts(value)
	if !ok {
		return "?"
	}
	for i, part := range parts {
		parts[i] = redactAccessEntry(part)
	}
	return strings.Join(parts, ", ")
}

func redactAccessEntry(part string) string {
	part = strings.TrimSpace(part)
	if accessEntry.MatchString(part) {
		return part
	}
	return "?"
}

func redactOperatorInfo(value string, identifiers map[string]bool) string {
	if value == "" {
		return ""
	}
	parts, ok := splitPlanParts(value)
	if !ok {
		return "?"
	}
	result := make([]string, 0, len(parts))
	previousRange := false
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lower := strings.ToLower(part)
		switch {
		case strings.HasPrefix(lower, "range:"):
			rangeValue, valid := redactRange(strings.TrimSpace(part[len("range:"):]))
			result = append(result, "range:"+rangeValue)
			previousRange = valid
		case previousRange && (strings.HasPrefix(part, "[") || strings.HasPrefix(part, "(")):
			rangeValue, valid := redactRange(part)
			result = append(result, rangeValue)
			previousRange = valid
		case lower == "keep order:true" || lower == "keep order:false" || lower == "desc" || lower == "asc" ||
			lower == "inner join" || lower == "left outer join" || lower == "right outer join":
			result = append(result, lower)
			previousRange = false
		case strings.HasPrefix(lower, "table:") || strings.HasPrefix(lower, "index:") || strings.HasPrefix(lower, "clustered index:") || strings.HasPrefix(lower, "partition:"):
			result = append(result, redactAccessEntry(part))
			previousRange = false
		case strings.HasPrefix(lower, "data:"):
			operator := strings.TrimSpace(part[len("data:"):])
			if planID.MatchString(operator) {
				result = append(result, "data:"+operator)
			} else {
				result = append(result, "?")
			}
			previousRange = false
		default:
			result = append(result, redactPlanExpression(part, identifiers, 0))
			previousRange = false
		}
	}
	return strings.Join(result, ", ")
}

func redactRange(value string) (string, bool) {
	if len(value) < 3 || (value[0] != '[' && value[0] != '(') || (value[len(value)-1] != ']' && value[len(value)-1] != ')') {
		return "?", false
	}
	ends, ok := splitPlanParts(value[1 : len(value)-1])
	if !ok || len(ends) != 2 || strings.TrimSpace(ends[0]) == "" || strings.TrimSpace(ends[1]) == "" {
		return "?", false
	}
	return value[:1] + "?,?" + value[len(value)-1:], true
}

func redactPlanExpression(value string, identifiers map[string]bool, depth int) string {
	value = strings.TrimSpace(value)
	if depth > 20 {
		return "?"
	}
	if match := functionCall.FindStringSubmatch(value); match != nil {
		name := strings.ToLower(match[1])
		switch name {
		case "eq", "ne", "lt", "le", "gt", "ge", "nulleq", "in", "not", "isnull", "and", "or", "like":
		default:
			return "?"
		}
		args, ok := splitPlanParts(match[2])
		if !ok || len(args) == 0 {
			return "?"
		}
		for i, arg := range args {
			args[i] = redactPlanExpression(arg, identifiers, depth+1)
		}
		return name + "(" + strings.Join(args, ", ") + ")"
	}
	if qualifiedName.MatchString(value) {
		for _, part := range strings.Split(value, ".") {
			if !identifiers[strings.ToLower(part)] {
				return "?"
			}
		}
		return value
	}
	return "?"
}

// Split commas outside expressions, ranges, and quoted strings. TiDB ranges
// can have mixed endpoints such as (1,2], so bracket depth is tracked without
// requiring matching bracket types. Malformed input is never emitted verbatim.
func splitPlanParts(text string) ([]string, bool) {
	var result []string
	start, depth := 0, 0
	var quote byte
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if quote != 0 {
			if ch == '\\' {
				i++
				if i >= len(text) {
					return nil, false
				}
			} else if ch == quote {
				if i+1 < len(text) && text[i+1] == quote {
					i++
				} else {
					quote = 0
				}
			}
			continue
		}
		switch ch {
		case '\'', '"', '`':
			quote = ch
		case '(', '[':
			depth++
		case ')', ']':
			depth--
			if depth < 0 {
				return nil, false
			}
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(text[start:i]))
				start = i + 1
			}
		}
	}
	if quote != 0 || depth != 0 || strings.TrimSpace(text) == "" {
		return nil, false
	}
	return append(result, strings.TrimSpace(text[start:])), true
}
