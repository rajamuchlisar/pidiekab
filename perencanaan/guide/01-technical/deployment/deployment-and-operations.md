# Deployment And Operations

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `deployment-architecture.md`
- `systemd-and-process-management.md`
- `deployment-runbook.md`
- `database-backup-and-recovery.md`

---

## Materi sumber: `deployment-architecture.md`

# Deployment Architecture — PidieKab

**Status:** Baseline deployment specification  
**Target OS:** Ubuntu Server 24.04 LTS  
**Application:** One Go binary, Go `net/http`  
**Database:** PostgreSQL  
**Frontend runtime:** Static files served by Go  
**Architecture:** Single application deployment; database placement configurable

## 1. Deployment principles

- Go `net/http` menjadi satu-satunya application web server.
- Tidak menggunakan Nginx, Caddy, Apache, LiteSpeed, atau reverse proxy aplikasi.
- Satu Go binary melayani REST API, public SSG files, CMS static files, redirect, dan health endpoints.
- SSG worker berjalan dalam proses aplikasi yang sama pada MVP.
- PostgreSQL adalah source of truth dan durable job queue.
- Tidak ada Redis, message broker, atau microservices pada MVP.
- Node.js tidak dipasang sebagai production web runtime.
- Secrets tidak disimpan di Git atau frontend build output.

## 2. Logical topology

```text
Visitor / Admin Browser
          |
       Internet
          |
    TCP 80 / 443
          |
   Go PidieKab Service
   ├── Public SSG files
   ├── CMS static files
   ├── Public REST API
   ├── Admin REST API
   ├── Redirect / SEO files
   └── SSG generation worker
          |
    ┌─────┴─────────┐
    |               |
 PostgreSQL     Media Storage
```

PostgreSQL dapat berada pada host yang sama atau host/database server terpisah. Untuk MVP, keputusan mengikuti kapasitas VPS, kebutuhan backup, dan network security. Jika satu host, PostgreSQL hanya bind ke loopback/private socket. Jika terpisah, gunakan private network, TLS database connection, dan firewall allowlist dari application host.

## 3. Ubuntu baseline

- Ubuntu Server 24.04 LTS.
- Dedicated non-root Linux user, misalnya `pidiekab`.
- Application binary dan static assets berada pada direktori release yang tidak writable oleh runtime user.
- Media uploads berada pada direktori data persisten terpisah dari release.
- Configuration/secrets berada di `/etc/pidiekab/` dengan permission ketat atau systemd credentials.
- Logs dikirim ke journald; retention ditetapkan.
- Security updates diterapkan dengan prosedur terkontrol.

## 4. Directory layout

Contoh layout:

```text
/opt/pidiekab/
├── releases/
│   ├── 2026xxxx-xxxx/
│   │   ├── pidiekab
│   │   ├── public/
│   │   └── cms/
│   └── ...
└── current -> /opt/pidiekab/releases/<active-release>

/var/lib/pidiekab/
├── media/
├── ssg-releases/
└── tmp/

/etc/pidiekab/
├── pidiekab.env
└── tls/                 # bila certificate files dikelola di sini

/var/log/                # journald; bukan log file aplikasi wajib
```

Path dapat disesuaikan. Release code/static assets tidak boleh bercampur dengan mutable media uploads. Active SSG release dan app release dapat dikelola terpisah agar generation tidak mengubah executable release.

## 5. Network and firewall

Minimum inbound:
- TCP 80: HTTP redirect ke HTTPS dan/atau ACME challenge bila metode yang dipilih memerlukannya.
- TCP 443: HTTPS Go server.
- SSH: dibatasi ke IP/admin access yang diperlukan.

Jangan expose PostgreSQL `5432` ke public internet. Jangan membuka port worker atau internal diagnostics ke publik.

Outbound yang dibutuhkan:
- DNS resolution;
- package/security updates;
- ACME certificate authority dan DNS provider API bila DNS-01 dipilih;
- backup destination;
- optional SMTP provider jika reset password/email notification nanti diaktifkan.

## 6. TLS termination

TLS diakhiri langsung oleh Go server. Tidak ada TLS reverse proxy.

Go server melayani:
- HTTP listener port 80 untuk redirect ke HTTPS (dan ACME challenge bila diperlukan);
- HTTPS listener port 443 untuk seluruh public/admin/API traffic.

Certificate provisioning/renewal harus berjalan terpisah dari request handling dan tidak boleh mengharuskan menambah reverse proxy. Kandidat:
- DNS-01 automation dengan DNS provider API;
- ACME client yang menyimpan certificate/key pada path aman;
- Go server memuat certificate terbaru melalui `tls.Config.GetCertificate` atau mekanisme reload yang teruji.

Jika domain memakai Cloudflare DNS, DNS-01 dapat menjadi opsi yang tidak memerlukan inbound challenge path. API token harus dibatasi ke zone/DNS edit yang diperlukan dan disimpan sebagai secret. Tool ACME dan renewal mechanism belum dikunci; uji renewal end-to-end sebelum production.

## 7. Deployment release process

1. Build frontend public SSG dan CMS pada CI/build machine.
2. Build Go binary untuk target Ubuntu/Linux architecture.
3. Jalankan tests, typecheck, lint, migration checks, dan SSG output validation.
4. Upload artifact ke release directory baru.
5. Validasi konfigurasi dan file permission.
6. Jalankan database migration sebagai step terpisah dan terkontrol.
7. Restart/reload service dengan systemd.
8. Health/readiness check.
9. Smoke test public homepage, search, CMS login, dan asset delivery.
10. Tandai release aktif setelah validasi.

Rollback aplikasi harus mempertahankan kompatibilitas schema. Migration destructive memerlukan backup dan rencana forward-fix; rollback binary tidak otomatis membalikkan schema.

## 8. Availability behavior

- Jika generation gagal, public site tetap melayani SSG release terakhir yang sukses.
- Jika PostgreSQL tidak tersedia, Admin/API yang memerlukan DB tidak ready; file publik yang sudah digenerate idealnya tetap dapat dilayani oleh Go process selama proses hidup.
- Jika media storage unavailable, jangan menghapus/mengganti media lama; tampilkan fallback untuk media yang gagal dibaca dan catat error.
- Restart aplikasi tidak boleh menghilangkan media atau SSG releases.

## 9. Deferred infrastructure

Tidak ada Kubernetes, Docker Swarm, service mesh, Redis, message broker, CDN requirement, atau multi-region pada MVP. Docker Compose bukan keputusan wajib untuk deployment production; systemd native dapat digunakan untuk satu service Go di Ubuntu.

---

## Materi sumber: `systemd-and-process-management.md`

# systemd & Process Management — PidieKab

**Status:** Baseline operational specification

## 1. Service model

Jalankan Go binary sebagai satu systemd service, misalnya `pidiekab.service`. Runtime memakai dedicated non-root user. Worker SSG berjalan dalam proses yang sama pada MVP.

## 2. Example unit file

Sesuaikan paths, user, group, environment file, dan binary sebelum dipakai:

```ini
[Unit]
Description=PidieKab Application
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=pidiekab
Group=pidiekab
WorkingDirectory=/opt/pidiekab/current
EnvironmentFile=/etc/pidiekab/pidiekab.env
ExecStart=/opt/pidiekab/current/pidiekab
Restart=on-failure
RestartSec=5
TimeoutStartSec=30
TimeoutStopSec=90
KillSignal=SIGTERM

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/pidiekab
ReadOnlyPaths=/etc/pidiekab

LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Systemd hardening perlu diuji dengan storage path dan certificate access aktual. Jika app perlu membaca certificate dari path khusus, berikan read-only permission eksplisit. Jangan menjalankan service sebagai root untuk mengatasi permission problem.

## 3. Environment and secrets

`/etc/pidiekab/pidiekab.env` harus dimiliki root dan hanya dapat dibaca root/service sesuai pendekatan yang dipilih (misalnya group khusus atau systemd credentials). Permission ketat, umumnya `0640` atau lebih ketat. Jangan menyimpan secret dalam unit file, Git, atau command-line arguments yang dapat terlihat pada process list.

Environment minimum mengikuti backend bootstrap spec:
- APP environment/address/base URL;
- DATABASE URL/pool;
- storage paths;
- session/CSRF secret;
- SSG paths/worker limits;
- logging level.

Jangan menyalin contoh credential ke production. Rotate secret jika pernah terekspos.

## 4. Service lifecycle

Operasi umum:
- `systemctl daemon-reload` setelah unit berubah;
- `systemctl enable --now pidiekab` untuk enable/start;
- `systemctl status pidiekab` untuk status;
- `journalctl -u pidiekab` untuk log;
- `systemctl restart pidiekab` untuk restart terkontrol.

Aplikasi wajib menangani SIGTERM dengan graceful shutdown sesuai backend spec.

## 5. Resource controls

Tetapkan CPU/memory limits setelah load test, bukan menebak. Awali dengan monitoring RSS, CPU, open files, database pool usage, generation duration, dan disk. Jika menggunakan systemd resource controls, uji agar generation worker tidak menghabiskan seluruh resource hingga HTTP serving terganggu.

## 6. Health checks and restart

- `/healthz`: process liveness.
- `/readyz`: readiness dependency kritis.
- Restart policy `on-failure`; hindari restart loop tanpa log/alert.
- Health check eksternal dapat memantau HTTPS homepage dan API readiness.
- Jangan expose diagnostic endpoints yang berisi detail internal.

---

## Materi sumber: `deployment-runbook.md`

# Deployment & Operations Runbook — PidieKab

**Status:** Baseline runbook; commands must be adapted and tested on staging

## 1. Pre-deployment checklist

- Domain DNS mengarah ke host.
- Port 80/443 terbuka sesuai kebutuhan.
- SSH dibatasi ke administrator.
- Go binary dan frontend artifacts sudah lulus CI.
- PostgreSQL tersedia dan backup terbaru tervalidasi.
- Environment variables/secrets tersedia dengan permission aman.
- Media dan SSG data directories writable oleh service user.
- TLS certificate valid dan renewal procedure telah diuji.
- Migration telah direview dan memiliki backup/forward-fix plan.

## 2. Release deployment

1. Buat release directory baru.
2. Upload binary, public SSG artifacts, dan CMS static build.
3. Atur owner/group dan permission.
4. Jalankan migration terkontrol bila ada.
5. Update active application release pointer.
6. Restart service.
7. Periksa `systemctl status`.
8. Periksa `/healthz` dan `/readyz`.
9. Smoke test homepage ID/EN, category/detail page, search, CMS login, image/media, dan sitemap.
10. Pantau log dan disk usage.

Jangan menghapus release sebelumnya sebelum release baru lolos smoke test dan rollback window selesai.

## 3. Rollback

- Hentikan deployment jika health/readiness gagal.
- Kembalikan active application release ke binary/static assets sebelumnya.
- Restart service dan smoke test.
- Jika migration sudah berjalan, jangan otomatis downgrade database. Evaluasi kompatibilitas schema dan gunakan forward-fix atau restore sesuai prosedur.
- Public SSG release terakhir yang sukses tetap tersedia saat generation baru gagal.

## 4. Operational monitoring

Pantau:
- service up/down dan restart count;
- HTTPS certificate expiry;
- HTTP 5xx rate dan response latency;
- PostgreSQL availability, connections, disk, slow queries;
- disk space untuk media, SSG releases, temp, backups;
- generation queue pending/failed age;
- backup success/failure dan restore drill;
- CPU/RAM dan worker concurrency.

MVP dapat menggunakan journald dan external uptime check sederhana. Tidak wajib memasang observability stack kompleks.

## 5. Routine maintenance

- Security updates Ubuntu dan Go/frontend dependencies secara terjadwal.
- Review Go module/npm dependency vulnerabilities.
- Periksa disk usage dan cleanup temporary/old release sesuai retention.
- Verifikasi backup job.
- Uji TLS renewal.
- Audit Admin account aktif dan akses server.
- Rotasi secrets jika ada indikasi exposure.
- Jalankan restore drill berkala.

## 6. Incident response

### Application unavailable
1. Periksa service status dan journal.
2. Periksa readiness/database connectivity.
3. Periksa disk space dan permission.
4. Rollback ke release terakhir yang diketahui baik jika deployment baru menjadi penyebab.
5. Catat request ID dan waktu kejadian.

### Database unavailable
1. Jangan menghapus/recreate data directory.
2. Periksa PostgreSQL service, disk, logs, dan network.
3. Pulihkan konektivitas.
4. Restore backup hanya bila recovery decision disetujui dan backup tervalidasi.

### TLS failure
1. Periksa certificate validity/path/permission.
2. Periksa ACME renewal logs dan DNS provider API.
3. Jangan menonaktifkan HTTPS sebagai solusi permanen.
4. Perbaiki renewal dan reload certificate Go server.

### Disk full
1. Identifikasi media, SSG releases, temp, logs, backups.
2. Hapus hanya file temporary/retention-expired yang telah diverifikasi.
3. Jangan menghapus media user atau backup terakhir.
4. Pastikan backup dan generation berjalan kembali.

---

## Materi sumber: `database-backup-and-recovery.md`

# Database Backup & Recovery — PidieKab

**Status:** Baseline operational specification

## 1. Data to protect

Backup harus mencakup:
1. PostgreSQL database (semua schema/data PidieKab).
2. Media storage original files dan metadata yang tersimpan di DB.
3. SSG active release dan konfigurasi deployment yang diperlukan untuk recovery cepat.
4. Application configuration/secrets melalui secret backup terkontrol, bukan disimpan bersama backup publik.
5. Migration files dan build artifacts/source release.

SSG HTML dapat dibangun ulang dari database dan media, tetapi backup release terakhir mempercepat recovery.

## 2. Backup strategy

MVP baseline:
- PostgreSQL logical backup harian menggunakan `pg_dump` format custom.
- Backup media incremental/daily sesuai volume dan kemampuan storage.
- Salinan backup disimpan di luar VPS/host utama.
- Backup terenkripsi saat transit dan at-rest.
- Retention awal yang diusulkan: 7 daily, 4 weekly, 3 monthly. Sesuaikan kapasitas dan kebutuhan bisnis.
- Backup job mengirim status sukses/gagal ke log/monitoring.
- Backup tidak dianggap berhasil hanya karena command exit 0; validasi ukuran, exit status, dan integritas file.

Untuk database besar atau kebutuhan RPO lebih ketat, evaluasi physical backup dan WAL archiving/PITR. Jangan mengimplementasikan kompleksitas PITR sebelum target RPO/RTO disepakati.

## 3. RPO and RTO

Tetapkan target sebelum production:
- RPO: berapa lama data terbaru yang boleh hilang.
- RTO: berapa lama layanan boleh tidak tersedia saat recovery.

Usulan awal untuk MVP non-critical public directory: RPO maksimum 24 jam dan RTO 4–8 jam sebagai target awal, bukan jaminan. Jika content update harian perlu dipertahankan lebih ketat, tingkatkan frekuensi backup atau aktifkan WAL archiving.

## 4. PostgreSQL backup example

Contoh pola command (gunakan credential file/secret mechanism, jangan menaruh password literal di command history):

```bash
pg_dump --format=custom --no-owner --no-acl \
  --file="/backup/pidiekab-$(date +%F).dump" \
  "$DATABASE_URL"
```

Restore ke database baru untuk pengujian:

```bash
createdb pidiekab_restore_test
pg_restore --no-owner --no-acl \
  --dbname=pidiekab_restore_test \
  /backup/pidiekab-YYYY-MM-DD.dump
```

Jangan restore langsung menimpa production sebelum backup terkini dan recovery plan diverifikasi.

## 5. Media consistency

Database dan media file harus dipulihkan sebagai satu recovery point yang konsisten. Karena media upload dapat terjadi bersamaan dengan backup, gunakan strategi snapshot/versioned backup atau prosedur quiesce/snapshot. Pastikan orphan file dan missing media dapat dideteksi.

Shared media tidak ikut terhapus saat content dihapus dari CMS; backup tetap melindungi media yang masih digunakan.

## 6. Restore drill

Sebelum production dan berkala:
1. pilih backup;
2. restore ke environment terisolasi;
3. jalankan migrations/status validation;
4. verifikasi jumlah record utama;
5. verifikasi media file dan referensi DB;
6. jalankan Go app pada restore environment;
7. jalankan homepage, search, CMS login, detail page, image check;
8. catat durasi dan kendala;
9. hapus test environment setelah disetujui.

Backup tanpa restore test belum membuktikan kemampuan recovery.

## 7. Security

- Backup terenkripsi.
- Akses backup dibatasi.
- Backup destination terpisah dari host utama.
- Jangan mengirim backup ke repository publik.
- Jangan mencatat credential backup.
- Retention dan penghapusan backup lama mengikuti kebijakan.

---
