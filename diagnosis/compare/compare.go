package compare

import (
	"context"
	"fmt"
	"html"
	"slices"
	"strings"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/util"
)

// SQLInfo contains sql info recorded in the `information_schema.cluster_statement_summary` table.
// However, it contains only part of the info in one row.
type SQLInfo struct {
	Sample
	SQLDigest  string
	PlanDigest string
	ExecCount  uint64
	ExecTime   uint64 // SUM_LATENCY, in nanoseconds.
	PlanHint   string
	Plan       string
}

type CompareRepository interface {
	GetSQLInfo(context.Context) ([]SQLInfo, error)
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
				return nil, util.NewStackErrorf("plan header must contain id and task columns")
			}
			continue
		}
		if len(fields) <= max(idColumn, taskColumn) || fields[idColumn] == "" || fields[taskColumn] == "" {
			return nil, util.NewStackErrorf("plan row is missing id or task: %q", line)
		}
		operators = append(operators, planOperator{id: operatorName(fields[idColumn]), task: fields[taskColumn]})
	}
	if len(operators) == 0 {
		return nil, util.NewStackErrorf("plan has no operators")
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
	statement     SQLInfo
	newPlan       string
	newPlanDigest string
	bindings      string
}

func (c *Comparer) Compare(ctx context.Context) (string, error) {
	// Finish collecting the snapshot before EXPLAIN adds more summary entries.
	sqlInfos, err := c.repository.GetSQLInfo(ctx)
	if err != nil {
		return "", err
	}
	sqlInfos = aggregateSqlInfos(sqlInfos)
	type explanation struct{ plan, digest string }
	explained := make(map[Sample]explanation)
	var changed []comparedPlan
	for _, sqlInfo := range sqlInfos {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		oldOperators, err := planOperators(sqlInfo.Plan)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s, plan digest %s: read current plan: %w", sqlInfo.SQLDigest, sqlInfo.PlanDigest, err)
		}
		newPlan, ok := explained[sqlInfo.Sample]
		if !ok {
			if repository, ok := c.repository.(planDigestRepository); ok {
				newPlan.plan, newPlan.digest, err = repository.ExplainPlanWithDigest(ctx, sqlInfo.Sample)
			} else {
				newPlan.plan, err = c.repository.ExplainPlan(ctx, sqlInfo.Sample)
				// Repositories that only provide EXPLAIN text are mainly useful
				// for tests and adapters; preserve a visible digest in that case.
				newPlan.digest = sqlInfo.PlanDigest
			}
			if err != nil {
				return "", fmt.Errorf("SQL digest %s: %w", sqlInfo.SQLDigest, err)
			}
			explained[sqlInfo.Sample] = newPlan
		}
		newOperators, err := planOperators(newPlan.plan)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s: read new plan: %w", sqlInfo.SQLDigest, err)
		}
		if slices.Equal(oldOperators, newOperators) {
			continue
		}
		binding, err := currentPlanBinding(sqlInfo.SQL, sqlInfo.PlanHint)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s: build current plan binding: %w", sqlInfo.SQLDigest, err)
		}
		changed = append(changed, comparedPlan{
			statement: sqlInfo, newPlan: newPlan.plan, newPlanDigest: newPlan.digest, bindings: binding,
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

func aggregateSqlInfos(sqlInfos []SQLInfo) []SQLInfo {
	type key struct{ schema, digest, planDigest string }
	indices := make(map[key]int)
	result := make([]SQLInfo, 0, len(sqlInfos))
	for _, sqlInfo := range sqlInfos {
		k := key{sqlInfo.Schema, sqlInfo.SQLDigest, sqlInfo.PlanDigest}
		if i, ok := indices[k]; ok {
			result[i].ExecCount += sqlInfo.ExecCount
			result[i].ExecTime += sqlInfo.ExecTime
		} else {
			indices[k] = len(result)
			result = append(result, sqlInfo)
		}
	}
	slices.SortFunc(result, func(a, b SQLInfo) int {
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
	report.WriteString("| SQL Digest | Total ExecTime | ExecCount | Current Plan | New Plan | Plan Change | Binding of the Current Plan |\n")
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
			title := "SQL: " + sqlDigestPrefix
			sqlAnchor = uniqueHeadingAnchor("sql-"+sqlDigestPrefix, anchors)
			sqlAnchors[sqlKey] = sqlAnchor
			fmt.Fprintf(details, "\n<a id=\"%s\"></a>\n\n## %s\n\n", html.EscapeString(sqlAnchor), title)
			fmt.Fprintf(details, "Schema: %s  \nSQL Digest: %s\n\n",
				markdownCell(statement.Schema), markdownCell(statement.SQLDigest))
			writeCodeBlock(details, statement.SQL)
		}

		currentPlanPrefix := firstEight(statement.PlanDigest)
		newPlanPrefix := firstEight(plan.newPlanDigest)
		currentTitle := "Current Plan: " + currentPlanPrefix
		newTitle := "New Plan: " + newPlanPrefix
		bindingTitle := "Binding Stmt: " + sqlDigestPrefix + "_" + currentPlanPrefix
		currentAnchor := uniqueHeadingAnchor("current-plan-"+sqlDigestPrefix+"-"+currentPlanPrefix, anchors)
		newAnchor := uniqueHeadingAnchor("new-plan-"+sqlDigestPrefix+"-"+newPlanPrefix, anchors)
		bindingAnchor := uniqueHeadingAnchor("binding-stmt-"+sqlDigestPrefix+"-"+currentPlanPrefix, anchors)
		currentAnchor = writeDetailWithAnchor(details, currentAnchor, currentTitle, statement, statement.PlanDigest, withoutPlanColumns(statement.Plan, "actRows", "execution info", "memory", "disk"))
		newAnchor = writeDetailWithAnchor(details, newAnchor, newTitle, statement, plan.newPlanDigest, plan.newPlan)
		bindingText := plan.bindings
		if bindingText == "" {
			bindingText = "No PLAN_HINT available."
		} else {
			bindingText = bindingStatement(statement.SQL, bindingText)
		}
		bindingAnchor = writeDetailWithAnchor(details, bindingAnchor, bindingTitle, statement, statement.PlanDigest, bindingText)
		fmt.Fprintf(&report, "| [%s](#%s) | %s | %d | [%s](#%s) | [%s](#%s) | N/A | [%s](#%s) |\n",
			markdownCell(sqlDigestPrefix), sqlAnchor, formatExecTime(statement.ExecTime), statement.ExecCount,
			markdownCell(firstEight(statement.PlanDigest)), currentAnchor,
			markdownCell(firstEight(plan.newPlanDigest)), newAnchor,
			markdownCell("binding stmt"), bindingAnchor)
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

func writeDetail(details *strings.Builder, title string, statement SQLInfo, planDigest, content string, anchors map[string]int) string {
	anchor := uniqueHeadingAnchor(title, anchors)
	return writeDetailWithAnchor(details, anchor, title, statement, planDigest, content)
}

func writeDetailWithAnchor(details *strings.Builder, anchor, title string, statement SQLInfo, planDigest, content string) string {
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
