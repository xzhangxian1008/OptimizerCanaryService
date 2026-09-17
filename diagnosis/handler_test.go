package diagnosis

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerValidatesAllSources(t *testing.T) {
	repo := &fakeRepository{samples: []Sample{{SQL: "select 1"}, {SQL: "select 2"}, {SQL: "select 3"}}}
	handler := NewHandler(NewValidator(repo), slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodPost, "/validate", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"statement_summary":{"sampled":3,"explained":3}`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"slow_query":{"sampled":3,"explained":3}`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"top_sql":{"sampled":3,"explained":3}`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestHandlerReturnsUnprocessableEntityOnFailure(t *testing.T) {
	handler := NewHandler(NewValidator(&fakeRepository{sampleErr: errors.New("unavailable")}), slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodPost, "/validate", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected HTTP 422, got %d", response.Code)
	}
}
