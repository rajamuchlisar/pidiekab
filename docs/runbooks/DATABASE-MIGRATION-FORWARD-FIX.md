# Database Migration Forward-Fix Runbook — PidieKab

**Baseline Standard:** Goose Migrations (`01-technical/database/migrations/`)  
**Prinsip Utama:** *No Down-Migration in Production* (Forward-Only Schema Evolution)

---

## 1. Filosofi Forward-Fix

1. **Larangan Down-Migration di Production:**
   Menjalankan `goose down` di environment production berisiko tinggi menghapus data transaksi, audit log, atau relasi foreign key yang telah tercipta setelah migrasi berjalan.
2. **Backward-Compatible Changes:**
   Setiap perubahan skema database harus backward-compatible dengan versi binary yang sedang aktif sebelum rilis baru dipromosikan (Expand and Contract Pattern).
3. **Penyelesaian Masalah Skema:**
   Jika ditemukan ketidaksesuaian indeks, constraint, atau kolom di production, solusi yang disahkan adalah membuat file migrasi baru `00X_forward_fix_description.sql` dan mengaplikasikannya via `goose up`.

---

## 2. Prosedur Pembuatan Forward-Fix

### Langkah 1: Diagnosis Masalah
Identifikasi constraint, index, atau kolom yang bermasalah menggunakan query katalog PostgreSQL:
```sql
-- Cek status migrasi
SELECT version_id, is_applied, tstamp FROM goose_db_version ORDER BY id DESC;

-- Cek constraint yang gagal
SELECT conname, contype, pg_get_constraintdef(c.oid)
FROM pg_constraint c
WHERE conrelid = 'contents'::regclass;
```

### Langkah 2: Buat File Migrasi Baru
Buat file migrasi dengan nomor versi berikutnya di `01-technical/database/migrations/`:
Contoh: `002_fix_content_slug_index.sql`

```sql
-- +goose Up
-- +goose StatementBegin
-- Terapkan perbaikan skema secara idempotent atau safe DDL
ALTER TABLE content_translations ADD CONSTRAINT chk_slug_not_empty CHECK (length(trim(slug)) > 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE content_translations DROP CONSTRAINT IF EXISTS chk_slug_not_empty;
-- +goose StatementEnd
```

### Langkah 3: Uji Coba pada Staging / Disposable DB
Sebelum dieksekusi di production:
```bash
goose -dir 01-technical/database/migrations postgres "$STAGING_DATABASE_URL" up
```
Pastikan seluruh unit & integration tests aplikasi lulus tanpa regresi.

### Langkah 4: Terapkan di Production
Jalankan migrasi terkontrol:
```bash
goose -dir 01-technical/database/migrations postgres "$PIDIEKAB_DATABASE_URL" up
```
Verifikasi log status Goose dan periksa readiness probe aplikasi:
```bash
curl -s http://127.0.0.1:8080/health/ready
```
