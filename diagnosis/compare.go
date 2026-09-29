package diagnosis

import (
	"context"
	"fmt"
	"html"
	"slices"
	"strings"
)

// StatementPlan is one plan recorded in the cluster statement summary.
type StatementPlan struct {
	Sample
	SQLDigest  string
	PlanDigest string
	ExecCount  uint64
	ExecTime   uint64 // SUM_LATENCY, in nanoseconds.
	PlanHint   string
	Plan       string
}

type CompareRepository interface {
	StatementPlans(context.Context) ([]StatementPlan, error)
	ExplainPlan(context.Context, Sample) (string, error)
}

type planDigestRepository interface {
	ExplainPlanWithDigest(context.Context, Sample) (string, string, error)
}

type Comparer struct {
	repository CompareRepository
}

func NewComparer(repository CompareRepository) *Comparer {
	return &Comparer{repository: repository}
}

type planOperator struct {
	id   string
	task string
}

// planOperators reads TiDB's tab-separated PLAN/EXPLAIN output. The summary's
// PLAN places task before estRows, whereas EXPLAIN places it after estRows.
// Ignore numeric operator ID suffixes; preserve task, branches and row order.
func planOperators(plan string) ([]planOperator, error) {
	var operators []planOperator
	idColumn, taskColumn := -1, -1
	for _, line := range strings.Split(plan, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tabSeparated := strings.ContainsRune(line, '\t')
		fields := strings.Fields(line)
		if tabSeparated {
			fields = strings.Split(line, "\t")
		}
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if idColumn < 0 {
			for i, field := range fields {
				switch strings.ToLower(field) {
				case "id":
					idColumn = i
				case "task":
					taskColumn = i
				}
			}
			if idColumn < 0 || taskColumn < 0 {
				return nil, newStackErrorf("plan header must contain id and task columns")
			}
			continue
		}
		if len(fields) <= max(idColumn, taskColumn) || fields[idColumn] == "" || fields[taskColumn] == "" {
			return nil, newStackErrorf("plan row is missing id or task: %q", line)
		}
		operators = append(operators, planOperator{id: operatorName(fields[idColumn]), task: fields[taskColumn]})
	}
	if len(operators) == 0 {
		return nil, newStackErrorf("plan has no operators")
	}
	return operators, nil
}

// operatorName removes only a trailing underscore followed entirely by digits.
func operatorName(id string) string {
	i := strings.LastIndexByte(id, '_')
	if i < 0 || i == len(id)-1 {
		return id
	}
	for _, c := range id[i+1:] {
		if c < '0' || c > '9' {
			return id
		}
	}
	return id[:i]
}

type comparedPlan struct {
	statement     StatementPlan
	newPlan       string
	newPlanDigest string
	bindings      string
}

func (c *Comparer) Compare(ctx context.Context) (string, error) {
	// Finish collecting the snapshot before EXPLAIN adds more summary entries.
	statements, err := c.repository.StatementPlans(ctx)
	if err != nil {
		return "", err
	}
	statements = aggregateStatementPlans(statements)
	type explanation struct{ plan, digest string }
	explained := make(map[Sample]explanation)
	var changed []comparedPlan
	for _, statement := range statements {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		oldOperators, err := planOperators(statement.Plan)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s, plan digest %s: read current plan: %w", statement.SQLDigest, statement.PlanDigest, err)
		}
		newPlan, ok := explained[statement.Sample]
		if !ok {
			if repository, ok := c.repository.(planDigestRepository); ok {
				newPlan.plan, newPlan.digest, err = repository.ExplainPlanWithDigest(ctx, statement.Sample)
			} else {
				newPlan.plan, err = c.repository.ExplainPlan(ctx, statement.Sample)
				// Repositories that only provide EXPLAIN text are mainly useful
				// for tests and adapters; preserve a visible digest in that case.
				newPlan.digest = statement.PlanDigest
			}
			if err != nil {
				return "", fmt.Errorf("SQL digest %s: %w", statement.SQLDigest, err)
			}
			explained[statement.Sample] = newPlan
		}
		newOperators, err := planOperators(newPlan.plan)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s: read new plan: %w", statement.SQLDigest, err)
		}
		if slices.Equal(oldOperators, newOperators) {
			continue
		}
		binding, err := currentPlanBinding(statement.SQL, statement.PlanHint)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s: build current plan binding: %w", statement.SQLDigest, err)
		}
		changed = append(changed, comparedPlan{
			statement: statement, newPlan: newPlan.plan, newPlanDigest: newPlan.digest, bindings: binding,
		})
	}
	slices.SortFunc(changed, func(a, b comparedPlan) int {
		if a.statement.ExecTime > b.statement.ExecTime {
			return -1
		}
		if a.statement.ExecTime < b.statement.ExecTime {
			return 1
		}
		if order := strings.Compare(a.statement.SQLDigest, b.statement.SQLDigest); order != 0 {
			return order
		}
		return strings.Compare(a.statement.PlanDigest, b.statement.PlanDigest)
	})
	return renderComparison(changed), nil
}

func aggregateStatementPlans(statements []StatementPlan) []StatementPlan {
	type key struct{ schema, digest, planDigest string }
	indices := make(map[key]int)
	result := make([]StatementPlan, 0, len(statements))
	for _, statement := range statements {
		k := key{statement.Schema, statement.SQLDigest, statement.PlanDigest}
		if i, ok := indices[k]; ok {
			result[i].ExecCount += statement.ExecCount
			result[i].ExecTime += statement.ExecTime
		} else {
			indices[k] = len(result)
			result = append(result, statement)
		}
	}
	slices.SortFunc(result, func(a, b StatementPlan) int {
		if order := strings.Compare(a.SQLDigest, b.SQLDigest); order != 0 {
			return order
		}
		if order := strings.Compare(a.Schema, b.Schema); order != 0 {
			return order
		}
		return strings.Compare(a.PlanDigest, b.PlanDigest)
	})
	return result
}

func renderComparison(plans []comparedPlan) string {
	var report strings.Builder
	report.WriteString("# SQL Plan Comparison\n\n")
	report.WriteString("| SQL Digest | ExecCount | ExecTime | Current Plan | New Plan | Plan Change | Binding of the Current Plan |\n")
	report.WriteString("| --- | ---: | ---: | --- | --- | --- | --- |\n")
	anchors := make(map[string]int)
	type sqlDetailKey struct {
		digest string
		schema string
		sql    string
	}
	sqlAnchors := make(map[sqlDetailKey]string)
	detailGroups := make(map[sqlDetailKey]*strings.Builder)
	var detailOrder []sqlDetailKey
	for _, plan := range plans {
		statement := plan.statement
		sqlDigestPrefix := firstEight(statement.SQLDigest)
		sqlKey := sqlDetailKey{digest: statement.SQLDigest, schema: statement.Schema, sql: statement.SQL}
		sqlAnchor, known := sqlAnchors[sqlKey]
		details := detailGroups[sqlKey]
		if !known {
			details = &strings.Builder{}
			detailGroups[sqlKey] = details
			detailOrder = append(detailOrder, sqlKey)
			title := sqlDigestPrefix + "_sql"
			sqlAnchor = uniqueHeadingAnchor(title, anchors)
			sqlAnchors[sqlKey] = sqlAnchor
			fmt.Fprintf(details, "\n<a id=\"%s\"></a>\n\n## %s\n\n", html.EscapeString(sqlAnchor), title)
			fmt.Fprintf(details, "Schema: %s  \nSQL Digest: %s\n\n",
				markdownCell(statement.Schema), markdownCell(statement.SQLDigest))
			writeCodeBlock(details, statement.SQL)
		}

		currentTitle := sqlDigestPrefix + "_" + firstEight(statement.PlanDigest) + "_current_plan"
		newTitle := sqlDigestPrefix + "_" + firstEight(plan.newPlanDigest) + "_new_plan"
		bindingTitle := sqlDigestPrefix + "_" + firstEight(statement.PlanDigest) + "_binding_info"
		currentAnchor := writeDetail(details, currentTitle, statement, statement.PlanDigest, withoutPlanColumns(statement.Plan, "actRows", "execution info", "memory", "disk"), anchors)
		newAnchor := writeDetail(details, newTitle, statement, plan.newPlanDigest, plan.newPlan, anchors)
		bindingText := plan.bindings
		if bindingText == "" {
			bindingText = "No PLAN_HINT available."
		}
		bindingAnchor := writeDetail(details, bindingTitle, statement, statement.PlanDigest, bindingText, anchors)
		fmt.Fprintf(&report, "| [%s](#%s) | %d | %s | [%s](#%s) | [%s](#%s) | N/A | [%s](#%s) |\n",
			markdownCell(sqlDigestPrefix), sqlAnchor, statement.ExecCount, formatExecTime(statement.ExecTime),
			markdownCell(firstEight(statement.PlanDigest)), currentAnchor,
			markdownCell(firstEight(plan.newPlanDigest)), newAnchor,
			markdownCell("bind info"), bindingAnchor)
	}
	if len(plans) == 0 {
		report.WriteString("\nNo plan differences found.\n")
	}
	for _, key := range detailOrder {
		report.WriteString(detailGroups[key].String())
	}
	return report.String()
}

func uniqueHeadingAnchor(title string, anchors map[string]int) string {
	anchor := strings.ToLower(title)
	anchors[anchor]++
	if anchors[anchor] > 1 {
		return fmt.Sprintf("%s-%d", anchor, anchors[anchor])
	}
	return anchor
}

func writeDetail(details *strings.Builder, title string, statement StatementPlan, planDigest, content string, anchors map[string]int) string {
	anchor := uniqueHeadingAnchor(title, anchors)
	fmt.Fprintf(details, "\n<a id=\"%s\"></a>\n\n### %s\n\n", html.EscapeString(anchor), title)
	fmt.Fprintf(details, "Schema: %s  \nSQL Digest: %s  \nPlan Digest: %s\n\n",
		markdownCell(statement.Schema), markdownCell(statement.SQLDigest), markdownCell(planDigest))
	writeCodeBlock(details, content)
	return anchor
}

func withoutPlanColumns(plan string, columnNames ...string) string {
	remove := make(map[int]struct{})
	lines := strings.Split(plan, "\n")
	headerLine := -1
	var header []string
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		for _, field := range fields {
			if strings.EqualFold(strings.TrimSpace(field), "id") {
				headerLine = i
				header = fields
				break
			}
		}
		if headerLine >= 0 {
			break
		}
	}
	if headerLine < 0 {
		return plan
	}
	for i, field := range header {
		for _, columnName := range columnNames {
			if strings.EqualFold(strings.TrimSpace(field), columnName) {
				remove[i] = struct{}{}
			}
		}
	}
	if len(remove) == 0 {
		return plan
	}
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) != len(header) {
			continue
		}
		kept := make([]string, 0, len(fields)-len(remove))
		for j, field := range fields {
			if _, ok := remove[j]; !ok {
				kept = append(kept, field)
			}
		}
		lines[i] = strings.Join(kept, "\t")
	}
	return strings.Join(lines, "\n")
}

func firstEight(value string) string {
	return value[:min(8, len(value))]
}

func preview(value string) string {
	runes := []rune(strings.Join(strings.Fields(value), " "))
	if len(runes) > 16 {
		return string(runes[:15]) + "…"
	}
	return string(runes)
}

func markdownCell(value string) string {
	value = html.EscapeString(value)
	return strings.NewReplacer(
		"\\", "\\\\", "|", "&#124;", "`", "\\`", "*", "\\*", "_", "\\_",
		"[", "\\[", "]", "\\]", "\r\n", "<br>", "\n", "<br>", "\r", "<br>",
	).Replace(value)
}

func writeCodeBlock(out *strings.Builder, text string) {
	fence := "```"
	for strings.Contains(text, fence) {
		fence += "`"
	}
	fmt.Fprintf(out, "%stext\n%s\n%s\n", fence, text, fence)
}
