# PidieKab — Final Documentation Review & Architecture Lock

**Versi:** v78  
**Status:** Review baseline — siap masuk Phase 0 setelah keputusan OPEN yang menjadi blocker disetujui  
**Dokumen acuan:** seluruh dokumentasi v77

## 1. Tujuan

Dokumen ini menjadi pemeriksaan akhir sebelum implementasi. Review berfokus pada:
- konsistensi product scope, content governance, UX, API, database, security, deployment, QA, dan roadmap;
- pemisahan keputusan yang sudah dikunci dari rekomendasi dan hal yang masih memerlukan persetujuan;
- identifikasi risiko yang dapat menyebabkan rework jika dibiarkan sampai implementasi.

Review ini tidak menggantikan spesifikasi rinci v77. Jika terjadi perbedaan, keputusan eksplisit dalam dokumen governance v78 menjadi catatan status terbaru; detail teknis v77 tetap menjadi baseline sampai perubahan disetujui.

## 2. Kesimpulan eksekutif

**Rekomendasi: CONDITIONAL GO untuk Phase 0.**

Dokumentasi telah cukup matang untuk memulai repository foundation, toolchain, konfigurasi lokal, CI checks, dan proof-of-concept teknis. Belum disarankan memulai fitur CMS atau public content sampai blocker berikut ditutup:
1. pemilihan dan pembuktian tool SSG;
2. format migrasi Goose yang benar-benar dapat dieksekusi dari baseline SQL;
3. keputusan keamanan minimum untuk Admin bootstrap, session cookie, password reset, dan password hashing;
4. mekanisme TLS issuance/renewal yang diuji;
5. keputusan awal deployment topology dan backup/restore.

Tidak semua keputusan OPEN harus menghambat Phase 0. Nilai kapasitas, alert threshold, dan tuning performa dapat ditentukan setelah pengukuran.

## 3. Hasil pemeriksaan konsistensi

| Area | Hasil review | Tindakan |
|---|---|---|
| Product positioning | Konsisten sebagai platform promosi independen, bukan situs resmi pemerintah | Pertahankan disclaimer dan atribusi sumber |
| MVP content | Tiga kategori awal konsisten: Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah | Tidak menambahkan kategori Phase 2 ke MVP |
| Content governance | Admin internal, Draft → Preview → Publish, tanpa approval/verification workflow | Pertahankan; source requirement adalah referensi editorial, bukan verifikasi |
| Bahasa | ID dan EN, status publikasi independen | Uji locale isolation pada CMS, API, SSG, sitemap |
| URL & SEO | Prefix `/id/` dan `/en/`, localized slug, 301 saat slug berubah | Pastikan redirect tidak berantai/loop |
| Category model | Satu kategori utama per konten; kategori dinamis | Validasi aturan kategori nonaktif dan penghapusan |
| Public UX | Informasi/discovery; tanpa akun pengunjung, review, komentar, atau transaksi | Pertahankan batasan MVP |
| CMS | Satu Admin role, konten, halaman informasi, media, pengaturan | Jangan membangun RBAC kompleks di MVP |
| Search | PostgreSQL full-text search, hanya konten Published pada locale aktif | POC konfigurasi bahasa Indonesia dan Inggris |
| Location | Master kabupaten/kecamatan/gampong; koordinat tidak boleh direka | Sumber master data dan prosedur pembaruan perlu diputuskan |
| Backend | Go, `net/http`, Modular Monolith, satu binary | Konsisten; hindari framework besar |
| Database | Satu PostgreSQL, SQL eksplisit, tanpa ORM | Pertahankan |
| Frontend | Vue 3 + TypeScript; public SSG dan CMS SPA | SSG tool masih OPEN |
| Publication | Job PostgreSQL, worker `SKIP LOCKED`, temporary release, atomic activation | Perlu uji crash/retry/rollback |
| Infrastructure | Go menjadi satu-satunya application web server; tanpa reverse proxy | Pastikan deployment tidak menambah Nginx/Caddy |
| Security | Session cookie, CSRF, sanitasi rich text, upload aman | Parameter final dan bootstrap masih OPEN |
| QA | Test unit/integration/E2E/security/operational | Tambahkan release gate untuk seluruh keputusan blocker |
| Observability | `slog`, health checks, metrics dan alert | Ambang alert ditetapkan dari baseline aktual |
| Roadmap | Phase 0 sebelum implementasi fitur | Phase 0 boleh dimulai dengan batasan di bawah |

## 4. Catatan teknis yang wajib dibereskan

### 4.1 SSG dan server runtime
SSG tool belum dikunci. Tool harus menghasilkan HTML statis untuk localized dynamic routes, metadata SEO, canonical/hreflang, sitemap, dan redirects. Hasil build harus dapat disajikan langsung oleh Go tanpa Node runtime atau server aplikasi kedua. Pilih tool hanya setelah proof-of-concept.

### 4.2 Goose dan initial schema
Baseline `001_initial_schema.sql` perlu diperiksa terhadap format migration Goose, termasuk directive `-- +goose Up` dan `-- +goose Down` bila dipakai. Jangan menjalankan migration sebelum format, urutan, rollback strategy, dan hasil clean install diuji pada PostgreSQL versi target.

### 4.3 Password hashing dependency
Go standard library tidak menyediakan Argon2id. Jika Argon2id dipilih, `golang.org/x/crypto/argon2` menjadi dependency eksplisit yang dibenarkan. Alternatif bcrypt juga membutuhkan library eksternal. Jangan mengganti dengan SHA-256 atau hash cepat.

### 4.4 Cookie scope dan Preview
Cookie path harus memungkinkan Admin mengakses API dan preview yang memang membutuhkan autentikasi, tetapi tidak memperluas paparan tanpa alasan. Tetapkan path, SameSite, Secure, HttpOnly, expiry, CSRF strategy, dan Origin validation bersama-sama. Preview wajib authenticated dan `noindex`; draft tidak boleh bocor melalui static output atau API publik.

### 4.5 Atomic SSG release
Publikasi konten dan pembuatan job harus berada dalam satu transaksi database. Worker membangun ke direktori release sementara; release aktif hanya diganti setelah build sukses. Kegagalan build harus mempertahankan release sebelumnya. Definisikan juga pembersihan release lama, retry, dan penanganan job yang tertinggal setelah proses restart.

### 4.6 TLS tanpa reverse proxy
Go menangani HTTPS langsung. DNS-01 ACME merupakan kandidat, bukan keputusan final. Uji issuance, renewal, certificate reload, expiry alert, dan recovery. Port 80 hanya dibuka jika metode challenge/redirect yang dipilih memerlukannya.

### 4.7 Sumber data lokasi
Master data lokasi tidak boleh diisi dari tebakan. Tentukan sumber resmi/otoritatif, tanggal pengambilan, prosedur pembaruan, serta cara mengoreksi perubahan nama atau pemekaran wilayah.

### 4.8 API specification
OpenAPI saat ini adalah baseline awal, bukan kontrak lengkap seluruh endpoint. Endpoint yang benar-benar diimplementasikan harus ditambahkan dan divalidasi dalam CI agar dokumentasi tidak menyimpang dari perilaku API.

## 5. Batas pekerjaan Phase 0

Phase 0 boleh mengerjakan:
- repository structure dan Go module;
- frontend workspace dan build tooling;
- lint/format/test baseline;
- konfigurasi environment dan secret handling lokal;
- logger, request ID, graceful shutdown skeleton;
- PostgreSQL local development dan migration proof;
- SSG proof-of-concept localized route;
- CI pipeline dasar dan dokumentasi cara menjalankan proyek.

Phase 0 **belum** boleh:
- membangun fitur CMS penuh;
- membuat alur publish production sebelum atomic release teruji;
- mengunci tool SSG tanpa hasil POC;
- memasukkan Redis, message broker, microservices, atau reverse proxy;
- mengisi data lokasi/koordinat yang tidak memiliki sumber;
- menambahkan fitur Phase 2.

## 6. Exit criteria Final Review

Final review dianggap selesai apabila:
- seluruh keputusan berstatus LOCKED atau OPEN dengan owner/aksi yang jelas;
- blocker sebelum CMS/production ditutup;
- migration clean install berhasil;
- SSG POC memenuhi kriteria;
- security baseline Admin disetujui;
- TLS dan backup/restore memiliki prosedur teruji;
- MVP scope disepakati dan tidak melebar.

## 7. Rekomendasi keputusan

Gunakan `architecture-decision-register.md` dan `open-decisions-and-recommendations.md` sebagai sumber status. Jangan menafsirkan kata “recommended” sebagai keputusan yang sudah disetujui.
