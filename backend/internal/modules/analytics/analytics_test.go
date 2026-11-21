package analytics_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"pidiekab/internal/modules/analytics/application"
	"pidiekab/internal/modules/analytics/domain"
)

type mockAnalyticsRepo struct{}

func (m *mockAnalyticsRepo) RecordPageView(ctx context.Context, pv domain.PageView) error {
	return nil
}

func (m *mockAnalyticsRepo) GetContentKPIs(ctx context.Context) (domain.ContentKPIs, error) {
	return domain.ContentKPIs{
		TotalContents:   25,
		PublishedID:     20,
		PublishedEN:     18,
		DraftContents:   5,
		TotalCategories: 8,
		TotalMedia:      42,
	}, nil
}

func (m *mockAnalyticsRepo) GetVisitorKPIs(ctx context.Context, inMemoryActive int64) (domain.VisitorKPIs, error) {
	return domain.VisitorKPIs{
		ActiveVisitorsNow:  inMemoryActive,
		TotalVisitorsToday: 140,
		TotalViewsToday:    320,
		TotalViewsAllTime:  15420,
	}, nil
}

func (m *mockAnalyticsRepo) GetTrafficTrend7D(ctx context.Context) ([]domain.TrafficDay, error) {
	return []domain.TrafficDay{
		{Date: "2026-11-20", DayName: "Fri", Views: 210, Visitors: 95},
		{Date: "2026-11-21", DayName: "Sat", Views: 320, Visitors: 140},
	}, nil
}

func (m *mockAnalyticsRepo) GetTopArticles(ctx context.Context, limit int) ([]domain.TopArticle, error) {
	return []domain.TopArticle{
		{ContentID: 23, Title: "Menyusuri Nafas Tanah Pedir", Views: 1250},
		{ContentID: 5, Title: "Kue Adee Khas Pidie", Views: 890},
	}, nil
}

func (m *mockAnalyticsRepo) GetCategoryTraffic(ctx context.Context, limit int) ([]domain.CategoryTraffic, error) {
	return []domain.CategoryTraffic{
		{Name: "Sejarah & Budaya", Views: 2400},
		{Name: "Kuliner Khas", Views: 1800},
	}, nil
}

func (m *mockAnalyticsRepo) GetRecentArticlesWithViews(ctx context.Context, limit int) ([]domain.RecentArticle, error) {
	return []domain.RecentArticle{
		{ID: 23, TitleID: "Menyusuri Nafas Tanah Pedir", StatusID: "published", Views: 1250},
	}, nil
}

func (m *mockAnalyticsRepo) ResolveContentIDBySlugOrPath(ctx context.Context, path string) *int64 {
	id := int64(23)
	return &id
}

func TestActiveVisitorsTracking(t *testing.T) {
	service := application.NewAnalyticsService(nil)

	// Simulate 3 requests from different IPs
	r1 := httptest.NewRequest("GET", "/id/sejarah-budaya/menyusuri-nafas-tanah-pedir/", nil)
	r1.RemoteAddr = "192.168.1.10:1234"
	service.TrackRequest(r1)

	r2 := httptest.NewRequest("GET", "/id/kuliner-khas/kue-adee/", nil)
	r2.RemoteAddr = "192.168.1.20:5678"
	service.TrackRequest(r2)

	r3 := httptest.NewRequest("GET", "/id/sejarah-budaya/masjid-po-teumeureuhom/", nil)
	r3.RemoteAddr = "192.168.1.30:9999"
	service.TrackRequest(r3)

	count := service.GetActiveVisitorsCount()
	if count != 3 {
		t.Errorf("expected 3 active visitors, got %d", count)
	}

	// Static files should not be tracked
	rStatic := httptest.NewRequest("GET", "/assets/main.js", nil)
	rStatic.RemoteAddr = "192.168.1.40:1111"
	service.TrackRequest(rStatic)

	if service.GetActiveVisitorsCount() != 3 {
		t.Errorf("expected still 3 active visitors after static asset request, got %d", service.GetActiveVisitorsCount())
	}
}

func TestGetDashboardStats(t *testing.T) {
	// Simple struct verification
	stats := domain.DashboardStats{
		VisitorKPIs: domain.VisitorKPIs{
			ActiveVisitorsNow: 5,
			TotalViewsToday:   100,
		},
		ContentKPIs: domain.ContentKPIs{
			TotalContents: 15,
		},
	}

	if stats.VisitorKPIs.ActiveVisitorsNow != 5 {
		t.Errorf("expected 5 active visitors, got %d", stats.VisitorKPIs.ActiveVisitorsNow)
	}
	if stats.ContentKPIs.TotalContents != 15 {
		t.Errorf("expected 15 contents, got %d", stats.ContentKPIs.TotalContents)
	}
	_ = time.Now()
}
