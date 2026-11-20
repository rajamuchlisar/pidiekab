package sitegeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

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
	LogoURL        string
	FaviconURL     string
	SocialLinks    map[string]string
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

type MenuItemRender struct {
	ID           int64
	TitleID      string
	TitleEN      string
	ItemType     string // static_page, content, category, external
	TargetID     *int64
	ExternalURL  string
	OpenInNewTab bool
	SortOrder    int
}

type CurationData struct {
	FeaturedIDs   []int64
	LeadID        *int64
	SupportingIDs []int64
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

	// 4b. Fetch navigation menu items
	menuItems, _ := g.fetchMenuItems(ctx)

	// 4c. Fetch curation data for ID and EN
	curationID, err := g.fetchCurationData(ctx, "id")
	if err != nil {
		return fmt.Errorf("failed to load curation data for ID: %w", err)
	}
	curationEN, err := g.fetchCurationData(ctx, "en")
	if err != nil {
		return fmt.Errorf("failed to load curation data for EN: %w", err)
	}

	// 5. Render Indonesian site
	if err := g.renderLocaleSite(targetDir, "id", siteCtx, categoriesID, categoriesEN, contentsID, pagesID, menuItems, curationID); err != nil {
		return fmt.Errorf("failed to render ID site: %w", err)
	}

	// 6. Render English site
	if err := g.renderLocaleSite(targetDir, "en", siteCtx, categoriesEN, categoriesID, contentsEN, pagesEN, menuItems, curationEN); err != nil {
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
			COALESCE(s.contact_email, ''), COALESCE(s.contact_phone, ''), COALESCE(s.contact_address, ''),
			COALESCE(s.logo_url, CASE WHEN m_logo.storage_key IS NOT NULL THEN '/media/' || m_logo.storage_key ELSE '' END, ''),
			COALESCE(s.favicon_url, CASE WHEN m_fav.storage_key IS NOT NULL THEN '/media/' || m_fav.storage_key ELSE '' END, ''),
			COALESCE(s.social_links::text, '{}')
		FROM site_settings s
		LEFT JOIN media_assets m_logo ON s.logo_media_id = m_logo.id
		LEFT JOIN media_assets m_fav ON s.favicon_media_id = m_fav.id
		WHERE s.id = 1
	`
	var socialStr string
	_ = g.pool.QueryRow(ctx, query).Scan(&sc.ContactEmail, &sc.ContactPhone, &sc.ContactAddress, &sc.LogoURL, &sc.FaviconURL, &socialStr)
	sc.SocialLinks = make(map[string]string)
	if socialStr != "" {
		_ = json.Unmarshal([]byte(socialStr), &sc.SocialLinks)
	}

	tRows, err := g.pool.Query(ctx, `SELECT locale, site_name, COALESCE(tagline, ''), COALESCE(default_seo_title, '') FROM site_setting_translations WHERE site_settings_id = 1`)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var loc, name, tag, seo string
			if err := tRows.Scan(&loc, &name, &tag, &seo); err == nil {
				switch loc {
				case "id":
					if name != "" {
						sc.SiteNameID = name
					}
					sc.TaglineID = tag
					sc.DefaultSeoID = seo
				case "en":
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

func (g *Generator) fetchMenuItems(ctx context.Context) ([]MenuItemRender, error) {
	query := `
		SELECT id, title_id, title_en, item_type, target_id, COALESCE(external_url, ''), open_in_new_tab, sort_order
		FROM menu_items
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, id ASC
	`
	rows, err := g.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MenuItemRender
	for rows.Next() {
		var it MenuItemRender
		if err := rows.Scan(
			&it.ID, &it.TitleID, &it.TitleEN, &it.ItemType, &it.TargetID,
			&it.ExternalURL, &it.OpenInNewTab, &it.SortOrder,
		); err == nil {
			items = append(items, it)
		}
	}
	return items, nil
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
		switch cr.Locale {
		case "id":
			catsID[cr.ID] = cr
		case "en":
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

		switch item.Locale {
		case "id":
			contentsID = append(contentsID, item)
		case "en":
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
		switch p.Locale {
		case "id":
			pagesID = append(pagesID, p)
		case "en":
			pagesEN = append(pagesEN, p)
		}
	}

	return pagesID, pagesEN, nil
}

func (g *Generator) fetchCurationData(ctx context.Context, locale string) (*CurationData, error) {
	cd := &CurationData{}

	// 1. Featured contents ordered by position
	fRows, err := g.pool.Query(ctx, `SELECT content_id FROM featured_contents ORDER BY position ASC`)
	if err == nil {
		defer fRows.Close()
		for fRows.Next() {
			var cid int64
			if err := fRows.Scan(&cid); err == nil {
				cd.FeaturedIDs = append(cd.FeaturedIDs, cid)
			}
		}
	}

	// 2. Story items for this locale ordered by position
	sRows, err := g.pool.Query(ctx, `SELECT content_id, role FROM homepage_story_items WHERE locale = $1 ORDER BY position ASC`, locale)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var cid int64
			var role string
			if err := sRows.Scan(&cid, &role); err == nil {
				if role == "lead" {
					cidCopy := cid
					cd.LeadID = &cidCopy
				} else if role == "supporting" {
					cd.SupportingIDs = append(cd.SupportingIDs, cid)
				}
			}
		}
	}

	return cd, nil
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

func buildHeaderNavHTML(
	locale string,
	menus []MenuItemRender,
	cats map[int64]CategoryRender,
	contents []ContentRender,
	pages []StaticPageRender,
) string {
	var sb strings.Builder
	for _, m := range menus {
		title := m.TitleID
		if locale == "en" && m.TitleEN != "" {
			title = m.TitleEN
		}

		href := ""
		switch m.ItemType {
		case "static_page":
			if m.TargetID != nil {
				for _, p := range pages {
					if p.ID == *m.TargetID {
						href = fmt.Sprintf("/%s/%s/", locale, p.Slug)
						break
					}
				}
			}
		case "content":
			if m.TargetID != nil {
				for _, c := range contents {
					if c.ID == *m.TargetID {
						href = fmt.Sprintf("/%s/%s/%s/", locale, c.CategorySlug, c.Slug)
						break
					}
				}
			}
		case "category":
			if m.TargetID != nil {
				if cat, ok := cats[*m.TargetID]; ok {
					href = fmt.Sprintf("/%s/%s/", locale, cat.Slug)
				}
			}
		case "external":
			href = m.ExternalURL
			if href == "/" || href == "/id" || href == "/en" || href == "/id/" || href == "/en/" {
				href = fmt.Sprintf("/%s/", locale)
			}
		}

		if href == "" {
			continue
		}

		targetAttr := ""
		if m.OpenInNewTab {
			targetAttr = ` target="_blank" rel="nofollow noopener noreferrer"`
		} else if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			targetAttr = ` rel="nofollow noopener noreferrer"`
		}

		sb.WriteString(fmt.Sprintf(`<li class="header-nav-item"><a href="%s"%s>%s</a></li>`, template.HTMLEscapeString(href), targetAttr, template.HTMLEscapeString(title)))
	}

	if sb.Len() == 0 {
		homeText := "Beranda"
		aboutText := "Tentang Pidie"
		if locale == "en" {
			homeText = "Home"
			aboutText = "About Pidie"
		}
		sb.WriteString(fmt.Sprintf(`<li class="header-nav-item"><a href="/%s/">%s</a></li>`, locale, homeText))
		for _, cat := range cats {
			sb.WriteString(fmt.Sprintf(`<li class="header-nav-item"><a href="/%s/%s/">%s</a></li>`, locale, cat.Slug, template.HTMLEscapeString(cat.Name)))
		}
		if len(pages) > 0 {
			sb.WriteString(fmt.Sprintf(`<li class="header-nav-item"><a href="/%s/%s/">%s</a></li>`, locale, pages[0].Slug, aboutText))
		}
	}

	return sb.String()
}

func (g *Generator) renderLocaleSite(
	targetDir, locale string,
	sc *SiteContext,
	cats map[int64]CategoryRender,
	altCats map[int64]CategoryRender,
	contents []ContentRender,
	pages []StaticPageRender,
	menus []MenuItemRender,
	curation *CurationData,
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
    gap: 1rem;
    position: relative;
}
.header-left {
    display: flex;
    align-items: center;
    min-width: 0;
    flex-shrink: 0;
}
.header-nav {
    display: flex;
    align-items: center;
}
.header-nav-list {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 1.25rem;
    align-items: center;
}
.header-nav-item a {
    color: #334155;
    text-decoration: none;
    font-size: 0.875rem;
    font-weight: 600;
    padding: 0.35rem 0.2rem;
    position: relative;
    transition: color 0.15s ease;
    white-space: nowrap;
}
.header-nav-item a:hover {
    color: #047857;
}
.header-right {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    flex-shrink: 0;
}
.mobile-nav-toggle {
    display: none;
    background: #f8fafc;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    padding: 0.35rem 0.45rem;
    color: #1e293b;
    cursor: pointer;
    align-items: center;
    justify-content: center;
}
@media (max-width: 860px) {
    .mobile-nav-toggle {
        display: inline-flex;
    }
    .header-nav {
        display: none;
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: #ffffff;
        border-bottom: 1px solid #e2e8f0;
        box-shadow: 0 4px 6px -1px rgba(0,0,0,0.08);
        padding: 1rem 1.5rem;
        z-index: 49;
    }
    .header-nav.nav-open {
        display: block;
    }
    .header-nav-list {
        flex-direction: column;
        align-items: flex-start;
        gap: 0.75rem;
    }
    .header-nav-item a {
        font-size: 0.95rem;
        display: block;
        padding: 0.4rem 0;
    }
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
.brand-logo-img {
    height: 36px;
    width: auto;
    max-width: 180px;
    object-fit: contain;
    vertical-align: middle;
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

/* Category Chips */
.category-nav-section {
    margin: 1rem 0 2rem;
}
.category-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    align-items: center;
}
.cat-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.4rem 0.85rem;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 9999px;
    font-size: 0.825rem;
    font-weight: 600;
    color: #334155;
    text-decoration: none;
    transition: all 0.15s ease;
    box-shadow: 0 1px 2px rgba(0,0,0,0.02);
}
.cat-chip:hover {
    border-color: #059669;
    color: #047857;
    background: #ecfdf5;
    transform: translateY(-1px);
}
.cat-chip-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #10b981;
}

/* Curated Stories Section */
.curated-stories-section {
    margin-bottom: 2.75rem;
}
.stories-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.25rem;
}
@media (min-width: 860px) {
    .stories-grid {
        grid-template-columns: 1.4fr 1fr;
        gap: 1.5rem;
    }
}
.lead-card {
    background: #ffffff;
    border: 2px solid #a7f3d0;
    border-radius: 14px;
    padding: 1.75rem;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    box-shadow: 0 4px 15px rgba(5, 150, 105, 0.08);
}
.lead-badges {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    flex-wrap: wrap;
}
.lead-badge {
    background: #059669;
    color: #ffffff;
    padding: 0.25rem 0.65rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
}
.lead-card h3 {
    margin: 0.25rem 0 0.75rem;
    font-size: clamp(1.2rem, 2.5vw, 1.55rem);
    line-height: 1.35;
    font-weight: 800;
    color: #064e3b;
}
.lead-card h3 a {
    color: inherit;
    text-decoration: none;
}
.lead-card h3 a:hover {
    color: #047857;
    text-decoration: underline;
}
.lead-card p {
    font-size: 0.95rem;
    color: #475569;
    line-height: 1.6;
    margin: 0 0 1.25rem;
}
.supporting-column {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}
.supporting-heading {
    margin: 0 0 0.5rem;
    font-size: 0.95rem;
    font-weight: 700;
    color: #475569;
    text-transform: uppercase;
    letter-spacing: 0.04em;
}
.supporting-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    flex: 1;
}
.supporting-card {
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    padding: 1rem 1.15rem;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    box-shadow: 0 1px 3px rgba(0,0,0,0.03);
    transition: transform 0.15s, border-color 0.15s;
}
.supporting-card:hover {
    border-color: #cbd5e1;
    transform: translateY(-2px);
    box-shadow: 0 4px 10px rgba(0,0,0,0.05);
}
.supporting-card h4 {
    margin: 0.2rem 0;
    font-size: 0.95rem;
    line-height: 1.4;
    font-weight: 700;
    color: #0f172a;
}
.supporting-card h4 a {
    color: inherit;
    text-decoration: none;
}
.supporting-card h4 a:hover {
    color: #047857;
    text-decoration: underline;
}
.supporting-card p {
    margin: 0;
    font-size: 0.8125rem;
    color: #64748b;
    line-height: 1.45;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

/* Featured Section */
.featured-section {
    margin-bottom: 2.75rem;
}
.section-header-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 1.25rem;
    flex-wrap: wrap;
    gap: 0.5rem;
}
.section-subtitle-tag {
    font-size: 0.8125rem;
    font-weight: 600;
    color: #059669;
    background: #ecfdf5;
    padding: 0.2rem 0.55rem;
    border-radius: 6px;
    border: 1px solid #d1fae5;
}
.card-featured {
    border-color: #fde68a;
    background: linear-gradient(180deg, #fffbeb 0%, #ffffff 40%);
}
.badge-featured {
    background: #fef3c7;
    color: #92400e;
    border: 1px solid #fde68a;
}
.all-articles-section {
    margin-bottom: 2.5rem;
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
    margin-top: 4.5rem;
    background: #0f172a;
    color: #94a3b8;
    font-size: 0.875rem;
    border-top: 1px solid #1e293b;
    padding-top: 3.5rem;
}
.footer-grid {
    display: grid;
    grid-template-columns: 1.4fr 1fr 1fr 1.2fr;
    gap: 2.5rem;
    padding-bottom: 2.5rem;
    text-align: left;
}
@media (max-width: 900px) {
    .footer-grid {
        grid-template-columns: 1fr 1fr;
        gap: 2rem;
    }
}
@media (max-width: 600px) {
    .footer-grid {
        grid-template-columns: 1fr;
        gap: 2rem;
    }
}
.footer-col h4 {
    color: #f8fafc;
    font-size: 0.875rem;
    font-weight: 700;
    margin: 0 0 1.25rem 0;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 2px solid #059669;
    padding-bottom: 0.4rem;
    display: inline-block;
}
.footer-brand .brand-logo {
    color: #10b981;
    margin-bottom: 0.75rem;
}
.footer-brand p {
    margin: 0.75rem 0;
    line-height: 1.6;
    color: #94a3b8;
    font-size: 0.85rem;
}
.footer-badge-disclaimer {
    display: inline-block;
    margin-top: 0.5rem;
    padding: 0.3rem 0.65rem;
    background: rgba(16, 185, 129, 0.1);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.25);
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 600;
    line-height: 1.4;
}
.footer-links {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
}
.footer-links li a {
    color: #cbd5e1;
    text-decoration: none;
    font-size: 0.875rem;
    transition: all 0.15s ease;
    display: inline-block;
}
.footer-links li a:hover {
    color: #34d399;
    padding-left: 4px;
}
.footer-contact {
    list-style: none;
    padding: 0;
    margin: 0 0 1.5rem 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}
.footer-contact-item {
    display: flex;
    gap: 0.65rem;
    align-items: flex-start;
    color: #cbd5e1;
    font-size: 0.85rem;
    line-height: 1.5;
}
.footer-contact-item a {
    color: #cbd5e1;
    text-decoration: none;
    transition: color 0.15s;
}
.footer-contact-item a:hover {
    color: #34d399;
}
.footer-icon {
    width: 17px;
    height: 17px;
    min-width: 17px;
    margin-top: 2px;
    color: #10b981;
}
.social-section {
    margin-top: 1rem;
}
.social-section h4 {
    margin-bottom: 0.75rem;
}
.social-links {
    display: flex;
    gap: 0.65rem;
    align-items: center;
}
.social-btn {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: #1e293b;
    color: #cbd5e1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
    border: 1px solid #334155;
    text-decoration: none;
}
.social-btn svg {
    width: 18px;
    height: 18px;
}
.social-btn:hover {
    background: #059669;
    color: #ffffff;
    border-color: #059669;
    transform: translateY(-2px);
}
.footer-bottom {
    border-top: 1px solid #1e293b;
    padding: 1.5rem 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 1rem;
    font-size: 0.8rem;
    color: #64748b;
    text-align: left;
}
.footer-bottom p {
    margin: 0;
}
.footer-bottom-links {
    display: flex;
    gap: 1.25rem;
}
.footer-bottom-links a {
    color: #94a3b8;
    text-decoration: none;
    font-size: 0.8rem;
}
.footer-bottom-links a:hover {
    color: #34d399;
}
`

	brandLogoHTML := `<span class="brand-logo"><span class="brand-logo-badge">PK</span>PidieKab</span>`
	if sc.LogoURL != "" {
		brandLogoHTML = fmt.Sprintf(`<span class="brand-logo"><img src="%s" alt="%s" class="brand-logo-img" /></span>`, template.HTMLEscapeString(sc.LogoURL), template.HTMLEscapeString(siteName))
	}

	headerNavHTML := buildHeaderNavHTML(locale, menus, cats, contents, pages)

	buildHeader := func(idURL, enURL string) string {
		return fmt.Sprintf(`
    <header class="site-header">
        <div class="site-wrapper">
            <div class="header-inner">
                <div class="header-left">
                    <a href="/%s/" class="brand-group">
                        %s
                        <span class="brand-tagline">%s</span>
                    </a>
                </div>
                <nav class="header-nav" id="header-menu-nav" aria-label="Navigasi Utama">
                    <ul class="header-nav-list">
                        %s
                    </ul>
                </nav>
                <div class="header-right">
                    <nav class="lang-switcher" aria-label="Language selector">
                        <a href="%s" class="lang-btn %s"><span class="lang-short">ID</span><span class="lang-long">Bahasa Indonesia</span></a>
                        <a href="%s" class="lang-btn %s"><span class="lang-short">EN</span><span class="lang-long">English</span></a>
                    </nav>
                    <button type="button" class="mobile-nav-toggle" aria-label="Buka Menu" onclick="document.getElementById('header-menu-nav').classList.toggle('nav-open')">
                        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16"/></svg>
                    </button>
                </div>
            </div>
        </div>
    </header>
`, locale, brandLogoHTML, template.HTMLEscapeString(siteName), headerNavHTML, idURL, idClass, enURL, enClass)
	}

	headerHTML := buildHeader("/id/", "/en/")

	quickLinksTitle := "Kategori Pilihan"
	legalTitle := "Informasi & Legal"
	contactTitle := "Kontak & Alamat"
	socialTitle := "Media Sosial Resmi"
	communityDisclaimer := "Platform independen promosi potensi daerah & direktori informasi Kabupaten Pidie, Aceh. Bukan website resmi pemerintah."
	allRightsReserved := "Hak Cipta Dilindungi."
	homeNavText := "Beranda"
	if locale == "en" {
		quickLinksTitle = "Featured Categories"
		legalTitle = "Information & Legal"
		contactTitle = "Contact & Secretariat"
		socialTitle = "Official Social Media"
		communityDisclaimer = "Independent regional promotion & digital visitor guide for Pidie Regency, Aceh. Not an official government website."
		allRightsReserved = "All Rights Reserved."
		homeNavText = "Home"
	}

	var catLinksHTML strings.Builder
	var sortedCats []CategoryRender
	for _, c := range cats {
		sortedCats = append(sortedCats, c)
	}
	sort.Slice(sortedCats, func(i, j int) bool {
		return sortedCats[i].ID < sortedCats[j].ID
	})
	for _, c := range sortedCats {
		catLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/%s/">%s</a></li>`, locale, c.Slug, template.HTMLEscapeString(c.Name)))
	}

	var pageLinksHTML strings.Builder
	if len(pages) > 0 {
		for _, p := range pages {
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/%s/">%s</a></li>`, locale, p.Slug, template.HTMLEscapeString(p.Title)))
		}
	} else {
		if locale == "en" {
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/privacy-policy/">Privacy Policy</a></li>`, locale))
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/terms-and-conditions/">Terms and Conditions</a></li>`, locale))
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/about-us/">About Us</a></li>`, locale))
		} else {
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/kebijakan-privasi/">Kebijakan Privasi</a></li>`, locale))
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/syarat-dan-ketentuan/">Syarat dan Ketentuan</a></li>`, locale))
			pageLinksHTML.WriteString(fmt.Sprintf(`<li><a href="/%s/tentang-kami/">Tentang Kami</a></li>`, locale))
		}
	}

	var contactHTML strings.Builder
	if sc.ContactAddress != "" {
		contactHTML.WriteString(fmt.Sprintf(`<li class="footer-contact-item"><svg class="footer-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M9.69 18.933l.003.001C9.89 19.02 10 19 10 19s.11.02.308-.066l.002-.001.006-.003.018-.008a5.741 5.741 0 00.281-.14c.186-.096.446-.24.757-.433.62-.384 1.445-.966 2.274-1.765C15.302 14.988 17 12.493 17 9A7 7 0 103 9c0 3.492 1.698 5.988 3.36 7.587a15.42 15.42 0 002.273 1.765 11.75 11.75 0 001.04.573l.018.008.006.003zM10 11.25a2.25 2.25 0 100-4.5 2.25 2.25 0 000 4.5z" clip-rule="evenodd"/></svg><span>%s</span></li>`, template.HTMLEscapeString(sc.ContactAddress)))
	}
	if sc.ContactEmail != "" {
		contactHTML.WriteString(fmt.Sprintf(`<li class="footer-contact-item"><svg class="footer-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M3 4a2 2 0 00-2 2v1.161l8.441 4.221a1.25 1.25 0 001.118 0L19 7.162V6a2 2 0 00-2-2H3z"/><path d="M19 8.839l-7.77 3.885a2.75 2.75 0 01-2.46 0L1 8.839V14a2 2 0 002 2h14a2 2 0 002-2V8.839z"/></svg><a href="mailto:%s">%s</a></li>`, template.HTMLEscapeString(sc.ContactEmail), template.HTMLEscapeString(sc.ContactEmail)))
	}
	if sc.ContactPhone != "" {
		contactHTML.WriteString(fmt.Sprintf(`<li class="footer-contact-item"><svg class="footer-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M2 3.5A1.5 1.5 0 013.5 2h1.148a1.5 1.5 0 011.465 1.175l.716 3.223a1.5 1.5 0 01-1.052 1.767l-.933.267c-.41.117-.643.555-.48.95a11.542 11.542 0 006.254 6.254c.395.163.833-.07.95-.48l.267-.933a1.5 1.5 0 011.767-1.052l3.223.716A1.5 1.5 0 0118 15.352V16.5a1.5 1.5 0 01-1.5 1.5H15c-1.149 0-2.263-.15-3.326-.43A13.022 13.022 0 012.43 8.326 13.019 13.019 0 012 5V3.5z" clip-rule="evenodd"/></svg><a href="tel:%s">%s</a></li>`, template.HTMLEscapeString(sc.ContactPhone), template.HTMLEscapeString(sc.ContactPhone)))
	}
	if contactHTML.Len() == 0 {
		contactHTML.WriteString(`<li class="footer-contact-item"><svg class="footer-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M9.69 18.933l.003.001C9.89 19.02 10 19 10 19s.11.02.308-.066l.002-.001.006-.003.018-.008a5.741 5.741 0 00.281-.14c.186-.096.446-.24.757-.433.62-.384 1.445-.966 2.274-1.765C15.302 14.988 17 12.493 17 9A7 7 0 103 9c0 3.492 1.698 5.988 3.36 7.587a15.42 15.42 0 002.273 1.765 11.75 11.75 0 001.04.573l.018.008.006.003zM10 11.25a2.25 2.25 0 100-4.5 2.25 2.25 0 000 4.5z" clip-rule="evenodd"/></svg><span>Sigli, Kabupaten Pidie, Aceh, Indonesia</span></li><li class="footer-contact-item"><svg class="footer-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M3 4a2 2 0 00-2 2v1.161l8.441 4.221a1.25 1.25 0 001.118 0L19 7.162V6a2 2 0 00-2-2H3z"/><path d="M19 8.839l-7.77 3.885a2.75 2.75 0 01-2.46 0L1 8.839V14a2 2 0 002 2h14a2 2 0 002-2V8.839z"/></svg><a href="mailto:kontak@pidiekab.id">kontak@pidiekab.id</a></li>`)
	}

	var socialLinksHTML strings.Builder
	if ig, ok := sc.SocialLinks["instagram"]; ok && strings.TrimSpace(ig) != "" {
		igURL := strings.TrimSpace(ig)
		if !strings.HasPrefix(igURL, "http") {
			igURL = "https://instagram.com/" + strings.TrimPrefix(igURL, "@")
		}
		socialLinksHTML.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="nofollow noopener noreferrer" class="social-btn" title="Instagram" aria-label="Instagram"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838c-3.403 0-6.162 2.759-6.162 6.162s2.759 6.163 6.162 6.163 6.162-2.759 6.162-6.163c0-3.403-2.759-6.162-6.162-6.162zm0 10.162c-2.209 0-4-1.79-4-4 0-2.209 1.791-4 4-4s4 1.791 4 4c0 2.21-1.791 4-4 4zm6.406-11.845c-.796 0-1.441.645-1.441 1.44s.645 1.44 1.441 1.44c.795 0 1.439-.645 1.439-1.44s-.644-1.44-1.439-1.44z"/></svg></a>`, template.HTMLEscapeString(igURL)))
	}
	if fb, ok := sc.SocialLinks["facebook"]; ok && strings.TrimSpace(fb) != "" {
		fbURL := strings.TrimSpace(fb)
		if !strings.HasPrefix(fbURL, "http") {
			fbURL = "https://" + fbURL
		}
		socialLinksHTML.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="nofollow noopener noreferrer" class="social-btn" title="Facebook" aria-label="Facebook"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg></a>`, template.HTMLEscapeString(fbURL)))
	}
	if yt, ok := sc.SocialLinks["youtube"]; ok && strings.TrimSpace(yt) != "" {
		ytURL := strings.TrimSpace(yt)
		if !strings.HasPrefix(ytURL, "http") {
			ytURL = "https://" + ytURL
		}
		socialLinksHTML.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="nofollow noopener noreferrer" class="social-btn" title="YouTube" aria-label="YouTube"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg></a>`, template.HTMLEscapeString(ytURL)))
	}
	if tw, ok := sc.SocialLinks["twitter"]; ok && strings.TrimSpace(tw) != "" {
		twURL := strings.TrimSpace(tw)
		if !strings.HasPrefix(twURL, "http") {
			twURL = "https://x.com/" + strings.TrimPrefix(twURL, "@")
		}
		socialLinksHTML.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="nofollow noopener noreferrer" class="social-btn" title="X / Twitter" aria-label="X"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg></a>`, template.HTMLEscapeString(twURL)))
	}
	if socialLinksHTML.Len() == 0 {
		socialLinksHTML.WriteString(`<a href="https://instagram.com" target="_blank" rel="nofollow noopener noreferrer" class="social-btn" title="Instagram" aria-label="Instagram"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838c-3.403 0-6.162 2.759-6.162 6.162s2.759 6.163 6.162 6.163 6.162-2.759 6.162-6.163c0-3.403-2.759-6.162-6.162-6.162zm0 10.162c-2.209 0-4-1.79-4-4 0-2.209 1.791-4 4-4s4 1.791 4 4c0 2.21-1.791 4-4 4zm6.406-11.845c-.796 0-1.441.645-1.441 1.44s.645 1.44 1.441 1.44c.795 0 1.439-.645 1.439-1.44s-.644-1.44-1.439-1.44z"/></svg></a>`)
	}

	footerHTML := fmt.Sprintf(`
    <footer class="site-footer">
        <div class="site-wrapper">
            <div class="footer-grid">
                <!-- Column 1: Brand & Disclaimer -->
                <div class="footer-col footer-brand">
                    <div class="brand-group">
                        %s
                    </div>
                    <p>%s</p>
                    <span class="footer-badge-disclaimer">%s</span>
                </div>

                <!-- Column 2: Quicklinks (Categories) -->
                <div class="footer-col">
                    <h4>%s</h4>
                    <ul class="footer-links">
                        %s
                    </ul>
                </div>

                <!-- Column 3: Legal & Static Pages -->
                <div class="footer-col">
                    <h4>%s</h4>
                    <ul class="footer-links">
                        %s
                    </ul>
                </div>

                <!-- Column 4: Contact & Social Media -->
                <div class="footer-col">
                    <h4>%s</h4>
                    <ul class="footer-contact">
                        %s
                    </ul>
                    <div class="social-section">
                        <h4>%s</h4>
                        <div class="social-links">
                            %s
                        </div>
                    </div>
                </div>
            </div>

            <div class="footer-bottom">
                <p>&copy; %d PidieKab. %s</p>
                <div class="footer-bottom-links">
                    <a href="/%s/">%s</a>
                    <a href="/sitemap.xml">Sitemap</a>
                </div>
            </div>
        </div>
    </footer>
`,
		brandLogoHTML,
		template.HTMLEscapeString(tagline),
		communityDisclaimer,
		quickLinksTitle,
		catLinksHTML.String(),
		legalTitle,
		pageLinksHTML.String(),
		contactTitle,
		contactHTML.String(),
		socialTitle,
		socialLinksHTML.String(),
		time.Now().Year(),
		allRightsReserved,
		locale,
		homeNavText,
	)

	faviconTag := ""
	if sc.FaviconURL != "" {
		faviconTag = fmt.Sprintf("\n    <link rel=\"icon\" href=\"%s\">", template.HTMLEscapeString(sc.FaviconURL))
	}

	// Build quick category chips
	var catChipsBuilder strings.Builder
	for _, c := range sortedCats {
		catChipsBuilder.WriteString(fmt.Sprintf(
			`<a href="/%s/%s/" class="cat-chip"><span class="cat-chip-dot"></span>%s</a>`,
			locale, c.Slug, template.HTMLEscapeString(c.Name),
		))
	}

	// Prepare curated contents
	contentByID := make(map[int64]ContentRender, len(contents))
	for _, c := range contents {
		contentByID[c.ID] = c
	}

	var leadStory *ContentRender
	var supportingStories []ContentRender
	var featuredStories []ContentRender
	handledStoryIDs := make(map[int64]bool)

	if curation != nil {
		if curation.LeadID != nil {
			if ls, ok := contentByID[*curation.LeadID]; ok {
				leadStory = &ls
				handledStoryIDs[*curation.LeadID] = true
			}
		}
		for _, sID := range curation.SupportingIDs {
			if handledStoryIDs[sID] {
				continue
			}
			if ss, ok := contentByID[sID]; ok {
				supportingStories = append(supportingStories, ss)
				handledStoryIDs[sID] = true
			}
		}
		for _, fID := range curation.FeaturedIDs {
			if fc, ok := contentByID[fID]; ok {
				featuredStories = append(featuredStories, fc)
			}
		}
	}

	storiesSectionTitle := "Cerita tentang Pidie"
	leadBadgeText := "Cerita Utama"
	readLeadText := "Baca Kisah Utama &rarr;"
	supportingHeading := "Sorotan Terkait"
	featuredSectionTitle := "Pilihan Redaksi"
	featuredBadgeText := "Pilihan Redaksi"
	featuredTagText := "Rekomendasi Terverifikasi"
	allSectionTitle := "Eksplorasi Destinasi & Potensi Daerah"
	if locale == "en" {
		storiesSectionTitle = "Stories of Pidie"
		leadBadgeText = "Lead Story"
		readLeadText = "Read Lead Story &rarr;"
		supportingHeading = "Related Highlights"
		featuredSectionTitle = "Editor's Picks"
		featuredBadgeText = "Editor's Pick"
		featuredTagText = "Verified Recommendations"
		allSectionTitle = "Explore All Destinations & Heritage"
	}

	// Build Curated Stories Section (Lead & Supporting)
	var storiesHTML strings.Builder
	if leadStory != nil || len(supportingStories) > 0 {
		storiesHTML.WriteString(fmt.Sprintf(`
        <section class="curated-stories-section">
            <h2 class="section-title">%s</h2>
            <div class="stories-grid">`, storiesSectionTitle))

		if leadStory != nil {
			leadDetailPath := fmt.Sprintf("/%s/%s/%s/", locale, leadStory.CategorySlug, leadStory.Slug)
			storiesHTML.WriteString(fmt.Sprintf(`
                <article class="lead-card">
                    <div>
                        <div class="lead-badges">
                            <span class="lead-badge">%s</span>
                            <span class="badge">%s</span>
                        </div>
                        <h3><a href="%s">%s</a></h3>
                        <p>%s</p>
                    </div>
                    <div>
                        <a href="%s" class="card-action">%s</a>
                    </div>
                </article>`,
				leadBadgeText,
				template.HTMLEscapeString(leadStory.CategoryName),
				leadDetailPath,
				template.HTMLEscapeString(leadStory.Title),
				template.HTMLEscapeString(leadStory.Summary),
				leadDetailPath,
				readLeadText,
			))
		}

		if len(supportingStories) > 0 {
			storiesHTML.WriteString(fmt.Sprintf(`
                <div class="supporting-column">
                    <h3 class="supporting-heading">%s</h3>
                    <div class="supporting-list">`, supportingHeading))
			for _, ss := range supportingStories {
				sDetailPath := fmt.Sprintf("/%s/%s/%s/", locale, ss.CategorySlug, ss.Slug)
				storiesHTML.WriteString(fmt.Sprintf(`
                        <article class="supporting-card">
                            <span class="badge">%s</span>
                            <h4><a href="%s">%s</a></h4>
                            <p>%s</p>
                        </article>`,
					template.HTMLEscapeString(ss.CategoryName),
					sDetailPath,
					template.HTMLEscapeString(ss.Title),
					template.HTMLEscapeString(ss.Summary),
				))
			}
			storiesHTML.WriteString(`
                    </div>
                </div>`)
		}

		storiesHTML.WriteString(`
            </div>
        </section>`)
	}

	// Build Curated Featured Section
	var featuredHTML strings.Builder
	if len(featuredStories) > 0 {
		featuredHTML.WriteString(fmt.Sprintf(`
        <section class="featured-section">
            <div class="section-header-row">
                <h2 class="section-title">%s</h2>
                <span class="section-subtitle-tag">%s</span>
            </div>
            <div class="card-grid">`, featuredSectionTitle, featuredTagText))

		for _, fc := range featuredStories {
			fDetailPath := fmt.Sprintf("/%s/%s/%s/", locale, fc.CategorySlug, fc.Slug)
			featuredHTML.WriteString(fmt.Sprintf(`
                <article class="card card-featured">
                    <div>
                        <div class="lead-badges">
                            <span class="badge badge-featured">%s</span>
                            <span class="badge">%s</span>
                        </div>
                        <h3><a href="%s">%s</a></h3>
                        <p>%s</p>
                    </div>
                    <div>
                        <a href="%s" class="card-action">%s</a>
                    </div>
                </article>`,
				featuredBadgeText,
				template.HTMLEscapeString(fc.CategoryName),
				fDetailPath,
				template.HTMLEscapeString(fc.Title),
				template.HTMLEscapeString(fc.Summary),
				fDetailPath,
				readMoreText,
			))
		}

		featuredHTML.WriteString(`
            </div>
        </section>`)
	}

	// Build General / Remaining Contents Section
	var remainingContents []ContentRender
	hasCuration := (leadStory != nil || len(supportingStories) > 0 || len(featuredStories) > 0)
	for _, c := range contents {
		if hasCuration && handledStoryIDs[c.ID] {
			continue
		}
		remainingContents = append(remainingContents, c)
	}
	if len(remainingContents) == 0 && len(contents) > 0 {
		remainingContents = contents
	}

	var allArticlesHTML strings.Builder
	if len(remainingContents) > 0 {
		displayTitle := allSectionTitle
		if !hasCuration {
			displayTitle = sectionTitle
		}
		allArticlesHTML.WriteString(fmt.Sprintf(`
        <section class="all-articles-section">
            <h2 class="section-title">%s</h2>
            <div class="card-grid">`, displayTitle))

		for _, c := range remainingContents {
			detailPath := fmt.Sprintf("/%s/%s/%s/", locale, c.CategorySlug, c.Slug)
			allArticlesHTML.WriteString(fmt.Sprintf(`
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

		allArticlesHTML.WriteString(`
            </div>
        </section>`)
	}

	// 1. Render Homepage: <locale>/index.html
	homeHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s — %s</title>
    <meta name="description" content="%s">%s
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
        <section class="category-nav-section" aria-label="Kategori">
            <div class="category-chips">
%s
            </div>
        </section>
%s
%s
%s
    </main>
%s
</body>
</html>`,
		locale,
		template.HTMLEscapeString(siteName),
		template.HTMLEscapeString(tagline),
		template.HTMLEscapeString(tagline),
		faviconTag,
		g.baseURL, locale,
		g.baseURL, g.baseURL, g.baseURL,
		commonCSS,
		headerHTML,
		template.HTMLEscapeString(tagline),
		template.HTMLEscapeString(heroSubtitle),
		catChipsBuilder.String(),
		storiesHTML.String(),
		featuredHTML.String(),
		allArticlesHTML.String(),
		footerHTML,
	)

	if err := os.WriteFile(filepath.Join(localeDir, "index.html"), []byte(homeHTML), 0644); err != nil {
		return err
	}

	// 2. Render Category Archive Pages: <locale>/<cat-slug>/index.html
	for _, cat := range cats {
		catDir := filepath.Join(localeDir, cat.Slug)
		if err := os.MkdirAll(catDir, 0755); err != nil {
			return err
		}

		var catContents []ContentRender
		for _, c := range contents {
			if c.CategorySlug == cat.Slug {
				catContents = append(catContents, c)
			}
		}

		altSlug := ""
		if altCat, ok := altCats[cat.ID]; ok && altCat.Slug != "" {
			altSlug = altCat.Slug
		}

		catIDURL := fmt.Sprintf("/id/%s/", cat.Slug)
		catENURL := fmt.Sprintf("/en/%s/", cat.Slug)
		if locale == "id" {
			if altSlug != "" {
				catENURL = fmt.Sprintf("/en/%s/", altSlug)
			} else {
				catENURL = "/en/"
			}
		} else {
			catENURL = fmt.Sprintf("/en/%s/", cat.Slug)
			if altSlug != "" {
				catIDURL = fmt.Sprintf("/id/%s/", altSlug)
			} else {
				catIDURL = "/id/"
			}
		}

		catHeaderHTML := buildHeader(catIDURL, catENURL)

		catCanonical := fmt.Sprintf("%s/%s/%s/", g.baseURL, locale, cat.Slug)
		catAlternateID := fmt.Sprintf("%s/id/%s/", g.baseURL, cat.Slug)
		catAlternateEN := fmt.Sprintf("%s/en/%s/", g.baseURL, cat.Slug)
		if locale == "id" && altSlug != "" {
			catAlternateEN = fmt.Sprintf("%s/en/%s/", g.baseURL, altSlug)
		} else if locale == "en" && altSlug != "" {
			catAlternateID = fmt.Sprintf("%s/id/%s/", g.baseURL, altSlug)
		}

		catDesc := cat.Description
		if catDesc == "" {
			catDesc = tagline
		}

		categoryBadgeLabel := "Kategori"
		articlesSectionTitle := "Daftar Artikel & Panduan Destinasi"
		emptyTitle := "Belum Ada Artikel dalam Kategori Ini"
		emptyDesc := "Konten editorial dan panduan terverifikasi untuk kategori ini sedang disiapkan oleh tim redaksi."
		if locale == "en" {
			categoryBadgeLabel = "Category"
			articlesSectionTitle = "Articles & Discovery Guides"
			emptyTitle = "No Articles in This Category Yet"
			emptyDesc = "Verified editorial guides and articles for this category are currently being curated."
		}

		var catCardsBuilder strings.Builder
		if len(catContents) > 0 {
			catCardsBuilder.WriteString(`<div class="card-grid">`)
			for _, c := range catContents {
				detailPath := fmt.Sprintf("/%s/%s/%s/", locale, c.CategorySlug, c.Slug)
				catCardsBuilder.WriteString(fmt.Sprintf(`
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
			catCardsBuilder.WriteString(`</div>`)
		} else {
			catCardsBuilder.WriteString(fmt.Sprintf(`
            <div style="background:#ffffff;border:1px dashed #cbd5e1;border-radius:12px;padding:3rem 1.5rem;text-align:center;color:#64748b;">
                <svg style="width:48px;height:48px;margin:0 auto 1rem;color:#94a3b8;display:block;" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
                <h3 style="font-size:1.1rem;font-weight:700;color:#334155;margin:0 0 0.5rem;">%s</h3>
                <p style="font-size:0.9rem;margin:0;max-width:480px;margin-left:auto;margin-right:auto;">%s</p>
            </div>`,
				template.HTMLEscapeString(emptyTitle),
				template.HTMLEscapeString(emptyDesc),
			))
		}

		categoryHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s — %s</title>
    <meta name="description" content="%s">%s
    <link rel="canonical" href="%s">
    <link rel="alternate" hreflang="id" href="%s">
    <link rel="alternate" hreflang="en" href="%s">
    <link rel="alternate" hreflang="x-default" href="%s">
    <style>%s</style>
</head>
<body>
%s
    <main class="site-wrapper" style="padding-top: 1.5rem;">
        <p><a href="/%s/" class="back-btn">%s</a></p>
        <section class="hero" style="margin-top: 0.5rem;">
            <span class="badge" style="background: rgba(5,150,105,0.15); color: #047857;">%s</span>
            <h2>%s</h2>
            <p>%s</p>
        </section>
        <section style="margin-top: 2rem;">
            <h2 class="section-title">%s</h2>
            %s
        </section>
    </main>
%s
</body>
</html>`,
			locale,
			template.HTMLEscapeString(cat.Name),
			template.HTMLEscapeString(siteName),
			template.HTMLEscapeString(catDesc),
			faviconTag,
			catCanonical,
			catAlternateID,
			catAlternateEN,
			catAlternateID,
			commonCSS,
			catHeaderHTML,
			locale,
			backHomeText,
			categoryBadgeLabel,
			template.HTMLEscapeString(cat.Name),
			template.HTMLEscapeString(catDesc),
			template.HTMLEscapeString(articlesSectionTitle),
			catCardsBuilder.String(),
			footerHTML,
		)

		if err := os.WriteFile(filepath.Join(catDir, "index.html"), []byte(categoryHTML), 0644); err != nil {
			return err
		}
	}

	// 3. Render Detail Pages: <locale>/<cat-slug>/<content-slug>/index.html
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
    <meta name="description" content="%s">%s
    <link rel="canonical" href="%s">
    <style>%s</style>
</head>
<body>
%s
    <main class="site-wrapper" style="max-width: 800px; padding-top: 1.5rem;">
        <p style="display:flex;align-items:center;gap:0.4rem;flex-wrap:wrap;font-size:0.85rem;margin-bottom:1.25rem;">
            <a href="/%s/" class="back-btn" style="margin-bottom:0;">%s</a>
            <span style="color:#94a3b8;">/</span>
            <a href="/%s/%s/" class="back-btn" style="margin-bottom:0;">%s</a>
        </p>
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
			faviconTag,
			canonical,
			commonCSS,
			headerHTML,
			locale,
			backHomeText,
			locale,
			c.CategorySlug,
			template.HTMLEscapeString(c.CategoryName),
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

	// 4. Render Static Pages: <locale>/<page-slug>/index.html
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
    <meta name="description" content="%s">%s
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
			faviconTag,
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

	// Categories ID
	for catID, c := range catsID {
		url := fmt.Sprintf("%s/id/%s/", g.baseURL, c.Slug)
		altURL := ""
		if alt, ok := catsEN[catID]; ok && alt.Slug != "" {
			altURL = fmt.Sprintf("%s/en/%s/", g.baseURL, alt.Slug)
		}
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <xhtml:link rel="alternate" hreflang="id" href="%s"/>`, url, url))
		if altURL != "" {
			sb.WriteString(fmt.Sprintf(`
    <xhtml:link rel="alternate" hreflang="en" href="%s"/>`, altURL))
		}
		sb.WriteString(`
    <changefreq>daily</changefreq>
    <priority>0.9</priority>
  </url>
`)
	}

	// Categories EN
	for catID, c := range catsEN {
		url := fmt.Sprintf("%s/en/%s/", g.baseURL, c.Slug)
		altURL := ""
		if alt, ok := catsID[catID]; ok && alt.Slug != "" {
			altURL = fmt.Sprintf("%s/id/%s/", g.baseURL, alt.Slug)
		}
		sb.WriteString(fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <xhtml:link rel="alternate" hreflang="en" href="%s"/>`, url, url))
		if altURL != "" {
			sb.WriteString(fmt.Sprintf(`
    <xhtml:link rel="alternate" hreflang="id" href="%s"/>`, altURL))
		}
		sb.WriteString(`
    <changefreq>daily</changefreq>
    <priority>0.9</priority>
  </url>
`)
	}

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
