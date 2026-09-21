package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis"
	"go.uber.org/zap"
)

const (
	dbPingTimeout    = 5 * time.Second
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 60 * time.Second
)

type fileConfig struct {
	TiDB struct {
		DSN string `toml:"dsn"`
	} `toml:"tidb"`
}

func resolveTiDBDSN(commandLineDSN, configPath string) (string, error) {
	if commandLineDSN != "" && configPath != "" {
		return "", errors.New("-dsn and -config cannot be used together")
	}
	if configPath == "" {
		if commandLineDSN == "" {
			return "", errors.New("either -dsn or -config is required")
		}
		return commandLineDSN, nil
	}

	var config fileConfig
	metadata, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return "", fmt.Errorf("read config %q: %w", configPath, err)
	}
	if unknown := metadata.Undecoded(); len(unknown) != 0 {
		return "", fmt.Errorf("unknown config key %q", unknown[0])
	}
	if config.TiDB.DSN == "" {
		return "", errors.New("config must contain a non-empty [tidb] dsn")
	}
	return config.TiDB.DSN, nil
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	tidbDSN := flag.String("dsn", "", "Diagnostic TiDB MySQL DSN")
	configPath := flag.String("config", "", "TOML config file containing the Diagnostic TiDB DSN")
	httpAddress := flag.String("http-addr", "", "HTTP listen address in host:port form")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "usage: %s (-dsn <tidb-dsn> | -config <file.toml>) -http-addr <host:port>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *httpAddress == "" || flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	resolvedDSN, err := resolveTiDBDSN(*tidbDSN, *configPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}

	db, err := diagnosis.OpenDB(resolvedDSN)
	if err != nil {
		logger.Error("open TiDB connection", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer startupCancel()
	if err := db.PingContext(startupCtx); err != nil {
		logger.Error("connect to Diagnostic TiDB", zap.Error(err))
		os.Exit(1)
	}

	repository := diagnosis.NewSQLRepository(db, logger)
	validator := diagnosis.NewValidator(repository)
	handler := diagnosis.NewHandler(validator, logger)

	mux := http.NewServeMux()
	mux.Handle("POST /validate", handler)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
	})

	server := &http.Server{
		Addr:              *httpAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("shut down HTTP server", zap.Error(err))
		}
	}()

	logger.Info("Diagnostic Service started", zap.String("address", *httpAddress))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("serve HTTP", zap.Error(err))
		os.Exit(1)
	}
}
