package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis"
	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/index_advisor"
	"go.uber.org/zap"
)

type jobOptions struct {
	dsn                string
	diagnosticType     string
	output             string
	timeout            time.Duration
	slowQuerySQLColumn string
}

func registerJobFlags(flags *flag.FlagSet) *jobOptions {
	options := &jobOptions{}
	flags.StringVar(&options.diagnosticType, "job-type", "", "diagnosis to run: compare or recommend-index (job mode)")
	flags.StringVar(&options.output, "job-output", "", "report file path; an existing file is replaced (job mode)")
	flags.DurationVar(&options.timeout, "job-timeout", 0, "overall job timeout; 0 disables the deadline (job mode)")
	flags.StringVar(&options.slowQuerySQLColumn, "recommend-index-sql-column", "", "slow-query SQL column: query or in_memory_original_query (job mode)")
	return options
}

type jobDiagnostics interface {
	Compare(context.Context) (string, error)
	RecommendIndex(context.Context) ([]index_advisor.RecommendIndexResult, error)
}

func runJob(ctx context.Context, options jobOptions, logger *zap.Logger) error {
	if options.diagnosticType != "compare" && options.diagnosticType != "recommend-index" {
		return errors.New("job-type must be compare or recommend-index")
	}
	if strings.TrimSpace(options.output) == "" {
		return errors.New("job-output must not be empty")
	}
	if options.timeout < 0 {
		return errors.New("job-timeout must not be negative")
	}
	if !index_advisor.IsSupportedSlowQuerySQLColumn(options.slowQuerySQLColumn) {
		return errors.New("recommend-index-sql-column must be query or in_memory_original_query")
	}
	if strings.TrimSpace(options.dsn) == "" {
		return errors.New("dsn must not be empty")
	}
	if options.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.timeout)
		defer cancel()
	}

	connections := diagnosis.NewConnectionManager(logger)
	defer connections.Close()
	result := connections.Replace(ctx, options.dsn)
	if result.Err != nil {
		return fmt.Errorf("connect to diagnostic TiDB: %w", result.Err)
	}
	if err := executeJob(ctx, options, connections); err != nil {
		return err
	}
	logger.Info("diagnostic job finished", zap.String("job", options.diagnosticType))
	return nil
}

func executeJob(ctx context.Context, options jobOptions, diagnostics jobDiagnostics) error {
	var report string
	var err error
	switch options.diagnosticType {
	case "compare":
		report, err = diagnostics.Compare(ctx)
	case "recommend-index":
		var results []index_advisor.RecommendIndexResult
		if options.slowQuerySQLColumn != "" {
			if configurable, ok := diagnostics.(interface {
				RecommendIndexWithSQLColumn(context.Context, string) ([]index_advisor.RecommendIndexResult, error)
			}); ok {
				results, err = configurable.RecommendIndexWithSQLColumn(ctx, options.slowQuerySQLColumn)
			} else {
				results, err = diagnostics.RecommendIndex(ctx)
			}
		} else {
			results, err = diagnostics.RecommendIndex(ctx)
		}
		if err == nil {
			report = index_advisor.RenderRecommendIndex(results)
		}
	default:
		return errors.New("job-type must be compare or recommend-index")
	}
	if err != nil {
		return fmt.Errorf("run %s diagnosis: %w", options.diagnosticType, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := writeReport(options.output, report); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

// writeReport replaces path with report. The destination is replaced only after
// the new contents are fully written, so a write failure leaves an existing file unchanged.
func writeReport(path, report string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".report-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_, writeErr := tmp.WriteString(report)
	closeErr := tmp.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
