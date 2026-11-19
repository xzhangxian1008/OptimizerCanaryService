package util

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// SQLStringLiteral encodes text independently of the session's quoting mode.
func SQLStringLiteral(value string) string {
	return "_utf8mb4 X'" + hex.EncodeToString([]byte(value)) + "'"
}

func SQLValueLiteral(value any) (string, error) {
	switch value := value.(type) {
	case nil:
		return "NULL", nil
	case string:
		return SQLStringLiteral(value), nil
	case []byte:
		if value == nil {
			return "NULL", nil
		}
		return "X'" + hex.EncodeToString(value) + "'", nil
	case int:
		return strconv.Itoa(value), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case uint64:
		return strconv.FormatUint(value, 10), nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return "", fmt.Errorf("cannot encode non-finite SQL number %v", value)
		}
		literal := strconv.FormatFloat(value, 'g', -1, 64)
		if !strings.Contains(literal, "e") {
			literal += "e0"
		}
		return literal, nil
	default:
		return "", fmt.Errorf("unsupported SQL value type %T", value)
	}
}
