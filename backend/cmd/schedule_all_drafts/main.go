package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EditorialArticle struct {
	ContentID      int64
	ScheduleDate   time.Time // in UTC
	IDTitle        string
	IDSlug         string
	IDSummary      string
	IDBodyHTML     string
	IDSEOTitle     string
	IDSEODesc      string
	ENTitle        string
	ENSlug         string
	ENSummary      string
	ENBodyHTML     string
	ENSEOTitle     string
	ENSEODesc      string
	SourceName     string
	SourceType     string
	SourceURL      string
	SourceNote     string
}

func countWords(text string) int {
	clean := strings.ReplaceAll(text, "<", " <")
	clean = strings.ReplaceAll(clean, ">", "> ")
	var words []string
	inTag := false
	var cur strings.Builder
	for _, r := range clean {
		if r == '<' {
			inTag = true
			if cur.Len() > 0 {
				w := strings.TrimSpace(cur.String())
				if w != "" {
					words = append(words, strings.Fields(w)...)
				}
				cur.Reset()
			}
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		w := strings.TrimSpace(cur.String())
		if w != "" {
			words = append(words, strings.Fields(w)...)
		}
	}
	return len(words)
}

func getArticles() []EditorialArticle {
	var list []EditorialArticle
	list = append(list, getBatch1()...)
	list = append(list, getBatch2()...)
	list = append(list, getBatch3()...)
	return list
}

func main() {
	dbURL := os.Getenv("PIDIEKAB_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_dev?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	var adminID int64
	err = pool.QueryRow(ctx, "SELECT id FROM admins WHERE email = 'admin@pidiekab.web.id' LIMIT 1").Scan(&adminID)
	if err != nil {
		_ = pool.QueryRow(ctx, "SELECT id FROM admins WHERE is_active = true ORDER BY id ASC LIMIT 1").Scan(&adminID)
	}
	if adminID == 0 {
		log.Fatalf("no admin user found")
	}

	articles := getArticles()

	fmt.Println("================================================================================")
	fmt.Printf("SCHEDULE ALL DRAFT CONTENTS (Total: %d Articles)\n", len(articles))
	fmt.Println("================================================================================")

	for idx, art := range articles {
		idWords := countWords(art.IDBodyHTML)
		enWords := countWords(art.ENBodyHTML)

		fmt.Printf("[%2d/%2d] Content #%d | Sched: %s WIB | ID: %d words | EN: %d words\n",
			idx+1, len(articles), art.ContentID,
			art.ScheduleDate.In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02 15:04"),
			idWords, enWords,
		)

		if idWords < 500 {
			log.Fatalf("ERROR: Indonesian body for Content #%d is under 500 words (%d words)", art.ContentID, idWords)
		}
		if enWords < 500 {
			log.Fatalf("ERROR: English body for Content #%d is under 500 words (%d words)", art.ContentID, enWords)
		}

		// Ensure citation source exists
		var srcCount int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM content_sources WHERE content_id = $1", art.ContentID).Scan(&srcCount)
		if srcCount == 0 {
			_, err = pool.Exec(ctx, `
				INSERT INTO content_sources (content_id, source_name, source_type, source_url, notes, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
			`, art.ContentID, art.SourceName, art.SourceType, art.SourceURL, art.SourceNote, adminID)
			if err != nil {
				log.Fatalf("failed to insert source for content #%d: %v", art.ContentID, err)
			}
		}

		// Update Indonesian Translation
		idBodyJSON, _ := json.Marshal(map[string]any{"html": art.IDBodyHTML, "markdown": art.IDBodyHTML})
		_, err = pool.Exec(ctx, `
			INSERT INTO content_translations (
				content_id, locale, title, slug, summary, body, seo_title, seo_description,
				status, scheduled_at, published_by, updated_by, search_vector, created_at, updated_at
			)
			VALUES (
				$1, 'id', $2, $3, $4, $5, $6, $7,
				'scheduled', $8, $9, $9,
				setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'),
				NOW(), NOW()
			)
			ON CONFLICT (content_id, locale) DO UPDATE
			SET title = EXCLUDED.title,
			    slug = EXCLUDED.slug,
			    summary = EXCLUDED.summary,
			    body = EXCLUDED.body,
			    seo_title = EXCLUDED.seo_title,
			    seo_description = EXCLUDED.seo_description,
			    status = 'scheduled',
			    scheduled_at = EXCLUDED.scheduled_at,
			    published_by = EXCLUDED.published_by,
			    updated_by = EXCLUDED.updated_by,
			    search_vector = EXCLUDED.search_vector,
			    updated_at = NOW()
		`, art.ContentID, art.IDTitle, art.IDSlug, art.IDSummary, idBodyJSON, art.IDSEOTitle, art.IDSEODesc, art.ScheduleDate, adminID)
		if err != nil {
			log.Fatalf("failed to upsert ID translation for #%d: %v", art.ContentID, err)
		}

		// Update English Translation
		enBodyJSON, _ := json.Marshal(map[string]any{"html": art.ENBodyHTML, "markdown": art.ENBodyHTML})
		_, err = pool.Exec(ctx, `
			INSERT INTO content_translations (
				content_id, locale, title, slug, summary, body, seo_title, seo_description,
				status, scheduled_at, published_by, updated_by, search_vector, created_at, updated_at
			)
			VALUES (
				$1, 'en', $2, $3, $4, $5, $6, $7,
				'scheduled', $8, $9, $9,
				setweight(to_tsvector('simple', COALESCE($2, '')), 'A') || setweight(to_tsvector('simple', COALESCE($4, '')), 'B'),
				NOW(), NOW()
			)
			ON CONFLICT (content_id, locale) DO UPDATE
			SET title = EXCLUDED.title,
			    slug = EXCLUDED.slug,
			    summary = EXCLUDED.summary,
			    body = EXCLUDED.body,
			    seo_title = EXCLUDED.seo_title,
			    seo_description = EXCLUDED.seo_description,
			    status = 'scheduled',
			    scheduled_at = EXCLUDED.scheduled_at,
			    published_by = EXCLUDED.published_by,
			    updated_by = EXCLUDED.updated_by,
			    search_vector = EXCLUDED.search_vector,
			    updated_at = NOW()
		`, art.ContentID, art.ENTitle, art.ENSlug, art.ENSummary, enBodyJSON, art.ENSEOTitle, art.ENSEODesc, art.ScheduleDate, adminID)
		if err != nil {
			log.Fatalf("failed to upsert EN translation for #%d: %v", art.ContentID, err)
		}
	}

	fmt.Println("\n[SUCCESS] All 15 draft contents have been expanded (>500 words) and scheduled daily!")
}
