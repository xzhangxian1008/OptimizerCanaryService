package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/index_advisor"
	"go.uber.org/zap"
)

type fakeJobDiagnostics struct {
	called string
	err    error
}

func TestRunJobRequiresDSNArgument(t *testing.T) {
	t.Setenv("DIAGNOSTIC_TIDB_DSN", "root@tcp(127.0.0.1:4000)/")
	err := runJob(context.Background(), jobOptions{
		diagnosticType: "compare",
		output:         filepath.Join(t.TempDir(), "report.md"),
	}, zap.NewNop())
	if err == nil || err.Error() != "dsn must not be empty" {
		t.Fatalf("expected missing DSN error, got %v", err)
	}
}

func (f *fakeJobDiagnostics) Compare(context.Context) (string, error) {
	f.called = "compare"
	return "# Compare\n", f.err
}

func (f *fakeJobDiagnostics) RecommendIndex(context.Context) ([]index_advisor.RecommendIndexResult, error) {
	f.called = "recommend-index"
	return nil, f.err
}

func TestExecuteJob(t *testing.T) {
	for _, tc := range []struct {
		kind string
		want string
	}{
		{"compare", "# Compare"},
		{"recommend-index", "# Index Recommendation Overview"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "report.md")
			diagnostics := &fakeJobDiagnostics{}
			if err := executeJob(context.Background(), jobOptions{diagnosticType: tc.kind, output: path}, diagnostics); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(data), tc.want) || diagnostics.called != tc.kind {
				t.Fatalf("called %q, report %q, error %v", diagnostics.called, data, err)
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("report mode: %v, error: %v", info, err)
			}
		})
	}
}

func TestExecuteJobFailureLeavesReportAndSuccessReplacesIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.md")
	options := jobOptions{diagnosticType: "compare", output: path}
	diagnostics := &fakeJobDiagnostics{err: errors.New("diagnosis failed")}
	if err := executeJob(context.Background(), options, diagnostics); err == nil {
		t.Fatal("expected diagnosis failure")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("report exists after failure: %v", err)
	}
	if err := os.WriteFile(path, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	diagnostics.err = nil
	if err := executeJob(context.Background(), options, diagnostics); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "# Compare\n" {
		t.Fatalf("report %q, error: %v", data, err)
	}
}

func TestExecuteJobCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	path := filepath.Join(t.TempDir(), "report.md")
	err := executeJob(ctx, jobOptions{diagnosticType: "compare", output: path}, &fakeJobDiagnostics{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("report exists after cancellation: %v", err)
	}
}
