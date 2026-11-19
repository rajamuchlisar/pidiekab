package sitegeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Generator struct {
	pool    *pgxpool.Pool
	baseURL string
}

func NewGenerator(pool *pgxpool.Pool, baseURL string) *Generator {
	if baseURL == "" {
		baseURL = "https://pidiekab.web.id"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Generator{pool: pool, baseURL: baseURL}
}

type SiteContext struct {
	BaseURL        string
	SiteNameID     string
	SiteNameEN     string
	TaglineID      string
	TaglineEN      string
	DefaultSeoID   string
	DefaultSeoEN   string
	ContactEmail   string
	ContactPhone   string
	ContactAddress string
}

type CategoryRender struct {
	ID          int64
	Locale      string
	Name        string
	Description string
	Slug        string
}

type ContentRender struct {
	ID           int64
	Locale       string
	CategorySlug string
	CategoryName string
	Title        string
	Slug         string
	Summary      string
	BodyHTML     template.HTML
	Address      string
	Latitude     *float64
	Longitude    *float64
	Sources      []SourceRender
}

type SourceRender struct {
	Name string
	URL  string
	Type string
}

type StaticPageRender struct {
	ID       int64
	Locale   string
	Title    string
	Slug     string
	BodyHTML template.HTML
	SeoTitle string
	SeoDesc  string
}

func (g *Generator) Generate(ctx context.Context, targetDir string) error {
	// 1. Fetch site settings
	siteCtx, err := g.fetchSiteContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to load site context: %w", err)
	}

	// 2. Fetch categories
	categoriesID, categoriesEN, err := g.fetchCategories(ctx)
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	// 3. Fetch published contents
	contentsID, contentsEN, err := g.fetchPublishedContents(ctx)
	if err != nil {
		return fmt.Errorf("failed to load contents: %w", err)
	}

	// 4. Fetch published static pages
	pagesID, pagesEN, err := g.fetchPublishedPages(ctx)
	if err != nil {
		return fmt.Errorf("failed to load static pages: %w", err)
	}

	// 5. Render Indonesian site
	if err := g.renderLocaleSite(targetDir, "id", siteCtx, categoriesID, contentsID, pagesID); err != nil {
		return fmt.Errorf("failed to render ID site: %w", err)
	}

	// 6. Render English site
	if err := g.renderLocaleSite(targetDir, "en", siteCtx, categoriesEN, contentsEN, pagesEN); err != nil {
		return fmt.Errorf("failed to render EN site: %w", err)
	}

	// 7. Render Root Redirect index.html (redirects / to /id/)
	rootHTML := `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="0;url=/id/"><script>location.replace('/id/');</script></head><body><p>Redirecting to <a href="/id/">Bahasa Indonesia</a>...</p></body></html>`
	if err := os.WriteFile(filepath.Join(targetDir, "index.html"), []byte(rootHTML), 0644); err != nil {
		return err
	}

	// 8. Render sitemap.xml
	if err := g.renderSitemap(targetDir, categoriesID, categoriesEN, contentsID, contentsEN, pagesID, pagesEN); err != nil {
		return fmt.Errorf("failed to render sitemap: %w", err)
	}

	// 9. Render robots.txt
	robotsContent := fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n", g.baseURL)
	if err := os.WriteFile(filepath.Join(targetDir, "robots.txt"), []byte(robotsContent), 0644); err != nil {
		return err
	}

	return nil
}

func (g *Generator) fetchSiteContext(ctx context.Context) (*SiteContext, error) {
	sc := &SiteContext{
		BaseURL:    g.baseURL,
		SiteNameID: "PidieKab — Platform Promosi & Potensi Daerah Pidie",
		SiteNameEN: "PidieKab — Regional Promotion & Discovery Platform",
	}

	query := `
		SELECT
			COALESCE(contact_email, ''), COALESCE(contact_phone, ''), COALESCE(contact_address, '')
		FROM site_settings WHERE id = 1
	`
	_ = g.pool.QueryRow(ctx, query).Scan(&sc.ContactEmail, &sc.ContactPhone, &sc.ContactAddress)

	tRows, err := g.pool.Query(ctx, `SELECT locale, site_name, COALESCE(tagline, ''), COALESCE(default_seo_title, '') FROM site_setting_translations WHERE site_settings_id = 1`)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var loc, name, tag, seo string
			if err := tRows.Scan(&loc, &name, &tag, &seo); err == nil {
				if loc == "id" {
					if name != "" {
						sc.SiteNameID = name
					}
					sc.TaglineID = tag
					sc.DefaultSeoID = seo
				} else if loc == "en" {
					if name != "" {
						sc.SiteNameEN = name
					}
					sc.TaglineEN = tag
					sc.DefaultSeoEN = seo
				}
			}
		}
	}

	return sc, nil
}

func (g *Generator) fetchCategories(ctx context.Context) (map[int64]CategoryRender, map[int64]CategoryRender, error) {
	query := `
		SELECT c.id, ct.locale, ct.name, COALESCE(ct.description, ''), ct.slug
		FROM categories c
		JOIN category_translations ct ON ct.category_id = c.id
		WHERE c.is_active = TRUE
		ORDER BY c.id ASC
	`
	rows, err := g.pool.Query(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	catsID := make(map[int64]CategoryRender)
	catsEN := make(map[int64]CategoryRender)

	for rows.Next() {
		var cr CategoryRender
		if err := rows.Scan(&cr.ID, &cr.Locale, &cr.Name, &cr.Description, &cr.Slug); err != nil {
			return nil, nil, err
		}
		if cr.Locale == "id" {
			catsID[cr.ID] = cr
		} else if cr.Locale == "en" {
			catsEN[cr.ID] = cr
		}
	}

	return catsID, catsEN, nil
}

func (g *Generator) fetchPublishedContents(ctx context.Context) ([]ContentRender, []ContentRender, error) {
	query := `
		SELECT
			c.id, ct.locale, ct.title, ct.slug, COALESCE(ct.summary, ''), ct.body,
			COALESCE(c.address, ''), c.latitude, c.longitude,
			cat_t.slug, cat_t.name
		FROM contents c
		JOIN content_translations ct ON ct.content_id = c.id
		JOIN categories cat ON cat.id = c.category_id AND cat.is_active = TRUE
		JOIN category_translations cat_t ON cat_t.category_id = cat.id AND cat_t.locale = ct.locale
		WHERE ct.status = 'published'
		ORDER BY ct.published_at DESC NULLS LAST, c.id DESC
	`
	rows, err := g.pool.Query(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var contentsID, contentsEN []ContentRender
	for rows.Next() {
		var item ContentRender
		var bodyBytes []byte
		if err := rows.Scan(
			&item.ID, &item.Locale, &item.Title, &item.Slug, &item.Summary, &bodyBytes,
			&item.Address, &item.Latitude, &item.Longitude,
			&item.CategorySlug, &item.CategoryName,
		); err != nil {
			return nil, nil, err
		}

		item.BodyHTML = renderBodyBlocksToHTML(bodyBytes)

		if item.Locale == "id" {
			contentsID = append(contentsID, item)
		} else if item.Locale == "en" {
			contentsEN = append(contentsEN, item)
		}
	}

	return contentsID, contentsEN, nil
}

func (g *Generator) fetchPublishedPages(ctx context.Context) ([]StaticPageRender, []StaticPageRender, error) {
	query := `
		SELECT sp.id, spt.locale, spt.title, spt.slug, spt.body, COALESCE(spt.seo_title, ''), COALESCE(spt.seo_description, '')
		FROM static_pages sp
		JOIN static_page_translations spt ON spt.static_page_id = sp.id
		WHERE spt.status = 'published'
		ORDER BY sp.id ASC
	`
	rows, err := g.pool.Query(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var pagesID, pagesEN []StaticPageRender
	for rows.Next() {
		var p StaticPageRender
		var bodyBytes []byte
		if err := rows.Scan(&p.ID, &p.Locale, &p.Title, &p.Slug, &bodyBytes, &p.SeoTitle, &p.SeoDesc); err != nil {
			return nil, nil, err
		}
		p.BodyHTML = renderBodyBlocksToHTML(bodyBytes)
		if p.Locale == "id" {
			pagesID = append(pagesID, p)
		} else if p.Locale == "en" {
			pagesEN = append(pagesEN, p)
		}
	}

	return pagesID, pagesEN, nil
}

var (
	linkTagRegex    = regexp.MustCompile(`(?i)<a\s+([^>]+)>`)
	hrefAttrRegex   = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)
	relAttrRegex    = regexp.MustCompile(`(?i)\brel\s*=\s*["'][^"']*["']`)
	targetAttrRegex = regexp.MustCompile(`(?i)\btarget\s*=\s*["'][^"']*["']`)
)

// enforceNofollow ensures all external outbound links strictly carry rel="nofollow noopener noreferrer" and target="_blank".
func enforceNofollow(htmlStr string) string {
	if htmlStr == "" {
		return ""
	}

	return linkTagRegex.ReplaceAllStringFunc(htmlStr, func(fullTag string) string {
		hrefMatches := hrefAttrRegex.FindStringSubmatch(fullTag)
		if len(hrefMatches) < 2 {
			return fullTag
		}

		href := strings.TrimSpace(hrefMatches[1])
		// External links start with http://, https://, or //, and do NOT point to internal pidiekab.web.id
		isExternal := (strings.HasPrefix(strings.ToLower(href), "http://") ||
			strings.HasPrefix(strings.ToLower(href), "https://") ||
			strings.HasPrefix(href, "//")) &&
			!strings.Contains(strings.ToLower(href), "pidiekab.web.id")

		if !isExternal {
			return fullTag
		}

		// Remove existing rel and target attributes
		cleaned := relAttrRegex.ReplaceAllString(fullTag, "")
		cleaned = targetAttrRegex.ReplaceAllString(cleaned, "")
		cleaned = strings.TrimSuffix(strings.TrimSpace(cleaned), ">")

		return fmt.Sprintf(`%s rel="nofollow noopener noreferrer" target="_blank">`, cleaned)
	})
}

func renderBodyBlocksToHTML(bodyBytes []byte) template.HTML {
	if len(bodyBytes) == 0 {
		return ""
	}

	var parsed map[string]any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return template.HTML(enforceNofollow("<p>" + template.HTMLEscapeString(string(bodyBytes)) + "</p>"))
	}

	if htmlStr, ok := parsed["html"].(string); ok && htmlStr != "" {
		return template.HTML(enforceNofollow(htmlStr))
	}

	blocks, ok := parsed["blocks"].([]any)
	if !ok {
		if mdStr, ok := parsed["markdown"].(string); ok && mdStr != "" {
			return template.HTML(enforceNofollow("<p>" + template.HTMLEscapeString(mdStr) + "</p>"))
		}
		return ""
	}

	var sb strings.Builder
	for _, b := range blocks {
		bm, ok := b.(map[string]any)
		if !ok {
			continue
		}
		bType, _ := bm["type"].(string)
		data, _ := bm["data"].(map[string]any)
		if data == nil {
			continue
		}

		switch bType {
		case "paragraph":
			text, _ := data["text"].(string)
			sb.WriteString(fmt.Sprintf("<p>%s</p>\n", template.HTMLEscapeString(text)))
		case "header":
			text, _ := data["text"].(string)
			level := 2
			if lvl, ok := data["level"].(float64); ok && lvl >= 1 && lvl <= 6 {
				level = int(lvl)
			}
			sb.WriteString(fmt.Sprintf("<h%d>%s</h%d>\n", level, template.HTMLEscapeString(text), level))
		}
	}

	return template.HTML(enforceNofollow(sb.String()))
}

func (g *Generator) renderLocaleSite(
	targetDir, locale string,
	sc *SiteContext,
	cats map[int64]CategoryRender,
	contents []ContentRender,
	pages []StaticPageRender,
) error {
	localeDir := filepath.Join(targetDir, locale)
	if err := os.MkdirAll(localeDir, 0755); err != nil {
		return err
	}

	siteName := sc.SiteNameID
	tagline := sc.TaglineID
	backHomeText := "&larr; Kembali ke Beranda"
	sectionTitle := "Destinasi & Kebudayaan Pidie"
	heroSubtitle := "Jelajahi keindahan alam, kekayaan budaya, dan kuliner khas Kabupaten Pidie, Aceh."
	readMoreText := "Selengkapnya &rarr;"
	locationLabel := "Lokasi"
	sourcesHeader := "Sumber & Referensi Terverifikasi:"
	footerText := "&copy; PidieKab — Inisiatif Platform Digital Independen Masyarakat Pidie. Bukan website resmi pemerintah."

	idClass := "active"
	enClass := ""
	if locale == "en" {
		siteName = sc.SiteNameEN
		tagline = sc.TaglineEN
		backHomeText = "&larr; Back to Home"
		sectionTitle = "Pidie Destinations & Culture"
		heroSubtitle = "Discover the natural splendor, culinary treasures, and cultural heritage of Pidie, Aceh."
		readMoreText = "Explore &rarr;"
		locationLabel = "Location"
		sourcesHeader = "Verified Editorial Sources & Citations:"
		footerText = "&copy; PidieKab — Independent Regional Promotion & Visitor Discovery Platform. Not an official government website."
		idClass = ""
		enClass = "active"
	}

	commonCSS := `
*, *::before, *::after { box-sizing: border-box; }
html { font-size: 16px; -webkit-text-size-adjust: 100%; scroll-behavior: smooth; }
body {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    line-height: 1.6;
    color: #1e293b;
    background-color: #f8fafc;
    -webkit-font-smoothing: antialiased;
}
.site-wrapper {
    width: 100%;
    max-width: 1040px;
    margin: 0 auto;
    padding: 0 1rem;
}
@media (min-width: 640px) {
    .site-wrapper { padding: 0 1.5rem; }
}
.site-header {
    background: #ffffff;
    border-bottom: 1px solid #e2e8f0;
    position: sticky;
    top: 0;
    z-index: 50;
    box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}
.header-inner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0;
    gap: 0.75rem;
}
.brand-group {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    text-decoration: none;
    min-width: 0;
}
.brand-logo {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 1.25rem;
    font-weight: 800;
    color: #047857;
    text-decoration: none;
    letter-spacing: -0.02em;
    white-space: nowrap;
}
.brand-logo-badge {
    background: #047857;
    color: #ffffff;
    font-size: 0.75rem;
    font-weight: 800;
    padding: 0.2rem 0.45rem;
    border-radius: 6px;
    letter-spacing: 0.05em;
}
.brand-tagline {
    display: none;
    font-size: 0.8125rem;
    color: #64748b;
    border-left: 1px solid #cbd5e1;
    padding-left: 0.5rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
@media (min-width: 768px) {
    .brand-tagline { display: inline-block; }
    .brand-logo { font-size: 1.35rem; }
}

/* Language Switcher */
.lang-switcher {
    display: inline-flex;
    align-items: center;
    background: #f1f5f9;
    border: 1px solid #e2e8f0;
    border-radius: 9999px;
    padding: 3px;
    flex-shrink: 0;
}
.lang-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.35rem 0.65rem;
    font-size: 0.75rem;
    font-weight: 700;
    border-radius: 9999px;
    color: #64748b;
    text-decoration: none;
    transition: all 0.15s ease-in-out;
    min-width: 36px;
    line-height: 1;
}
@media (min-width: 640px) {
    .lang-btn {
        padding: 0.4rem 0.85rem;
        font-size: 0.8125rem;
    }
}
.lang-btn.active {
    background: #047857;
    color: #ffffff !important;
    box-shadow: 0 1px 2px rgba(4, 120, 87, 0.2);
}
.lang-btn:hover:not(.active) {
    color: #0f172a;
    background: #e2e8f0;
}
.lang-short { display: inline; }
.lang-long { display: none; }
@media (min-width: 640px) {
    .lang-short { display: none; }
    .lang-long { display: inline; }
}

/* Hero Section */
.hero {
    background: linear-gradient(135deg, #ecfdf5 0%, #f0fdf4 100%);
    border: 1px solid #d1fae5;
    border-left: 4px solid #059669;
    border-radius: 12px;
    padding: 1.25rem;
    margin: 1.25rem 0 1.75rem;
}
@media (min-width: 640px) {
    .hero { padding: 1.75rem 2rem; margin: 2rem 0; border-radius: 16px; }
}
.hero h2 {
    font-size: clamp(1.2rem, 3.5vw, 1.75rem);
    font-weight: 800;
    color: #064e3b;
    margin: 0 0 0.5rem;
    line-height: 1.3;
}
.hero p {
    font-size: 0.9375rem;
    color: #334155;
    margin: 0;
    line-height: 1.5;
}
@media (min-width: 640px) {
    .hero p { font-size: 1.05rem; }
}

/* Card Grid */
.section-title {
    font-size: 1.25rem;
    font-weight: 800;
    color: #0f172a;
    margin: 0 0 1.25rem;
    letter-spacing: -0.01em;
}
.card-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1rem;
}
@media (min-width: 640px) {
    .card-grid { grid-template-columns: repeat(2, 1fr); gap: 1.25rem; }
}
@media (min-width: 1024px) {
    .card-grid { grid-template-columns: repeat(3, 1fr); gap: 1.5rem; }
}
.card {
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    box-shadow: 0 1px 3px rgba(0,0,0,0.03);
    transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.06);
    border-color: #cbd5e1;
}
.badge {
    display: inline-block;
    align-self: flex-start;
    background: #e0e7ff;
    color: #3730a3;
    padding: 0.25rem 0.55rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    margin-bottom: 0.5rem;
}
.card h3 {
    margin: 0 0 0.5rem;
    font-size: 1.05rem;
    line-height: 1.4;
    color: #065f46;
}
.card h3 a {
    color: inherit;
    text-decoration: none;
}
.card h3 a:hover {
    color: #047857;
    text-decoration: underline;
}
.card p {
    margin: 0 0 1rem;
    font-size: 0.875rem;
    color: #475569;
    line-height: 1.55;
    flex-grow: 1;
}
.card-action {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.875rem;
    font-weight: 700;
    color: #047857;
    text-decoration: none;
}
.card-action:hover {
    color: #065f46;
    text-decoration: underline;
}

/* Detail page */
.back-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.45rem 0.85rem;
    background: #f1f5f9;
    color: #334155;
    border-radius: 8px;
    font-size: 0.8125rem;
    font-weight: 600;
    text-decoration: none;
    margin-bottom: 1.25rem;
    transition: background 0.15s;
}
.back-btn:hover {
    background: #e2e8f0;
    color: #0f172a;
}
.article-header {
    margin-bottom: 1.5rem;
    padding-bottom: 1.25rem;
    border-bottom: 1px solid #e2e8f0;
}
.article-title {
    font-size: clamp(1.35rem, 4vw, 2.1rem);
    font-weight: 800;
    color: #0f172a;
    line-height: 1.25;
    margin: 0.6rem 0 0.75rem;
}
.article-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.75rem;
    font-size: 0.85rem;
    color: #64748b;
}
.article-body {
    font-size: 1rem;
    line-height: 1.75;
    color: #334155;
}
.article-body p {
    margin: 0 0 1.25rem;
}
.article-body h2 {
    font-size: 1.35rem;
    font-weight: 700;
    color: #0f172a;
    margin: 2rem 0 0.75rem;
}
.article-body h3 {
    font-size: 1.15rem;
    font-weight: 700;
    color: #0f172a;
    margin: 1.5rem 0 0.5rem;
}
.article-body img {
    max-width: 100%;
    height: auto;
    border-radius: 10px;
}
.source-box {
    margin-top: 2.5rem;
    padding: 1.25rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-left: 4px solid #047857;
    border-radius: 8px;
    font-size: 0.875rem;
}
.source-box strong {
    color: #0f172a;
    display: block;
    margin-bottom: 0.5rem;
}
.source-list {
    margin: 0;
    padding-left: 1.25rem;
}
.source-list li {
    margin-bottom: 0.35rem;
    word-break: break-word;
    overflow-wrap: break-word;
}
.source-list a {
    color: #047857;
    text-decoration: underline;
}

/* Footer */
.site-footer {
    margin-top: 3.5rem;
    padding: 2rem 0;
    border-top: 1px solid #e2e8f0;
    color: #64748b;
    font-size: 0.8125rem;
    text-align: center;
    line-height: 1.6;
}
`

	headerHTML := fmt.Sprintf(`
    <header class="site-header">
        <div class="site-wrapper">
            <div class="header-inner">
                <a href="/%s/" class="brand-group">
                    <span class="brand-logo"><span class="brand-logo-badge">PK</span>PidieKab</span>
                    <span class="brand-tagline">%s</span>
                </a>
                <nav class="lang-switcher" aria-label="Language selector">
                    <a href="/id/" class="lang-btn %s"><span class="lang-short">ID</span><span class="lang-long">Bahasa Indonesia</span></a>
                    <a href="/en/" class="lang-btn %s"><span class="lang-short">EN</span><span class="lang-long">English</span></a>
                </nav>
            </div>
        </div>
    </header>
`, locale, template.HTMLEscapeString(siteName), idClass, enClass)

	footerHTML := fmt.Sprintf(`
    <footer class="site-footer">
        <div class="site-wrapper">
            <p>%s</p>
        </div>
    </footer>
`, footerText)

	// 1. Render Homepage: <locale>/index.html
	homeHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s — %s</title>
    <meta name="description" content="%s">
    <link rel="canonical" href="%s/%s/">
    <link rel="alternate" hreflang="id" href="%s/id/">
    <link rel="alternate" hreflang="en" href="%s/en/">
    <link rel="alternate" hreflang="x-default" href="%s/id/">
    <style>%s</style>
</head>
<body>
%s
    <main class="site-wrapper">
        <section class="hero">
            <h2>%s</h2>
            <p>%s</p>
        </section>
        <section>
            <h2 class="section-title">%s</h2>
            <div class="card-grid">
`,
		locale,
		template.HTMLEscapeString(siteName),
		template.HTMLEscapeString(tagline),
		template.HTMLEscapeString(tagline),
		g.baseURL, locale,
		g.baseURL, g.baseURL, g.baseURL,
		commonCSS,
		headerHTML,
		template.HTMLEscapeString(tagline),
		template.HTMLEscapeString(heroSubtitle),
		template.HTMLEscapeString(sectionTitle),
	)

	var cardsBuilder strings.Builder
	for _, c := range contents {
		detailPath := fmt.Sprintf("/%s/%s/%s/", locale, c.CategorySlug, c.Slug)
		cardsBuilder.WriteString(fmt.Sprintf(`
                <article class="card">
                    <div>
                        <span class="badge">%s</span>
                        <h3><a href="%s">%s</a></h3>
                        <p>%s</p>
                    </div>
                    <div>
                        <a href="%s" class="card-action">%s</a>
                    </div>
                </article>`,
			template.HTMLEscapeString(c.CategoryName),
			detailPath,
			template.HTMLEscapeString(c.Title),
			template.HTMLEscapeString(c.Summary),
			detailPath,
			readMoreText,
		))
	}

	homeHTML += cardsBuilder.String() + fmt.Sprintf(`
            </div>
        </section>
    </main>
%s
</body>
</html>`, footerHTML)

	if err := os.WriteFile(filepath.Join(localeDir, "index.html"), []byte(homeHTML), 0644); err != nil {
		return err
	}

	// 2. Render Detail Pages: <locale>/<cat-slug>/<content-slug>/index.html
	for _, c := range contents {
		contentDir := filepath.Join(localeDir, c.CategorySlug, c.Slug)
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			return err
		}

		var sourcesHTML strings.Builder
		if len(c.Sources) > 0 {
			sourcesHTML.WriteString(fmt.Sprintf(`
        <div class="source-box">
            <strong>%s</strong>
            <ul class="source-list">`, sourcesHeader))
			for _, src := range c.Sources {
				if src.URL != "" {
					sourcesHTML.WriteString(fmt.Sprintf(`<li>%s — <a href="%s" target="_blank" rel="nofollow noopener noreferrer">%s</a></li>`, template.HTMLEscapeString(src.Name), template.HTMLEscapeString(src.URL), template.HTMLEscapeString(src.URL)))
				} else {
					sourcesHTML.WriteString(fmt.Sprintf(`<li>%s</li>`, template.HTMLEscapeString(src.Name)))
				}
			}
			sourcesHTML.WriteString(`</ul></div>`)
		}

		canonical := fmt.Sprintf("%s/%s/%s/%s/", g.baseURL, locale, c.CategorySlug, c.Slug)
		detailHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s — %s</title>
    <meta name="description" content="%s">
    <link rel="canonical" href="%s">
    <style>%s</style>
</head>
<body>
%s
    <main class="site-wrapper" style="max-width: 800px; padding-top: 1.5rem;">
        <p><a href="/%s/" class="back-btn">%s</a></p>
        <article class="article-header">
            <span class="badge">%s</span>
            <h1 class="article-title">%s</h1>
            <div class="article-meta">📍 %s: %s</div>
        </article>
        <article class="article-body">
            %s
            %s
        </article>
    </main>
%s
</body>
</html>`,
			locale,
			template.HTMLEscapeString(c.Title),
			template.HTMLEscapeString(siteName),
			template.HTMLEscapeString(c.Summary),
			canonical,
			commonCSS,
			headerHTML,
			locale,
			backHomeText,
			template.HTMLEscapeString(c.CategoryName),
			template.HTMLEscapeString(c.Title),
			locationLabel,
			template.HTMLEscapeString(c.Address),
			string(c.BodyHTML),
			sourcesHTML.String(),
			footerHTML,
		)

		if err := os.WriteFile(filepath.Join(contentDir, "index.html"), []byte(detailHTML), 0644); err != nil {
			return err
		}
	}

	// 3. Render Static Pages: <locale>/<page-slug>/index.html
	for _, p := range pages {
		pageDir := filepath.Join(localeDir, p.Slug)
		if err := os.MkdirAll(pageDir, 0755); err != nil {
			return err
		}

		canonical := fmt.Sprintf("%s/%s/%s/", g.baseURL, locale, p.Slug)
		pageHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s — %s</title>
    <meta name="description" content="%s">
    <link rel="canonical" href="%s">
    <style>%s</style>
</head>
<body>
%s
    <main class="site-wrapper" style="max-width: 800px; padding-top: 1.5rem;">
        <p><a href="/%s/" class="back-btn">%s</a></p>
        <article class="article-header">
            <h1 class="article-title">%s</h1>
        </article>
        <article class="article-body">
            %s
        </article>
    </main>
%s
</body>
</html>`,
			locale,
			template.HTMLEscapeString(p.Title),
			template.HTMLEscapeString(siteName),
			template.HTMLEscapeString(p.SeoDesc),
			canonical,
			commonCSS,
			headerHTML,
			locale,
			backHomeText,
			template.HTMLEscapeString(p.Title),
			string(p.BodyHTML),
			footerHTML,
		)

		if err := os.WriteFile(filepath.Join(pageDir, "index.html"), []byte(pageHTML), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) renderSitemap(
	targetDir string,
	catsID, catsEN map[int64]CategoryRender,
	contentsID, contentsEN []ContentRender,
	pagesID, pagesEN []StaticPageRender,
) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
`)

	// Homepage ID & EN
	sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s/id/</loc>
    <xhtml:link rel="alternate" hreflang="id" href="%s/id/"/>
    <xhtml:link rel="alternate" hreflang="en" href="%s/en/"/>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>%s/en/</loc>
    <xhtml:link rel="alternate" hreflang="id" href="%s/id/"/>
    <xhtml:link rel="alternate" hreflang="en" href="%s/en/"/>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
`, g.baseURL, g.baseURL, g.baseURL, g.baseURL, g.baseURL, g.baseURL))

	// Contents ID
	for _, c := range contentsID {
		url := fmt.Sprintf("%s/id/%s/%s/", g.baseURL, c.CategorySlug, c.Slug)
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
`, url))
	}

	// Contents EN
	for _, c := range contentsEN {
		url := fmt.Sprintf("%s/en/%s/%s/", g.baseURL, c.CategorySlug, c.Slug)
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
`, url))
	}

	// Pages ID & EN
	for _, p := range pagesID {
		url := fmt.Sprintf("%s/id/%s/", g.baseURL, p.Slug)
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <changefreq>monthly</changefreq>
    <priority>0.5</priority>
  </url>
`, url))
	}
	for _, p := range pagesEN {
		url := fmt.Sprintf("%s/en/%s/", g.baseURL, p.Slug)
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <changefreq>monthly</changefreq>
    <priority>0.5</priority>
  </url>
`, url))
	}

	sb.WriteString("</urlset>\n")

	return os.WriteFile(filepath.Join(targetDir, "sitemap.xml"), []byte(sb.String()), 0644)
}
