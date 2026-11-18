package audit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pidiekab/internal/modules/audit/domain"
	auditHttp "pidiekab/internal/modules/audit/transport/http"
)

func TestAuditLogJSONMarshalling(t *testing.T) {
	adminID := int64(42)
	entityID := int64(101)
	ip := "192.168.1.50"
	now := time.Now().UTC()

	item := domain.LogItem{
		ID:               1,
		AdminID:          &adminID,
		AdminEmail:       ptr("admin@pidiekab.go.id"),
		AdminDisplayName: ptr("Super Admin"),
		Action:           "content.update",
		EntityType:       "content",
		EntityID:         &entityID,
		Metadata: map[string]any{
			"category_id": float64(5),
		},
		IPAddress: &ip,
		CreatedAt: now,
	}

	b, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal LogItem: %v", err)
	}

	var parsed domain.LogItem
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("failed to unmarshal LogItem: %v", err)
	}

	if parsed.Action != "content.update" {
		t.Errorf("expected action 'content.update', got %q", parsed.Action)
	}
	if parsed.AdminDisplayName == nil || *parsed.AdminDisplayName != "Super Admin" {
		t.Errorf("expected AdminDisplayName 'Super Admin', got %v", parsed.AdminDisplayName)
	}
}

func TestAuditHandler_RequiresAuth(t *testing.T) {
	handler := auditHttp.NewAuditHandler(nil, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Since authService is nil or no session cookie, should return 401 Unauthorized
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", rec.Code)
	}
}

func ptr[T any](v T) *T {
	return &v
}
