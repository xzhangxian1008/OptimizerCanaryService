package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis"
)

const (
	dbPingTimeout    = 5 * time.Second
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 60 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	tidbDSN := flag.String("dsn", "", "Diagnostic TiDB MySQL DSN")
	httpAddress := flag.String("http-addr", "", "HTTP listen address in host:port form")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "usage: %s -dsn <tidb-dsn> -http-addr <host:port>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *tidbDSN == "" || *httpAddress == "" || flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}

	db, err := diagnosis.OpenDB(*tidbDSN)
	if err != nil {
		logger.Error("open TiDB connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer startupCancel()
	if err := db.PingContext(startupCtx); err != nil {
		logger.Error("connect to Diagnostic TiDB", "error", err)
		os.Exit(1)
	}

	repository := diagnosis.NewSQLRepository(db)
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
			logger.Error("shut down HTTP server", "error", err)
		}
	}()

	logger.Info("Diagnostic Service started", "address", *httpAddress)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("serve HTTP", "error", err)
		os.Exit(1)
	}
}
