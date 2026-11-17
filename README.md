# PidieKab

**Independent Regional Promotion & Visitor Discovery Platform — Kabupaten Pidie, Aceh**

PidieKab adalah platform digital independen (bukan website resmi pemerintah) yang dibangun untuk mempromosikan potensi daerah Kabupaten Pidie—khususnya destinasi wisata, kuliner & UMKM, serta seni budaya & sejarah—kepada masyarakat luas dan wisatawan domestik maupun mancanegara.

---

## Status Proyek — Phase 0: Repository & Development Foundation

- **Backend:** Go 1.24 (`net/http`, Go standard library, Modular Monolith)
- **Database:** PostgreSQL 15+ (Didukung migrasi via Goose v3)
- **Frontend Public Site:** Vue 3 + TypeScript + Vite + Static Site Generation (`vite-ssg`)
- **Frontend Admin CMS:** Vue 3 + TypeScript + Vite (Single Page Application)
- **Reverse Proxy:** **None** (Go bertindak sebagai satu-satunya application web server di production)
- **CI / Quality Gate:** GitHub Actions (`.github/workflows/ci.yml`)

---

## Struktur Repository

```text
PidieKab/
├── backend/                         # Backend Go application
│   ├── cmd/pidiekab/main.go         # Application composition root & entrypoint
│   └── internal/
│       └── platform/
│           ├── config/              # Environment config loader & strict validation
│           └── httpserver/          # HTTP server, routing & middleware
│               └── middleware/      # RequestID, logging, security headers, recovery
├── frontend/
│   ├── public-site/                 # Public website (Vue 3 + TS + Vite + vite-ssg)
│   └── admin-cms/                   # Admin CMS (Vue 3 + TS + Vite SPA + Pinia)
├── migrations/                      # PostgreSQL DDL migrations (Goose format)
├── deploy/
│   ├── systemd/                     # systemd service unit configuration
│   └── scripts/                     # Automated deployment script
├── docs/
│   └── reports/                     # Implementation & evaluation reports
├── tasks/                           # Planning & execution tracking
├── .github/workflows/               # CI pipeline workflow
├── .env.example                     # Environment variables template
├── .gitignore                       # Git ignore definitions
└── README.md                        # Project documentation
```

---

## Panduan Menjalankan dari Mesin Bersih (Clean Machine Setup)

### 1. Prasyarat Sistem
- **Go:** Versi 1.24 atau lebih baru
- **Node.js:** Versi 22.x atau 24.x & npm 10+
- **PostgreSQL:** Versi 15, 16, atau 17
- **Goose CLI:** `go install github.com/pressly/goose/v3/cmd/goose@latest`

### 2. Konfigurasi Lingkungan
Salin file konfigurasi lingkungan:
```bash
cp .env.example .env
```

### 3. Menjalankan Backend Go
```bash
cd backend
# Menjalankan unit tests
go test -v ./...

# Memeriksa static analysis / vet
go vet ./...

# Menjalankan server backend
go run ./cmd/pidiekab
```
Server akan aktif di `http://127.0.0.1:8080`.
- Health Live: `GET http://127.0.0.1:8080/health/live` (mengembalikan 200 OK)
- Health Ready: `GET http://127.0.0.1:8080/health/ready` (mengembalikan 503 Not Ready - Database belum tersambung pada Phase 0)
- Public Root: `GET http://127.0.0.1:8080/` (301 redirect otomatis ke `/id/`)

### 4. Menjalankan & Membangun Frontend Public Site (SSG)
```bash
cd frontend/public-site
npm ci
npm run typecheck
npm run lint
npm run build
# Menjalankan acceptance test SSG POC
node scripts/test-ssg-poc.mjs
```
Hasil build berupa HTML statis mandiri siap saji tersimpan di `frontend/public-site/dist/`.

### 5. Menjalankan & Membangun Frontend Admin CMS (SPA)
```bash
cd frontend/admin-cms
npm ci
npm run typecheck
npm run lint
npm run build
```
Hasil build SPA tersimpan di `frontend/admin-cms/dist/`.

### 6. Menjalankan Migrasi Database (PostgreSQL & Goose)
Pastikan PostgreSQL berjalan dan database target (misalnya `pidiekab_dev`) telah dibuat:
```bash
# Migrasi ke versi terbaru (Up)
goose -dir migrations postgres "postgres://postgres:password@localhost:5432/pidiekab_dev?sslmode=disable" up

# Memeriksa status migrasi
goose -dir migrations postgres "postgres://postgres:password@localhost:5432/pidiekab_dev?sslmode=disable" status

# Menguji rollback (Down)
goose -dir migrations postgres "postgres://postgres:password@localhost:5432/pidiekab_dev?sslmode=disable" down
```

---

## Quality Gates & CI

Setiap pull request dan merge ke branch utama diverifikasi otomatis oleh GitHub Actions (`.github/workflows/ci.yml`):
1. **Backend Gate:** `gofmt` verification, `go vet ./...`, `go test -v -race ./...`, build binary.
2. **Frontend Public Site Gate:** `npm ci`, TypeScript typecheck (`vue-tsc`), lint, SSG build, acceptance tests.
3. **Frontend Admin CMS Gate:** `npm ci`, TypeScript typecheck (`vue-tsc`), lint, SPA build.
4. **Database Migration Gate:** Clean install `goose up`, rollback `goose down`, dan re-up idempotency test pada PostgreSQL 17.
