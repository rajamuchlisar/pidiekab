package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pidiekab/internal/modules/location/domain"
)

// Repository manages persistence for locations in PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new location Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListDistricts returns all districts sorted alphabetically by name.
func (r *Repository) ListDistricts(ctx context.Context) ([]*domain.District, error) {
	query := `
		SELECT id, regency_id, name, code, created_at, updated_at
		FROM districts
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query districts: %w", err)
	}
	defer rows.Close()

	var districts []*domain.District
	for rows.Next() {
		var d domain.District
		if err := rows.Scan(&d.ID, &d.RegencyID, &d.Name, &d.Code, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan district row: %w", err)
		}
		districts = append(districts, &d)
	}

	return districts, rows.Err()
}

// ListVillagesByDistrict returns all villages in a district sorted alphabetically.
func (r *Repository) ListVillagesByDistrict(ctx context.Context, districtID int64) ([]*domain.Village, error) {
	query := `
		SELECT id, district_id, name, code, created_at, updated_at
		FROM villages
		WHERE district_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, districtID)
	if err != nil {
		return nil, fmt.Errorf("failed to query villages: %w", err)
	}
	defer rows.Close()

	var villages []*domain.Village
	for rows.Next() {
		var v domain.Village
		if err := rows.Scan(&v.ID, &v.DistrictID, &v.Name, &v.Code, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan village row: %w", err)
		}
		villages = append(villages, &v)
	}

	return villages, rows.Err()
}

// GetVillageByID finds a village by its ID.
func (r *Repository) GetVillageByID(ctx context.Context, id int64) (*domain.Village, error) {
	query := `
		SELECT id, district_id, name, code, created_at, updated_at
		FROM villages
		WHERE id = $1
	`

	var v domain.Village
	err := r.pool.QueryRow(ctx, query, id).Scan(&v.ID, &v.DistrictID, &v.Name, &v.Code, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query village by id: %w", err)
	}

	return &v, nil
}

// SeedPidieBaselineIfEmpty populates Kabupaten Pidie and official 23 kecamatan if the regency table is empty.
func (r *Repository) SeedPidieBaselineIfEmpty(ctx context.Context) error {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM regencies`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check regency count: %w", err)
	}
	if count > 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var regencyID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO regencies (name, code)
		VALUES ('Kabupaten Pidie', '11.07')
		RETURNING id
	`).Scan(&regencyID)
	if err != nil {
		return fmt.Errorf("failed to seed Kabupaten Pidie: %w", err)
	}

	// 23 official kecamatan in Kabupaten Pidie
	districts := []struct {
		name string
		code string
	}{
		{"Batee", "11.07.03"},
		{"Delima", "11.07.04"},
		{"Geumpang", "11.07.05"},
		{"Geulumpang Baro", "11.07.29"},
		{"Glumpang Tiga", "11.07.06"},
		{"Grong-Grong", "11.07.21"},
		{"Indra Jaya", "11.07.07"},
		{"Kembang Tanjong", "11.07.08"},
		{"Keumala", "11.07.18"},
		{"Kota Sigli", "11.07.09"},
		{"Mane", "11.07.31"},
		{"Mila", "11.07.10"},
		{"Muara Tiga", "11.07.11"},
		{"Mutiara", "11.07.13"},
		{"Mutiara Timur", "11.07.24"},
		{"Padang Tiji", "11.07.14"},
		{"Peukan Baro", "11.07.15"},
		{"Pidie", "11.07.16"},
		{"Sakti", "11.07.17"},
		{"Simpang Tiga", "11.07.19"},
		{"Tangse", "11.07.22"},
		{"Tiro / Truseb", "11.07.25"},
		{"Titeue", "11.07.30"},
	}

	for _, d := range districts {
		_, err := tx.Exec(ctx, `
			INSERT INTO districts (regency_id, name, code)
			VALUES ($1, $2, $3)
			ON CONFLICT (regency_id, LOWER(name)) DO NOTHING
		`, regencyID, d.name, d.code)
		if err != nil {
			return fmt.Errorf("failed to seed district %s: %w", d.name, err)
		}
	}

	return tx.Commit(ctx)
}
