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
	report := RenderRecommendIndex([]RecommendIndexResult{{Digest: "abcdefgh123", AvgMem: 250, AvgRequestUnitRead: 2.5, NewPlan: line}})
	for _, want := range []string{"# Overview", "Exec Count | AVG_MEM | AVG_REQ_UNIT_READ | Est Improvement | Index Advice", "# Recommendation List", line} {
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
