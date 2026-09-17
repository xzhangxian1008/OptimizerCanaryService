package diagnosis

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	validator *Validator
	logger    *slog.Logger
}

func NewHandler(validator *Validator, logger *slog.Logger) *Handler {
	return &Handler{validator: validator, logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := h.validator.Validate(r.Context())
	status := http.StatusOK
	if response.Status == "failed" {
		status = http.StatusUnprocessableEntity
	}
	h.logger.Info("validation completed", "status", response.Status, "reason", response.Reason)
	writeJSON(w, status, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
