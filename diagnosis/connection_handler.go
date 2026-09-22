package diagnosis

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const testEndpointWarning = "TEST ONLY: this endpoint will be removed in a future release"

type connectionReplacer interface {
	Replace(context.Context, string) ConnectionStatus
}

type ConnectionHandler struct {
	connections connectionReplacer
	logger      *zap.Logger
}

func NewConnectionHandler(connections connectionReplacer, logger *zap.Logger) *ConnectionHandler {
	return &ConnectionHandler{connections: connections, logger: logger}
}

type connectionRequest struct {
	DSN string `json:"dsn"`
}

type connectionResponse struct {
	Status          string `json:"status"`
	Message         string `json:"message"`
	Warning         string `json:"warning"`
	DSN             string `json:"dsn"`
	Connected       bool   `json:"connected"`
	ConnectionError string `json:"connection_error,omitempty"`
}

func (h *ConnectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request connectionRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"status":  "failed",
			"message": "request body must be JSON with a non-empty dsn",
			"warning": testEndpointWarning,
		})
		return
	}
	if strings.TrimSpace(request.DSN) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"status":  "failed",
			"message": "request body must contain a non-empty dsn",
			"warning": testEndpointWarning,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result := h.connections.Replace(ctx, request.DSN)
	response := connectionResponse{
		Status:    "failed",
		Message:   "TiDB connection was not established",
		Warning:   testEndpointWarning,
		DSN:       result.DSN,
		Connected: result.Connected,
	}
	status := http.StatusUnprocessableEntity
	if result.Connected {
		response.Status = "success"
		response.Message = "TiDB connection established and is now active"
		status = http.StatusOK
	} else if result.Err != nil {
		response.ConnectionError = result.Err.Error()
	}
	h.logger.Info("test TiDB connection completed",
		zap.String("dsn", result.DSN),
		zap.Bool("connected", result.Connected),
		zap.Error(result.Err),
	)
	writeJSON(w, status, response)
}

var _ connectionReplacer = (*ConnectionManager)(nil)
