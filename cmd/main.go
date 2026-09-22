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

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis"
	"go.uber.org/zap"
)

const (
	httpReadTimeout  = 10 * time.Second
	httpWriteTimeout = 60 * time.Second
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	httpAddress := flag.String("http-addr", "", "HTTP listen address in host:port form")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "usage: %s -http-addr <host:port>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *httpAddress == "" || flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	connections := diagnosis.NewConnectionManager(logger)
	defer connections.Close()

	validator := diagnosis.NewValidator(connections)
	handler := diagnosis.NewHandler(validator, logger)
	connectionHandler := diagnosis.NewConnectionHandler(connections, logger)

	mux := http.NewServeMux()
	mux.Handle("POST /validate", handler)
	mux.Handle("POST /test/connect", connectionHandler)
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
