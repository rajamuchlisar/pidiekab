# PidieKab

**Independent Regional Promotion & Visitor Discovery Platform — Kabupaten Pidie, Aceh**

PidieKab adalah platform digital independen (bukan website resmi pemerintah) yang dibangun untuk mempromosikan potensi daerah Kabupaten Pidie—khususnya destinasi wisata, kuliner & UMKM, serta seni budaya & sejarah—kepada masyarakat luas dan wisatawan domestik maupun mancanegara.

---

## Status Proyek — Production Ready (Phase 0 – Phase 9 Complete)

- **Backend:** Go 1.24 (`net/http`, Go standard library, Modular Monolith Architecture)
- **Database:** PostgreSQL 17 (Migrasi terkelola via Goose v3)
- **Frontend Public Site:** Vue 3 + TypeScript + Vite + Static Site Generation (`vite-ssg`)
- **Frontend Admin CMS:** Vue 3 + TypeScript + Vite (Single Page Application dengan Pinia & Vue Router)
- **Web Server & TLS:** Native Go `net/http` dengan modern TLS 1.2+ listener dan port 80 HTTP-to-HTTPS redirect (tanpa Nginx/Apache/reverse proxy)
- **Keamanan:** Argon2id password hashing, SHA-256 session tokens, double-submit CSRF, rate limiter, Content-Security-Policy (CSP), HSTS
- **CI / Quality Gate:** GitHub Actions (`.github/workflows/ci.yml`) dengan 100% test passing di backend, frontend, dan database migrations

---

## Struktur Repository

```text
PidieKab/
├── backend/                         # Backend Go application (Modular Monolith)
│   ├── cmd/pidiekab/main.go         # Application composition root & entrypoint
│   └── internal/                    # Domain modules & platform infrastructure
├── frontend/                        # Frontend applications
│   ├── public-site/                 # Public website (Vue 3 + TS + Vite + SSG)
│   └── admin-cms/                   # Admin CMS (Vue 3 + TS + Vite SPA)
├── migrations/                      # PostgreSQL DDL migrations (Goose format)
├── deploy/                          # Deployment & operational assets
│   ├── systemd/                     # systemd service unit configuration
│   └── scripts/                     # Deployment, backup & restore scripts
├── perencanaan/guide/               # Dokumentasi perencanaan, arsitektur & panduan (v79)
│   ├── 00-discovery/                # Riset & product discovery
│   ├── 01-technical/                # Dokumen arsitektur, API, database & security
│   ├── 02-implementation/           # Implementation roadmap & backlog
│   ├── 03-governance/               # ADR & tata kelola proyek
│   ├── docs/                        # Reports & operational runbooks
│   └── DOKUMENTASI-GUIDE.md         # Indeks master dokumentasi v79
├── .github/workflows/               # CI pipeline workflow
├── .env.example                     # Environment variables template
├── .gitignore                       # Git ignore definitions
└── README.md                        # Project overview & running instructions
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
- Health Ready: `GET http://127.0.0.1:8080/health/ready` (mengembalikan 200 OK jika PostgreSQL terkoneksi, 503 jika belum siap)
- Public Root: `GET http://127.0.0.1:8080/` (301 redirect otomatis ke `/id/`)
- Admin CMS: `GET http://127.0.0.1:8080/cms/` (Dashboard Administrasi PidieKab)

Untuk membuat akun Administrator awal:
```bash
./bin/pidiekab admin create --email admin@pidiekab.go.id --name "Super Administrator" --password "Admin12345!"
```

### 4. Inisialisasi Data Baseline & Konten Autentik (Seed CLI)
Untuk mengisi database dengan data master wilayah (Kabupaten Pidie + 23 kecamatan), kategori dwibahasa, pengaturan, halaman statis, artikel autentik terverifikasi, dan memicu kompilasi statis (SSG):
```bash
./bin/pidiekab seed
```
Akun default yang terbentuk:
- **Email:** `admin@pidiekab.go.id`
- **Password:** `AdminPidie123!`

Untuk membuat akun administrator tambahan:
```bash
./bin/pidiekab admin create --email nama@pidiekab.go.id --name "Nama Admin" --password "PasswordKuat123!"
```

### 5. Menjalankan & Membangun Frontend Public Site (SSG)
```bash
cd frontend/public-site
npm ci
npm run typecheck
npm run lint
npm run build
# Menjalankan acceptance test SSG POC
npm run test:ssg
```
Hasil build berupa HTML statis mandiri siap saji tersimpan di `frontend/public-site/dist/` lengkap dengan `sitemap.xml`.

### 6. Menjalankan & Membangun Frontend Admin CMS (SPA)
```bash
cd frontend/admin-cms
npm ci
npm run typecheck
npm run lint
npm run build
```
Hasil build SPA tersimpan di `frontend/admin-cms/dist/`.

### 7. Menjalankan Migrasi Database (PostgreSQL & Goose)
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

## Kompilasi Binary Produksi (Cross-Platform)

```bash
cd backend

# Binary untuk Windows
go build -o ../bin/pidiekab.exe ./cmd/pidiekab

# Binary untuk Linux Ubuntu 24.04 (ELF 64-bit x86-64)
GOOS=linux GOARCH=amd64 go build -o ../bin/pidiekab ./cmd/pidiekab
```

---

## Operasional & Deployment Produksi (Ubuntu 24.04 LTS)

Platform PidieKab dirancang untuk berjalan sebagai **Single Go Binary** native di bawah pengawasan `systemd` tanpa memerlukan Nginx, Apache, Redis, atau container bloat.

### 1. Inisialisasi Host / Provisioning Otomatis
Jalankan skrip provisioning host untuk membuat akun sistem `pidiekab` (tanpa shell login), struktur direktori dengan hak akses ketat (`/opt/pidiekab`, `/var/lib/pidiekab`, `/var/backups/pidiekab`), unit service systemd, dan penjadwalan cron backup harian:
```bash
sudo bash deploy/scripts/setup-server.sh
```
Setelah skrip selesai, sesuaikan kredensial database dan domain di `/etc/pidiekab/pidiekab.env`:
```bash
sudo nano /etc/pidiekab/pidiekab.env
```

### 2. Rilis Tanpa Downtime (Automated Deployment)
Skrip `deploy/scripts/deploy.sh` melakukan:
1. Pembuatan direktori rilis baru berbasis timestamp di `/opt/pidiekab/releases/<tag>/`.
2. Penyalinan binary Linux dan aset statis `dist/`.
3. Menjalankan migrasi database via Goose.
4. Pengalihan symlink atomik `/opt/pidiekab/current`.
5. Restart layanan `systemctl restart pidiekab`.
6. Health check probe (`/health/ready`). Jika gagal, otomatis melakukan rollback ke rilis sebelumnya.
7. Pembersihan rilis lama (menyimpan 5 rilis terakhir).

```bash
export PIDIEKAB_DATABASE_URL="postgres://pidiekab:password@127.0.0.1:5432/pidiekab_prod?sslmode=disable"
bash deploy/scripts/deploy.sh
```

### 3. Backup & Disaster Recovery Drill
- **Backup Terjadwal:**
  ```bash
  export PIDIEKAB_DATABASE_URL="postgres://pidiekab:password@127.0.0.1:5432/pidiekab_prod?sslmode=disable"
  bash deploy/scripts/backup.sh
  ```
  Menghasilkan logical dump PostgreSQL (`.dump`), arsip media (`.tar.gz`), dan verifikasi checksum SHA-256 di `/var/backups/pidiekab/`.

- **Restore & Recovery Drill:**
  ```bash
  bash deploy/scripts/restore.sh /var/backups/pidiekab/pidiekab_db_YYYYMMDD_HHMMSS.dump "postgres://.../target_db"
  ```

---

## Quality Gates & CI

Setiap pull request dan merge ke branch utama diverifikasi otomatis oleh GitHub Actions (`.github/workflows/ci.yml`):
1. **Backend Gate:** `gofmt` verification, `go vet ./...`, `go test -v -race ./...`, build binary.
2. **Frontend Public Site Gate:** `npm ci`, TypeScript typecheck (`vue-tsc`), lint, SSG build, acceptance tests.
3. **Frontend Admin CMS Gate:** `npm ci`, TypeScript typecheck (`vue-tsc`), lint, SPA build.
4. **Database Migration Gate:** Clean install `goose up`, rollback `goose down`, dan re-up idempotency test pada PostgreSQL 17.
