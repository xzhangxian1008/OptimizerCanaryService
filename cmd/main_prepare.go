package main

import (
	"context"
	"flag"
	"time"

	"github.com/xzhangxian1008/OptimizerCanaryService/prepare"
	"go.uber.org/zap"
)

type prepareOptions struct {
	dsn         string
	schema      string
	concurrency int
	duration    time.Duration
}

func registerPrepareFlags(flags *flag.FlagSet) *prepareOptions {
	options := &prepareOptions{}
	flags.StringVar(&options.dsn, "dsn", "", "TiDB DSN (prepare mode)")
	flags.StringVar(&options.schema, "prepare-schema", prepare.DefaultSchema, "schema for the compact TPC-C fixture")
	flags.IntVar(&options.concurrency, "prepare-concurrency", prepare.DefaultConcurrency, "number of concurrent prepare workload workers")
	flags.DurationVar(&options.duration, "prepare-duration", prepare.DefaultDuration, "duration of the concurrent prepare workload")
	return options
}

func runPrepare(ctx context.Context, options prepareOptions, logger *zap.Logger) error {
	return prepare.Run(ctx, prepare.Config{
		DSN:         options.dsn,
		Schema:      options.schema,
		Concurrency: options.concurrency,
		Duration:    options.duration,
	}, logger)
}
