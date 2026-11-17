# Architecture Baseline

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `backend-modular-monolith-architecture.md`
- `frontend-architecture-vue-vite-ssg.md`
- `go-backend-directory-structure.md`
- `go-backend-library-baseline.md`
- `go-dependency-strategy.md`
- `go-net-http-only-web-server.md`
- `postgresql-database-architecture.md`

---

## Materi sumber: `backend-modular-monolith-architecture.md`

# ARCH-003 — Pragmatic Modular Monolith Backend

**Status: Disepakati.**

## Keputusan
Backend PidieKab menggunakan **Pragmatic Modular Monolith** berbasis Go.

Aplikasi tetap sederhana secara operasional: satu codebase, satu binary/deployable Go application, satu HTTP server `net/http`, dan satu PostgreSQL database. Kode dipisahkan ke dalam module berdasarkan business capability agar tanggung jawab tetap jelas dan perubahan lebih mudah dipelihara.

## Prinsip
- Satu aplikasi Go yang dapat dijalankan sebagai satu service.
- Satu codebase dan satu deployment unit backend.
- Satu PostgreSQL database sebagai sumber data utama.
- Modul dipisahkan berdasarkan domain/business capability.
- Business logic tidak ditempatkan di HTTP handler.
- Pemanggilan antar modul menggunakan function/application service call dalam proses yang sama.
- Tidak ada komunikasi HTTP antar modul.
- Tidak menggunakan microservices pada MVP.
- Tidak membuat interface atau abstraction tanpa kebutuhan nyata.
- Tidak menambahkan Redis, message broker, atau distributed infrastructure tanpa kebutuhan yang terbukti.
- Transaksi database dikendalikan oleh application use case.
- Background jobs menggunakan mekanisme sederhana di aplikasi Go; kebutuhan durable job/outbox dievaluasi berdasarkan risiko dan kebutuhan SSG.

## Modul domain awal
1. Identity & Access.
2. Content.
3. Category.
4. Media.
5. Location.
6. Editorial Curation.
7. Static Pages.
8. Website Settings.
9. Search.
10. Publication & Site Generation.
11. Audit Log.

Daftar modul dapat disesuaikan saat Technical Specification, tetapi batas domain harus tetap jelas. Modul tidak harus menjadi service atau deployment terpisah.

## Lapisan konseptual
- Transport: HTTP routes, handlers, middleware, request/response DTO.
- Application: use cases dan orchestration.
- Domain: entities dan business rules.
- Infrastructure: PostgreSQL, filesystem/object storage bila dipilih, SSG runner, logging, TLS/server.

Penerapan lapisan bersifat pragmatis. Tidak wajib membuat semua lapisan/file untuk setiap operasi sederhana.

## Prinsip transaksi dan integrasi
- Gunakan transaksi PostgreSQL untuk perubahan yang harus atomik.
- Jangan menahan transaksi database ketika melakukan pekerjaan lambat atau I/O eksternal.
- Perubahan publikasi dan permintaan regenerasi SSG harus memiliki mekanisme konsistensi yang jelas.
- Kegagalan generate tidak boleh menggantikan website publik terakhir yang berhasil.
- Pemanggilan antar modul tidak boleh membentuk circular dependency.
- Audit untuk aksi penting dilakukan secara konsisten.

## Yang tidak digunakan pada MVP tanpa alasan khusus
- Microservices.
- Redis.
- Kafka/RabbitMQ atau message broker lain.
- Kubernetes.
- Distributed tracing stack.
- Generic repository/interface layer untuk semua entitas.
- Event-driven architecture kompleks.

## Keputusan teknis lanjutan
Struktur direktori/package Go, router, PostgreSQL driver/query layer, migration tool, autentikasi/sesi, background job, SSG regeneration, TLS, observability, dan deployment detail akan ditetapkan pada Technical Specification.

---

## Materi sumber: `frontend-architecture-vue-vite-ssg.md`

# ARCH-001 — Frontend Vue 3 + TypeScript + Vite dengan SSG

**Status: Disepakati.**

## Keputusan
Frontend PidieKab menggunakan:
- Vue 3.
- TypeScript.
- Vite sebagai build tool.
- Static Site Generation (SSG) untuk website publik.
- Go REST API sebagai backend/data source.
- PostgreSQL sebagai database utama.

Nuxt tidak ditetapkan sebagai framework wajib. Prioritas arsitektur adalah ringan, cepat, sederhana untuk deployment, dan hemat resource server.

## Pemisahan aplikasi
### Website publik
- Dibangun menggunakan Vue 3 + TypeScript + Vite.
- Halaman publik dihasilkan menjadi HTML statis melalui proses SSG/build.
- File hasil build disajikan langsung oleh Go `net/http` sesuai keputusan ARCH-002.
- Data konten diambil dari Go REST API pada proses generate/build.
- Halaman yang dihasilkan harus menyertakan metadata SEO sesuai locale.
- Website mendukung URL bilingual `/id/` dan `/en/` sesuai keputusan URL yang sudah disepakati.

### CMS Admin
- Tetap merupakan aplikasi dinamis berbasis Vue 3 + TypeScript.
- Berkomunikasi dengan Go REST API.
- Tidak memerlukan SSR untuk area CMS.
- Memerlukan autentikasi dan otorisasi Admin.
- CMS tidak boleh mengekspos endpoint preview/draft kepada publik.

## Mekanisme pembaruan konten
Karena halaman publik dihasilkan secara statis:
- Publish, unpublish, archive, perubahan konten, perubahan slug, perubahan kategori, atau perubahan pengaturan SEO yang memengaruhi halaman publik harus memicu proses regenerasi halaman terkait.
- Regenerasi dapat dilakukan melalui mekanisme build/deploy otomatis atau proses incremental generation yang aman; pilihan implementasi ditetapkan pada Technical Specification.
- Setelah proses berhasil, Nginx harus menyajikan hasil terbaru.
- Jika proses generate gagal, versi situs publik terakhir yang berhasil harus tetap tersedia; jangan mengganti situs aktif dengan hasil build yang tidak lengkap.
- Preview Admin untuk Draft dan perubahan yang belum dipublikasikan menggunakan mekanisme terautentikasi dan tidak boleh masuk ke hasil SSG publik.

## SEO dan rendering
- HTML statis harus memuat konten utama, title, meta description, canonical, hreflang yang valid, dan metadata sosial yang diperlukan.
- Sitemap dan robots directives dihasilkan berdasarkan konten yang Published serta aturan locale/SEO yang sudah disepakati.
- Tidak boleh mengandalkan client-side JavaScript saja untuk menampilkan konten utama yang dibutuhkan mesin pencari.
- Halaman kategori dan detail konten harus tersedia sebagai HTML yang sudah dirender saat pertama kali dimuat.

## Deployment dan runtime
- Go `net/http` melayani file statis frontend.
- Go melayani REST API dan file statis frontend melalui satu HTTP server aplikasi; tidak ada reverse proxy/web server tambahan.
- Node.js digunakan pada tahap development/build, bukan sebagai server runtime publik.
- Tidak diperlukan Node.js process permanen untuk menyajikan website publik; Go `net/http` melayani hasil build statis.
- Build pipeline harus menghasilkan artifact statis yang dapat dipasang tanpa downtime atau dengan strategi atomic switch.

## Batasan dan hal yang belum diputuskan
- Strategi full rebuild vs incremental regeneration.
- Pemicu regenerasi: webhook internal, job queue, atau command terjadwal.
- Strategi cache dan invalidasi.
- Penanganan preview draft melalui token/route terproteksi.
- Apakah CMS dan website publik berada dalam satu repository atau repository terpisah.
- Detail library SSG untuk Vue/Vite dan struktur proyek.
Semua hal tersebut ditetapkan pada Technical Specification.

## Alasan keputusan
- Frontend statis dapat dilayani Nginx tanpa runtime Node.js.
- Cocok untuk website regional berbasis konten dengan kebutuhan SEO.
- Backend Go tetap menjadi pusat business logic dan akses data.
- Pemisahan frontend/backend mengurangi coupling dan memudahkan pengembangan.
- Beban server publik lebih ringan dibanding menjalankan SSR Node.js terus-menerus.

---

## Materi sumber: `go-backend-directory-structure.md`

# TECH-001 — Usulan Struktur Direktori Backend Go

**Status: Rekomendasi awal untuk dibahas pada Technical Specification.**

Struktur berikut menerapkan Pragmatic Modular Monolith. Tujuannya menjaga batas domain tanpa menambah lapisan dan abstraksi yang tidak diperlukan.

```text
pidiekab/
├── cmd/
│   └── pidiekab/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── config.go
│   ├── platform/
│   │   ├── httpserver/
│   │   ├── database/
│   │   ├── logger/
│   │   └── storage/
│   ├── identity/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   ├── content/
│   ├── category/
│   ├── media/
│   ├── location/
│   ├── curation/
│   ├── staticpage/
│   ├── setting/
│   ├── search/
│   ├── publication/
│   └── audit/
├── migrations/
├── web/
│   ├── public/       # Vue SSG build output
│   └── admin/        # CMS build output
├── go.mod
├── go.sum
└── README.md
```

## Aturan struktur
- `cmd/pidiekab/main.go` hanya menjadi composition root: membaca konfigurasi, membangun dependency, menjalankan server, dan menangani graceful shutdown.
- `internal/platform` berisi komponen teknis bersama, bukan business logic.
- Setiap domain module memiliki implementasi yang diperlukan saja; tidak harus memiliki file handler/service/repository/model lengkap jika tidak relevan.
- Package domain tidak boleh mengimpor package handler atau transport.
- Hindari import silang yang membentuk circular dependency.
- Module lain tidak boleh mengakses tabel melalui repository internal module lain secara langsung; gunakan fungsi/application service yang diekspos module tersebut.
- `migrations` menyimpan skema database yang versioned.
- `web/public` dan `web/admin` adalah hasil build frontend, bukan source Vue.
- Jika static assets di-embed dengan `embed.FS`, lokasi dan proses packaging ditentukan pada implementasi final.

## Catatan
Ini bukan keputusan final atas nama package atau jumlah file. Struktur final dapat diringkas atau diubah setelah kontrak antar modul, router, database layer, dan SSG workflow dirinci.

---

## Materi sumber: `go-backend-library-baseline.md`

# TECH-002 — Baseline Library Backend Go

**Status: Rekomendasi teknis untuk baseline implementasi.**

Berdasarkan ARCH-004, baseline yang direkomendasikan:

| Kebutuhan | Pilihan | Alasan |
|---|---|---|
| HTTP server & router | Go `net/http` (`ServeMux`) | Sudah cukup untuk routing method/path dan middleware sederhana; tanpa router framework tambahan |
| PostgreSQL driver | `github.com/jackc/pgx/v5` | Driver PostgreSQL native Go, mendukung connection pool melalui `pgxpool` |
| Query layer | SQL eksplisit dengan `pgx` | Kontrol query dan schema; tanpa ORM |
| Database migrations | `pressly/goose` dengan SQL files | Migration versioned, sederhana, mendukung SQL migration |
| Configuration | `os` + `encoding`/parser kecil hanya bila diperlukan | Hindari config framework besar |
| Logging | `log/slog` | Standard library structured logging |
| HTTP middleware | Middleware kecil buatan sendiri | Cukup untuk request ID, logging, recovery, auth, CORS, timeout |
| Password hashing | `golang.org/x/crypto/argon2` atau `bcrypt` | Gunakan algoritma password hashing khusus; jangan hash password dengan SHA-256 biasa |
| Testing | `testing`, `httptest` | Standard library |

## Aturan implementasi
- Gunakan Go version yang masih didukung dan tetapkan versi minimum pada Technical Specification.
- Gunakan `http.ServeMux` terlebih dahulu; tambahkan Chi hanya jika kebutuhan routing terbukti melampaui kemampuan standard library.
- Gunakan `pgxpool` dengan batas koneksi yang dikonfigurasi.
- Query SQL disimpan dekat dengan module pemilik domain atau pada lokasi query module tersebut.
- Hindari generic repository untuk seluruh model.
- Migration dijalankan sebagai langkah deployment terkontrol, bukan otomatis tanpa kendali setiap kali server restart.
- Jangan menambahkan dependency untuk fitur yang dapat ditangani standard library secara sederhana.

## Catatan keamanan
- Password disimpan menggunakan password hashing adaptif seperti Argon2id atau bcrypt dengan parameter yang sesuai.
- Jangan menyimpan secret di source code.
- Gunakan TLS langsung pada Go HTTP server sesuai ARCH-002.
- Terapkan HTTP timeouts, request body limits, graceful shutdown, dan validasi input.
- Versi dependency dikunci dan diperbarui secara berkala.

## Hal yang masih perlu ditetapkan
- Versi minimum Go.
- Pilihan final Argon2id vs bcrypt beserta parameter.
- Format konfigurasi environment.
- Batas pool PostgreSQL.
- Strategi migration saat deploy.

---

## Materi sumber: `go-dependency-strategy.md`

# ARCH-004 — Dependency Strategy Go

**Status: Disepakati.**

## Prinsip dependency
Strategi dependency backend PidieKab:
- Utamakan Go Standard Library.
- Hindari framework backend besar.
- Tambahkan library hanya jika memberikan manfaat nyata.
- Tidak menggunakan Redis pada MVP.
- Tidak menggunakan message broker pada MVP.
- Hindari abstraksi berlebihan.

## Penerapan
- HTTP server dan routing menggunakan kemampuan standard library Go.
- Database access menggunakan driver/library PostgreSQL yang ringan dan terawat.
- Query database ditulis eksplisit; tidak menggunakan ORM pada MVP.
- Database migration menggunakan SQL migration files yang versioned.
- Dependency tambahan harus memiliki alasan teknis yang jelas, dipakai pada batas yang tepat, dan tidak menduplikasi kemampuan standard library.
- Setiap dependency baru ditinjau dari kebutuhan, maintenance, lisensi, ukuran dependency tree, keamanan, dan kompatibilitas Go.

## Batasan MVP
Tidak memasukkan Redis, Kafka, RabbitMQ, atau message broker lain. Background work dan SSG regeneration menggunakan mekanisme internal aplikasi terlebih dahulu. Jika kemudian dibutuhkan durability atau throughput lebih tinggi, keputusan dependency baru dibuat berdasarkan kebutuhan terukur.

## Konsekuensi
- Backend tetap kecil dan mudah dipelihara.
- Mengurangi beban operasional dan jumlah service.
- Tidak mengadopsi framework atau dependency hanya karena popularitas.
- Memilih standard library jika memenuhi kebutuhan dengan kualitas yang memadai.

---

## Materi sumber: `go-net-http-only-web-server.md`

# ARCH-002 — Go net/http sebagai Satu-satunya Web Server

**Status: Disepakati.**

## Keputusan
PidieKab menggunakan HTTP server bawaan Go dari standard library `net/http` sebagai satu-satunya web server aplikasi.

Tidak menggunakan Nginx, Caddy, Apache, LiteSpeed, atau web server/reverse proxy lain dalam arsitektur aplikasi PidieKab.

## Tanggung jawab Go HTTP Server
Satu aplikasi/service Go menangani:
- REST API untuk website publik dan CMS.
- Penyajian file statis hasil Vue SSG.
- Routing URL website publik.
- Routing API, misalnya prefix `/api/`.
- Penanganan fallback/routing untuk kebutuhan halaman statis yang ditetapkan dalam Technical Specification.
- HTTPS/TLS secara langsung melalui kemampuan standard library Go, apabila deployment menggunakan TLS langsung pada aplikasi.

Go `net/http` sudah menyediakan HTTP server dan handler untuk menyajikan file statis, sehingga tidak diperlukan web server terpisah.

## Frontend
- Website publik: Vue 3 + TypeScript + Vite dengan SSG.
- CMS Admin: Vue 3 + TypeScript sebagai aplikasi dinamis.
- Hasil build website publik dan CMS disajikan oleh Go HTTP Server.
- CMS dan website publik mengakses REST API pada aplikasi Go.
- Node.js hanya digunakan untuk development/build frontend, bukan runtime server produksi.

## Database
- PostgreSQL sebagai database utama.
- Go mengakses PostgreSQL melalui driver/library yang dipilih pada Technical Specification.
- Database tidak diekspos langsung ke internet.

## HTTPS dan keamanan
- Tidak ada TLS termination oleh Nginx/Caddy/Apache pada arsitektur aplikasi.
- Jika HTTPS ditangani langsung oleh Go, aplikasi menggunakan fasilitas TLS standard library dan sertifikat yang dikelola melalui mekanisme deployment.
- Cara penerbitan, pembaruan, dan penyimpanan sertifikat TLS masih perlu ditetapkan pada Technical Specification.
- Wajib menerapkan timeout server, batas ukuran request, graceful shutdown, logging, dan konfigurasi keamanan HTTP.
- Detail trusted proxy hanya diperlukan jika kelak ada infrastruktur eksternal di depan aplikasi; bukan bagian dari web server aplikasi yang dipilih.

## Deployment
- Go dijalankan sebagai satu service utama.
- File hasil build Vue SSG dan CMS ditempatkan pada lokasi yang dapat dibaca service Go atau di-embed sebagai static assets.
- PostgreSQL berjalan sebagai service database terpisah.
- Tidak ada service Nginx/Caddy/Apache.
- Strategi file release, atomic update, backup, health check, dan restart tanpa downtime ditetapkan pada Technical Specification.

## Batasan dan keputusan yang belum ditetapkan
- Apakah static assets disimpan pada filesystem atau di-embed menggunakan `embed.FS`.
- Strategi SSG regeneration dan publikasi hasil build.
- Strategi TLS certificate provisioning/renewal.
- Port dan domain routing.
- Apakah CMS dan public site menggunakan satu domain dengan route terpisah atau subdomain.
- Batas timeout, concurrency, dan limit request.
Hal-hal tersebut tidak mengubah keputusan bahwa Go `net/http` adalah satu-satunya web server aplikasi.

## Alasan
- Mengurangi jumlah komponen dan service yang harus dipelihara.
- Menghindari reverse proxy/web server tambahan.
- Menggunakan HTTP server yang sudah tersedia di Go standard library.
- Menjaga deployment tetap sederhana dan hemat resource.
- Backend Go tetap menjadi pusat routing, API, dan penyajian frontend statis.

---

## Materi sumber: `postgresql-database-architecture.md`

# DB-001 — Arsitektur Database PostgreSQL PidieKab

**Status: Rekomendasi arsitektur dasar untuk Technical Specification.**

## 1. Prinsip utama

PidieKab menggunakan **satu database PostgreSQL** untuk seluruh modul Modular Monolith.

- Satu database, satu sumber kebenaran.
- Modul Go memiliki ownership atas tabel/domain masing-masing.
- Relasi antar modul menggunakan foreign key dan application service yang jelas.
- Tidak membuat database terpisah per modul.
- Tidak menggandakan data bilingual atau data master.
- Normalisasi untuk data inti; JSONB hanya untuk metadata/field tambahan yang memang fleksibel.
- Semua waktu disimpan sebagai `timestamptz` dalam UTC; timezone tampilan diatur pada aplikasi.
- Primary key internal menggunakan `BIGINT GENERATED ... AS IDENTITY` untuk efisiensi dan kesederhanaan.
- URL publik memakai slug, bukan ID database.
- Foreign key, unique constraint, check constraint, dan transaksi digunakan untuk menjaga konsistensi.

## 2. Strategi bilingual

Satu identitas konten memiliki dua record terjemahan pada `content_translations`, masing-masing locale `id` dan `en`.

Field bersama disimpan pada tabel induk `contents`, antara lain kategori utama, lokasi, media utama, dan timestamps. Field yang berbeda bahasa disimpan di tabel terjemahan: title, slug, summary, body, SEO metadata, publication status, published_at.

Kategori menggunakan pola yang sama: `categories` untuk identitas dan status global; `category_translations` untuk nama, deskripsi, dan slug ID/EN.

Tidak ada fallback yang menyamarkan konten bahasa lain sebagai terjemahan. Query publik hanya mengambil terjemahan locale yang diminta dan berstatus Published.

## 3. Domain dan ownership tabel

| Domain/module | Tabel utama |
|---|---|
| Identity & Access | `admins`, `admin_sessions` |
| Category | `categories`, `category_translations` |
| Content | `contents`, `content_translations`, `content_sources` |
| Location | `regencies`, `districts`, `villages` |
| Media | `media_assets` |
| Editorial Curation | `featured_contents`, `homepage_story_items` |
| Static Pages | `static_pages`, `static_page_translations` |
| Website Settings | `site_settings`, `site_setting_translations` |
| Publication | `site_generation_jobs`, `url_redirects` |
| Audit | `audit_logs` |

Search MVP menggunakan PostgreSQL full-text search pada data `content_translations`; tidak memerlukan search engine eksternal.

## 4. Relasi penting

- `categories 1:N contents`.
- `categories 1:N category_translations`, satu per locale.
- `contents 1:N content_translations`, satu per locale.
- `contents 1:N content_sources`.
- `regencies 1:N districts 1:N villages`.
- `villages 1:N contents` (nullable pada konten yang lokasinya tidak diketahui/tidak relevan).
- `media_assets 1:N contents` melalui `contents.main_media_id` (nullable).
- `contents 1:N featured_contents` hanya jika ditetapkan sebagai satu record per konten; gunakan unique constraint pada `content_id`.
- `contents 1:N homepage_story_items`, dengan kurasi per locale.
- `admins 1:N admin_sessions`.
- `admins 1:N audit_logs`.
- `admins 1:N content_translations` melalui created_by/updated_by/published_by (nullable untuk akun yang sudah tidak aktif/dihapus sesuai kebijakan).
- `site_generation_jobs` mencatat permintaan generate setelah perubahan publikasi.

## 5. Status dan aturan publikasi

- Status kategori Aktif/Nonaktif bersifat global.
- Status publikasi setiap versi bahasa konten independen: Draft, Published, Archived.
- Konten publik harus memiliki kategori Aktif.
- Konten yang terhubung ke kategori Nonaktif tidak tampil di listing, pencarian, sitemap, atau detail publik.
- Tidak ada status Verified atau approval sebelum publikasi.
- Setiap konten tetap wajib memiliki minimal satu sumber yang dapat ditelusuri sesuai DET-002; pencatatan sumber bukan proses verifikasi, bukan approval, dan tidak menciptakan status Verified.
- Perubahan konten publik harus membuat job regenerasi dalam transaksi yang sama dengan perubahan data.

## 6. Penghapusan dan integritas

- Penghapusan konten melalui CMS bersifat permanen; tidak ada soft delete/recycle bin.
- Penghapusan kategori hanya jika tidak memiliki konten terkait dalam status apa pun.
- Foreign key mencegah penghapusan kategori yang masih direferensikan.
- Penghapusan media yang masih menjadi media utama konten atau halaman harus ditolak.
- `audit_logs` mencatat identitas aksi dan entity ID, bukan salinan penuh konten yang dihapus.
- Gunakan `RESTRICT` untuk relasi yang tidak boleh hilang diam-diam; hindari cascading delete pada konten/media/kategori.

## 7. Pencarian

Gunakan PostgreSQL Full Text Search untuk MVP:
- Search vector per `content_translations`.
- Bobot: title tertinggi, summary berikutnya, body berikutnya.
- Index GIN.
- Filter hanya locale aktif, status Published, dan kategori Aktif.
- Filter kategori mempertahankan relevance ordering.
- Urutan tie-breaker deterministik: relevance, `published_at` terbaru, lalu ID.
- Pagination numbered, default 12 hasil per halaman sesuai rancangan awal.
- Tidak menggunakan Elasticsearch/AI Search pada MVP.

Konfigurasi bahasa, normalisasi aksen, dan dukungan stemming Indonesia/Inggris perlu diuji dalam Technical Specification. Jangan mengasumsikan konfigurasi text search default sudah optimal untuk Bahasa Indonesia.

## 8. SSG dan konsistensi publikasi

Gunakan tabel internal `site_generation_jobs` sebagai antrean persisten sederhana, bukan Redis/message broker.

Dalam satu transaksi:
1. Simpan perubahan konten/terjemahan/kategori/SEO.
2. Tambahkan job regenerasi yang relevan.
3. Commit.

Worker Go mengambil job secara aman, menjalankan proses build/generate, dan menandai hasil. Build harus menghasilkan release baru di direktori sementara; hanya setelah sukses, aplikasi mengaktifkan release tersebut secara atomik. Jika gagal, release publik terakhir tetap disajikan.

Strategi full rebuild vs incremental generation masih harus diputuskan melalui pengukuran. MVP boleh memulai dari full build bila ukuran konten masih kecil dan waktu build memenuhi target.

## 9. Indexing dan performa

Index awal yang diperlukan:
- Unique `(locale, slug)` pada `content_translations`.
- Unique `(locale, slug)` pada `category_translations`.
- Index `contents(category_id)`, `contents(village_id)`.
- Index `content_translations(status, locale, published_at)`.
- GIN pada search vector.
- Index `site_generation_jobs(status, available_at, created_at)`.
- Unique `url_redirects(old_path)`.
- Index foreign key pada tabel relasi/curation.

Tambahkan index berdasarkan query nyata dan `EXPLAIN (ANALYZE, BUFFERS)`, bukan menambahkan index pada setiap kolom.

## 10. Batasan desain

- Tidak menggunakan PostGIS pada MVP; koordinat numerik cukup untuk detail/map link.
- Tidak membuat tabel terpisah untuk setiap jenis konten.
- Tidak menggunakan EAV.
- JSONB hanya untuk metadata atau atribut kategori opsional yang jarang difilter/diurutkan; field inti tetap kolom relasional.
- Tidak membuat multi-tenant database.
- Tidak membuat schema PostgreSQL terpisah untuk setiap modul.

---
