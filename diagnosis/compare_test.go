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
			{Sample: changed, SQLDigest: "2222222222222222", PlanDigest: "bbbbbbbbbbbbbbbb", ExecCount: 3, Plan: oldPlan},
			{Sample: changed, SQLDigest: "2222222222222222", PlanDigest: "bbbbbbbbbbbbbbbb", ExecCount: 4, Plan: oldPlan},
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
	if strings.Contains(report, "11111111_aaaaaaaa") {
		t.Fatalf("unchanged plan was included: %s", report)
	}
	if !strings.Contains(report, "| [22222222](#22222222_sql) | 7 | [bbbbbbbb](#22222222_bbbbbbbb_old_plan) | [bbbbbbbb](#22222222_bbbbbbbb_new_plan) |") {
		t.Fatalf("changed plan or aggregated execution count missing: %s", report)
	}
	if !strings.Contains(report, "## 22222222SELECT * FROM t WHERE id = 2") {
		t.Fatalf("SQL detail heading missing: %s", report)
	}
	if !strings.Contains(report, "## 22222222_bbbbbbbb_old_plan") ||
		!strings.Contains(report, "## 22222222_bbbbbbbb_new_plan") ||
		!strings.Contains(report, "## 22222222_bbbbbbbb_binding_info") {
		t.Fatalf("detail headings missing: %s", report)
	}
}

func TestPlanOperatorsSupportsWhitespacePlans(t *testing.T) {
	operators, err := planOperators("id task estRows\n└─TableReader_1 root 1")
	if err != nil {
		t.Fatalf("planOperators returned error: %v", err)
	}
	if len(operators) != 1 || operators[0].id != "└─TableReader_1" || operators[0].task != "root" {
		t.Fatalf("unexpected operators: %+v", operators)
	}
}
