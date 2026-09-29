package diagnosis

import (
	"context"
	"strings"
	"testing"
)

type fakeCompareRepository struct {
	statements []StatementPlan
	newPlans   map[Sample]string
	bindings   map[string]string
}

func (f *fakeCompareRepository) StatementPlans(context.Context) ([]StatementPlan, error) {
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
		statements: []StatementPlan{
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
	if !strings.Contains(report, "SELECT /*+ HASH_AGG() */ * FROM t WHERE id = 2") {
		t.Fatal("missing current plan hinted SQL")
	}
	if strings.Contains(report, "11111111_aaaaaaaa") {
		t.Fatalf("unchanged plan was included: %s", report)
	}
	if !strings.Contains(report, "| [22222222](#22222222_sql) | 7 | 3.000000000s | [bbbbbbbb](#22222222_bbbbbbbb_current_plan) | [bbbbbbbb](#22222222_bbbbbbbb_new_plan) | N/A |") {
		t.Fatalf("changed plan or aggregated execution count missing: %s", report)
	}
	if !strings.Contains(report, "## 22222222_sql") || strings.Contains(report, "## 22222222SELECT * FROM t WHERE id = 2") {
		t.Fatalf("SQL detail heading missing: %s", report)
	}
	if !strings.Contains(report, "### 22222222_bbbbbbbb_current_plan") ||
		!strings.Contains(report, "### 22222222_bbbbbbbb_new_plan") ||
		!strings.Contains(report, "### 22222222_bbbbbbbb_binding_info") {
		t.Fatalf("detail headings missing: %s", report)
	}
}

func TestPlanOperatorsSupportsWhitespacePlans(t *testing.T) {
	operators, err := planOperators("id task estRows\n└─TableReader_1 root 1")
	if err != nil {
		t.Fatalf("planOperators returned error: %v", err)
	}
	if len(operators) != 1 || operators[0].id != "└─TableReader" || operators[0].task != "root" {
		t.Fatalf("unexpected operators: %+v", operators)
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
		statements: []StatementPlan{
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

func TestOperatorNamePreservesNonNumericSuffix(t *testing.T) {
	for input, want := range map[string]string{
		"HashAgg_12": "HashAgg", "└─TableReader_14": "└─TableReader",
		"Op_12_extra": "Op_12_extra", "Op_": "Op_", "Op": "Op",
	} {
		if got := operatorName(input); got != want {
			t.Fatalf("%q: got %q, want %q", input, got, want)
		}
	}
}
