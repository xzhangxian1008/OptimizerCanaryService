package index_advisor

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func TestAggregateWeightedMetrics(t *testing.T) {
	rows := []recommendIndexSample{
		{Sample: Sample{Schema: "app", SQL: "select 1"}, Digest: "a", ExecCount: 2, ExecTime: 10, AvgMem: 100, AvgRequestUnitRead: 1, LastSeen: "2026-01-01 01:00:00"},
		{Sample: Sample{Schema: "app", SQL: "select 3"}, Digest: "a", ExecCount: 6, ExecTime: 20, AvgMem: 300, AvgRequestUnitRead: 3, LastSeen: "2026-01-01 02:00:00"},
		{Sample: Sample{SQL: "select 2"}, Digest: "b", ExecCount: 1, ExecTime: 100},
		{Sample: Sample{Schema: "other", SQL: "select 1"}, Digest: "a", ExecCount: 1, ExecTime: 1},
	}
	got := aggregateAndSortSamples(rows)
	if len(got) != 3 || got[0].Digest != "b" {
		t.Fatalf("unexpected order: %+v", got)
	}
	a := got[1]
	if a.ExecCount != 8 || a.ExecTime != 30 || a.AvgMem != 250 || a.AvgRequestUnitRead != 2.5 || a.SQL != "select 3" {
		t.Fatalf("unexpected aggregation: %+v", a)
	}
}

func TestHasOnlyUserTables(t *testing.T) {
	for _, tc := range []struct {
		tables string
		want   bool
	}{
		{"app.orders,other.customers", true},
		{"app.orders,INFORMATION_SCHEMA.tables", false},
		{"mysql.user", false},
		{"performance_schema.events_statements_current", false},
		{"metrics_schema.tidb_query", false},
		{"inspection_schema.cluster_info", false},
		{"sys.metrics", false},
		{"", false},
		{"table_without_schema", false},
	} {
		if got := hasOnlyUserTables(tc.tables); got != tc.want {
			t.Errorf("hasOnlyUserTables(%q) = %v, want %v", tc.tables, got, tc.want)
		}
	}
}

func TestMergeSummaryWindowsUsesCompleteHistory(t *testing.T) {
	current := []recommendIndexSample{
		{Instance: "tidb-1", WindowStart: "10:00", ExecCount: 2},
		{Instance: "tidb-2", WindowStart: "10:00", ExecCount: 3},
	}
	history := []recommendIndexSample{
		{Instance: "tidb-1", WindowStart: "10:00", ExecCount: 4},
		{Instance: "tidb-1", WindowStart: "10:00", ExecCount: 5},
		{Instance: "tidb-1", WindowStart: "09:30", ExecCount: 6},
	}
	got := mergeSummaryWindows(current, history)
	if len(got) != 4 || got[0].ExecCount != 4 || got[1].ExecCount != 5 || got[2].ExecCount != 6 || got[3].ExecCount != 3 {
		t.Fatalf("unexpected windows: %+v", got)
	}
}

type fakeRecommendationRunner struct {
	calls    []string
	failures map[string]error
}

func (r *fakeRecommendationRunner) recommendFor(_ context.Context, sample Sample) ([]IndexRecommendation, error) {
	r.calls = append(r.calls, sample.SQL)
	if err := r.failures[sample.SQL]; err != nil {
		return nil, err
	}
	if len(r.calls) <= 2 {
		return nil, nil
	}
	return []IndexRecommendation{{}}, nil
}

func TestRecommendSamplesSkipsMissingObjects(t *testing.T) {
	runner := &fakeRecommendationRunner{failures: map[string]error{
		"missing": fmt.Errorf("select schema: %w", &mysql.MySQLError{Number: 1049}),
	}}
	samples := []recommendIndexSample{
		{Sample: Sample{SQL: "missing"}},
		{Sample: Sample{SQL: "no advice"}},
		{Sample: Sample{SQL: "valid"}},
	}
	results, err := recommendSamples(context.Background(), samples, runner, zap.NewNop())
	if err != nil || len(results) != 1 || results[0].SQL != "valid" {
		t.Fatalf("results=%+v calls=%v err=%v", results, runner.calls, err)
	}
}

func (*fakeRecommendationRunner) getSQLPlan(context.Context, Sample, string) (string, error) {
	return "plan", nil
}

func TestRecommendSamplesStopsAfterTwentySQLsWithAdvice(t *testing.T) {
	var samples []recommendIndexSample
	for i := 0; i < 30; i++ {
		samples = append(samples, recommendIndexSample{Sample: Sample{SQL: fmt.Sprintf("SELECT %d", i)}})
	}
	runner := &fakeRecommendationRunner{}
	results, err := recommendSamples(context.Background(), samples, runner, zap.NewNop())
	if err != nil || len(results) != 20 || len(runner.calls) != 22 || runner.calls[21] != "SELECT 21" {
		t.Fatalf("results=%d calls=%v err=%v", len(results), runner.calls, err)
	}
}

func TestPlanIndentationAndOverview(t *testing.T) {
	line := "  └─IndexRangeScan_14"
	if got := planCellString([]byte(line)); got != line {
		t.Fatalf("indentation lost: %q", got)
	}
	report := RenderRecommendIndex([]RecommendIndexResult{{Digest: "abcdefgh123", AvgMem: 250, AvgRequestUnitRead: 2.5, Recommendations: []IndexRecommendation{{newPlan: "id\ttask\n" + line + "\troot"}}}})
	for _, want := range []string{"# Index Recommendation Overview", "<th>Count</th><th>Database</th><th>Table</th><th>Est Cost Change</th>", "# Recommendation List", line} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestAlignPlanColumns(t *testing.T) {
	input := "id\testRows\ttask\taccess object\nProjection_7\t1.00\troot\t\n  └─IndexRangeScan_14\t100.00\tcop[tikv]\ttable:t"
	got := alignPlanColumns(input)
	want := "id                     estRows  task       access object\nProjection_7           1.00     root       \n  └─IndexRangeScan_14  100.00   cop[tikv]  table:t"
	if got != want {
		t.Fatalf("unexpected alignment:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(got, "\t") {
		t.Fatal("plan still contains tabs")
	}
}

func TestPlanBlockPreservesTextAndEscapesHTML(t *testing.T) {
	var out strings.Builder
	writePlanBlock(&out, "  └─Selection\nlt(a, 3) <script>&")
	got := out.String()
	for _, want := range []string{"white-space: pre", "overflow-x: auto", "  └─Selection\n", "&lt;script&gt;&amp;"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
}

func TestReportCostChangeAndCreateStatement(t *testing.T) {
	report := RenderRecommendIndex([]RecommendIndexResult{{Digest: "abc", OldPlan: "id\testCost\nRoot_1\t123.45\n  Child_2\t999", Recommendations: []IndexRecommendation{{newPlan: "id\testCost\nRoot_1\t10.00", createStatement: "CREATE INDEX idx ON t(a);"}}}})
	for _, want := range []string{"123.45 -&gt; 10.00", "#### Create Index Statement", "create index `idx` on `t` ( `a` )"} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing %q", want)
		}
	}
	for _, removed := range []string{"AVG_MEM", "AVG_REQ_UNIT_READ", "### Index Recommendation", "Top Impacted Queries", "#### Recommendation"} {
		if strings.Contains(report, removed) {
			t.Fatalf("unexpected %q", removed)
		}
	}
	if topPlanCost("id\ttask\nRoot\troot") != "N/A" {
		t.Fatal("expected N/A for missing cost")
	}
}

func TestMultipleAdviceRowsAndSections(t *testing.T) {
	low, high := 0.1, 0.9
	report := RenderRecommendIndex([]RecommendIndexResult{{Digest: "abcdef123", Schema: "test", ExecCount: 7, Recommendations: []IndexRecommendation{
		{Database: "test", Table: "first", improvement: &low, createStatement: "CREATE INDEX first ON t(a);"},
		{Database: "test", Table: "second", improvement: &high, createStatement: "CREATE INDEX second ON t(b);"},
	}}})
	if strings.Count(report, "rowspan=\"2\"") != 3 || strings.Count(report, "### Advice\n") != 2 {
		t.Fatalf("missing merged cells or advice sections: %s", report)
	}
	if strings.Index(report, "<td>first</td>") > strings.Index(report, "<td>second</td>") {
		t.Fatal("advice order changed")
	}
	for _, want := range []string{"#index-advice-abcdef123-74657374-1", "#index-advice-abcdef123-74657374-2", "create index `first`", "create index `second`"} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if strings.Contains(report, "Schema:") {
		t.Fatal("schema line retained")
	}
}

func TestReportAnchorsDistinguishSchemas(t *testing.T) {
	results := []RecommendIndexResult{
		{Schema: "db1", Digest: "abcd", Recommendations: []IndexRecommendation{{}}},
		{Schema: "db2", Digest: "abcd", Recommendations: []IndexRecommendation{{}}},
	}
	report := RenderRecommendIndex(results)
	for _, want := range []string{"href=\"#sql-abcd-646231\"", "id=\"sql-abcd-646231\"", "href=\"#sql-abcd-646232\"", "id=\"sql-abcd-646232\""} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing anchor %q", want)
		}
	}
}

func TestReportRedactsSQLAndPlans(t *testing.T) {
	plan := "id\testRows\testCost\ttask\taccess object\toperator info\n" +
		"IndexRangeScan_7\t1.00\t2.00\tcop[tikv]\ttable:orders\trange:[975318,975318], eq(email, 'secret@example.com')"
	result := RecommendIndexResult{
		Schema:  "app",
		Digest:  "abcdef123",
		SQL:     "select * from orders where id=975318 and email='secret@example.com' [arguments: 864209]",
		OldPlan: plan,
		Recommendations: []IndexRecommendation{{
			Database: "app", Table: "orders", createStatement: "CREATE INDEX idx ON orders(id)",
			newSQL:  "select /*+ USE_INDEX(orders,idx) */ * from orders where id=975318 and email='secret@example.com'",
			newPlan: plan,
		}},
	}
	report := RenderRecommendIndex([]RecommendIndexResult{result})
	for _, secret := range []string{"975318", "864209", "secret@example.com"} {
		if strings.Contains(report, secret) {
			t.Fatalf("report contains secret %q: %s", secret, report)
		}
	}
	for _, want := range []string{"select * from `orders`", "range:[?,?]", "table:orders", "IndexRangeScan_7", "2.00 -&gt; 2.00"} {
		if !strings.Contains(report, want) {
			t.Fatalf("report missing %q: %s", want, report)
		}
	}
}

func TestReportOmitsMalformedSQLAndPlan(t *testing.T) {
	result := RecommendIndexResult{
		Digest: "abcdef", SQL: "select * from t where x='private-secret", OldPlan: "private-plan",
		Recommendations: []IndexRecommendation{{newSQL: "select 'private-hint", newPlan: "private-plan", createStatement: "private-ddl"}},
	}
	report := RenderRecommendIndex([]RecommendIndexResult{result})
	if strings.Contains(report, "private-") || strings.Count(report, "[content omitted: unable to redact]") < 5 {
		t.Fatalf("report did not fail closed: %s", report)
	}
}
