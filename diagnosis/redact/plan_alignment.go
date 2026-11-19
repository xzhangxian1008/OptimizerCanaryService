package redact

import (
	"strings"
	"unicode"

	"golang.org/x/text/width"
)

func alignPlanColumns(plan string) string {
	lines := strings.Split(plan, "\n")
	rows := make([][]string, len(lines))
	widths := make([]int, strings.Count(lines[0], "\t")+1)
	for i, line := range lines {
		rows[i] = strings.Split(line, "\t")
		for j, cell := range rows[i] {
			// Preserve the leading indentation of the operator tree.
			cell = strings.TrimRight(cell, " ")
			rows[i][j] = cell
			widths[j] = max(widths[j], planDisplayWidth(cell))
		}
	}

	var out strings.Builder
	for i, row := range rows {
		if i > 0 {
			out.WriteByte('\n')
		}
		for j, cell := range row {
			out.WriteString(cell)
			if j < len(row)-1 {
				out.WriteString(strings.Repeat(" ", widths[j]-planDisplayWidth(cell)+2))
			}
		}
	}
	return out.String()
}

func planDisplayWidth(text string) int {
	columns := 0
	for _, char := range text {
		if unicode.Is(unicode.Mn, char) || unicode.Is(unicode.Me, char) || unicode.Is(unicode.Cf, char) {
			continue
		}
		switch width.LookupRune(char).Kind() {
		case width.EastAsianFullwidth, width.EastAsianWide:
			columns += 2
		default:
			columns++
		}
	}
	return columns
}
