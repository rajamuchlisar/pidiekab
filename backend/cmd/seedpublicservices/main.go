package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ContentSeed struct {
	Number     int
	Lat        float64
	Lng        float64
	Address    string
	IDTitle    string
	IDSlug     string
	IDSummary  string
	IDBodyHTML string
	ENTitle    string
	ENSlug     string
	ENSummary  string
	ENBodyHTML string
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://pidiekab:pidiekab_strong_pass_2026!@127.0.0.1:5432/pidiekab_production?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Ping DB failed: %v", err)
	}
	fmt.Println("Connected to database successfully.")

	// 1. Get or verify Admin ID
	var adminID int64
	err = pool.QueryRow(ctx, "SELECT id FROM admins WHERE is_active = true ORDER BY id ASC LIMIT 1").Scan(&adminID)
	if err != nil {
		adminID = 1
	}

	// 2. Get category ID for layanan-publik
	var categoryID int64
	err = pool.QueryRow(ctx, "SELECT category_id FROM category_translations WHERE locale = 'id' AND slug = 'layanan-publik' LIMIT 1").Scan(&categoryID)
	if err != nil || categoryID == 0 {
		log.Fatalf("Category 'layanan-publik' not found in database: %v", err)
	}
	fmt.Printf("Target Category 'layanan-publik' ID: %d\n", categoryID)

	// 3. Update Menu Items so all main categories and pages appear cleanly
	fmt.Println("Configuring navigation menu_items...")
	_ = configureMenuItems(ctx, pool)

	// 4. Seed/Upsert the 12 public service articles
	articles := get12Articles()
	fmt.Printf("Seeding %d Public Service articles...\n", len(articles))

	baseTime := time.Date(2026, 11, 28, 20, 0, 0, 0, time.UTC)

	for _, art := range articles {
		// Priority ordering: Article 1 has newest published_at, Article 12 has oldest
		pubTime := baseTime.Add(-time.Duration(art.Number-1) * time.Hour)

		// Check if content already exists by slug in id
		var contentID int64
		err = pool.QueryRow(ctx, `
			SELECT c.id FROM contents c
			JOIN content_translations ct ON c.id = ct.content_id
			WHERE ct.locale = 'id' AND ct.slug = $1
			LIMIT 1
		`, art.IDSlug).Scan(&contentID)

		if err != nil || contentID == 0 {
			// Insert content master record
			err = pool.QueryRow(ctx, `
				INSERT INTO contents (category_id, address, latitude, longitude, location_accuracy, main_media_id, created_by, updated_by, created_at, updated_at)
				VALUES ($1, $2, $3, $4, 'accurate', 2, $5, $5, $6, $6)
				RETURNING id
			`, categoryID, art.Address, art.Lat, art.Lng, adminID, pubTime).Scan(&contentID)
			if err != nil {
				log.Fatalf("Failed to insert content #%d: %v", art.Number, err)
			}
			fmt.Printf(" [+] Created Content #%d (ID: %d): %s\n", art.Number, contentID, art.IDTitle)
		} else {
			// Update content master record
			_, err = pool.Exec(ctx, `
				UPDATE contents
				SET category_id = $1, address = $2, latitude = $3, longitude = $4, updated_by = $5, updated_at = $6
				WHERE id = $7
			`, categoryID, art.Address, art.Lat, art.Lng, adminID, pubTime, contentID)
			if err != nil {
				log.Fatalf("Failed to update content #%d: %v", art.Number, err)
			}
			fmt.Printf(" [=] Updated Content #%d (ID: %d): %s\n", art.Number, contentID, art.IDTitle)
		}

		// Prepare JSON body
		idBodyJSON, _ := json.Marshal(map[string]string{
			"html":     art.IDBodyHTML,
			"markdown": art.IDBodyHTML,
		})
		enBodyJSON, _ := json.Marshal(map[string]string{
			"html":     art.ENBodyHTML,
			"markdown": art.ENBodyHTML,
		})

		// Upsert Indonesian Translation
		_, err = pool.Exec(ctx, `
			INSERT INTO content_translations (
				content_id, locale, title, slug, summary, body, seo_title, seo_description, status, published_at, created_by, updated_by, published_by, created_at, updated_at
			) VALUES (
				$1, 'id', $2, $3, $4, $5, $6, $7, 'published', $8, $9, $9, $9, $8, $8
			)
			ON CONFLICT (content_id, locale) DO UPDATE SET
				title = EXCLUDED.title,
				slug = EXCLUDED.slug,
				summary = EXCLUDED.summary,
				body = EXCLUDED.body,
				seo_title = EXCLUDED.seo_title,
				seo_description = EXCLUDED.seo_description,
				status = 'published',
				published_at = EXCLUDED.published_at,
				published_by = EXCLUDED.published_by,
				updated_by = EXCLUDED.updated_by,
				updated_at = EXCLUDED.updated_at
		`, contentID, art.IDTitle, art.IDSlug, art.IDSummary, idBodyJSON, art.IDTitle+" — Layanan & Info Publik Pidie", art.IDSummary, pubTime, adminID)
		if err != nil {
			log.Fatalf("Failed to upsert ID translation for #%d: %v", art.Number, err)
		}

		// Upsert English Translation
		_, err = pool.Exec(ctx, `
			INSERT INTO content_translations (
				content_id, locale, title, slug, summary, body, seo_title, seo_description, status, published_at, created_by, updated_by, published_by, created_at, updated_at
			) VALUES (
				$1, 'en', $2, $3, $4, $5, $6, $7, 'published', $8, $9, $9, $9, $8, $8
			)
			ON CONFLICT (content_id, locale) DO UPDATE SET
				title = EXCLUDED.title,
				slug = EXCLUDED.slug,
				summary = EXCLUDED.summary,
				body = EXCLUDED.body,
				seo_title = EXCLUDED.seo_title,
				seo_description = EXCLUDED.seo_description,
				status = 'published',
				published_at = EXCLUDED.published_at,
				published_by = EXCLUDED.published_by,
				updated_by = EXCLUDED.updated_by,
				updated_at = EXCLUDED.updated_at
		`, contentID, art.ENTitle, art.ENSlug, art.ENSummary, enBodyJSON, art.ENTitle+" — Public Services Pidie", art.ENSummary, pubTime, adminID)
		if err != nil {
			log.Fatalf("Failed to upsert EN translation for #%d: %v", art.Number, err)
		}
	}

	fmt.Println("All 12 Public Service articles successfully seeded and published!")
}

func configureMenuItems(ctx context.Context, pool *pgxpool.Pool) error {
	// Query categories
	catMap := make(map[string]int64)
	rows, err := pool.Query(ctx, "SELECT category_id, slug FROM category_translations WHERE locale = 'id'")
	if err == nil {
		for rows.Next() {
			var id int64
			var slug string
			if err := rows.Scan(&id, &slug); err == nil {
				catMap[slug] = id
			}
		}
		rows.Close()
	}

	// Query static pages
	var aboutID, contactID int64
	_ = pool.QueryRow(ctx, "SELECT static_page_id FROM static_page_translations WHERE locale = 'id' AND slug = 'tentang-pidie' LIMIT 1").Scan(&aboutID)
	if aboutID == 0 {
		_ = pool.QueryRow(ctx, "SELECT static_page_id FROM static_page_translations WHERE locale = 'id' AND slug = 'tentang-kami' LIMIT 1").Scan(&aboutID)
	}
	_ = pool.QueryRow(ctx, "SELECT static_page_id FROM static_page_translations WHERE locale = 'id' AND slug = 'hubungi' LIMIT 1").Scan(&contactID)

	type navItem struct {
		titleID   string
		titleEN   string
		itemType  string
		targetID  *int64
		extURL    string
		sortOrder int
	}

	navs := []navItem{
		{titleID: "Beranda", titleEN: "Home", itemType: "external", extURL: "/", sortOrder: 1},
	}

	if id, ok := catMap["wisata-alam"]; ok {
		i := id
		navs = append(navs, navItem{titleID: "Wisata Alam", titleEN: "Nature", itemType: "category", targetID: &i, sortOrder: 2})
	}
	if id, ok := catMap["sejarah-budaya"]; ok {
		i := id
		navs = append(navs, navItem{titleID: "Sejarah & Budaya", titleEN: "History & Culture", itemType: "category", targetID: &i, sortOrder: 3})
	}
	if id, ok := catMap["kuliner-khas"]; ok {
		i := id
		navs = append(navs, navItem{titleID: "Kuliner Khas", titleEN: "Culinary", itemType: "category", targetID: &i, sortOrder: 4})
	}
	if id, ok := catMap["layanan-publik"]; ok {
		i := id
		navs = append(navs, navItem{titleID: "Layanan & Info Publik", titleEN: "Public Services", itemType: "category", targetID: &i, sortOrder: 5})
	}
	if aboutID > 0 {
		i := aboutID
		navs = append(navs, navItem{titleID: "Tentang Pidie", titleEN: "About Pidie", itemType: "static_page", targetID: &i, sortOrder: 6})
	}
	if contactID > 0 {
		i := contactID
		navs = append(navs, navItem{titleID: "Hubungi", titleEN: "Contact Us", itemType: "static_page", targetID: &i, sortOrder: 7})
	}

	// Clean and re-seed menu_items
	_, _ = pool.Exec(ctx, "TRUNCATE menu_items RESTART IDENTITY")
	for _, n := range navs {
		_, err := pool.Exec(ctx, `
			INSERT INTO menu_items (title_id, title_en, item_type, target_id, external_url, open_in_new_tab, sort_order, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, false, $6, true, NOW(), NOW())
		`, n.titleID, n.titleEN, n.itemType, n.targetID, n.extURL, n.sortOrder)
		if err != nil {
			fmt.Printf("Warning: failed to insert menu item %s: %v\n", n.titleID, err)
		}
	}
	fmt.Printf("Inserted %d menu items.\n", len(navs))
	return nil
}
