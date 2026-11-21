package domain

import (
	"time"
)

type PageView struct {
	ID         int64     `json:"id"`
	Path       string    `json:"path"`
	Locale     string    `json:"locale"`
	ContentID  *int64    `json:"content_id,omitempty"`
	IPHash     string    `json:"ip_hash"`
	UserAgent  string    `json:"user_agent"`
	DeviceType string    `json:"device_type"`
	Referrer   string    `json:"referrer"`
	CreatedAt  time.Time `json:"created_at"`
}

type TopArticle struct {
	ContentID    int64  `json:"content_id"`
	Title        string `json:"title"`
	CategoryName string `json:"category_name"`
	Slug         string `json:"slug"`
	Locale       string `json:"locale"`
	Views        int64  `json:"views"`
}

type TrafficDay struct {
	Date     string `json:"date"`
	DayName  string `json:"day_name"`
	Views    int64  `json:"views"`
	Visitors int64  `json:"visitors"`
}

type CategoryTraffic struct {
	Name  string `json:"name"`
	Views int64  `json:"views"`
}

type ContentKPIs struct {
	TotalContents    int64 `json:"total_contents"`
	PublishedID      int64 `json:"published_id"`
	PublishedEN      int64 `json:"published_en"`
	DraftContents    int64 `json:"draft_contents"`
	TotalCategories  int64 `json:"total_categories"`
	TotalMedia       int64 `json:"total_media"`
	TotalDownloads   int64 `json:"total_downloads"`
	TotalStaticPages int64 `json:"total_static_pages"`
}

type VisitorKPIs struct {
	ActiveVisitorsNow   int64   `json:"active_visitors_now"`
	TotalVisitorsToday  int64   `json:"total_visitors_today"`
	TotalViewsToday     int64   `json:"total_views_today"`
	TotalViewsAllTime   int64   `json:"total_views_all_time"`
	MobilePercentage    float64 `json:"mobile_percentage"`
	DesktopPercentage   float64 `json:"desktop_percentage"`
	IndonesianTrafficPct float64 `json:"indonesian_traffic_pct"`
	EnglishTrafficPct   float64 `json:"english_traffic_pct"`
}

type DashboardStats struct {
	VisitorKPIs     VisitorKPIs       `json:"visitor_kpis"`
	ContentKPIs     ContentKPIs       `json:"content_kpis"`
	TrafficTrend7D  []TrafficDay      `json:"traffic_trend_7d"`
	TopArticles     []TopArticle      `json:"top_articles"`
	CategoryTraffic []CategoryTraffic `json:"category_traffic"`
	RecentArticles  []RecentArticle   `json:"recent_articles"`
}

type RecentArticle struct {
	ID           int64     `json:"id"`
	TitleID      string    `json:"title_id"`
	TitleEN      string    `json:"title_en"`
	CategoryName string    `json:"category_name"`
	StatusID     string    `json:"status_id"`
	StatusEN     string    `json:"status_en"`
	Views        int64     `json:"views"`
	UpdatedAt    time.Time `json:"updated_at"`
}
