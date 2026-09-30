package diagnosis

import (
	"net/http"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/util"
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
	util.WriteJSON(w, status, response)
}
