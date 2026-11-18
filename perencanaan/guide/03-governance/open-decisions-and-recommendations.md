# PidieKab — Open Decisions & Resolutions Status

**Versi:** v79 (Final Production Baseline)  
**Tujuan:** Catatan status keputusan arsitektural dan operasional platform PidieKab beserta referensi implementasi konkret.

| ID | Keputusan / Area | Status | Resolusi & Implementasi Konkret | Referensi Implementasi |
|---|---|---|---|---|
| **OPEN-001** | Tool SSG Vue | **LOCKED** | Menggunakan `vite-ssg` yang menghasilkan 100% HTML statis dwibahasa (ID & EN) mandiri, canonical link, hreflang, sitemap.xml, dan disajikan langsung oleh Go standard library tanpa Node.js di production. | `frontend/public-site/`, `scripts/test-ssg-poc.mjs` |
| **OPEN-002** | Goose migration format | **LOCKED** | Menggunakan format Goose SQL dengan directive `-- +goose Up` dan `-- +goose Down`. Teruji stabil pada PostgreSQL 17 dengan idempotency check Up-Down-Up di CI. | `migrations/00001_initial_schema.sql` |
| **OPEN-003** | Password hashing | **LOCKED** | Argon2id melalui `golang.org/x/crypto/argon2` dengan parameter aman: Time 1, Memory 64MB, Threads 4, Salt 16B, Key 32B. | `backend/internal/platform/authn/password.go` |
| **OPEN-004** | Password policy | **LOCKED** | Minimal 8 karakter (rekomendasi 12+), mendukung passphrase panjang, tanpa aturan komposisi artifisial yang merusak entropi. | `backend/internal/modules/identity/domain/admin.go` |
| **OPEN-005** | Admin bootstrap | **LOCKED** | Perintah CLI aman `pidiekab admin create --email <email> --name <name> --password <pwd>`. Tidak ada akun default hardcoded di kode sumber. | `backend/cmd/pidiekab/main.go` |
| **OPEN-006** | Password reset | **LOCKED** | Reset password terkontrol operator via CLI `pidiekab admin reset-password --email <email> --password <new>` yang otomatis mencabut seluruh sesi login aktif admin tersebut. | `backend/cmd/pidiekab/main.go`, `identity/application/auth_service.go` |
| **OPEN-007** | Session lifetime | **LOCKED** | Cookie session berumur absolut 12 jam, idle timeout 30 menit. Hash-at-rest di tabel `admin_sessions`, terhapus saat logout atau password reset. | `backend/internal/modules/identity/domain/session.go` |
| **OPEN-008** | Cookie & CSRF | **LOCKED** | Cookie `pidiekab_session` dengan flag `HttpOnly`, `SameSite=Lax`, dan `Secure` otomatis saat HTTPS aktif. Proteksi CSRF ganda (token header `X-CSRF-Token` + validasi Origin/Referer). | `backend/internal/platform/httpserver/middleware/csrf.go` |
| **OPEN-009** | 2FA (MFA) | **DEFERRED (POST-MVP)** | Desain session dan identity bersih dan mendukung ekstensi WebAuthn/TOTP di fase pengembangan berikutnya tanpa mengubah skema tabel utama. | `backend/internal/modules/identity/` |
| **OPEN-010** | TLS issuance & renewal | **LOCKED** | Native Go `net/http` TLS listener (TLS 1.2 & 1.3) langsung melayani port 443 dengan redirect otomatis dari port 80 tanpa Nginx. | `backend/internal/platform/httpserver/server.go` |
| **OPEN-011** | Database topology | **LOCKED** | Single-host PostgreSQL 17 pada host yang sama dengan private unix socket / loopback TCP `127.0.0.1:5432` dengan hak akses ketat di bawah user `pidiekab`. | `deploy/systemd/pidiekab.service`, `deploy/scripts/setup-server.sh` |
| **OPEN-012** | Backup retention, RPO & RTO | **LOCKED** | Skrip `deploy/scripts/backup.sh` membuat logical dump terkompresi PostgreSQL + arsip media lokal dengan SHA-256 checksum harian (retensi 7 hari). Dilengkapi drill restore `deploy/scripts/restore.sh`. | `deploy/scripts/backup.sh`, `deploy/scripts/restore.sh` |
| **OPEN-013** | Location master source | **LOCKED** | Master data wilayah resmi Kabupaten Pidie (1 kabupaten, 23 kecamatan) diinisialisasi melalui migrasi dan seed command. | `migrations/00001_initial_schema.sql`, `cmd/pidiekab/seed.go` |
| **OPEN-014** | Coordinate accuracy | **LOCKED** | Koordinat opsional dengan validasi rentang geografis (-90..90, -180..180) dan label akurasi eksplisit (`accurate` atau `approximate`). | `backend/internal/modules/content/domain/content.go` |
| **OPEN-015** | Multilingual full-text search | **LOCKED** | Pencarian teks terindeks PostgreSQL FTS pada judul, ringkasan, dan isi konten dwibahasa (ID & EN) dengan ranking relevansi dan sanitasi query. | `backend/internal/modules/search/infrastructure/postgres/search_repository.go` |
| **OPEN-016** | Upload size & body limits | **LOCKED** | Batas upload gambar 10 MB, validasi MIME type via magic bytes, penyimpanan konten aman tanpa izin eksekusi. | `backend/internal/modules/media/application/media_service.go` |
| **OPEN-017** | Rate limit, timeouts & pool | **LOCKED** | Rate limiting in-memory berbasis IP pada endpoint sensitif (login), timeout server ketat (Read 15s, Write 30s, Idle 60s), dan connection pool konfigurabel. | `backend/internal/platform/httpserver/middleware/ratelimit.go`, `config/config.go` |
| **OPEN-018** | SSG build strategy | **LOCKED** | Atomic release symlink switching: build dilakukan di staging temporary directory, diverifikasi, lalu dialihkan via symlink tanpa downtime. Rilis lama disimpan (default 5 rilis). | `backend/internal/workers/sitegeneration/release_manager.go` |
| **OPEN-019** | Permanent URL Redirects | **LOCKED** | Modul 301 URL Redirects dengan persistensi tabel `url_redirects`, penanganan di static fallback handler, auto-redirect otomatis saat slug konten/halaman diubah, serta view manajemen di Admin CMS. | `backend/internal/modules/redirect/`, `frontend/admin-cms/src/views/redirects/` |
| **OPEN-020** | API contract coverage | **LOCKED** | Spesifikasi resmi OpenAPI 3.1 di `api/openapi.yaml`, disajikan langsung melalui endpoint `GET /api/v1/openapi.yaml` dengan unit test otomatis. | `api/openapi.yaml`, `backend/internal/platform/httpserver/server.go` |
| **OPEN-021** | Contact & social URL policy | **LOCKED** | Validasi skema wajib `https://` untuk seluruh tautan eksternal media sosial guna mencegah injeksi `javascript:`/`http:`, serta validasi format RFC email dan nomor telepon. | `backend/internal/modules/settings/domain/settings.go` |
| **OPEN-022** | Metrics & alert thresholds | **LOCKED** | Endpoint observabilitas runtime `GET /health/metrics` melaporkan status DB, uptime, goroutines, dan alokasi memori. Diagnostik CLI `pidiekab db status`. Liveness `GET /health/live` dan readiness `GET /health/ready`. | `backend/internal/platform/httpserver/server.go`, `cmd/pidiekab/operator_cli.go` |

---

## Kesimpulan Tata Kelola
Seluruh item keputusan pembuka (OPEN-001 hingga OPEN-022) telah diselesaikan, diuji, dan dikunci dalam implementasi kode sumber platform PidieKab. Tidak ada blocker fungsional atau arsitektural yang tersisa untuk fase peluncuran produksi.
