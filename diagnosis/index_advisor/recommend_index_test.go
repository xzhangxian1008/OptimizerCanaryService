package index_advisor

import (
	"strings"
	"testing"
)

func TestAggregateWeightedMetrics(t *testing.T) {
	rows := []recommendIndexSample{
		{Sample: Sample{SQL: "select 1"}, Digest: "a", ExecCount: 2, ExecTime: 10, AvgMem: 100, AvgRequestUnitRead: 1},
		{Sample: Sample{SQL: "select 1"}, Digest: "a", ExecCount: 6, ExecTime: 20, AvgMem: 300, AvgRequestUnitRead: 3},
		{Sample: Sample{SQL: "select 2"}, Digest: "b", ExecCount: 1, ExecTime: 100},
	}
	got := aggregateAndSortSamples(rows)
	if len(got) != 2 || got[0].Digest != "b" {
		t.Fatalf("unexpected order: %+v", got)
	}
	a := got[1]
	if a.ExecCount != 8 || a.ExecTime != 30 || a.AvgMem != 250 || a.AvgRequestUnitRead != 2.5 {
		t.Fatalf("unexpected aggregation: %+v", a)
	}
}

func TestPlanIndentationAndOverview(t *testing.T) {
	line := "  └─IndexRangeScan_14"
	if got := planCellString([]byte(line)); got != line {
		t.Fatalf("indentation lost: %q", got)
	}
	report := RenderRecommendIndex([]RecommendIndexResult{{Digest: "abcdefgh123", AvgMem: 250, AvgRequestUnitRead: 2.5, Recommendations: []IndexRecommendation{{newPlan: line}}}})
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
	for _, want := range []string{"123.45 -&gt; 10.00", "#### Create Index Statement", "CREATE INDEX idx ON t(a);"} {
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
	for _, want := range []string{"#index-advice-abcdef123-1", "#index-advice-abcdef123-2", "CREATE INDEX first", "CREATE INDEX second"} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if strings.Contains(report, "Schema:") {
		t.Fatal("schema line retained")
	}
}
