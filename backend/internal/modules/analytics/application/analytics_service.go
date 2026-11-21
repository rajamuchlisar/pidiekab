package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"pidiekab/internal/modules/analytics/domain"
	"pidiekab/internal/modules/analytics/infrastructure/postgres"
)

type AnalyticsService struct {
	repo         *postgres.AnalyticsRepository
	queue        chan domain.PageView
	activeMu     sync.RWMutex
	activeVisits map[string]time.Time // ipHash -> lastSeen
}

func NewAnalyticsService(repo *postgres.AnalyticsRepository) *AnalyticsService {
	s := &AnalyticsService{
		repo:         repo,
		queue:        make(chan domain.PageView, 1000),
		activeVisits: make(map[string]time.Time),
	}

	// Background worker to flush queued page views without blocking HTTP request lifecycle
	go s.workerLoop()

	// Periodic cleanup of active sessions older than 15 minutes
	go s.cleanupLoop()

	return s
}

func (s *AnalyticsService) workerLoop() {
	for pv := range s.queue {
		if s.repo != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = s.repo.RecordPageView(ctx, pv)
			cancel()
		}
	}
}

func (s *AnalyticsService) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.activeMu.Lock()
		cutoff := time.Now().Add(-15 * time.Minute)
		for ipHash, lastSeen := range s.activeVisits {
			if lastSeen.Before(cutoff) {
				delete(s.activeVisits, ipHash)
			}
		}
		s.activeMu.Unlock()
	}
}

func (s *AnalyticsService) TrackRequest(r *http.Request) {
	if r == nil || r.Method != http.MethodGet && r.Method != http.MethodHead {
		return
	}

	path := r.URL.Path
	// Ignore static assets, health checks, cms administration, api endpoints, and robots/sitemaps
	if strings.HasPrefix(path, "/cms") ||
		strings.HasPrefix(path, "/api") ||
		strings.HasPrefix(path, "/health") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".webp") ||
		strings.HasSuffix(path, ".svg") ||
		strings.HasSuffix(path, ".woff2") ||
		strings.HasSuffix(path, ".map") ||
		path == "/robots.txt" ||
		path == "/sitemap.xml" {
		return
	}

	ipHash := hashIP(getClientIP(r))
	ua := r.UserAgent()
	device := "desktop"
	uaLower := strings.ToLower(ua)
	if strings.Contains(uaLower, "mobile") || strings.Contains(uaLower, "android") || strings.Contains(uaLower, "iphone") {
		device = "mobile"
	} else if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		device = "tablet"
	}

	locale := "id"
	if strings.HasPrefix(path, "/en/") || path == "/en" {
		locale = "en"
	}

	now := time.Now()

	// Update active memory tracker
	s.activeMu.Lock()
	s.activeVisits[ipHash] = now
	s.activeMu.Unlock()

	// Extract content ID if it matches an article
	var contentID *int64
	if s.repo != nil {
		contentID = s.repo.ResolveContentIDBySlugOrPath(r.Context(), path)
	}

	pv := domain.PageView{
		Path:       path,
		Locale:     locale,
		ContentID:  contentID,
		IPHash:     ipHash,
		UserAgent:  ua,
		DeviceType: device,
		Referrer:   r.Referer(),
		CreatedAt:  now,
	}

	// Non-blocking send to queue
	select {
	case s.queue <- pv:
	default:
		// Queue full, drop gracefully
	}
}

func (s *AnalyticsService) GetActiveVisitorsCount() int64 {
	s.activeMu.RLock()
	defer s.activeMu.RUnlock()

	var count int64
	cutoff := time.Now().Add(-5 * time.Minute)
	for _, lastSeen := range s.activeVisits {
		if lastSeen.After(cutoff) {
			count++
		}
	}
	return count
}

func (s *AnalyticsService) GetDashboardStats(ctx context.Context) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	// 1. Content KPIs
	cKPIs, err := s.repo.GetContentKPIs(ctx)
	if err == nil {
		stats.ContentKPIs = cKPIs
	}

	// 2. Visitor KPIs
	activeNow := s.GetActiveVisitorsCount()
	vKPIs, err := s.repo.GetVisitorKPIs(ctx, activeNow)
	if err == nil {
		stats.VisitorKPIs = vKPIs
	}

	// 3. Traffic Trend 7D
	trend, err := s.repo.GetTrafficTrend7D(ctx)
	if err == nil {
		stats.TrafficTrend7D = trend
	}

	// 4. Top Viewed Articles
	top, err := s.repo.GetTopArticles(ctx, 6)
	if err == nil {
		stats.TopArticles = top
	}

	// 5. Category Traffic
	catTraffic, err := s.repo.GetCategoryTraffic(ctx, 5)
	if err == nil {
		stats.CategoryTraffic = catTraffic
	}

	// 6. Recent Articles with Views
	recent, err := s.repo.GetRecentArticlesWithViews(ctx, 6)
	if err == nil {
		stats.RecentArticles = recent
	}

	return stats, nil
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func hashIP(ip string) string {
	if ip == "" {
		return "anonymous"
	}
	h := sha256.Sum256([]byte(ip + "-pidiekab-salt"))
	return hex.EncodeToString(h[:16])
}
