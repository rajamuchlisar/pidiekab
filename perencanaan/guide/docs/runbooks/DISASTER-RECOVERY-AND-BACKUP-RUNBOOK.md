# Disaster Recovery & Backup Runbook — PidieKab

**Target RPO:** $\le 24$ Jam (Daily Automated Logical Dump + Incremental Media Sync)  
**Target RTO:** 1–4 Jam  
**Metode:** Logical Dump (`pg_dump`), Tarball Media Storage, SHA-256 Checksum, dan Automated Restore Drill.

---

## 1. Jadwal Backup Otomatis (Cron Job)

Backup otomatis dijadwalkan setiap hari pukul 02:00 WIB via cron user `root` atau `pidiekab`:

```cron
# /etc/cron.d/pidiekab-backup
0 2 * * * root /opt/pidiekab/current/deploy/scripts/backup.sh >> /var/log/pidiekab-backup.log 2>&1
```

Setiap file backup menghasilkan:
- `pidiekab_db_YYYYMMDD_HHMMSS.dump` (PostgreSQL custom format dump)
- `pidiekab_db_YYYYMMDD_HHMMSS.dump.sha256` (Integritas checksum)
- `pidiekab_media_YYYYMMDD_HHMMSS.tar.gz` (Arsip media storage)
- `pidiekab_media_YYYYMMDD_HHMMSS.tar.gz.sha256` (Integritas checksum)

---

## 2. Off-Host Replication (Disaster Boundary)

File backup pada `/var/backups/pidiekab` harus segera direplikasi ke penyimpanan off-host (S3-compatible, rsync server terisolasi, atau off-site storage):

```bash
# Contoh rsync ke backup server terisolasi
rsync -avz --delete -e 'ssh -i /etc/pidiekab/backup_id_ed25519' \
  /var/backups/pidiekab/ \
  backupuser@remote-backup.internal:/secure-backups/pidiekab/
```

---

## 3. Prosedur Pemulihan Bencana (Disaster Recovery Drill)

Jika server utama mengalami kegagalan total atau data terkorupsi:

### Langkah 1: Siapkan Server Baru / Database Baru
Pastikan PostgreSQL 17 terinstal dan database baru telah dibuat:
```bash
createdb -U postgres pidiekab_recovered
```

### Langkah 2: Verifikasi Integritas File Backup
```bash
cd /var/backups/pidiekab
sha256sum -c pidiekab_db_20261117_020000.dump.sha256
sha256sum -c pidiekab_media_20261117_020000.tar.gz.sha256
```
> [!IMPORTANT]
> Jangan pernah merestore file backup yang checksum-nya tidak cocok.

### Langkah 3: Eksekusi Restore Menggunakan Script
```bash
./deploy/scripts/restore.sh \
  /var/backups/pidiekab/pidiekab_db_20261117_020000.dump \
  postgres://postgres:password@127.0.0.1:5432/pidiekab_recovered \
  /var/backups/pidiekab/pidiekab_media_20261117_020000.tar.gz
```

### Langkah 4: Pulihkan Media Storage
Pindahkan hasil ekstraksi media ke direktori storage aktif:
```bash
sudo rsync -av /tmp/pidiekab_restore_media/media/ /var/lib/pidiekab/media/
sudo chown -R pidiekab:pidiekab /var/lib/pidiekab/media
```

### Langkah 5: Start Service & Validasi
Arahkan `PIDIEKAB_DATABASE_URL` ke database baru di `/etc/pidiekab/pidiekab.env`, kemudian restart service:
```bash
sudo systemctl restart pidiekab.service
curl -s http://127.0.0.1:8080/health/ready
```
Periksa halaman publik dan CMS login untuk memverifikasi kelengkapan data.

