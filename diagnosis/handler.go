package diagnosis

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	validator *Validator
	logger    *zap.Logger
}

func NewHandler(validator *Validator, logger *zap.Logger) *Handler {
	return &Handler{validator: validator, logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := h.validator.Validate(r.Context())
	status := http.StatusOK
	if response.Status == "failed" {
		status = http.StatusUnprocessableEntity
	}
	h.logger.Info("validation completed",
		zap.String("status", response.Status),
		zap.String("reason", response.Reason),
	)
	writeJSON(w, status, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
