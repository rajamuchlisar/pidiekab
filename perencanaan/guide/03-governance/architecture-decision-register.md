# PidieKab — Architecture Decision Register

**Versi:** v78  
**Status:** Register konsolidasi  
**Legenda:** LOCKED = keputusan final; PROPOSED = rekomendasi menunggu persetujuan; OPEN = belum diputuskan.

## 1. Product & scope

| ID | Keputusan | Status | Catatan |
|---|---|---|---|
| ADR-001 | PidieKab adalah Independent Regional Promotion & Visitor Discovery Platform | LOCKED | Bukan situs resmi/representasi pemerintah |
| ADR-002 | MVP berisi Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah | LOCKED | Akomodasi, Panduan Wisata, Berita & Agenda masuk Phase 2 |
| ADR-003 | Admin internal mengelola konten; tidak ada public submission | LOCKED | Tidak ada akun operator publik |
| ADR-004 | Tidak ada approval/verification workflow | LOCKED | Source wajib sebagai referensi, bukan proses verifikasi |
| ADR-005 | Bahasa Indonesia dan Inggris sejak MVP | LOCKED | Status publikasi per bahasa independen |
| ADR-006 | Pengunjung publik tidak perlu login | LOCKED | Tanpa review, komentar, rating, favorit, atau profil |
| ADR-007 | Satu kategori utama per konten | LOCKED | Tidak ada multi-category/tag MVP |
| ADR-008 | Delete konten permanen dari CMS dengan konfirmasi dan audit metadata | LOCKED | Media bersama tidak ikut terhapus |

## 2. Application architecture

| ID | Keputusan | Status | Catatan |
|---|---|---|---|
| ADR-009 | Backend Go | LOCKED | Satu codebase dan binary |
| ADR-010 | Go standard library `net/http` sebagai satu-satunya application web server | LOCKED | Tidak ada Nginx, Caddy, Apache, LiteSpeed, atau reverse proxy |
| ADR-011 | Pragmatic Modular Monolith | LOCKED | Komunikasi modul in-process |
| ADR-012 | PostgreSQL sebagai satu database utama | LOCKED | Satu source of truth |
| ADR-013 | Tidak memakai Redis atau message broker pada MVP | LOCKED | Evaluasi ulang hanya jika kebutuhan terbukti |
| ADR-014 | SQL eksplisit, tanpa ORM | LOCKED | `pgx/v5` dan `pgxpool` sebagai baseline |
| ADR-015 | Migration memakai Goose | LOCKED | Terbukti kompatibel dengan DDL baseline via Goose v3.24 (Up, Down, Re-up idempotent) |
| ADR-016 | Logging memakai `log/slog`; testing memakai `testing` dan `httptest` | LOCKED | Library tambahan harus memberi manfaat nyata |
| ADR-017 | Password hashing Argon2id | LOCKED | Terbukti dan diimplementasikan via `golang.org/x/crypto/argon2` dengan constant-time comparison |
| ADR-018 | In-process SSG worker pada satu service | LOCKED | PostgreSQL job queue; tanpa broker eksternal |
| ADR-019 | SSG build menggunakan temporary release dan atomic activation | LOCKED | Detail tool/generator masih OPEN |

## 3. Frontend & public delivery

| ID | Keputusan | Status | Catatan |
|---|---|---|---|
| ADR-020 | Vue 3 + TypeScript + Vite | LOCKED | Public site dan Admin CMS dipisah sebagai dua app frontend |
| ADR-021 | Public website menggunakan SSG | LOCKED | SEO dan HTML awal harus tersedia tanpa client rendering penuh |
| ADR-022 | Admin CMS menggunakan Vue SPA dinamis | LOCKED | Same-origin REST API |
| ADR-023 | Node.js hanya untuk development/build | LOCKED | Tidak berjalan sebagai production web server |
| ADR-024 | Tool SSG memakai vite-ssg (dengan fallback prerender script) | LOCKED | Lolos 22/22 kriteria POC: localized ID/EN, SEO meta, hreflang, sitemap, draft exclusion, Go static serving |
| ADR-025 | Public static files disajikan oleh Go | LOCKED | Tidak ada runtime web server tambahan |

## 4. Database, content & search

| ID | Keputusan | Status | Catatan |
|---|---|---|---|
| ADR-026 | Satu PostgreSQL database; tabel dimiliki modul | LOCKED | Tidak ada database/schema terpisah per modul |
| ADR-027 | ID menggunakan BIGINT identity; timestamp `timestamptz` UTC | LOCKED | FK, unique/check constraints |
| ADR-028 | Content identity dengan translation rows ID/EN | LOCKED | Data faktual bersama tidak diduplikasi |
| ADR-029 | PostgreSQL full-text search dengan GIN | LOCKED | Tidak memakai Elasticsearch/AI search MVP |
| ADR-030 | Konfigurasi tokenizer dan ranking ID/EN | OPEN | Perlu POC dan acceptance test |
| ADR-031 | Location master: kabupaten/kecamatan/gampong | LOCKED | Data sumber dan update cadence masih OPEN |
| ADR-032 | Koordinat hanya disimpan bila memiliki dasar; akurasi ditandai | LOCKED | Tidak boleh mengarang koordinat |
| ADR-033 | PostGIS tidak digunakan pada MVP | LOCKED | Tidak ada spatial search/route planner |

## 5. Security & operations

| ID | Keputusan | Status | Catatan |
|---|---|---|---|
| ADR-034 | Admin memakai server-side session dan opaque token hash-at-rest | LOCKED | Token tidak disimpan di localStorage |
| ADR-035 | Cookie HttpOnly, Secure, CSRF protection | LOCKED | Parameter cookie/expiry final masih OPEN |
| ADR-036 | Rich text dibatasi dan disanitasi server-side | LOCKED | Tidak menerima free HTML/JS |
| ADR-037 | Upload image dibatasi tipe, divalidasi konten, disimpan non-executable | LOCKED | Batas ukuran final masih OPEN |
| ADR-038 | Service berjalan non-root di Ubuntu 24.04 + systemd | LOCKED | Satu Go service |
| ADR-039 | TLS ditangani Go secara langsung | LOCKED | ACME challenge/renewal/reload masih OPEN |
| ADR-040 | Backup database dan media, termasuk salinan off-host terenkripsi | LOCKED | Retention, provider, RPO/RTO final masih OPEN |
| ADR-041 | Satu instance aplikasi sebagai baseline MVP | PROPOSED | Konsisten dengan in-memory rate limit; validasi kapasitas |
| ADR-042 | Observability memakai structured logs, health checks, metrics | LOCKED | Threshold/SLO final berdasarkan pengukuran |

## 6. Aturan pengelolaan keputusan

1. LOCKED hanya boleh berubah melalui keputusan eksplisit dan pembaruan register.
2. PROPOSED tidak boleh dianggap final sebelum disetujui.
3. OPEN harus memiliki tindakan penyelesaian dan fase paling lambat kapan dibutuhkan.
4. Keputusan yang berdampak ke schema harus diperbarui pada DDL dan migration.
5. Keputusan yang berdampak ke API harus diperbarui pada OpenAPI.
6. Setiap perubahan arsitektur harus diperiksa dampaknya ke QA, deployment, security, dan roadmap.
