package sitegeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Generator struct {
	pool    *pgxpool.Pool
	baseURL string
}

func NewGenerator(pool *pgxpool.Pool, baseURL string) *Generator {
	if baseURL == "" {
		baseURL = "https://pidiekab.go.id"
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
		SiteNameID: "Portal Wisata & Budaya Kabupaten Pidie",
		SiteNameEN: "Pidie Regency Tourism & Cultural Portal",
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

func renderBodyBlocksToHTML(bodyBytes []byte) template.HTML {
	if len(bodyBytes) == 0 {
		return ""
	}

	var parsed map[string]any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return template.HTML("<p>" + template.HTMLEscapeString(string(bodyBytes)) + "</p>")
	}

	if htmlStr, ok := parsed["html"].(string); ok && htmlStr != "" {
		return template.HTML(htmlStr)
	}

	blocks, ok := parsed["blocks"].([]any)
	if !ok {
		if mdStr, ok := parsed["markdown"].(string); ok && mdStr != "" {
			return template.HTML("<p>" + template.HTMLEscapeString(mdStr) + "</p>")
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

	return template.HTML(sb.String())
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
	if locale == "en" {
		siteName = sc.SiteNameEN
		tagline = sc.TaglineEN
	}

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
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; max-width: 1024px; margin: 0 auto; padding: 1.5rem; color: #1a1a1a; }
        header { border-bottom: 2px solid #047857; padding-bottom: 1rem; margin-bottom: 2rem; }
        .hero { background: #f0fdf4; padding: 2rem; border-radius: 8px; margin-bottom: 2rem; border-left: 4px solid #059669; }
        .card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1.5rem; }
        .card { border: 1px solid #e5e7eb; border-radius: 8px; padding: 1.25rem; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,0.05); }
        .card h3 { margin-top: 0; color: #065f46; }
        .badge { display: inline-block; background: #e0e7ff; color: #3730a3; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: bold; }
        footer { margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid #e5e7eb; font-size: 0.875rem; color: #6b7280; }
        a { color: #047857; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <header>
        <div style="display:flex; justify-content:space-between; align-items:center;">
            <h1><a href="/%s/">%s</a></h1>
            <nav>
                <a href="/id/">Bahasa Indonesia</a> | <a href="/en/">English</a>
            </nav>
        </div>
    </header>
    <main>
        <section class="hero">
            <h2>%s</h2>
            <p>Jelajahi keindahan alam, kekayaan budaya, dan kuliner khas Kabupaten Pidie, Aceh.</p>
        </section>
        <section>
            <h2>Destinasi & Budaya</h2>
            <div class="card-grid">
`, locale, template.HTMLEscapeString(siteName), template.HTMLEscapeString(tagline), template.HTMLEscapeString(tagline), g.baseURL, locale, g.baseURL, g.baseURL, g.baseURL, locale, template.HTMLEscapeString(siteName), template.HTMLEscapeString(tagline))

	var cardsBuilder strings.Builder
	for _, c := range contents {
		detailPath := fmt.Sprintf("/%s/%s/%s/", locale, c.CategorySlug, c.Slug)
		cardsBuilder.WriteString(fmt.Sprintf(`
                <article class="card">
                    <span class="badge">%s</span>
                    <h3><a href="%s">%s</a></h3>
                    <p>%s</p>
                    <p><a href="%s">Selengkapnya &rarr;</a></p>
                </article>`,
			template.HTMLEscapeString(c.CategoryName),
			detailPath,
			template.HTMLEscapeString(c.Title),
			template.HTMLEscapeString(c.Summary),
			detailPath,
		))
	}

	homeHTML += cardsBuilder.String() + `
            </div>
        </section>
    </main>
    <footer>
        <p>&copy; Pemerintah Kabupaten Pidie — Disparbud. Hak Cipta Dilindungi.</p>
    </footer>
</body>
</html>`

	if err := os.WriteFile(filepath.Join(localeDir, "index.html"), []byte(homeHTML), 0644); err != nil {
		return err
	}

	// 2. Render Detail Pages: <locale>/<cat-slug>/<content-slug>/index.html
	for _, c := range contents {
		contentDir := filepath.Join(localeDir, c.CategorySlug, c.Slug)
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			return err
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
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; max-width: 800px; margin: 0 auto; padding: 1.5rem; color: #1a1a1a; }
        header { border-bottom: 2px solid #047857; padding-bottom: 1rem; margin-bottom: 2rem; }
        .meta { color: #4b5563; font-size: 0.9rem; margin-bottom: 1.5rem; }
        .badge { background: #e0e7ff; color: #3730a3; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem; }
        footer { margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid #e5e7eb; font-size: 0.875rem; color: #6b7280; }
        a { color: #047857; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <header>
        <p><a href="/%s/">&larr; Kembali ke Beranda</a></p>
        <span class="badge">%s</span>
        <h1>%s</h1>
        <div class="meta">Lokasi: %s</div>
    </header>
    <main>
        <article>
            %s
        </article>
    </main>
    <footer>
        <p>&copy; Pemerintah Kabupaten Pidie — Disparbud.</p>
    </footer>
</body>
</html>`,
			locale,
			template.HTMLEscapeString(c.Title),
			template.HTMLEscapeString(siteName),
			template.HTMLEscapeString(c.Summary),
			canonical,
			locale,
			template.HTMLEscapeString(c.CategoryName),
			template.HTMLEscapeString(c.Title),
			template.HTMLEscapeString(c.Address),
			string(c.BodyHTML),
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
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; max-width: 800px; margin: 0 auto; padding: 1.5rem; }
        a { color: #047857; text-decoration: none; }
    </style>
</head>
<body>
    <header>
        <p><a href="/%s/">&larr; Kembali ke Beranda</a></p>
        <h1>%s</h1>
    </header>
    <main>
        <article>%s</article>
    </main>
</body>
</html>`,
			locale,
			template.HTMLEscapeString(p.Title),
			template.HTMLEscapeString(siteName),
			template.HTMLEscapeString(p.SeoDesc),
			canonical,
			locale,
			template.HTMLEscapeString(p.Title),
			string(p.BodyHTML),
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
