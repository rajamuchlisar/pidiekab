# Go Backend Implementation Guide

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `go-project-structure.md`
- `backend-bootstrap-and-dependency-wiring.md`
- `http-middleware-and-routing.md`

---

## Materi sumber: `go-project-structure.md`

# Go Project Structure — PidieKab

**Status:** Baseline implementation architecture  
**Backend:** Go + standard library `net/http`  
**Architecture:** Pragmatic Modular Monolith  
**Database:** PostgreSQL via `pgx/v5`  
**Deployment:** One Go binary/service

## 1. Design goals

Repository harus mudah dipahami dan dirawat oleh satu developer atau tim kecil. Struktur mengikuti business capability, dengan layer pragmatis.

- Satu Go module dan satu binary aplikasi.
- Satu proses HTTP server menggunakan `net/http`.
- Satu PostgreSQL database.
- Module berkomunikasi melalui pemanggilan Go function/interface internal, bukan HTTP.
- Hindari package `common`, `utils`, atau `helpers` yang menjadi tempat menumpuk kode tanpa ownership.
- Interface dibuat pada sisi consumer saat diperlukan untuk testing atau mengganti dependency; bukan untuk setiap struct.
- Hindari circular imports.
- Domain tidak mengimpor HTTP, PostgreSQL driver, Vue, atau infrastructure.
- SQL ditulis eksplisit dan disimpan dekat module pemilik data.

## 2. Proposed repository tree

```text
pidiekab/
├── cmd/pidiekab/main.go
├── internal/
│   ├── app/                         # composition root dan route wiring
│   ├── platform/
│   │   ├── config/
│   │   ├── database/
│   │   ├── httpserver/
│   │   │   └── middleware/
│   │   ├── authn/
│   │   ├── storage/
│   │   └── logging/
│   ├── modules/
│   │   ├── identity/
│   │   ├── content/
│   │   ├── category/
│   │   ├── media/
│   │   ├── location/
│   │   ├── curation/
│   │   ├── staticpage/
│   │   ├── settings/
│   │   ├── search/
│   │   ├── publication/
│   │   └── audit/
│   ├── web/
│   │   ├── public/
│   │   ├── admin/
│   │   └── ssg/
│   └── workers/sitegeneration/
├── frontend/
│   ├── public-site/                 # Vue 3 + TS + Vite + SSG
│   └── admin-cms/                   # Vue 3 + TS + Vite
├── migrations/
├── deploy/
│   ├── systemd/
│   └── scripts/
├── docs/
├── go.mod
├── go.sum
├── Makefile
├── .env.example
├── .gitignore
└── README.md
```

Tree ini adalah target struktur. Folder module dibuat bertahap ketika implementasi dimulai; jangan membuat banyak file kosong hanya untuk memenuhi diagram.

## 3. Package responsibilities

### `cmd/pidiekab`
Composition root executable: membaca config, membuat logger, membuka PostgreSQL pool, membangun dependency graph, membuat router/server, menjalankan worker, dan menangani graceful shutdown. Tidak berisi business logic.

### `internal/app`
Merangkai module dan mendaftarkan routes. Tempat wiring dependency konkret, bukan business logic.

### `internal/platform`
Cross-cutting technical capabilities:
- `config`: environment parsing dan validation;
- `database`: pool PostgreSQL dan transaction helper;
- `httpserver`: response/error encoding dan server lifecycle;
- `authn`: password/session/CSRF primitives;
- `storage`: penyimpanan file media;
- `logging`: konfigurasi `slog`.

Platform tidak boleh mengetahui aturan bisnis PidieKab.

### `internal/modules/{capability}`
Setiap module memiliki ownership atas use case dan data bisnisnya. Subfolder yang disarankan:
- `domain`: entity/value object/rule yang memiliki invariant;
- `application`: use cases dan orchestration;
- `infrastructure/postgres`: repository/query SQL;
- `transport/http`: handler, request DTO, response DTO, route registration.

Tidak semua module wajib memiliki seluruh layer jika belum dibutuhkan. Module tidak mengakses tabel module lain secara langsung melalui repository internal. Gunakan exported application query/service atau read-model yang ownership-nya jelas.

### `internal/web`
`public`: static website delivery; `admin`: CMS static files; `ssg`: HTML rendering/build. Tidak berisi business rules atau SQL langsung.

### `internal/workers/sitegeneration`
Worker membaca `site_generation_jobs`, memanggil application service/query module dan renderer. Tidak menyalin business logic.

### `frontend`
Dua aplikasi terpisah dalam satu repository. Node.js hanya diperlukan untuk development/build, bukan production runtime.

## 4. Dependency direction

```text
cmd/pidiekab
   -> internal/app
      -> module transport/http
         -> module application
            -> module domain
            -> consumer-owned interfaces/ports
               <- module infrastructure/postgres
      -> internal/web and workers
```

Handler memanggil use case; use case mengatur aturan aplikasi/transaksi; repository menjalankan SQL; domain menjaga invariant. Hindari domain model yang hanya menjadi salinan row database tanpa perilaku.

## 5. Cross-module transactions

Contoh publish content memerlukan category aktif, media valid, source tersedia, audit record, dan generation job. Orchestration berada pada application use case publikasi. Satu transaksi PostgreSQL digunakan untuk validasi dan perubahan yang harus atomic. `pgx.Tx` diberikan ke repository methods terkait; repository tidak membuka transaksi sendiri pada setiap operasi.

## 6. SQL ownership

Query SQL berada dekat module pemilik, misalnya `internal/modules/content/infrastructure/postgres/queries/`. Migration schema tetap terpusat di `/migrations`. Migration yang sudah diterapkan pada environment bersama tidak diedit; buat migration baru.

Gunakan query eksplisit, tanpa ORM dan generic repository.

## 7. Testing layout

- Unit test berdampingan dengan source (`*_test.go`).
- HTTP test memakai `net/http/httptest`.
- Repository integration test menggunakan PostgreSQL test database.
- Migration test membangun schema dari kosong dan menguji upgrade.
- SSG test memeriksa HTML, canonical, hreflang, sitemap, dan redirects.
- End-to-end test memverifikasi Admin → publish → job → static release.

## 8. Explicitly out of scope

Jangan membuat module/folder untuk microservices, event bus, Redis/cache service, GraphQL, Elasticsearch, public user accounts, operator portal, analytics pipeline, atau recommendation engine pada MVP.

---

## Materi sumber: `backend-bootstrap-and-dependency-wiring.md`

# Backend Bootstrap & Dependency Wiring — PidieKab

**Status:** Baseline implementation specification

## 1. Entry point

`cmd/pidiekab/main.go`:
1. membaca dan memvalidasi config;
2. membuat structured logger (`log/slog`);
3. membuat PostgreSQL pool;
4. melakukan dependency wiring;
5. membuat HTTP server;
6. menjalankan site-generation worker;
7. menangani OS signal dan graceful shutdown;
8. menutup worker dan database pool.

Main tidak berisi SQL, handler logic, atau business rules.

## 2. Process model

MVP memakai satu executable dan satu deployment service. HTTP server dan worker berjalan sebagai goroutine dalam proses yang sama; concurrency worker dibatasi konfigurasi. Jika kelak perlu pemisahan process, binary yang sama dapat mendukung mode `http`, `worker`, atau `all`, tetapi jangan menambah deployment complexity sebelum diperlukan.

Tetap satu HTTP server Go `net/http`, tanpa reverse proxy aplikasi tambahan.

## 3. Dependencies

Composition root membuat concrete dependencies:
- `pgxpool.Pool`;
- storage implementation;
- password hasher dan session service;
- repositories per module;
- application use cases;
- HTTP handlers;
- SSG renderer;
- generation worker.

Gunakan constructor eksplisit. Hindari global mutable singleton dan package-level database variable.

## 4. Environment configuration

Kelompok konfigurasi minimum:
- `APP_ENV`, `APP_ADDR`, `APP_PUBLIC_BASE_URL`, `APP_CMS_BASE_PATH`;
- `APP_READ_TIMEOUT`, `APP_WRITE_TIMEOUT`, `APP_IDLE_TIMEOUT`, `APP_SHUTDOWN_TIMEOUT`;
- `DATABASE_URL`, `DATABASE_MAX_CONNS`, `DATABASE_MIN_CONNS`, `DATABASE_MAX_CONN_LIFETIME`;
- `MEDIA_STORAGE_DRIVER`, `MEDIA_LOCAL_ROOT`, `MEDIA_PUBLIC_BASE_PATH`;
- `SESSION_COOKIE_NAME`, `SESSION_TTL`, `SESSION_IDLE_TIMEOUT`, `SESSION_SECURE_COOKIE`;
- `CSRF_SECRET` atau mekanisme token yang ditetapkan security spec;
- `SSG_OUTPUT_ROOT`, `SSG_RELEASES_TO_KEEP`, `SSG_WORKER_CONCURRENCY`, `SSG_JOB_MAX_ATTEMPTS`;
- `LOG_LEVEL`.

`.env` hanya untuk local development. Production membaca environment dari service manager/container. `.env.example` hanya berisi placeholder. Jangan commit secret.

Config kritis divalidasi saat startup; jika invalid, aplikasi gagal startup dengan pesan aman.

## 5. Startup order

1. Parse/validate config.
2. Initialize logger.
3. Open PostgreSQL pool.
4. Ping database dengan timeout.
5. Initialize storage dan cek path bila local.
6. Construct repositories/services.
7. Recover stale generation jobs.
8. Construct HTTP router.
9. Start worker loop.
10. Start HTTP server.
11. Emit startup log tanpa secret.

Migration tidak dijalankan otomatis setiap restart. Jalankan sebagai deployment step terkontrol.

## 6. Graceful shutdown

Saat menerima `SIGTERM`/`SIGINT`:
- hentikan penerimaan request baru;
- beri waktu request aktif selesai;
- hentikan claim job baru;
- tunggu job aktif sampai batas aman;
- job yang belum selesai harus dapat di-retry setelah lease kedaluwarsa;
- tutup HTTP server dan storage;
- tutup PostgreSQL pool terakhir.

Generation job harus idempotent dan aman diulang setelah proses mati mendadak.

## 7. HTTP server

Gunakan `http.Server` dengan `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`, dan graceful `Shutdown(ctx)`. Upload memiliki batas body sendiri. Publish hanya enqueue job dan mengembalikan `202`; proses build tidak dilakukan dalam request.

Go dapat melayani HTTPS langsung. Provisioning/renewal certificate, key permissions, TLS policy, dan HTTP→HTTPS redirect harus ditetapkan pada deployment specification sebelum production.

## 8. Dependency injection pattern

```go
type Dependencies struct {
    DB      *pgxpool.Pool
    Logger  *slog.Logger
    Storage storage.Store
}

func NewContentService(
    repo ContentRepository,
    category CategoryReader,
    media MediaReader,
    publisher PublicationQueue,
) *ContentService {
    // application orchestration
}
```

Contoh ilustratif. Definisikan interface pada sisi consumer saat diperlukan; jangan membuat interface untuk setiap concrete repository tanpa alasan.

## 9. Failure policy

- Database tidak tersedia: startup gagal atau readiness false.
- Storage wajib tidak tersedia: startup gagal.
- SSG generation gagal: public site tetap menyajikan release terakhir yang sukses.
- Worker job gagal: retry terbatas; satu job tidak boleh crash-kan HTTP process.
- Panic request: recovery middleware mencatat stack trace server-side dan mengirim 500 generik.
- Secret/config error: jangan mencetak nilai secret.

---

## Materi sumber: `http-middleware-and-routing.md`

# HTTP Middleware & Routing — PidieKab

**Status:** Baseline implementation specification  
**HTTP:** Go standard library `net/http`

## 1. Middleware order

Susunan konseptual:
1. Request ID
2. Panic recovery
3. Access logging
4. Security headers
5. CORS policy (same-origin default; aktifkan hanya bila dibutuhkan)
6. Global request/body limits
7. Router
8. Route-specific authentication, CSRF, rate limit, locale/validation

Urutan final harus memastikan access logger mencatat panic response dan request ID tersedia pada semua log.

## 2. Middleware responsibilities

### Request ID
Buat ID server-side untuk setiap request. Jika menerima ID dari client/proxy, validasi format dan panjang. Sertakan pada response header, context, structured logs, dan problem response.

### Panic recovery
Tangkap panic pada request boundary. Catat stack trace hanya di server log. Kirim problem response 500 generik bila response belum ditulis. Jangan menulis response kedua jika header/body sudah terkirim.

### Access logging
Catat method, sanitized path, status, duration, request ID, dan Admin ID bila tersedia. Jangan log password, cookie, authorization header, CSRF token, upload content, atau query sensitif.

### Security headers
Baseline:
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy` ketat sesuai kebutuhan
- `X-Frame-Options: DENY` atau CSP `frame-ancestors` setara
- Content Security Policy disesuaikan asset/build aktual
- `Strict-Transport-Security` hanya pada HTTPS production setelah domain policy diuji

### CORS
Public website dan CMS idealnya same-origin dengan API. Default tidak mengaktifkan permissive CORS. Bila domain terpisah, gunakan allowlist origin eksplisit; jangan wildcard bersama credential cookie.

### Request size
Gunakan `http.MaxBytesReader`. Tetapkan limit JSON dan limit upload secara terpisah. Validasi isi file, MIME, ekstensi, ukuran, dan dimensi.

### Authentication and CSRF
Session middleware memvalidasi cookie terhadap session store dan status Admin aktif. State-changing request berbasis cookie wajib CSRF-protected. GET/HEAD/OPTIONS tidak boleh mengubah state.

### Rate limit
MVP minimal untuk login dan search. In-memory limiter cukup untuk satu instance; bila multi-instance, evaluasi shared store berdasarkan kebutuhan nyata. Jangan menambah Redis hanya untuk antisipasi.

## 3. Route groups

```text
GET  /healthz
GET  /readyz

GET  /api/v1/public/{locale}/search
GET  /api/v1/public/{locale}/categories

POST /api/v1/admin/auth/login
POST /api/v1/admin/auth/logout
GET  /api/v1/admin/auth/me
POST /api/v1/admin/auth/password/change

/api/v1/admin/...

/id/... and /en/...       generated public pages
/cms/...                  Admin CMS static application
/assets/...               static assets
/sitemap.xml
/robots.txt
```

Pilih satu CMS base path (`/cms` atau `/admin`) saat frontend routing ditetapkan; jangan gunakan keduanya. API tetap pada `/api/v1/admin`.

## 4. Go ServeMux

Gunakan `http.NewServeMux()` dan method-aware patterns pada versi Go yang ditetapkan di `go.mod`. Jangan membangun router abstraction sendiri kecuali ServeMux terbukti tidak cukup.

## 5. Static serving safety

- Resolve request path hanya di dalam active release root.
- Reject path traversal dan encoded traversal.
- Jangan expose directory listing.
- Set content type yang benar.
- Cache hashed immutable assets secara panjang; HTML memiliki cache policy yang memungkinkan update.
- Draft/preview tidak boleh berada pada public root.
- Redirect lookup dilakukan sebelum static file fallback.
- Jangan fallback semua unknown content route ke CMS SPA.
- CMS fallback hanya pada prefix CMS.
- Jangan membiarkan path publik yang tidak dikenal tertutup oleh SPA fallback.

## 6. Error handling

API memakai `application/problem+json`. Public HTML route mengembalikan HTTP status yang benar dengan halaman generik; custom branded error pages tidak termasuk MVP.

## 7. TLS

Go dapat melayani HTTPS langsung. Certificate issuance/renewal, key permissions, TLS policy, dan HTTP→HTTPS redirect harus didokumentasikan pada deployment spec. DNS provider tidak otomatis menjamin origin TLS.

---
