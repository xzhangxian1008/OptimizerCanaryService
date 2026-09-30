package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strconv"
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

type PlanNode struct {
	Operator    string
	ID          int
	Task        string
	Index       string
	accessTable string
	Children    []*PlanNode
}

func (n *PlanNode) String() string {
	value, _ := json.Marshal(struct {
		Operator string `json:"operator"`
		Task     string `json:"task"`
		Index    string `json:"index"`
		Table    string `json:"accessTable"`
	}{n.Operator, n.Task, n.Index, n.accessTable})
	return string(value)
}

type Plan struct {
	Root        *PlanNode
	PostOrder   []*PlanNode
	reader      []*PlanNode
	indexReader []*PlanNode
}

func (p *Plan) StringInPostOrder() string {
	parts := make([]string, 0, len(p.PostOrder))
	for _, node := range p.PostOrder {
		parts = append(parts, node.String())
	}
	return strings.Join(parts, "")
}

var operatorIDPattern = regexp.MustCompile(`^(.*)_([0-9]+)(?:\([^)]*\))?$`)

// TODO I think this function needs more and more tests
func parsePlan(planText string) (*Plan, error) {
	idColumn, taskColumn, accessColumn, infoColumn := -1, -1, -1, -1
	type planRow struct {
		node  *PlanNode
		depth int
	}
	rows := make([]planRow, 0)
	baseDepth := -1
	for _, line := range strings.Split(planText, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		depth, content := planDepth(line)
		fields := strings.Fields(content)
		if strings.ContainsRune(line, '\t') {
			fields = strings.Split(content, "\t")
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
				case "access object":
					accessColumn = i
				case "operator info":
					infoColumn = i
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
		id, operator, err := parseOperatorID(fields[idColumn])
		if err != nil {
			return nil, err
		}
		node := &PlanNode{Operator: operator, ID: id, Task: normalizeTask(fields[taskColumn])}
		if accessColumn >= 0 && accessColumn < len(fields) {
			node.Index = planIndex(fields[accessColumn])
		}
		if infoColumn >= 0 && infoColumn < len(fields) && node.Index == "" {
			node.Index = planIndex(fields[infoColumn])
		}
		if accessColumn >= 0 && accessColumn < len(fields) {
			node.accessTable = planAccessTable(fields[accessColumn])
		}
		if infoColumn >= 0 && infoColumn < len(fields) && node.accessTable == "" {
			node.accessTable = planAccessTable(fields[infoColumn])
		}
		// Some EXPLAIN renderers prefix the root row with └─. Use the first
		// row's depth as the baseline so that form and the usual unprefixed
		// root form produce the same tree.
		if baseDepth < 0 {
			baseDepth = depth
		}
		depth -= baseDepth
		rows = append(rows, planRow{node, depth})
	}
	if len(rows) == 0 {
		return nil, util.NewStackErrorf("plan has no operators")
	}
	result := &Plan{}
	stack := make([]*PlanNode, 0)
	for _, row := range rows {
		if isReaderOperator(row.node.Operator) {
			result.reader = append(result.reader, row.node)
		}
		if isIndexReaderOperator(row.node.Operator) {
			result.indexReader = append(result.indexReader, row.node)
		}
		if row.depth == 0 {
			if result.Root != nil {
				return nil, util.NewStackErrorf("plan has multiple roots")
			}
			result.Root = row.node
			stack = []*PlanNode{row.node}
			continue
		}
		if row.depth > len(stack) {
			return nil, util.NewStackErrorf("plan has invalid tree depth %d", row.depth)
		}
		stack = stack[:row.depth]
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, row.node)
		stack = append(stack, row.node)
	}
	postOrder(result.Root, &result.PostOrder)
	return result, nil
}

// isReaderOperator identifies physical operators that read rows from a
// storage engine or an index. Operators such as Selection and Projection are
// deliberately excluded because they only transform rows produced by one of
// these readers.
func isReaderOperator(operator string) bool {
	return operator == "TableReader" ||
		operator == "TableFullScan" ||
		operator == "TableScan" ||
		operator == "PointGet" ||
		operator == "BatchPointGet" ||
		operator == "Point_Get" ||
		operator == "Batch_Point_Get"
}

func isIndexReaderOperator(operator string) bool {
	return operator == "IndexReader" ||
		operator == "IndexLookUpReader" ||
		operator == "IndexMergeReader" ||
		strings.HasSuffix(operator, "IndexFullScan") ||
		strings.HasSuffix(operator, "IndexScan") ||
		operator == "IndexLookUp" ||
		operator == "IndexLookup"
}

func planDepth(line string) (int, string) {
	depth := 0
	// PLAN values read from cluster_statements_summary may contain one
	// formatting tab before every physical row. It is not a plan column and
	// must be removed before looking for the tree branch markers.
	for strings.HasPrefix(line, "\t") {
		line = line[1:]
	}
	for {
		if strings.HasPrefix(line, "│ ") {
			depth++
			line = line[len("│ "):]
			continue
		}
		if strings.HasPrefix(line, "  ") {
			depth++
			line = line[len("  "):]
			continue
		}
		break
	}
	if strings.HasPrefix(line, "├─") {
		line = line[len("├─"):]
		depth++
	} else if strings.HasPrefix(line, "└─") {
		line = line[len("└─"):]
		depth++
	}
	return depth, line
}

func parseOperatorID(value string) (int, string, error) {
	value = strings.TrimSpace(value)
	match := operatorIDPattern.FindStringSubmatch(value)
	if match == nil {
		// A real TiDB operator normally has a numeric suffix. Keeping a
		// suffix-less value usable makes the parser robust for explain
		// adapters and tests that use symbolic operator IDs; IDs are ignored
		// when plans are compared.
		return 0, value, nil
	}
	id, err := strconv.Atoi(match[2])
	if err != nil {
		return 0, "", util.NewStackErrorf("parse operator id %q: %w", value, err)
	}
	return id, match[1], nil
}

func normalizeTask(task string) string {
	task = strings.TrimSpace(task)
	for _, prefix := range []string{"cop[", "mpp["} {
		if strings.HasPrefix(task, prefix) && strings.HasSuffix(task, "]") {
			return task[len(prefix) : len(task)-1]
		}
	}
	return task
}

func planIndex(value string) string {
	position := strings.Index(strings.ToLower(value), "index:")
	if position < 0 {
		return ""
	}
	value = strings.TrimSpace(value[position+len("index:"):])
	if comma := strings.IndexByte(value, ','); comma >= 0 {
		value = value[:comma]
	}
	return strings.TrimSpace(value)
}

func planAccessTable(value string) string {
	position := strings.Index(strings.ToLower(value), "table:")
	if position < 0 {
		return ""
	}
	value = strings.TrimSpace(value[position+len("table:"):])
	if comma := strings.IndexByte(value, ','); comma >= 0 {
		value = value[:comma]
	}
	return strings.TrimSpace(value)
}

func postOrder(node *PlanNode, result *[]*PlanNode) {
	for _, child := range node.Children {
		postOrder(child, result)
	}
	*result = append(*result, node)
}

func plansEqual(left, right *Plan) bool {
	equal, _ := comparePlans(left, right)
	return equal
}

// comparePlans deliberately checks the reader collections before the rest of
// the tree. A storage-engine change is the most useful explanation to show in
// the report, so later comparisons are skipped once it is found.
func comparePlans(left, right *Plan) (bool, string) {
	if !equalReaderNodes(left.reader, right.reader) {
		return false, readerChangeReason(left.reader, right.reader, left.indexReader, right.indexReader)
	}
	if !equalReaderNodes(left.indexReader, right.indexReader) {
		return false, readerChangeReason(left.indexReader, right.indexReader, nil, nil)
	}
	if !planNodesEqual(left.Root, right.Root) {
		return false, "others"
	}
	return true, ""
}

func planReadersEqual(left, right *Plan) bool {
	return equalReaderNodes(left.reader, right.reader) && equalReaderNodes(left.indexReader, right.indexReader)
}

func equalReaderNodes(left, right []*PlanNode) bool {
	leftStrings := make([]string, 0, len(left))
	for _, node := range left {
		leftStrings = append(leftStrings, node.String())
	}
	rightStrings := make([]string, 0, len(right))
	for _, node := range right {
		rightStrings = append(rightStrings, node.String())
	}
	slices.Sort(leftStrings)
	slices.Sort(rightStrings)
	return slices.Equal(leftStrings, rightStrings)
}

func readerChangeReason(left, right, leftOther, rightOther []*PlanNode) string {
	leftNodes := append(append([]*PlanNode(nil), left...), leftOther...)
	rightNodes := append(append([]*PlanNode(nil), right...), rightOther...)
	if len(leftNodes) != len(rightNodes) {
		return "others"
	}
	sortPlanNodes(leftNodes)
	sortPlanNodes(rightNodes)
	for i := range leftNodes {
		oldNode, newNode := leftNodes[i], rightNodes[i]
		if oldNode.Operator != newNode.Operator {
			return strings.ToLower(oldNode.Operator) + "->" + strings.ToLower(newNode.Operator)
		}
		if oldNode.Task != newNode.Task {
			return strings.ToLower(oldNode.Task) + "->" + strings.ToLower(newNode.Task)
		}
	}
	return "others"
}

func sortPlanNodes(nodes []*PlanNode) {
	slices.SortFunc(nodes, func(left, right *PlanNode) int {
		return strings.Compare(left.String(), right.String())
	})
}

func planNodesEqual(left, right *PlanNode) bool {
	if left == nil || right == nil {
		return left == right
	}
	if left.Operator != right.Operator || left.Task != right.Task || left.Index != right.Index || left.accessTable != right.accessTable || len(left.Children) != len(right.Children) {
		return false
	}
	for i := range left.Children {
		if !planNodesEqual(left.Children[i], right.Children[i]) {
			return false
		}
	}
	return true
}

type comparedPlan struct {
	statement     SQLInfo
	newPlan       string
	newPlanDigest string
	bindings      string
	planChange    string
}

type reportSummary struct {
	checkedSQLs    int
	totalExecCount uint64
	changedSQLs    int
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

	type sqlKey struct{ schema, digest string }
	checkedSQLs := make(map[sqlKey]struct{})
	changedSQLs := make(map[sqlKey]struct{})

	var totalExecCount uint64
	var changed []comparedPlan

	for _, sqlInfo := range sqlInfos {
		totalExecCount += sqlInfo.ExecCount

		key := sqlKey{schema: sqlInfo.Schema, digest: sqlInfo.SQLDigest}
		checkedSQLs[key] = struct{}{}

		if err := ctx.Err(); err != nil {
			return "", err
		}

		oldPlan, err := parsePlan(sqlInfo.Plan)
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

		newPlanParsed, err := parsePlan(newPlan.plan)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s: read new plan: %w", sqlInfo.SQLDigest, err)
		}

		equal, planChange := comparePlans(oldPlan, newPlanParsed)
		if equal {
			continue
		}

		binding, err := currentPlanBinding(sqlInfo.SQL, sqlInfo.PlanHint)
		if err != nil {
			return "", fmt.Errorf("SQL digest %s: build current plan binding: %w", sqlInfo.SQLDigest, err)
		}

		changed = append(changed, comparedPlan{
			statement: sqlInfo, newPlan: newPlan.plan, newPlanDigest: newPlan.digest,
			bindings: binding, planChange: planChange,
		})
		changedSQLs[key] = struct{}{}
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

	return renderComparison(changed, reportSummary{
		checkedSQLs:    len(checkedSQLs),
		totalExecCount: totalExecCount,
		changedSQLs:    len(changedSQLs),
	}), nil
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

func renderComparison(plans []comparedPlan, summary reportSummary) string {
	var report strings.Builder
	report.WriteString("# SQL Plan Comparison\n\n")
	report.WriteString("## Summary\n\n")
	fmt.Fprintf(&report, "- SQLs Checked: %d\n", summary.checkedSQLs)
	fmt.Fprintf(&report, "- Total ExecCount: %d\n", summary.totalExecCount)
	fmt.Fprintf(&report, "- SQLs with Plan Changes: %d\n\n", summary.changedSQLs)
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
		fmt.Fprintf(&report, "| [%s](#%s) | %s | %d | [%s](#%s) | [%s](#%s) | %s | [%s](#%s) |\n",
			markdownCell(sqlDigestPrefix), sqlAnchor, formatExecTime(statement.ExecTime), statement.ExecCount,
			markdownCell(firstEight(statement.PlanDigest)), currentAnchor,
			markdownCell(firstEight(plan.newPlanDigest)), newAnchor,
			markdownCell(plan.planChange),
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
