# PidieKab — Open Decisions & Recommended Resolutions

**Versi:** v78  
**Tujuan:** daftar keputusan yang belum final, rekomendasi praktis, dan kapan keputusan harus ditutup.

| ID | Keputusan terbuka | Rekomendasi awal | Harus ditutup sebelum |
|---|---|---|---|
| OPEN-001 | Tool SSG Vue | Pilih tool yang menghasilkan HTML statis untuk dynamic localized routes dan dapat disajikan Go; lakukan POC, jangan memilih berdasarkan popularitas saja | Akhir Phase 0 |
| OPEN-002 | Goose migration format | Gunakan SQL migration dengan directive Goose; validasi `Up`, `Down`, clean install, dan rollback pada PostgreSQL target | Sebelum Phase 1 |
| OPEN-003 | Password hashing | Argon2id melalui `golang.org/x/crypto/argon2`; bcrypt sebagai alternatif bila parameter/operasional lebih sesuai | Sebelum implementasi Identity |
| OPEN-004 | Password policy | Minimum 12 karakter; izinkan passphrase panjang; batas maksimum cukup besar; jangan memaksa aturan komposisi yang tidak efektif | Sebelum implementasi Identity |
| OPEN-005 | Admin bootstrap | One-time CLI command, hanya bisa dijalankan operator server, tidak ada default credential | Sebelum Admin login |
| OPEN-006 | Password reset | Untuk MVP gunakan operator-controlled CLI reset; email reset ditunda sampai kanal email dan abuse protection siap | Sebelum Admin login |
| OPEN-007 | Session lifetime | Pertahankan proposal idle 30 menit dan absolute 12 jam; dapat dikonfigurasi; invalidasi saat logout/password reset/deactivation | Sebelum Admin login |
| OPEN-008 | Cookie & CSRF | Tentukan nama, Path, SameSite, Secure, HttpOnly; cookie scope harus mendukung API dan Preview; gunakan CSRF token + Origin validation | Sebelum Admin login |
| OPEN-009 | 2FA | Tunda dari MVP, tetapi desain session/auth tidak menghalangi penambahan TOTP/WebAuthn kemudian; aktifkan lebih awal bila risiko operasional menuntut | Sebelum go-live |
| OPEN-010 | TLS issuance/renewal | Evaluasi ACME DNS-01 dengan token DNS provider least-privilege; uji renewal otomatis dan reload sertifikat Go | Sebelum public launch |
| OPEN-011 | Database topology | Untuk MVP mulai satu host jika kapasitas dan backup memadai; pisahkan DB hanya jika kebutuhan resource/availability membenarkan dan jaringan private tersedia | Sebelum deployment production |
| OPEN-012 | Backup retention/RPO/RTO | Baseline usulan: backup harian, 7 daily/4 weekly/3 monthly; target awal RPO 24 jam dan RTO 4–8 jam; lakukan restore drill dan sesuaikan biaya/risiko | Sebelum go-live |
| OPEN-013 | Location master source | Gunakan sumber administrasi wilayah yang otoritatif; simpan asal, tanggal akses, versi/import; tetapkan proses pembaruan manual | Sebelum konten lokasi massal |
| OPEN-014 | Coordinate accuracy process | Admin hanya memasukkan koordinat dari sumber peta/kunjungan yang dapat dicatat; label accurate/approximate; jangan mengisi otomatis dari tebakan | Sebelum konten peta dipublikasikan |
| OPEN-015 | Indonesian/English full-text search | POC konfigurasi PostgreSQL; uji stemming/tokenization dan ranking pada corpus contoh ID/EN; bila konfigurasi built-in kurang, gunakan normalisasi sederhana teruji | Sebelum Search dianggap selesai |
| OPEN-016 | Upload size/body limits | Usulan awal gambar maksimal 10 MB; request body limit terpisah untuk API; konfigurasi dapat diubah setelah uji nyata | Sebelum Media module |
| OPEN-017 | Rate limit/timeouts/pool sizes | Jangan menebak angka final; tetapkan default konservatif, konfigurabel, lalu ukur load test dan pool saturation | Sebelum go-live |
| OPEN-018 | SSG full vs incremental | Mulai dengan rebuild terkontrol bila ukuran konten masih kecil; pertahankan release atomic; optimasi incremental hanya setelah waktu build terukur | Sebelum Publication module |
| OPEN-019 | Redirect retention | Simpan permanent redirects untuk slug/category path lama; cegah chain/loop dan reuse slug lama selama redirect aktif | Sebelum SEO launch |
| OPEN-020 | API contract coverage | Perluas OpenAPI seiring endpoint; CI harus memvalidasi file dan contract tests untuk endpoint yang selesai | Setiap endpoint merge |
| OPEN-021 | Contact/social URL policy | Izinkan hanya skema yang diperlukan (`https`, `mailto`, `tel`); validasi dan tampilkan tautan eksternal secara aman | Sebelum Settings module |
| OPEN-022 | Metrics/alert thresholds | Gunakan baseline server aktual; mulai dengan alert service down, DB unavailable, queue age, disk, backup overdue, TLS expiry | Sebelum go-live |

## Prioritas penutupan

### P0 — Blocker sebelum fitur inti
- OPEN-001 SSG tool POC
- OPEN-002 Goose migration proof
- OPEN-003 sampai OPEN-008 keamanan Admin
- OPEN-010 TLS
- OPEN-011 topology deployment

### P1 — Sebelum modul terkait selesai
- OPEN-013/014 master lokasi dan koordinat
- OPEN-015 search language behavior
- OPEN-016 upload limits
- OPEN-018 SSG build strategy
- OPEN-019 redirects

### P2 — Sebelum public launch atau berdasarkan pengukuran
- OPEN-009 2FA
- OPEN-012 backup RPO/RTO/retention
- OPEN-017 rate limit/timeouts/pool sizes
- OPEN-020 OpenAPI coverage
- OPEN-021 external URL policy
- OPEN-022 metrics/alert thresholds

## Prinsip penyelesaian

- Pilih solusi paling sederhana yang memenuhi acceptance criteria.
- Jangan menambah infrastruktur hanya untuk mengantisipasi skala yang belum ada.
- Semua nilai operasional harus dapat dikonfigurasi bila wajar.
- Uji restore backup, certificate renewal, dan atomic SSG release; dokumentasi saja tidak cukup.
- Jika keputusan belum diperlukan untuk fase berjalan, biarkan OPEN dengan deadline fase yang jelas.
