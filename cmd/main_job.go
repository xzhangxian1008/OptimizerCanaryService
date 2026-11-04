package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis"
	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/index_advisor"
	"go.uber.org/zap"
)

type jobOptions struct {
	diagnosticType string
	output         string
	timeout        time.Duration
}

func registerJobFlags(flags *flag.FlagSet) *jobOptions {
	options := &jobOptions{}
	flags.StringVar(&options.diagnosticType, "job-type", "", "diagnosis to run: compare or recommend-index (job mode)")
	flags.StringVar(&options.output, "job-output", "", "new report file path (job mode)")
	flags.DurationVar(&options.timeout, "job-timeout", 0, "overall job timeout; 0 disables the deadline (job mode)")
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
	dsn := os.Getenv("DIAGNOSTIC_TIDB_DSN")
	if strings.TrimSpace(dsn) == "" {
		return errors.New("DIAGNOSTIC_TIDB_DSN must not be empty")
	}
	if options.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.timeout)
		defer cancel()
	}

	connections := diagnosis.NewConnectionManager(logger)
	defer connections.Close()
	result := connections.Replace(ctx, dsn)
	if result.Err != nil {
		return fmt.Errorf("connect to diagnostic TiDB: %w", result.Err)
	}
	return executeJob(ctx, options, connections)
}

func executeJob(ctx context.Context, options jobOptions, diagnostics jobDiagnostics) error {
	var report string
	var err error
	switch options.diagnosticType {
	case "compare":
		report, err = diagnostics.Compare(ctx)
	case "recommend-index":
		var results []index_advisor.RecommendIndexResult
		results, err = diagnostics.RecommendIndex(ctx)
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

	// Do not overwrite an existing report or leave a partial file on failure.
	file, err := os.OpenFile(options.output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	_, writeErr := file.WriteString(report)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(options.output)
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}
