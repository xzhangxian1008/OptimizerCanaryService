package diagnosis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type fakeConnectionReplacer struct {
	result ConnectionStatus
	dsn    string
}

func (f *fakeConnectionReplacer) Replace(_ context.Context, dsn string) ConnectionStatus {
	f.dsn = dsn
	f.result.DSN = dsn
	return f.result
}

func TestConnectionHandlerReportsSuccessfulConnection(t *testing.T) {
	replacer := &fakeConnectionReplacer{result: ConnectionStatus{Connected: true}}
	handler := NewConnectionHandler(replacer, zap.NewNop())
	request := httptest.NewRequest(http.MethodPost, "/test/connect", strings.NewReader(`{"dsn":"root@tcp(tidb:4000)/"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"connected":true`) || !strings.Contains(body, `TEST ONLY`) {
		t.Fatalf("unexpected response: %s", body)
	}
	if replacer.dsn != "root@tcp(tidb:4000)/" {
		t.Fatalf("unexpected DSN passed to replacer: %q", replacer.dsn)
	}
}

func TestConnectionHandlerReportsFailedConnection(t *testing.T) {
	handler := NewConnectionHandler(&fakeConnectionReplacer{result: ConnectionStatus{Err: context.DeadlineExceeded}}, zap.NewNop())
	request := httptest.NewRequest(http.MethodPost, "/test/connect", strings.NewReader(`{"dsn":"root@tcp(unreachable:4000)/"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected HTTP 422, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"connected":false`) {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestConnectionHandlerRejectsEmptyDSN(t *testing.T) {
	handler := NewConnectionHandler(&fakeConnectionReplacer{}, zap.NewNop())
	request := httptest.NewRequest(http.MethodPost, "/test/connect", strings.NewReader(`{"dsn":""}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400, got %d", response.Code)
	}
}
