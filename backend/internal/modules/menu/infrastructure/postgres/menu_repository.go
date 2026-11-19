package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/menu/domain"
)

type MenuRepository struct {
	pool *pgxpool.Pool
}

func NewMenuRepository(pool *pgxpool.Pool) *MenuRepository {
	repo := &MenuRepository{pool: pool}
	// Automatically ensure schema on initialization
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = repo.EnsureSchema(ctx)
	return repo
}

func (r *MenuRepository) EnsureSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS menu_items (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		title_id TEXT NOT NULL,
		title_en TEXT NOT NULL DEFAULT '',
		item_type TEXT NOT NULL,
		target_id BIGINT,
		external_url TEXT,
		open_in_new_tab BOOLEAN NOT NULL DEFAULT FALSE,
		sort_order INT NOT NULL DEFAULT 0,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS menu_items_sort_idx ON menu_items (sort_order ASC, id ASC);
	`
	if _, err := r.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("failed to ensure menu_items table: %w", err)
	}

	// Check if empty; if so, seed baseline menus
	var count int
	_ = r.pool.QueryRow(ctx, "SELECT count(*) FROM menu_items").Scan(&count)
	if count == 0 {
		r.seedBaselineMenus(ctx)
	}

	return nil
}

func (r *MenuRepository) seedBaselineMenus(ctx context.Context) {
	// Query existing categories to link by ID
	catMap := make(map[string]int64)
	cRows, err := r.pool.Query(ctx, "SELECT category_id, slug FROM category_translations WHERE locale = 'id'")
	if err == nil {
		for cRows.Next() {
			var id int64
			var slug string
			if err := cRows.Scan(&id, &slug); err == nil {
				catMap[slug] = id
			}
		}
		cRows.Close()
	}

	// Query existing static page to link by ID
	var aboutPageID int64
	_ = r.pool.QueryRow(ctx, "SELECT static_page_id FROM static_page_translations WHERE locale = 'id' AND slug = 'tentang-kami' LIMIT 1").Scan(&aboutPageID)

	// Baseline items
	type baselineItem struct {
		titleID      string
		titleEN      string
		itemType     string
		targetID     *int64
		externalURL  string
		openInNewTab bool
		sortOrder    int
	}

	items := []baselineItem{
		{
			titleID:      "Beranda",
			titleEN:      "Home",
			itemType:     domain.TypeExternal,
			externalURL:  "/",
			openInNewTab: false,
			sortOrder:    1,
		},
	}

	if catID, ok := catMap["wisata-alam"]; ok {
		cID := catID
		items = append(items, baselineItem{
			titleID:      "Wisata Alam",
			titleEN:      "Nature",
			itemType:     domain.TypeCategory,
			targetID:     &cID,
			openInNewTab: false,
			sortOrder:    2,
		})
	}
	if catID, ok := catMap["sejarah-budaya"]; ok {
		cID := catID
		items = append(items, baselineItem{
			titleID:      "Sejarah & Budaya",
			titleEN:      "History & Heritage",
			itemType:     domain.TypeCategory,
			targetID:     &cID,
			openInNewTab: false,
			sortOrder:    3,
		})
	}
	if catID, ok := catMap["kuliner-khas"]; ok {
		cID := catID
		items = append(items, baselineItem{
			titleID:      "Kuliner Khas",
			titleEN:      "Culinary",
			itemType:     domain.TypeCategory,
			targetID:     &cID,
			openInNewTab: false,
			sortOrder:    4,
		})
	}
	if aboutPageID > 0 {
		pID := aboutPageID
		items = append(items, baselineItem{
			titleID:      "Tentang Pidie",
			titleEN:      "About Pidie",
			itemType:     domain.TypeStaticPage,
			targetID:     &pID,
			openInNewTab: false,
			sortOrder:    5,
		})
	}

	for _, it := range items {
		_, _ = r.pool.Exec(ctx, `
			INSERT INTO menu_items (title_id, title_en, item_type, target_id, external_url, open_in_new_tab, sort_order, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, NOW(), NOW())
		`, it.titleID, it.titleEN, it.itemType, it.targetID, it.externalURL, it.openInNewTab, it.sortOrder)
	}
}

func (r *MenuRepository) List(ctx context.Context) ([]domain.MenuItem, error) {
	query := `
		SELECT
			m.id, m.title_id, m.title_en, m.item_type, m.target_id, COALESCE(m.external_url, ''),
			m.open_in_new_tab, m.sort_order, m.is_active, m.created_at, m.updated_at,
			COALESCE(
				CASE
					WHEN m.item_type = 'static_page' THEN spt.title
					WHEN m.item_type = 'content' THEN ct.title
					WHEN m.item_type = 'category' THEN catt.name
					ELSE ''
				END, ''
			) AS target_title,
			COALESCE(
				CASE
					WHEN m.item_type = 'static_page' THEN '/id/' || spt.slug || '/'
					WHEN m.item_type = 'content' THEN '/id/' || catt_c.slug || '/' || ct.slug || '/'
					WHEN m.item_type = 'category' THEN '/id/' || catt.slug || '/'
					ELSE m.external_url
				END, ''
			) AS resolved_url
		FROM menu_items m
		LEFT JOIN static_pages sp ON m.item_type = 'static_page' AND sp.id = m.target_id
		LEFT JOIN static_page_translations spt ON spt.static_page_id = sp.id AND spt.locale = 'id'
		LEFT JOIN contents c ON m.item_type = 'content' AND c.id = m.target_id
		LEFT JOIN content_translations ct ON ct.content_id = c.id AND ct.locale = 'id'
		LEFT JOIN categories cat_c ON cat_c.id = c.category_id
		LEFT JOIN category_translations catt_c ON catt_c.category_id = cat_c.id AND catt_c.locale = 'id'
		LEFT JOIN categories cat ON m.item_type = 'category' AND cat.id = m.target_id
		LEFT JOIN category_translations catt ON catt.category_id = cat.id AND catt.locale = 'id'
		ORDER BY m.sort_order ASC, m.id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu_items: %w", err)
	}
	defer rows.Close()

	var items []domain.MenuItem
	for rows.Next() {
		var item domain.MenuItem
		err := rows.Scan(
			&item.ID, &item.TitleID, &item.TitleEN, &item.ItemType, &item.TargetID,
			&item.ExternalURL, &item.OpenInNewTab, &item.SortOrder, &item.IsActive,
			&item.CreatedAt, &item.UpdatedAt, &item.TargetTitle, &item.ResolvedURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu_item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *MenuRepository) GetByID(ctx context.Context, id int64) (*domain.MenuItem, error) {
	query := `
		SELECT
			id, title_id, title_en, item_type, target_id, COALESCE(external_url, ''),
			open_in_new_tab, sort_order, is_active, created_at, updated_at
		FROM menu_items
		WHERE id = $1
	`
	var item domain.MenuItem
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.TitleID, &item.TitleEN, &item.ItemType, &item.TargetID,
		&item.ExternalURL, &item.OpenInNewTab, &item.SortOrder, &item.IsActive,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMenuNotFound
		}
		return nil, fmt.Errorf("failed to query menu_item by id: %w", err)
	}
	return &item, nil
}

func (r *MenuRepository) Create(ctx context.Context, item *domain.MenuItem) error {
	query := `
		INSERT INTO menu_items (
			title_id, title_en, item_type, target_id, external_url,
			open_in_new_tab, sort_order, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		item.TitleID, item.TitleEN, item.ItemType, item.TargetID, item.ExternalURL,
		item.OpenInNewTab, item.SortOrder, item.IsActive,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert menu_item: %w", err)
	}
	return nil
}

func (r *MenuRepository) Update(ctx context.Context, item *domain.MenuItem) error {
	query := `
		UPDATE menu_items SET
			title_id = $2,
			title_en = $3,
			item_type = $4,
			target_id = $5,
			external_url = $6,
			open_in_new_tab = $7,
			sort_order = $8,
			is_active = $9,
			updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.pool.Exec(ctx, query,
		item.ID, item.TitleID, item.TitleEN, item.ItemType, item.TargetID,
		item.ExternalURL, item.OpenInNewTab, item.SortOrder, item.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to update menu_item: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrMenuNotFound
	}
	return nil
}

func (r *MenuRepository) Delete(ctx context.Context, id int64) error {
	cmd, err := r.pool.Exec(ctx, "DELETE FROM menu_items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete menu_item: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrMenuNotFound
	}
	return nil
}

func (r *MenuRepository) Reorder(ctx context.Context, ids []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for idx, id := range ids {
		_, err := tx.Exec(ctx, "UPDATE menu_items SET sort_order = $2, updated_at = NOW() WHERE id = $1", id, idx+1)
		if err != nil {
			return fmt.Errorf("failed to update order for id %d: %w", id, err)
		}
	}

	return tx.Commit(ctx)
}
