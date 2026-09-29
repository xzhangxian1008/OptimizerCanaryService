package diagnosis

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type comparisonGenerator interface {
	Compare(context.Context) (string, error)
}

type CompareHandler struct {
	comparer comparisonGenerator
	logger   *zap.Logger
}

func NewCompareHandler(comparer comparisonGenerator, logger *zap.Logger) *CompareHandler {
	return &CompareHandler{comparer: comparer, logger: logger}
}

func (h *CompareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// Comparing every recorded statement can exceed the server's default write
	// timeout. Individual database operations still have their own deadlines.
	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil && !errors.Is(err, http.ErrNotSupported) {
		h.logger.Warn("clear comparison response deadline", zap.Error(err))
	}
	report, err := h.comparer.Compare(r.Context())
	if err != nil {
		h.logger.Error("comparison failed", zap.Error(err))
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"status": "failed", "reason": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="compare.md"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(report))
}
