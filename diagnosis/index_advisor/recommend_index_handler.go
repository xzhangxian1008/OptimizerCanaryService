package index_advisor

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/util"
	"go.uber.org/zap"
)

type RecommendIndexHandler struct {
	repository RecommendIndexRepository
	logger     *zap.Logger
}

func NewRecommendIndexHandler(repository RecommendIndexRepository, logger *zap.Logger) *RecommendIndexHandler {
	return &RecommendIndexHandler{repository: repository, logger: logger}
}

func (h *RecommendIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil && !errors.Is(err, http.ErrNotSupported) {
		h.logger.Warn("clear recommend-index response deadline", zap.Error(err))
	}

	column := r.URL.Query().Get("slow_query_sql_column")
	if !IsSupportedSlowQuerySQLColumn(column) {
		http.Error(w, "slow_query_sql_column must be query or in_memory_original_query", http.StatusBadRequest)
		return
	}
	var results []RecommendIndexResult
	var err error
	if column != "" {
		if configurable, ok := h.repository.(ConfigurableRecommendIndexRepository); ok {
			results, err = configurable.RecommendIndexWithSQLColumn(r.Context(), strings.TrimSpace(column))
		} else {
			results, err = h.repository.RecommendIndex(r.Context())
		}
	} else {
		results, err = h.repository.RecommendIndex(r.Context())
	}
	if err != nil {
		h.logger.Error("recommend index failed", zap.Error(err))
		util.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"status": "failed", "reason": err.Error()})
		return
	}

	report := RenderRecommendIndex(results)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="recommend-index.md"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(report))
}
