package prepare

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		config Config
		ok     bool
	}{
		{name: "valid", config: Config{DSN: "root@tcp(127.0.0.1:4000)/", Schema: DefaultSchema, Concurrency: 4, Duration: time.Minute}, ok: true},
		{name: "missing DSN", config: Config{Schema: DefaultSchema, Concurrency: 4, Duration: time.Minute}},
		{name: "missing schema", config: Config{DSN: "root@tcp(127.0.0.1:4000)/", Concurrency: 4, Duration: time.Minute}},
		{name: "invalid concurrency", config: Config{DSN: "root@tcp(127.0.0.1:4000)/", Schema: DefaultSchema, Duration: time.Minute}},
		{name: "invalid duration", config: Config{DSN: "root@tcp(127.0.0.1:4000)/", Schema: DefaultSchema, Concurrency: 4}},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.validate()
			if test.ok && err != nil {
				t.Fatalf("validate() returned error: %v", err)
			}
			if !test.ok && err == nil {
				t.Fatal("validate() returned nil")
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	if got, want := quoteIdentifier("a`b"), "`a``b`"; got != want {
		t.Fatalf("quoteIdentifier() = %q, want %q", got, want)
	}
}

func TestTPCCWorkloadsUseSessionBindingMarkers(t *testing.T) {
	if got, want := len(tpccWorkloads), 36; got != want {
		t.Fatalf("TPC-C workload count = %d, want %d", got, want)
	}
	markers := make(map[string]struct{}, len(tpccWorkloads))
	weights := make(map[int]struct{}, len(tpccWorkloads))
	for _, workload := range tpccWorkloads {
		if workload.name == "" || workload.marker == "" || workload.sql == "" || workload.boundSQL == "" || workload.bindingHint == "" || workload.weight == 0 {
			t.Fatalf("incomplete workload: %+v", workload)
		}
		if workload.weight < 1 || workload.weight > 100 {
			t.Fatalf("workload %s weight = %d, want 1..100", workload.name, workload.weight)
		}
		if _, exists := weights[workload.weight]; exists {
			t.Fatalf("duplicate workload weight %d", workload.weight)
		}
		weights[workload.weight] = struct{}{}
		if !strings.Contains(workload.sql, workload.marker) || !strings.Contains(workload.boundSQL, workload.marker) {
			t.Fatalf("workload %s is missing its statement-summary marker", workload.name)
		}
		switch workload.bindingHint {
		case tpccBindingHintUseIndex:
			if !strings.Contains(workload.boundSQL, "USE_INDEX(") {
				t.Fatalf("workload %s is missing its USE_INDEX hint", workload.name)
			}
		case tpccBindingHintTiFlash:
			if !strings.Contains(workload.boundSQL, "READ_FROM_STORAGE(TIFLASH[") {
				t.Fatalf("workload %s is missing its TiFlash hint", workload.name)
			}
		default:
			t.Fatalf("workload %s has unsupported binding hint %q", workload.name, workload.bindingHint)
		}
		if _, exists := markers[workload.marker]; exists {
			t.Fatalf("duplicate workload marker %q", workload.marker)
		}
		markers[workload.marker] = struct{}{}
	}
}

func TestTPCCWorkloadBindingHintGeneration(t *testing.T) {
	const statement = `SELECT c.c_id, d.d_name
	FROM customer AS c
	JOIN district AS d ON d.d_w_id = c.c_w_id AND d.d_id = c.c_d_id`

	useIndex := newTPCCWorkloadWithHint("use_index_test", statement, tpccBindingHintUseIndex, "c", "d")
	if !strings.Contains(useIndex.boundSQL, "/*+ USE_INDEX(c) USE_INDEX(d) */") {
		t.Fatalf("unexpected USE_INDEX binding SQL: %s", useIndex.boundSQL)
	}

	tiFlash := newTPCCWorkloadWithHint("tiflash_test", statement, tpccBindingHintTiFlash, "c", "d")
	if !strings.Contains(tiFlash.boundSQL, "/*+ READ_FROM_STORAGE(TIFLASH[c, d]) */") {
		t.Fatalf("unexpected TiFlash binding SQL: %s", tiFlash.boundSQL)
	}
}

func TestWeightedTPCCPicker(t *testing.T) {
	workloads := []tpccWorkload{
		{name: "one", weight: 1},
		{name: "two", weight: 2},
		{name: "three", weight: 3},
	}
	picker, err := newWeightedTPCCPicker(workloads)
	if err != nil {
		t.Fatalf("newWeightedTPCCPicker() returned error: %v", err)
	}
	if got, want := picker.totalWeight, 6; got != want {
		t.Fatalf("total weight = %d, want %d", got, want)
	}
	for draw, want := range []int{0, 1, 1, 2, 2, 2} {
		if got := picker.pick(draw); got != want {
			t.Fatalf("pick(%d) = %d, want %d", draw, got, want)
		}
	}
}

func TestTPCCDefinesNineCoreTables(t *testing.T) {
	if got, want := len(tpccTableDDLs), 9; got != want {
		t.Fatalf("TPC-C table count = %d, want %d", got, want)
	}
}

func TestPendingTiFlashTables(t *testing.T) {
	statuses := make(map[string]tiFlashReplicaStatus, len(tpccTableDDLs))
	for _, table := range tpccTableDDLs {
		statuses[table.name] = tiFlashReplicaStatus{available: true, progress: 1}
	}
	statuses["customer"] = tiFlashReplicaStatus{progress: 0.75}
	delete(statuses, "history")

	got := pendingTiFlashTables(statuses)
	want := []string{"customer", "history"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("pendingTiFlashTables() = %v, want %v", got, want)
	}
}

func TestTPCCFixtureScale(t *testing.T) {
	if got, want := tpccWarehouseCount, 100; got != want {
		t.Fatalf("TPC-C warehouse count = %d, want %d", got, want)
	}
	if got, want := tpccFixtureRowCount(), 832100; got != want {
		t.Fatalf("TPC-C fixture row count = %d, want %d", got, want)
	}
}

func TestNormalizedOperatorID(t *testing.T) {
	for input, want := range map[string]string{
		"TableReader_12":  "TableReader",
		"Op_12_extra":     "Op_12_extra",
		"TableFullScan_3": "TableFullScan",
	} {
		if got := normalizedOperatorID(input); got != want {
			t.Fatalf("normalizedOperatorID(%q) = %q, want %q", input, got, want)
		}
	}
}
