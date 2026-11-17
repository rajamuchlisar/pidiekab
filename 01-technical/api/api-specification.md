# Api Specification

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `api-conventions.md`
- `public-api.md`
- `admin-api.md`
- `publication-api-and-ssg.md`

---

## Materi sumber: `api-conventions.md`

# API Conventions — PidieKab

**Status:** Baseline technical contract  
**API version:** `v1`  
**Base path:** `/api/v1`  
**Format:** JSON UTF-8  
**Backend:** Go `net/http`  
**Last updated:** 2026-11-16

## 1. Tujuan dan batas

Dokumen ini menetapkan konvensi REST API untuk website publik, CMS Admin, dan proses internal publikasi/SSG. API bukan pengganti halaman publik hasil SSG: halaman informasi utama dirender saat proses generation dan disajikan sebagai file statis oleh Go. API publik dinamis MVP terutama digunakan untuk pencarian.

Semua endpoint berada pada `/api/v1`. Perubahan yang tidak kompatibel memerlukan versi API baru; penambahan field opsional atau endpoint baru tidak dianggap breaking change.

## 2. Format request dan response

- Request body yang menerima JSON menggunakan `Content-Type: application/json`.
- Response JSON menggunakan `Content-Type: application/json; charset=utf-8`.
- Waktu menggunakan ISO 8601/RFC 3339 dengan timezone eksplisit; server menyimpan timestamp UTC (`timestamptz`).
- ID database dikirim sebagai string desimal dalam JSON untuk menghindari kehilangan presisi pada JavaScript.
- Nilai kosong dibedakan secara konsisten: field opsional dapat dihilangkan atau `null` sesuai kontrak endpoint; array koleksi kosong dikirim sebagai `[]`, bukan `null`.
- Jangan mengirim field internal seperti password hash, session hash, filesystem path, atau data audit sensitif ke public API.

Contoh response sukses satu resource:

```json
{
  "data": {
    "id": "123",
    "title": "Pantai Mantak Tari"
  }
}
```

Contoh response koleksi:

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 12,
    "total_items": 0,
    "total_pages": 0
  }
}
```

Response action yang tidak mengembalikan resource dapat menggunakan HTTP `204 No Content`.

## 3. Error format

Gunakan format `application/problem+json` yang konsisten dengan RFC 9457. Field minimum:

```json
{
  "type": "about:blank",
  "title": "Validation failed",
  "status": 422,
  "code": "VALIDATION_ERROR",
  "detail": "One or more fields are invalid.",
  "request_id": "req_..."
}
```

Untuk validasi field, tambahkan `errors`:

```json
{
  "type": "about:blank",
  "title": "Validation failed",
  "status": 422,
  "code": "VALIDATION_ERROR",
  "errors": {
    "title": ["Title is required."],
    "slug": ["Slug is already in use."]
  }
}
```

Status yang digunakan:
- `400` malformed JSON atau parameter tidak valid.
- `401` belum terautentikasi atau sesi tidak valid.
- `403` terautentikasi tetapi tidak diizinkan.
- `404` resource/route tidak ditemukan atau tidak boleh diungkapkan.
- `409` konflik state, slug, atau resource yang masih direferensikan.
- `413` ukuran upload melebihi batas.
- `415` media type tidak didukung.
- `422` validasi domain gagal.
- `429` rate limit.
- `500` kesalahan internal.

Jangan mengirim stack trace, SQL, path server, atau detail rahasia kepada client. Catat detail teknis pada server log dengan `request_id`.

## 4. Locale

Locale MVP hanya `id` dan `en`.

- Public localized endpoints menggunakan path `/public/{locale}/...`.
- `locale` harus persis `id` atau `en`; nilai lain menghasilkan `400 INVALID_LOCALE`.
- Tidak ada silent fallback ke bahasa lain.
- `Accept-Language` tidak mengubah locale konten.
- CMS mengirim locale secara eksplisit pada resource translation.
- URL website tetap memakai prefix `/id/` dan `/en/`; root website mengarahkan ke `/id/`.

## 5. Pagination, filter, dan sorting

Pagination menggunakan query:
- `page`: integer mulai 1, default `1`.
- `per_page`: default `12`, maksimum `50`.
- Nilai di luar batas ditolak dengan `400 INVALID_PAGINATION`, bukan diam-diam diubah.

Response koleksi memakai `meta.page`, `meta.per_page`, `meta.total_items`, `meta.total_pages`.

Filter dan query dipertahankan saat berpindah halaman. Endpoint menentukan sorting yang diperbolehkan; client tidak boleh mengirim SQL-like sort expression. Search memakai ranking relevansi yang ditentukan backend dan tie-breaker deterministik.

## 6. Authentication dan keamanan

CMS Admin memakai sesi server-side. Baseline desain:
- Login menghasilkan opaque random session token.
- Database hanya menyimpan hash token pada `admin_sessions`.
- Token disimpan pada cookie `HttpOnly`, `Secure` pada HTTPS, dan `SameSite` yang sesuai.
- Cookie tidak dapat diakses JavaScript.
- Endpoint yang mengubah state wajib memiliki proteksi CSRF bila autentikasi berbasis cookie.
- Logout menginvalidasi sesi di server.
- Sesi admin yang inactive/expired/revoked ditolak.
- Login, logout, perubahan password, reset password, perubahan status Admin, publish, archive, delete, dan perubahan pengaturan penting dicatat pada audit log.
- Password tidak pernah dikembalikan melalui API atau dicatat pada log.
- Rate limit diterapkan minimal pada login dan endpoint sensitif.

Nilai final seperti nama cookie, durasi sesi, idle timeout, CSRF token strategy, kebijakan password, 2FA, dan batas rate ditetapkan pada security implementation specification sebelum production.

## 7. Upload media

Upload menggunakan `multipart/form-data`. Server wajib:
- memvalidasi ukuran, MIME type berdasarkan isi file, ekstensi, dan dimensi gambar;
- membuat nama file internal yang aman dan tidak memakai nama upload sebagai path;
- menyimpan metadata dan informasi hak penggunaan;
- menolak SVG atau format aktif lain kecuali ada sanitasi khusus yang disetujui;
- membatasi ukuran request dan melakukan pemeriksaan file;
- tidak mempercayai MIME type dari browser.

Media yang sedang dipakai konten/halaman tidak dapat dihapus. File disimpan pada storage yang dikonfigurasi; detail local filesystem/object storage ditetapkan pada deployment spec.

## 8. Idempotency dan concurrency

- `GET`, `HEAD`, dan operasi baca tidak mengubah state.
- `PUT` untuk mengganti representasi translation/resource dirancang idempotent.
- Action `publish`, `archive`, dan `delete` memeriksa state saat ini dan mengembalikan konflik bila transisi tidak valid.
- Untuk mencegah lost update, admin write endpoint dapat memakai `updated_at`/version precondition. Mekanisme final (ETag/If-Match atau `version`) ditentukan saat implementasi.
- Publish harus atomic: validasi, update state, audit log, dan insert `site_generation_jobs` dalam satu transaksi PostgreSQL.

## 9. Caching dan request tracing

- Public search boleh memakai cache HTTP singkat setelah perilaku invalidasi dipastikan.
- Admin response menggunakan `Cache-Control: no-store`.
- Preview dan draft tidak boleh masuk cache publik atau sitemap.
- Setiap request diberi request ID; propagasikan ke log dan error response.
- Jangan cache response yang mengandung draft atau data admin.

## 10. Health endpoints

Internal operational endpoints:
- `GET /healthz`: proses HTTP hidup.
- `GET /readyz`: aplikasi siap melayani; memeriksa dependency kritis seperti koneksi PostgreSQL.
- Response minimal dan tidak membocorkan versi dependency/secret.
- Endpoint readiness tidak boleh menampilkan connection string atau diagnostic detail.

---

## Materi sumber: `public-api.md`

# Public API — PidieKab

**Status:** Baseline contract  
**Base path:** `/api/v1/public`  
**Audience:** Website publik dan proses pembentukan halaman publik

## 1. Prinsip delivery

Website publik utama menggunakan SSG. Halaman beranda, kategori, detail konten, dan halaman informasi dibangun menjadi HTML statis, lalu dilayani oleh Go `net/http`. Browser tidak perlu memanggil API untuk mengambil isi halaman tersebut pada setiap page load.

Public API disediakan untuk kebutuhan dinamis yang memang memerlukan request runtime, terutama Global Search. Data API publik hanya boleh berisi:
- konten berstatus `published`;
- locale yang diminta dan memang sudah dipublikasikan;
- kategori berstatus `active`;
- field yang aman untuk konsumsi publik.

Tidak ada endpoint public untuk draft, preview, arsip, data admin, audit log, atau sumber internal yang ditandai privat.

## 2. Endpoint MVP

| Method | Endpoint | Fungsi |
|---|---|---|
| GET | `/api/v1/public/{locale}/search` | Pencarian global |
| GET | `/api/v1/public/{locale}/categories` | Kategori aktif untuk kebutuhan client dinamis bila diperlukan |
| GET | `/api/v1/public/{locale}/categories/{slug}` | Detail kategori publik bila dibutuhkan integrasi |
| GET | `/api/v1/public/{locale}/contents/{id}` | Tidak menjadi endpoint website publik utama; hanya untuk integrasi publik yang memang memerlukan ID dan data terbit |

SSG tetap menjadi sumber delivery utama untuk halaman kategori/detail. Endpoint kategori dan detail dapat ditunda bila tidak ada consumer nyata; jangan membangun endpoint hanya karena tabel database tersedia.

## 3. Global Search

### Request

`GET /api/v1/public/{locale}/search?q={query}&category={category_slug}&page=1&per_page=12`

Parameter:
- `locale`: `id` atau `en`.
- `q`: wajib; trimmed, panjang minimum 2 karakter dan maksimum 120 karakter.
- `category`: opsional; slug kategori pada locale yang diminta. Nilai harus mengacu kategori aktif.
- `page`: default 1.
- `per_page`: default 12, maksimum 50.

### Response

```json
{
  "data": [
    {
      "id": "123",
      "title": "Pantai Mantak Tari",
      "category": {
        "slug": "wisata-destinasi",
        "name": "Wisata & Destinasi"
      },
      "location": {
        "district": "Kota Sigli",
        "village": "..."
      },
      "url": "/id/wisata-destinasi/pantai-mantak-tari"
    }
  ],
  "meta": {
    "locale": "id",
    "query": "pantai",
    "category": null,
    "page": 1,
    "per_page": 12,
    "total_items": 1,
    "total_pages": 1
  }
}
```

Field `location` boleh `null` bila tidak tersedia. Hasil minimum sesuai keputusan UX: judul, kategori, dan lokasi bila tersedia. Tidak menampilkan gambar atau ringkasan pada kartu hasil pencarian.

### Search behavior

- Search title, summary, dan body.
- Bobot tertinggi title, kemudian summary, kemudian body.
- Hasil hanya Published untuk locale aktif dan kategori Active.
- Filter kategori tidak mengubah urutan relevansi.
- Tie-breaker deterministik: publication date lalu content ID.
- PostgreSQL full-text/indexed matching; belum memakai Elasticsearch atau semantic AI search.
- Query tidak boleh mengembalikan konten dari locale lain.
- Pencarian harus diuji untuk konfigurasi bahasa Indonesia dan Inggris. Bila stemming/tokenization tidak memadai, lakukan normalisasi/index strategy yang terukur tanpa menambah search service.

## 4. Category listing

Jika endpoint ini dipakai oleh komponen client dinamis:

`GET /api/v1/public/{locale}/categories`

Hanya kategori aktif yang ditampilkan, dengan field:
- `slug`
- `name`
- `description` (boleh null)
- `content_count` (opsional, hanya bila dihitung dengan query efisien)

Kategori yang inactive tidak boleh muncul pada navigasi, filter, atau listing publik.

## 5. Static pages dan content detail

Halaman About, Contact, Privacy Policy, Disclaimer, kategori, dan detail konten disajikan melalui SSG. Tidak diperlukan endpoint browser-facing untuk masing-masing halaman pada MVP.

Jika integrasi eksternal nantinya membutuhkan public JSON, endpoint harus dirancang terpisah dan hanya mengekspos field published. Jangan mengembalikan body JSONB internal editor tanpa proses sanitasi dan serialisasi yang aman.

## 6. Sitemap dan robots

Sitemap XML dibentuk sebagai bagian dari SSG:
- hanya halaman Published dan indexable;
- hanya locale yang sudah Published;
- `hreflang` hanya bila versi ID dan EN sama-sama Published;
- canonical self-referential per locale;
- draft/archived/preview tidak masuk sitemap;
- robots directives mengikuti status publikasi/indexability.

Sitemap tidak perlu public JSON API. Go melayani file sitemap XML hasil generation.

## 7. Rate limit dan abuse protection

Search endpoint menerapkan batas request yang wajar per IP/session, panjang query, dan pagination. Rate limit harus memberi response `429` dengan `Retry-After` bila tersedia. Hindari menyimpan query pencarian yang dapat mengidentifikasi pengguna; analytics pencarian bukan bagian dari MVP.

---

## Materi sumber: `admin-api.md`

# Admin API — PidieKab CMS

**Status:** Baseline contract  
**Base path:** `/api/v1/admin`  
**Authentication:** Server-side Admin session

## 1. Prinsip

- Semua endpoint di dokumen ini memerlukan sesi Admin aktif, kecuali login.
- MVP hanya memiliki satu role Admin; tidak ada role hierarchy.
- Tidak ada public signup atau akun operator destinasi.
- Semua input divalidasi di server. Frontend validation hanya untuk UX.
- Data ID/EN dikelola terpisah; status publikasi tiap locale independen.
- API tidak boleh mengubah locale lain secara implisit.
- Setiap mutasi penting dicatat ke `audit_logs`.

## 2. Authentication

| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/auth/login` | Login Admin |
| POST | `/auth/logout` | Logout dan revoke sesi |
| GET | `/auth/me` | Admin dan status sesi saat ini |
| POST | `/auth/password/change` | Ganti password sendiri |

Login menerima email dan password. Response tidak mengembalikan token sesi di JSON; token disimpan sebagai cookie aman sesuai API Conventions. Pesan login gagal tidak boleh membocorkan apakah email terdaftar.

## 3. Dashboard

`GET /dashboard`

Mengembalikan ringkasan operasional:
- jumlah konten per kategori;
- jumlah konten Draft/Published/Archived;
- status ID dan EN secara terpisah;
- konten terakhir diubah;
- aktivitas Admin terbaru yang aman ditampilkan.

Tidak mengembalikan page views, visitor analytics, popularity, atau metrik pemasaran karena bukan scope MVP.

## 4. Content

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/contents` | List dengan search/filter/pagination |
| POST | `/contents` | Buat identity konten dan data bersama |
| GET | `/contents/{content_id}` | Detail CMS termasuk status translation |
| PUT | `/contents/{content_id}` | Ubah field bersama |
| DELETE | `/contents/{content_id}` | Permanent delete setelah konfirmasi |
| GET | `/contents/{content_id}/translations/{locale}` | Ambil versi locale |
| PUT | `/contents/{content_id}/translations/{locale}` | Buat/perbarui versi locale |
| POST | `/contents/{content_id}/translations/{locale}/preview` | Preview versi tersimpan/unsaved |
| POST | `/contents/{content_id}/translations/{locale}/publish` | Publish locale |
| POST | `/contents/{content_id}/translations/{locale}/archive` | Archive locale |

List filter yang didukung: `q`, `category_id`, `status`, `locale`, `page`, `per_page`. Status ID dan EN harus tetap terlihat terpisah pada list. Query tidak boleh mengubah status.

Field shared content meliputi kategori utama, lokasi, koordinat/accuracy, dan main media. Translation meliputi title, slug, summary, body, SEO title/description, locale publication status, dan metadata editor.

### Create/update

`POST /contents` membuat parent content dan menerima field shared minimum. Translation dibuat melalui endpoint translation. ID content dikirim sebagai string.

`PUT /contents/{id}/translations/{locale}` menyimpan satu locale saja. Perpindahan tab ID/EN di UI tidak boleh menyebabkan data locale lain tertimpa.

### Preview

Preview memerlukan autentikasi Admin. Preview harus:
- menampilkan rendering yang sama dengan website publik;
- dapat menerima unsaved content dari request body atau preview draft tersimpan;
- tidak mengubah versi Published;
- tidak membuat URL preview yang dapat diakses publik tanpa autentikasi;
- mengirim `Cache-Control: no-store` dan noindex;
- tidak masuk sitemap atau SSG release publik.

### Publish

Publish satu locale adalah transaksi:
1. validasi content dan locale translation;
2. pastikan category aktif;
3. pastikan minimal satu content source tercatat;
4. validasi media/rights dan field wajib;
5. ubah locale status menjadi Published dan set metadata publikasi;
6. tulis audit log;
7. enqueue site generation job;
8. commit seluruh perubahan bersama.

Jika validasi gagal, tidak ada perubahan parsial. Tidak ada Verified/Review/approval gate.

### Archive

Archive menghapus locale tersebut dari public output setelah SSG berhasil. Versi locale lain tidak berubah. Konten yang archived tetap berada di CMS.

### Permanent delete

Delete bersifat permanen untuk content identity dan seluruh locale versions serta content sources terkait. Harus ada konfirmasi eksplisit dari UI. Media bersama tidak ikut terhapus. Audit hanya menyimpan identitas referensi dan metadata tindakan, bukan salinan penuh konten yang dihapus. Database backup operasional berada di luar CMS.

## 5. Categories

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/categories` | List kategori |
| POST | `/categories` | Buat kategori |
| GET | `/categories/{category_id}` | Detail kategori |
| PUT | `/categories/{category_id}` | Ubah field bersama/status |
| PUT | `/categories/{category_id}/translations/{locale}` | Buat/perbarui translation |
| DELETE | `/categories/{category_id}` | Permanent delete jika tidak direferensikan |

Category status Active/Inactive bersifat global. Inactive tidak tampil publik dan tidak dapat dipilih untuk konten baru. Konten existing tidak otomatis dihapus/dipindahkan.

Delete ditolak `409 CATEGORY_IN_USE` bila masih ada konten Draft, Published, atau Archived yang mereferensikan kategori. Admin harus memindahkan seluruh konten secara eksplisit ke kategori aktif terlebih dahulu. Slug locale yang berubah memerlukan redirect 301 untuk URL kategori dan detail konten terdampak.

## 6. Content sources

Sumber disimpan bersama content:
- `research`
- `visit`
- `official`
- `social`

Sumber memuat nama, URL/referensi, tanggal publikasi/akses bila tersedia, notes, dan informasi asal/hak media bila relevan. Sumber merupakan traceability, bukan status verifikasi. Jangan menampilkan sumber internal yang ditandai private pada public page.

Endpoints:
- `GET /contents/{content_id}/sources`
- `POST /contents/{content_id}/sources`
- `PUT /contents/{content_id}/sources/{source_id}`
- `DELETE /contents/{content_id}/sources/{source_id}`

Publish mensyaratkan minimal satu sumber yang dapat ditelusuri.

## 7. Media Library

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/media` | List/search media |
| POST | `/media` | Upload multipart |
| GET | `/media/{media_id}` | Metadata |
| PUT | `/media/{media_id}` | Ubah alt text/rights metadata |
| DELETE | `/media/{media_id}` | Delete jika tidak digunakan |

Media metadata mencakup file name, storage key, MIME, size, dimensions, alt ID/EN, origin/source, license/permission, attribution, created timestamp, dan Admin uploader. File yang direferensikan konten/halaman tidak dapat dihapus (`409 MEDIA_IN_USE`).

## 8. Location master

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/locations/tree` | Kabupaten → Kecamatan → Gampong |
| GET | `/locations/districts` | Kecamatan |
| GET | `/locations/villages?district_id=` | Gampong berdasarkan kecamatan |

MVP menggunakan master data Kabupaten Pidie, kecamatan, dan gampong. API CRUD master lokasi tidak dibuka pada CMS MVP kecuali proses import/maintenance internal telah ditentukan. Admin memilih dari master bila data tersedia; alamat rinci dan koordinat dapat diisi manual. Koordinat tidak boleh dibuat-buat.

## 9. Homepage curation

Featured content:
- `GET /curation/featured`
- `PUT /curation/featured`

Story section:
- `GET /curation/stories/{locale}`
- `PUT /curation/stories/{locale}`

Curation hanya boleh memilih konten Published pada locale terkait. Featured dan Story adalah dua konfigurasi berbeda. Story mempunyai satu Lead Story dan Supporting Stories berurutan; Lead tidak boleh muncul lagi sebagai Supporting. Featured/Latest/Random di homepage tidak boleh menampilkan duplikasi yang dilarang oleh aturan homepage.

## 10. Static pages

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/static-pages` | List halaman |
| POST | `/static-pages` | Buat halaman |
| GET | `/static-pages/{page_id}` | Detail |
| PUT | `/static-pages/{page_id}` | Update field bersama |
| PUT | `/static-pages/{page_id}/translations/{locale}` | Update locale |
| POST | `/static-pages/{page_id}/translations/{locale}/publish` | Publish locale |
| POST | `/static-pages/{page_id}/translations/{locale}/archive` | Archive locale |
| DELETE | `/static-pages/{page_id}` | Permanent delete |

ID dan EN independen. Halaman informasi disajikan melalui SSG. Rich text harus disanitasi server-side; raw HTML/JavaScript tidak diterima.

## 11. Website settings

- `GET /settings`
- `PUT /settings`
- `GET /settings/translations/{locale}`
- `PUT /settings/translations/{locale}`

Mencakup identitas situs, logo/favicon, kontak, social URLs, bahasa, default SEO. Konten dapat memiliki SEO override atas default. Validasi URL dan file media wajib.

## 12. Administrator management

| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/administrators` | List Admin |
| POST | `/administrators` | Tambah Admin |
| GET | `/administrators/{admin_id}` | Detail |
| PUT | `/administrators/{admin_id}` | Update profil |
| POST | `/administrators/{admin_id}/activate` | Aktifkan |
| POST | `/administrators/{admin_id}/deactivate` | Nonaktifkan |
| POST | `/administrators/{admin_id}/password-reset` | Proses reset password |

Tidak ada self-registration. Admin inactive tidak dapat login. Sistem harus mencegah tindakan yang mengunci seluruh akses Admin aktif, termasuk menonaktifkan akun aktif terakhir tanpa pengganti. Detail pengiriman reset password dan bootstrap Admin ditetapkan pada security/deployment specification.

## 13. Audit log

`GET /audit-logs`

Filter minimum: `actor_admin_id`, `action`, `entity_type`, `date_from`, `date_to`, pagination. Audit log bersifat read-only dari CMS. Tidak ada endpoint edit/delete audit log.

## 14. Status implementasi

Endpoint di dokumen ini adalah contract baseline, bukan klaim bahwa seluruh endpoint wajib dibuat sekaligus. Implementasi mengikuti backlog MVP dan hanya membuat endpoint yang memiliki consumer nyata. Endpoint internal SSG dipisahkan dari Admin API.

---

## Materi sumber: `publication-api-and-ssg.md`

# Publication, Preview, and SSG Pipeline

**Status:** Baseline technical design  
**Runtime:** Go `net/http` + PostgreSQL  
**No Redis / message broker**

## 1. Delivery model

- Website publik: pre-rendered HTML/CSS/JS/assets hasil SSG.
- CMS Admin: Vue application dinamis yang menggunakan Admin REST API.
- Go `net/http`: melayani API dan file statis website/CMS.
- PostgreSQL: source of truth dan durable job queue.
- Tidak ada Nginx/Caddy/Apache/reverse proxy aplikasi tambahan.

Halaman publik utama tidak melakukan fetch API untuk setiap konten saat dibuka. Regeneration terjadi setelah perubahan publikasi/curation/settings yang memengaruhi output.

## 2. Generation trigger

Trigger internal:
- Publish translation;
- Archive translation;
- perubahan konten Published;
- perubahan kategori/slug/status yang memengaruhi URL atau listing;
- perubahan Featured/Story curation;
- perubahan static page Published;
- perubahan website settings yang memengaruhi output.

Draft save tidak memicu public generation, kecuali preview yang bersifat terpisah dan privat.

## 3. Atomic database transaction

Untuk operasi yang memengaruhi publikasi:
1. mulai transaksi PostgreSQL;
2. validasi state/domain;
3. simpan perubahan;
4. tulis audit event;
5. insert `site_generation_jobs` berstatus `pending` dengan payload perubahan;
6. commit.

Dengan pola transactional enqueue ini, perubahan Published tidak dapat tersimpan tanpa job generation yang tercatat.

## 4. Worker without message broker

Worker berada dalam Go application/deployment yang sama. Worker mengambil job dari PostgreSQL:
- transaksi singkat untuk claim;
- `SELECT ... FOR UPDATE SKIP LOCKED`;
- ubah status menjadi `processing`, set lock/attempt;
- commit claim sebelum proses build yang panjang;
- build output pada temporary release directory;
- validasi hasil build;
- aktifkan release baru secara atomic hanya jika build sukses;
- tandai job `succeeded`.

Jika gagal:
- simpan error ringkas yang aman dan metadata untuk diagnosis;
- increment attempts;
- retry terbatas dengan backoff;
- setelah batas retry, tandai `failed` dan tampilkan status pada CMS/operational log.

Job yang `processing` tetapi lock kedaluwarsa dapat di-claim ulang oleh worker recovery. Worker harus idempotent: job lama tidak boleh merusak release lebih baru. Gunakan generation sequence/release version untuk mencegah stale job mengaktifkan output yang lebih lama.

## 5. Build and release strategy

Jangan menulis file langsung ke direktori live saat generation berlangsung.

Pola:
1. buat direktori release unik;
2. render seluruh halaman yang terdampak (atau full site bila lebih aman untuk MVP);
3. tulis HTML dan asset manifest;
4. validasi file penting, sitemap, canonical, locale, dan link target;
5. atomically switch active release pointer/directory;
6. pertahankan release sebelumnya sampai release baru terbukti aktif.

Metode atomic switch harus disesuaikan dengan filesystem/deployment. Jangan mengasumsikan symlink swap aman pada semua platform; uji di server target.

Jika generation gagal, public website tetap menyajikan release terakhir yang sukses.

## 6. Incremental vs full generation

MVP boleh memakai full-site generation bila ukuran konten masih kecil dan durasi build dapat diterima. Arsitektur job tetap mencatat trigger dan payload agar incremental generation dapat ditambahkan kemudian. Jangan mengimplementasikan invalidation graph kompleks sebelum ada bukti kebutuhan.

## 7. Preview

Preview hanya untuk Admin authenticated:
- tidak menulis ke public release;
- tidak masuk sitemap;
- tidak dapat diakses anonymous;
- tidak boleh di-cache bersama public content;
- menampilkan draft/unsaved version sesuai request;
- renderer dan template sama dengan public page agar hasil representatif.

Preview dapat dilakukan on-demand oleh Go renderer atau preview build terisolasi. Pilih implementasi paling sederhana yang tidak mempublikasikan file preview ke public directory.

## 8. URL redirects

Setiap perubahan published slug atau category slug yang mengubah path harus membuat redirect record:
- old path;
- target canonical path;
- status 301;
- timestamp dan actor.

Redirect harus:
- menghindari loop;
- menghindari chain bila dapat diarahkan langsung ke final URL;
- mempertahankan redirect lama saat slug berubah berulang;
- menolak reuse old path selama masih dipakai sebagai redirect;
- diproses oleh Go router sebelum fallback static file/SPA route.

Pemindahan kategori konten yang mengubah detail URL juga memerlukan redirect.

## 9. Go HTTP routing order

Urutan konseptual request:
1. request ID, security headers, logging/recovery;
2. health/readiness routes;
3. `/api/v1/...` API routes;
4. redirect lookup untuk public paths;
5. serve static asset/file bila tersedia;
6. serve generated HTML page untuk route publik;
7. serve CMS entry point pada path CMS yang ditetapkan;
8. 404.

Static file path harus dinormalisasi dan dibatasi pada active release directory untuk mencegah path traversal. Jangan melakukan fallback SPA untuk URL konten publik yang tidak ditemukan jika dapat menutupi 404.

## 10. Operational visibility

CMS Dashboard atau halaman status operasional minimal dapat menampilkan:
- generation job terakhir;
- status pending/processing/succeeded/failed;
- waktu terakhir sukses;
- error ringkas untuk Admin;
- tombol retry untuk job failed (bila diimplementasikan).

Tidak perlu membangun observability platform eksternal pada MVP. Log terstruktur `slog` dan health/readiness endpoint cukup sebagai baseline.

---
