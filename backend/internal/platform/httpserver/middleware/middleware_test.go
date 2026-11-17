package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"pidiekab/internal/platform/httpserver/middleware"
)

func TestRequestIDMiddlewareGenerated(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.RequestIDFromContext(r.Context())
		if reqID == "" {
			t.Error("expected non-empty request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(rec, req)

	respReqID := rec.Header().Get(middleware.RequestIDHeader)
	if respReqID == "" {
		t.Error("expected X-Request-Id header in response")
	}
}

func TestRequestIDMiddlewarePreservesValid(t *testing.T) {
	const customID = "test-custom-request-id-12345"
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.RequestIDFromContext(r.Context())
		if reqID != customID {
			t.Errorf("expected context request ID %q, got %q", customID, reqID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(middleware.RequestIDHeader, customID)
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(middleware.RequestIDHeader) != customID {
		t.Errorf("expected header %q, got %q", customID, rec.Header().Get(middleware.RequestIDHeader))
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(rec, req)

	if val := rec.Header().Get("X-Content-Type-Options"); val != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", val)
	}
	if val := rec.Header().Get("Referrer-Policy"); val != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy: strict-origin-when-cross-origin, got %q", val)
	}
	if val := rec.Header().Get("X-Frame-Options"); val != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %q", val)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	handler := middleware.RequestID(
		middleware.Recovery(logger)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("simulated critical crash")
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 status code, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/problem+json; charset=utf-8" {
		t.Errorf("expected application/problem+json, got %q", contentType)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	if body["status"] != float64(500) {
		t.Errorf("expected status 500, got %v", body["status"])
	}
	if body["request_id"] == "" || body["request_id"] == nil {
		t.Error("expected request_id in error payload")
	}

	if !bytes.Contains(logBuf.Bytes(), []byte("panic recovered during http request")) {
		t.Error("expected server-side panic log")
	}
}
