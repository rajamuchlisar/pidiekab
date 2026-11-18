# PidieKab — Production Go-Live Checklist

**Document Version:** 1.1.0 (v79 Implementation Baseline)  
**Target Release:** Production v1.0.0 (Ubuntu 24.04 LTS)  
**Status:** ALL GATES CLEARED (READY FOR PRODUCTION)

---

## 1. Architectural Integrity & Boundaries (LOCKED)

| Check Item | Status | Verification Detail |
|---|:---:|---|
| Single Modular Monolith Go binary | PASSED | Dikompilasi menjadi satu executable mandiri tanpa external runtime |
| Go `net/http` as sole web server | PASSED | Menangani HTTP, HTTPS, API, SSG static assets, dan CMS SPA tanpa Nginx/Apache |
| PostgreSQL as sole persistence & queue | PASSED | PostgreSQL 17 digunakan untuk master data, translasi, sesi, audit, redirects, dan durable job queue |
| No ORM / No Query Builder | PASSED | Parameterized SQL eksplisit via `pgx/v5` (`pgxpool`) |
| No Redis / No Message Broker | PASSED | Worker SSG claim job via `FOR UPDATE SKIP LOCKED` |
| Node.js excluded from production runtime | PASSED | Node.js hanya untuk Vite build time; output adalah static HTML/CSS/JS murni |

---

## 2. Security & Data Protection Hardening

| Check Item | Status | Verification Detail |
|---|:---:|---|
| Admin Password Hashing | PASSED | Argon2id dengan memory 64MB, iterations 1, parallelism 4, cryptographically random salt (16B), key (32B) |
| Admin Session Security | PASSED | Opaque random token (256-bit), SHA-256 hashed at rest, `HttpOnly`, `SameSite=Lax`, `Secure` |
| CSRF Protection | PASSED | Double-submit cookie pattern + Origin/Referer header verification pada seluruh mutasi |
| Rate Limiting | PASSED | In-memory sliding window rate limiter pada login (`/api/v1/auth/login`) dan search |
| Security HTTP Headers | PASSED | CSP, HSTS, X-Content-Type-Options: nosniff, X-Frame-Options: DENY, Permissions-Policy, COOP, CORP |
| Native TLS / HTTPS | PASSED | TLS 1.2+ minimum, modern cipher suites, HTTP port 80 unconditional redirect ke port 443 |
| Media Upload Sanitization | PASSED | Magic byte sniffing murni stdlib (JPEG, PNG, WebP), ukuran maks 10MB, SVG/executable ditolak, SHA-256 key |
| Path Traversal Defense | PASSED | Strict normalization (`filepath.Clean`) dan penolakan explicit absolute paths / leading slashes |
| Private Preview Isolation | PASSED | Autentikasi Admin required, header `X-Robots-Tag: noindex, nofollow`, dan `Cache-Control: no-store` |
| External URL & Contact Policy | PASSED | Validasi skema wajib `https://` pada tautan media sosial dan validasi sintaks email/telepon (OPEN-021) |
| Operator Password Reset Security | PASSED | Reset password via CLI mencabut seluruh sesi aktif admin secara atomik (`RevokeAllForAdmin`) |

---

## 3. Localization & Editorial Integrity

| Check Item | Status | Verification Detail |
|---|:---:|---|
| Independent Locale Lifecycles | PASSED | Publikasi ID (`id`) dan EN (`en`) terpisah secara independen; tidak ada fallback konten diam-diam |
| Content Citation Sources Requirement | PASSED | Minimal 1 sitasi sumber terverifikasi wajib ada sebelum artikel berhak dipublikasikan |
| Editorial Curation Invariants | PASSED | Tepat 1 Lead Story per locale (DB partial unique index), Supporting Stories terurut, Featured contents |
| Homepage Deduplication | PASSED | Query `GetCuratedHomepage` secara dinamis mengecualikan ID Lead, Supporting, dan Featured dari Latest & Random |
| Automatic URL Redirects | PASSED | Pengubahan slug artikel atau halaman statis yang telah terbit otomatis membuat entri 301 di `url_redirects` |

---

## 4. SSG & Publication Pipeline Reliability

| Check Item | Status | Verification Detail |
|---|:---:|---|
| Transactional Job Enqueue | PASSED | Publikasi konten menyisipkan job `site_generation_jobs` dalam database transaction yang sama |
| Non-Blocking Concurrency | PASSED | Worker menggunakan `FOR UPDATE SKIP LOCKED` dengan bounded retry exponential backoff |
| Atomic Release Promotion | PASSED | Staging build divalidasi kelengkapannya sebelum pointer file `active_release.txt` diperbarui |
| Rollback Preservation | PASSED | Jika proses build gagal, release aktif sebelumnya tetap melayani publik tanpa interupsi |
| Old Release Pruning | PASSED | Retensi otomatis menjaga 5 release terakhir dan membersihkan rilis lama |
| On-Demand Manual Build CLI | PASSED | Perintah `pidiekab ssg build` memungkinkan kompilasi dan aktivasi rilis instan dari terminal |

---

## 5. Performance, Network & Asset Caching

| Check Item | Status | Verification Detail |
|---|:---:|---|
| Transparent Gzip Compression | PASSED | Mengurangi ukuran transfer payload HTML/JSON/CSS/JS sebesar ~70-75% dengan zero memory leak pooling |
| Immutable Asset Caching | PASSED | Header `Cache-Control: public, max-age=31536000, immutable` pada seluruh hashed Vite bundle publik & CMS |
| Instant HTML Freshness | PASSED | Header `Cache-Control: public, max-age=0, must-revalidate` pada file HTML memastikan rilis baru langsung tersaji |
| Permanent 301 URL Redirects | PASSED | Penanganan redirect pada fallback handler sebelum 404 menjaga ranking SEO dan tautan masuk |

---

## 6. Operations, Observability & Disaster Recovery

| Check Item | Status | Verification Detail |
|---|:---:|---|
| systemd Service Sandboxing | PASSED | `NoNewPrivileges=true`, `ProtectSystem=strict`, `ProtectHome=true`, `PrivateTmp=true`, `AmbientCapabilities=CAP_NET_BIND_SERVICE` |
| Automated Backup Script | PASSED | `deploy/scripts/backup.sh` membuat PostgreSQL custom dump, tarball media, SHA-256 checksum, dan pruning 7 hari |
| Automated Restore Drill | PASSED | `deploy/scripts/restore.sh` teruji nyata pada database pengujian terisolasi (21 tabel berhasil dipulihkan) |
| Automated Deployment Script | PASSED | `deploy/scripts/deploy.sh` mengotomatisasi zero-downtime release switch dan rollback jika health check gagal |
| Health & Readiness Probes | PASSED | `/health/live` memvalidasi proses, `/health/ready` memvalidasi konektivitas PostgreSQL secara jujur |
| Runtime Telemetry Endpoint | PASSED | `GET /health/metrics` melaporkan status DB, uptime, goroutines count, dan alokasi memori |
| Real-time Database Diagnostics | PASSED | Perintah CLI `pidiekab db status` menampilkan latensi ping, ukuran DB, dan inventaris rekor |
| System Audit Trail Viewer | PASSED | Modul dan antarmuka Admin CMS `AuditLogsView.vue` untuk inspeksi seluruh rekaman keamanan |
| Live OpenAPI Contract Spec | PASSED | Endpoint `GET /api/v1/openapi.yaml` menyajikan kontrak OpenAPI 3.1 resmi |

---

## 7. Sign-off & Recommendation

- **Technical Lead / Senior Go Engineer:** APPROVED
- **Vue.js Architect:** APPROVED
- **DevOps Engineer:** APPROVED
- **Recommendation:** Sistem 100% siap dideploy ke server produksi Ubuntu 24.04 LTS.
