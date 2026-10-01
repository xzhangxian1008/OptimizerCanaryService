package compare

import (
	"context"
	"strings"
	"testing"
)

type fakeCompareRepository struct {
	statements []SQLInfo
	newPlans   map[Sample]string
	bindings   map[string]string
}

type digestCountingRepository struct {
	fakeCompareRepository
	digestCalls int
}

func (r *digestCountingRepository) GetNewPlanDigest(context.Context, Sample, string) (string, error) {
	r.digestCalls++
	return "newdigest", nil
}

func (f *fakeCompareRepository) GetSQLInfo(context.Context) ([]SQLInfo, error) {
	return f.statements, nil
}

func (f *fakeCompareRepository) ExplainPlan(_ context.Context, sample Sample) (string, error) {
	return f.newPlans[sample], nil
}

func (f *fakeCompareRepository) Bindings(_ context.Context, digest, _ string) (string, error) {
	return f.bindings[digest], nil
}

func TestComparerOnlyReportsChangedIDAndTask(t *testing.T) {
	same := Sample{Schema: "app", SQL: "SELECT * FROM t WHERE id = 1"}
	changed := Sample{Schema: "app", SQL: "SELECT * FROM t WHERE id = 2"}
	oldPlan := "id\ttask\testRows\nA\troot\t1"
	newPlanWithDifferentColumns := "id\testRows\ttask\nA\t99\troot"
	repo := &fakeCompareRepository{
		statements: []SQLInfo{
			{Sample: same, SQLDigest: "1111111111111111", PlanDigest: "aaaaaaaaaaaaaaaa", ExecCount: 2, Plan: oldPlan},
			{Sample: changed, SQLDigest: "2222222222222222", PlanDigest: "bbbbbbbbbbbbbbbb", ExecCount: 3, ExecTime: 1000000000, PlanHint: "HASH_AGG()", Plan: oldPlan},
			{Sample: changed, SQLDigest: "2222222222222222", PlanDigest: "bbbbbbbbbbbbbbbb", ExecCount: 4, ExecTime: 2000000000, PlanHint: "HASH_AGG()", Plan: oldPlan},
		},
		newPlans: map[Sample]string{
			same:    newPlanWithDifferentColumns,
			changed: "id\testRows\ttask\nB\t1\troot",
		},
		bindings: map[string]string{"2222222222222222": "CREATE GLOBAL BINDING ..."},
	}
	report, err := NewComparer(repo).Compare(context.Background())
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if !strings.Contains(report, "## Summary\n\n- SQLs Checked: 2\n- Total ExecCount: 9\n- SQLs with Plan Changes: 1") {
		t.Fatalf("unexpected report summary: %s", report)
	}
	if !strings.Contains(report, "SELECT /*+ HASH_AGG() */ * FROM t WHERE id = 2") {
		t.Fatal("missing current plan hinted SQL")
	}
	if strings.Contains(report, "11111111_aaaaaaaa") {
		t.Fatalf("unchanged plan was included: %s", report)
	}
	if !strings.Contains(report, "| [22222222](#sql-22222222) | 3.00s | 7 | [bbbbbbbb](#current-plan-22222222-bbbbbbbb) | [bbbbbbbb](#new-plan-22222222-bbbbbbbb) | others |") {
		t.Fatalf("changed plan or aggregated execution count missing: %s", report)
	}
	if !strings.Contains(report, "## SQL: 22222222") || strings.Contains(report, "## 22222222SELECT * FROM t WHERE id = 2") {
		t.Fatalf("SQL detail heading missing: %s", report)
	}
	if !strings.Contains(report, "### Current Plan: bbbbbbbb") ||
		!strings.Contains(report, "### New Plan: bbbbbbbb") ||
		!strings.Contains(report, "### Binding Stmt: 22222222_bbbbbbbb") ||
		!strings.Contains(report, "CREATE GLOBAL BINDING FOR SELECT * FROM t WHERE id = 2 USING") {
		t.Fatalf("detail headings missing: %s", report)
	}
}

func TestPlanOperatorsSupportsWhitespacePlans(t *testing.T) {
	operators, err := parsePlan("id task estRows\n└─TableReader_1 root 1")
	if err != nil {
		t.Fatalf("parsePlan returned error: %v", err)
	}
	if operators == nil || operators.Root == nil || operators.Root.Operator != "TableReader" || operators.Root.Task != "root" {
		t.Fatalf("unexpected operators: %+v", operators)
	}
}

func TestParsePlanHandlesLeadingFormattingTabs(t *testing.T) {
	plan, err := parsePlan("\tid\ttask\n" +
		"\tHashAgg_1\troot\n" +
		"\t└─TableReader_2\troot\n" +
		"\t  └─TableFullScan_3\tcop[tikv]")
	if err != nil {
		t.Fatalf("parsePlan returned error: %v", err)
	}
	if plan.Root.Operator != "HashAgg" || len(plan.Root.Children) != 1 || plan.Root.Children[0].Operator != "TableReader" {
		t.Fatalf("unexpected plan tree: %+v", plan.Root)
	}
}

func TestParsePlanCollectsReaderOperators(t *testing.T) {
	plan, err := parsePlan("id\ttask\taccess object\n" +
		"HashJoin_1\troot\t\n" +
		"├─TableReader_2\troot\t\n" +
		"│ └─TableFullScan_3\tcop[tikv]\ttable:t\n" +
		"└─IndexReader_4\troot\t\n")
	if err != nil {
		t.Fatalf("parsePlan returned error: %v", err)
	}
	if len(plan.reader) != 2 {
		t.Fatalf("got %d reader operators, want 2", len(plan.reader))
	}
	for i, want := range []string{"TableReader", "TableFullScan"} {
		if plan.reader[i].Operator != want {
			t.Fatalf("reader[%d] = %q, want %q", i, plan.reader[i].Operator, want)
		}
	}
	if len(plan.indexReader) != 1 || plan.indexReader[0].Operator != "IndexReader" {
		t.Fatalf("unexpected index readers: %+v", plan.indexReader)
	}
	if got := plan.reader[1].accessTable; got != "t" {
		t.Fatalf("access table = %q, want %q", got, "t")
	}
}

func TestPlanReadersEqualSortsNodeStringsAndIgnoresIDs(t *testing.T) {
	left := &Plan{
		reader: []*PlanNode{
			{Operator: "TableReader", ID: 1, Task: "root", accessTable: "a"},
			{Operator: "TableFullScan", ID: 2, Task: "tikv", accessTable: "b"},
		},
		indexReader: []*PlanNode{{Operator: "IndexReader", ID: 3, Task: "root", Index: "idx", accessTable: "a"}},
	}
	right := &Plan{
		reader: []*PlanNode{
			{Operator: "TableFullScan", ID: 20, Task: "tikv", accessTable: "b"},
			{Operator: "TableReader", ID: 10, Task: "root", accessTable: "a"},
		},
		indexReader: []*PlanNode{{Operator: "IndexReader", ID: 30, Task: "root", Index: "idx", accessTable: "a"}},
	}
	if !planReadersEqual(left, right) {
		t.Fatal("reader and indexReader collections with the same nodes should compare equal")
	}
	right.indexReader[0].accessTable = "b"
	if planReadersEqual(left, right) {
		t.Fatal("different access tables should make reader collections different")
	}
}

func TestComparePlansReportsReaderChangeReasonBeforeNormalComparison(t *testing.T) {
	oldPlan := &Plan{reader: []*PlanNode{{Operator: "TableReader", Task: "tiflash"}}}
	newPlan := &Plan{reader: []*PlanNode{{Operator: "TableReader", Task: "tikv"}}}
	equal, reason := comparePlans(oldPlan, newPlan)
	if equal || reason != "tiflash->tikv" {
		t.Fatalf("got equal=%v reason=%q, want false and tiflash->tikv", equal, reason)
	}

	oldPlan = &Plan{reader: []*PlanNode{{Operator: "TableFullScan", Task: "tikv"}}}
	newPlan = &Plan{indexReader: []*PlanNode{{Operator: "IndexFullScan", Task: "tikv"}}}
	equal, reason = comparePlans(oldPlan, newPlan)
	if equal || reason != "tablefullscan->indexfullscan" {
		t.Fatalf("got equal=%v reason=%q, want false and tablefullscan->indexfullscan", equal, reason)
	}

	oldPlan = &Plan{indexReader: []*PlanNode{{Operator: "IndexReader", Task: "tikv"}}}
	newPlan = &Plan{indexReader: []*PlanNode{{Operator: "IndexReader", Task: "tiflash"}}}
	equal, reason = comparePlans(oldPlan, newPlan)
	if equal || reason != "tikv->tiflash" {
		t.Fatalf("got equal=%v reason=%q, want false and tikv->tiflash", equal, reason)
	}
}

func TestWithoutPlanColumns(t *testing.T) {
	plan := "id\ttask\testRows\tactRows\texecution info\tmemory\n" +
		"TableReader_1\troot\t1\t2\ttime:1ms\t1 KB"
	want := "id\ttask\testRows\tmemory\n" +
		"TableReader_1\troot\t1\t1 KB"
	if got := withoutPlanColumns(plan, "actRows", "execution info"); got != want {
		t.Fatalf("withoutPlanColumns() = %q, want %q", got, want)
	}
}

func TestNormalizeExplainSQLOnlyConvertsDoubleQuotedStringValues(t *testing.T) {
	got := normalizeExplainSQL(`EXPLAIN SELECT "column" FROM t WHERE name = "O'Reilly" AND note = 'a "quote"'`)
	want := `explain select "column" from t where name = 'O'Reilly' and note = 'a "quote"'`
	if got != want {
		t.Fatalf("normalizeExplainSQL() = %q, want %q", got, want)
	}
}

func TestComparerIgnoresOperatorNumbersAndSortsByExecTime(t *testing.T) {
	current := "id\ttask\nHashAgg_11\troot\n└─TableReader_12\troot\n  └─HashAgg_5\tcop[tikv]\n    └─TableFullScan_10\tcop[tikv]"
	same := "id\testRows\ttask\nHashAgg_13\t1\troot\n└─TableReader_14\t1\troot\n  └─HashAgg_6\t1\tcop[tikv]\n    └─TableFullScan_12\t1\tcop[tikv]"
	a, b, c := Sample{SQL: "select 1"}, Sample{SQL: "select 2"}, Sample{SQL: "select 3"}
	repo := &fakeCompareRepository{
		statements: []SQLInfo{
			{Sample: a, SQLDigest: "d8061f40", PlanDigest: "3880073a", Plan: current, ExecCount: 100, ExecTime: 100},
			{Sample: b, SQLDigest: "bbbbbbbb", PlanDigest: "bbbbbbbb", Plan: current, ExecCount: 50, ExecTime: 10},
			{Sample: c, SQLDigest: "cccccccc", PlanDigest: "cccccccc", Plan: current, ExecCount: 1, ExecTime: 20},
		},
		newPlans: map[Sample]string{a: same, b: strings.ReplaceAll(same, "cop[tikv]", "cop[tiflash]"), c: strings.ReplaceAll(same, "HashAgg", "StreamAgg")},
	}
	report, err := NewComparer(repo).Compare(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(report, "d8061f40") {
		t.Fatal("operator numbering alone must not produce a changed plan")
	}
	bi, ci := strings.Index(report, "| [bbbbbbbb]"), strings.Index(report, "| [cccccccc]")
	if bi < 0 || ci < 0 || ci > bi {
		t.Fatal("must report task/name changes ordered by ExecTime, not ExecCount")
	}
}

func TestComparerGetsNewDigestOnlyForChangedPlans(t *testing.T) {
	same, changed := Sample{SQL: "select 1"}, Sample{SQL: "select 2"}
	current := "id\ttask\nTableReader_1\troot"
	repo := &digestCountingRepository{fakeCompareRepository: fakeCompareRepository{
		statements: []SQLInfo{
			{Sample: same, SQLDigest: "same", PlanDigest: "old-same", Plan: current},
			{Sample: changed, SQLDigest: "changed", PlanDigest: "old-changed", Plan: current},
		},
		newPlans: map[Sample]string{
			same:    "id\ttask\nTableReader_2\troot",
			changed: "id\ttask\nIndexReader_2\troot",
		},
	}}
	if _, err := NewComparer(repo).Compare(context.Background()); err != nil {
		t.Fatal(err)
	}
	if repo.digestCalls != 1 {
		t.Fatalf("GetNewPlanDigest called %d times, want 1", repo.digestCalls)
	}
}
