package compare

import (
	"strings"
	"testing"
)

func TestCurrentPlanBinding(t *testing.T) {
	for _, tc := range []struct{ name, sql, want string }{
		{"simple", "select * from t", "select /*+ H() */ * from t"},
		{"cte", "WITH c AS (SELECT * FROM t), d AS (SELECT * FROM c) SELECT * FROM d", "WITH c AS (SELECT * FROM t), d AS (SELECT * FROM c) SELECT /*+ H() */ * FROM d"},
		{"union", "select 1 union all select 2", "select /*+ H() */ 1 union all select 2"},
		{"wrapped", "(select 1) union (select 2)", "(select /*+ H() */ 1) union (select 2)"},
		{"comments", "/* select */ -- SELECT\n# select\nSELECT 'select', 'A  B'", "/* select */ -- SELECT\n# select\nSELECT /*+ H() */ 'select', 'A  B'"},
		{"quoted cte", "WITH `select` AS (SELECT 'a select') SELECT * FROM `select`", "WITH `select` AS (SELECT 'a select') SELECT /*+ H() */ * FROM `select`"},
		{"existing hint", "SELECT /*+ OLD() */ * FROM t", "SELECT /*+ H() */ * FROM t"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getCurrentPlanBinding(tc.sql, "H()")
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
	if got, err := getCurrentPlanBinding("SELECT 1", ""); err != nil || got != "" {
		t.Fatalf("empty hint: %q %v", got, err)
	}
}

func TestComparisonReportColumnsAndHierarchy(t *testing.T) {
	sample := Sample{Schema: "test", SQL: "select 1"}
	report := renderComparison([]comparedPlan{
		{statement: SQLInfo{Sample: sample, SQLDigest: "aaaaaaaa1", PlanDigest: "11111111", ExecCount: 9, ExecTime: 1234567890}, bindings: "select /*+ H() */ 1"},
		{statement: SQLInfo{Sample: Sample{SQL: "select 2"}, SQLDigest: "bbbbbbbb2", PlanDigest: "22222222", ExecCount: 8}},
		{statement: SQLInfo{Sample: sample, SQLDigest: "aaaaaaaa1", PlanDigest: "33333333", ExecCount: 7}},
	}, reportSummary{checkedSQLs: 2, totalExecCount: 24, changedSQLs: 2})
	for _, text := range []string{
		"## Summary\n\n- SQLs Checked: 2\n- Total ExecCount: 24\n- SQLs with Plan Changes: 2",
		"| Total ExecTime | ExecCount | Current Plan | New Plan | Plan Change | Binding of the Current Plan |",
		"| 1.23s | 9 |",
		"[binding stmt](#binding-stmt-aaaaaaaa-11111111)",
		"### Current Plan: 11111111",
		"### Binding Stmt: aaaaaaaa_33333333",
		"CREATE GLOBAL BINDING FOR",
	} {
		if !strings.Contains(report, text) {
			t.Fatalf("missing %q in report", text)
		}
	}
	if strings.Contains(report, "_old_plan") || strings.Contains(report, "_current_plan") {
		t.Fatal("legacy plan title remains")
	}
	if strings.Index(report, "### Current Plan: 33333333") > strings.Index(report, "## SQL: bbbbbbbb") {
		t.Fatal("plan detail placed under the wrong SQL heading")
	}
}

func TestComparisonReportRedactsSQLBindingAndPlans(t *testing.T) {
	sql := "SELECT * FROM orders WHERE id = 975318 AND email = 'secret@example.com'"
	plan := "id\ttask\toperator info\nTableReader_1\troot\trange:[975318,975318], eq(email, 'secret@example.com')"
	report := renderComparison([]comparedPlan{{
		statement: SQLInfo{Sample: Sample{Schema: "app", SQL: sql}, SQLDigest: "abcdef123", PlanDigest: "11111111", Plan: plan},
		bindings:  "SELECT /*+ USE_INDEX(orders, idx) */ * FROM orders WHERE id = 975318 AND email = 'secret@example.com'",
		newPlan:   plan,
	}}, reportSummary{})
	for _, secret := range []string{"975318", "secret@example.com"} {
		if strings.Contains(report, secret) {
			t.Fatalf("report contains secret %q: %s", secret, report)
		}
	}
	for _, want := range []string{"select * from `orders`", "CREATE GLOBAL BINDING", "use_index", "TableReader_1", "range:[?,?]"} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing %q: %s", want, report)
		}
	}
}
