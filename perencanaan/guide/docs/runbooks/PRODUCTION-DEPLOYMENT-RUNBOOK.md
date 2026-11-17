# Production Deployment Runbook — PidieKab

**Target Environment:** Ubuntu Server 24.04 LTS  
**Application Stack:** Single Go Binary (`pidiekab`), Native `net/http` (TLS 1.2/1.3), PostgreSQL 17, systemd  
**Status:** Operational Production Standard

---

## 1. Arsitektur Deployment & Direktori Server

PidieKab berjalan sebagai satu proses mandiri tanpa reverse proxy tambahan (tanpa Nginx/Apache/Caddy) menggunakan model immutable release symlink:

```text
/opt/pidiekab/
├── releases/
│   ├── 20261117_120000/
│   │   ├── bin/pidiekab
│   │   ├── public/       # Static assets Vite SSG
│   │   └── cms/          # Vue SPA admin CMS
│   └── 20261117_130000/
└── current -> /opt/pidiekab/releases/20261117_130000

/var/lib/pidiekab/
├── media/               # Storage file upload publik (read/write non-executable)
└── ssg/                 # Output site generation aktif dan staging

/etc/pidiekab/
├── pidiekab.env         # Environment variables & secrets (chmod 0640, pidiekab:pidiekab)
└── tls/                 # Sertifikat SSL/TLS & private key (cert.pem, key.pem)
```

---

## 2. Persyaratan Awal (Server Provisioning)

### 2.1 Pembuatan Linux User Non-Root
```bash
sudo useradd -r -s /usr/sbin/nologin -d /opt/pidiekab pidiekab
sudo mkdir -p /opt/pidiekab/releases /var/lib/pidiekab/media /var/lib/pidiekab/ssg /etc/pidiekab/tls
sudo chown -R pidiekab:pidiekab /opt/pidiekab /var/lib/pidiekab /etc/pidiekab
```

### 2.2 Konfigurasi Environment File (`/etc/pidiekab/pidiekab.env`)
```bash
sudo tee /etc/pidiekab/pidiekab.env > /dev/null << 'EOF'
PIDIEKAB_ENV=production
PIDIEKAB_ADDR=:443
PIDIEKAB_HTTP_REDIRECT_ADDR=:80
PIDIEKAB_LOG_LEVEL=info
PIDIEKAB_DATABASE_URL=postgres://pidiekab_user:STRONG_PASSWORD_HERE@127.0.0.1:5432/pidiekab?sslmode=disable
PIDIEKAB_DATABASE_MAX_CONNS=20
PIDIEKAB_DATABASE_MIN_CONNS=4
PIDIEKAB_STORAGE_PATH=/var/lib/pidiekab/media
PIDIEKAB_SSG_OUTPUT_ROOT=/var/lib/pidiekab/ssg
PIDIEKAB_TLS_CERT_FILE=/etc/pidiekab/tls/fullchain.pem
PIDIEKAB_TLS_KEY_FILE=/etc/pidiekab/tls/privkey.pem
EOF

sudo chmod 0640 /etc/pidiekab/pidiekab.env
sudo chown root:pidiekab /etc/pidiekab/pidiekab.env
```

### 2.3 Pemasangan systemd Service
Salin file unit `deploy/systemd/pidiekab.service` ke `/etc/systemd/system/`:
```bash
sudo cp deploy/systemd/pidiekab.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable pidiekab.service
```

---

## 3. Prosedur Rilis Baru (Zero-Downtime Atomic Deployment)

Eksekusi deployment dapat dijalankan secara otomatis melalui script `deploy/scripts/deploy.sh` atau langkah manual berikut:

1. **Build Artifacts pada CI/Staging:**
   ```bash
   # Build Frontend
   cd frontend/public-site && npm run build
   cd ../admin-cms && npm run build
   
   # Build Go Linux Binary
   cd ../../backend
   GOOS=linux GOARCH=amd64 go build -mod=readonly -ldflags="-s -w" -o ../bin/pidiekab ./cmd/pidiekab
   ```

2. **Jalankan Deployment Script di Server:**
   ```bash
   ./deploy/scripts/deploy.sh
   ```

3. **Verifikasi Kesiapan (Post-Deployment Verification):**
   ```bash
   # 1. Periksa status unit systemd
   sudo systemctl status pidiekab.service

   # 2. Periksa health probe
   curl -I https://pidiekab.go.id/health/live
   curl -s https://pidiekab.go.id/health/ready

   # 3. Periksa security headers
   curl -I https://pidiekab.go.id/id/
   ```

---

## 4. Prosedur Rollback Cepat

Jika versi rilis baru menimbulkan anomali di level aplikasi:

1. **Kembalikan symlink `current` ke rilis sebelumnya:**
   ```bash
   sudo ln -sfn /opt/pidiekab/releases/<PREVIOUS_RELEASE_TIMESTAMP> /opt/pidiekab/current
   ```
2. **Restart service:**
   ```bash
   sudo systemctl restart pidiekab.service
   ```
3. **Verifikasi liveness:**
   ```bash
   curl -s http://127.0.0.1:8080/health/ready
   ```

