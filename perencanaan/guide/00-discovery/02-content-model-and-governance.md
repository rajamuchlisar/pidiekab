# 02 Content Model And Governance

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `content-governance.md`
- `content-category-management.md`
- `fixed-content-categories.md`
- `content-category-relationship.md`
- `content-category-status.md`
- `content-category-deletion-rules.md`
- `category-bilingual-fields.md`
- `category-localized-slugs.md`
- `category-specific-detail-fields.md`
- `category-content-validation.md`

---

## Materi sumber: `content-governance.md`

# Content Governance & Management Model

## Decision CG-001 --- Internal Admin-Managed Content

**Status: Agreed**

All public content on PidieKab is created, entered, edited, reviewed,
and published by the internal PidieKab admin/editor team.

### Product implications

1.  No public self-service account for tourism operators, business
    owners, community members, or general users in the MVP.
2.  No public submission portal for creating destination,
    culinary/business, cultural, or historical listings in the MVP.
3.  Internal admins/editors are responsible for content quality,
    completeness, publication, and updates.
4.  Public visitors can browse published content without an account,
    subject to later product decisions.
5.  A future suggestion/correction channel may be evaluated separately,
    but it is not part of the agreed content-management model for the
    MVP.
6.  Admin authentication and role permissions are required for content
    administration. Exact roles and permissions remain to be specified.

## Editorial Workflow --- Initial Hypothesis

The following workflow is a proposed operational baseline and requires
confirmation:

1.  Admin/editor gathers information and supporting sources.
2.  Admin/editor creates or updates a content record in draft state.
3.  Editor checks accuracy, completeness, source attribution, image
    rights, and relevance.
4.  Authorized admin/editor publishes the content.
5.  Content is reviewed periodically and updated when details change.
6.  Incorrect or outdated content can be unpublished or archived while
    being corrected.

## Content Trust Requirements

For content where applicable, store or display: - Source name and source
URL/reference. - Verification status. - Last verified date. - Last
content update date. - Editor or internal account responsible for the
change. - Image origin and usage permission/attribution information.

Do not imply that content is officially endorsed by a government
institution. When citing official public information, clearly identify
the source institution and link to its official channel where
appropriate.

## Out of Scope for MVP

-   Business-owner login and self-service listing management.
-   Public account registration for content submission.
-   Public listing creation or editing.
-   Automatic publication of user-submitted content.
-   Community moderation workflows.

These capabilities may be reconsidered in a later phase after the
operating model and demand are validated.

## Open Questions

-   How many internal content roles are needed (e.g. Super Admin,
    Editor, Author)?
-   Can an author create drafts while only an editor/admin can publish?
-   Is a separate reviewer/approval step required, or can one authorized
    editor publish?
-   What is the content update/re-verification cadence for each domain?
-   Will the platform provide a public correction/contact channel
    without allowing direct content submission?

---

## Materi sumber: `content-category-management.md`

# CAT-001 — Kategori Konten Dikelola melalui CMS

**Status: Disepakati.**

Kategori konten PidieKab dapat dikelola Admin melalui CMS menggunakan CRUD.
Tiga kategori awal MVP adalah Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.
Admin dapat menambah kategori dan menyesuaikan nama, deskripsi, slug, serta informasi
tampilan kategori. Kategori disimpan sebagai data referensi dinamis, bukan enum yang
harus diubah melalui deployment.

Setiap kategori memiliki status Aktif/Nonaktif. Aturan status dijelaskan pada
`content-category-status.md`.

Kategori yang masih memiliki konten terkait tidak dapat dihapus. Seluruh konten harus
dipindahkan terlebih dahulu ke kategori lain yang Aktif. Tidak ada pemindahan otomatis
atau kategori default. Rincian proses penghapusan ada pada
`content-category-deletion-rules.md`.

Kategori editorial seperti “Cerita tentang Pidie” dan “Konten Unggulan” tetap
merupakan kurasi/atribut konten, bukan kategori bawaan.

---

## Materi sumber: `fixed-content-categories.md`

# CAT-001 — Kategori Konten Dikelola melalui CMS

**Status: Disepakati.** Kategori konten dikelola Admin melalui CRUD. Tiga kategori awal adalah Wisata & Destinasi, Kuliner & UMKM, dan Budaya & Sejarah. Untuk aturan CRUD dan relasi konten, lihat `content-category-management.md`.

---

## Materi sumber: `content-category-relationship.md`

# CNT-001 — Relasi Konten dan Kategori

**Status: Disepakati.**

## Keputusan
Pada MVP, satu kategori dapat memiliki banyak konten, tetapi setiap konten hanya memiliki **satu kategori utama**.

Relasi:
- Category 1 → Many Content.
- Content Many → 1 Category.

Contoh:
- Pantai Mantak Tari → Wisata & Destinasi.
- Mie Caluek → Kuliner & UMKM.
- Seudati → Budaya & Sejarah.

## Aturan
- Setiap konten wajib memiliki tepat satu kategori utama.
- Satu kategori dapat digunakan oleh banyak konten.
- Admin memilih satu kategori ketika membuat atau mengubah konten.
- Konten tidak dapat dipublikasikan tanpa kategori yang valid dan Aktif.
- Perubahan kategori pada konten yang sudah dipublikasikan harus memperbarui konteks kategori pada halaman publik, navigasi, filter pencarian, dan breadcrumb.
- Pemindahan konten dari satu kategori ke kategori lain tidak menghapus konten.
- Tidak ada multi-category pada MVP.
- Tidak ada tag system pada MVP.
- Tidak ada kategori sekunder atau relasi many-to-many untuk konten.

## Dampak terhadap fitur publik
- Setiap konten tampil pada satu halaman kategori utama.
- Filter kategori pencarian menggunakan kategori utama konten.
- Breadcrumb menampilkan satu jalur kategori yang konsisten.
- Kartu konten menampilkan satu label kategori utama.
- Konten tidak diduplikasi di beberapa kategori karena relasi kategori.

## Catatan teknis
Relasi dapat direpresentasikan dengan foreign key `category_id` pada identitas konten. Constraint database harus memastikan setiap konten memiliki kategori valid. Detail perilaku saat kategori dinonaktifkan/dihapus, pemindahan kategori, URL, cache, dan indeks pencarian dirinci dalam Feature Specification dan Technical Specification.

---

## Materi sumber: `content-category-status.md`

# CAT-002 — Status Aktif dan Nonaktif Kategori

**Status: Disepakati.**

## Keputusan
Setiap kategori konten memiliki status **Aktif** atau **Nonaktif** yang dapat dikelola Admin melalui CMS.

Status ini melengkapi kemampuan CRUD kategori pada keputusan CAT-001.

## Perilaku kategori

### Aktif
- Kategori tersedia untuk dipilih ketika Admin membuat atau mengubah konten.
- Kategori dapat ditampilkan pada navigasi, daftar kategori, dan area publik lain sesuai konfigurasi website.
- Konten yang terhubung ke kategori aktif dapat dipublikasikan jika memenuhi validasi publikasi lainnya.

### Nonaktif
- Kategori tidak ditampilkan pada navigasi dan daftar kategori publik.
- Kategori tidak dapat dipilih untuk konten baru atau saat mengganti kategori konten.
- Kategori tetap tersimpan di CMS beserta relasi dan riwayat kontennya.
- Konten lama yang terhubung ke kategori nonaktif tidak otomatis dihapus atau dipindahkan.
- Admin perlu menangani konten terkait sebelum kategori dapat diaktifkan kembali atau dihapus.

## Aturan validasi
- Kategori yang masih digunakan oleh konten tidak boleh dihapus langsung.
- Sebelum penghapusan, Admin harus memindahkan seluruh konten terkait ke kategori lain yang aktif.
- Menonaktifkan kategori tidak menghapus konten yang terkait.
- Konten dalam kategori nonaktif tidak boleh tampil sebagai konten publik sampai dipindahkan ke kategori aktif dan memenuhi aturan publikasi.
- Jika kategori nonaktif memiliki konten Published, CMS harus memberikan peringatan yang jelas sebelum perubahan status dikonfirmasi.
- Perubahan status kategori harus tercatat pada audit log.

## Dampak pada website publik
- Navigasi, daftar kategori, dan filter pencarian hanya menawarkan kategori aktif.
- Hasil pencarian publik tidak boleh menampilkan konten yang masih terhubung ke kategori nonaktif.
- URL kategori nonaktif tidak ditampilkan sebagai halaman kategori publik. Respons URL dan penanganan SEO/redirect akan ditentukan pada Feature Specification dan Technical Specification.
- Konten yang dipindahkan ke kategori lain mengikuti URL, canonical, dan redirect rules yang akan ditetapkan secara terpisah.

## Batasan MVP
- Status kategori hanya dua: Aktif dan Nonaktif.
- Tidak ada penghapusan otomatis atau pemindahan otomatis konten.
- Tidak ada penjadwalan aktivasi/nonaktivasi.
- Tidak ada hierarki kategori atau subkategori yang diwajibkan oleh keputusan ini.

## Catatan teknis
Status dapat direpresentasikan sebagai boolean `is_active` atau enum sederhana. Pemilihan implementasi, transaksi perubahan kategori, validasi relasi, audit log, serta perilaku URL dan SEO dirinci dalam Technical Specification.

---

## Materi sumber: `content-category-deletion-rules.md`

# CAT-003 — Aturan Penghapusan Kategori

**Status: Disepakati.**

## Keputusan
Kategori hanya dapat dihapus secara permanen apabila tidak lagi memiliki konten yang terhubung. Admin wajib memindahkan seluruh konten terkait ke kategori lain yang berstatus Aktif sebelum menghapus kategori.

Tidak ada kategori default otomatis dan tidak ada pemindahan konten otomatis.

## Alur penghapusan
1. Admin memilih aksi Hapus pada kategori.
2. Sistem memeriksa seluruh konten terkait, termasuk Draft, Published, Archived, dan seluruh versi bahasa.
3. Jika masih ada konten, penghapusan ditolak. CMS menampilkan jumlah konten terkait dan meminta Admin memindahkannya ke kategori lain yang Aktif.
4. Admin memindahkan seluruh konten terkait.
5. Admin mengulangi aksi Hapus.
6. Sistem memeriksa ulang relasi untuk mencegah penghapusan jika masih ada konten terkait.
7. Jika sudah tidak ada relasi, CMS meminta konfirmasi penghapusan permanen.
8. Setelah dikonfirmasi, kategori dihapus dan aksi dicatat pada audit log.

## Aturan wajib
- Semua status konten diperhitungkan: Draft, Published, dan Archived.
- Relasi diperiksa pada level identitas konten, bukan hanya satu versi bahasa.
- Tujuan pemindahan harus kategori lain yang Aktif.
- Tidak ada kategori pengganti/default otomatis.
- Tidak ada penghapusan konten secara cascading akibat penghapusan kategori.
- Pemeriksaan relasi dilakukan kembali saat persistensi untuk menghindari race condition.
- Penghapusan permanen memerlukan konfirmasi eksplisit.

## Dampak URL dan SEO
Penghapusan kategori tidak menghapus konten. URL konten mengikuti aturan URL/slug yang berlaku. Penanganan URL halaman kategori yang dihapus, canonical, sitemap, dan redirect ditetapkan pada Feature Specification dan Technical Specification.

## Audit
Audit log mencatat Admin, ID dan nama kategori, waktu penghapusan, serta aksi. Audit log tidak perlu menyimpan salinan penuh kategori yang dihapus.

## Batasan MVP
- Tidak ada recycle bin atau pemulihan kategori melalui CMS.
- Tidak ada soft delete kategori.
- Tidak ada penghapusan otomatis konten terkait.
- Tidak ada kategori default otomatis.
- Backup operasional berada di luar fitur pemulihan CMS.

## Catatan teknis
Gunakan foreign key dan strategi transaksi yang mencegah kategori dihapus selama masih direferensikan konten. Detail locking, constraint, audit event, dan respons API ditetapkan pada Technical Specification.

---

## Materi sumber: `category-bilingual-fields.md`

# CAT-006 — Data Kategori Bilingual

**Status: Disepakati berdasarkan arahan untuk melanjutkan.**

## Keputusan
Setiap kategori memiliki data tampilan yang dikelola terpisah untuk Bahasa Indonesia (`id`) dan Bahasa Inggris (`en`):
- Nama kategori.
- Deskripsi kategori.
- Slug kategori.

Status Aktif/Nonaktif berlaku untuk kategori secara keseluruhan, bukan per bahasa.

## Perilaku bahasa
- Admin dapat mengisi dan menyimpan data ID dan EN secara independen.
- Perubahan nama, deskripsi, atau slug pada satu bahasa tidak mengubah data bahasa lainnya.
- Halaman kategori menggunakan data sesuai locale aktif.
- Website tidak boleh menampilkan nama atau deskripsi bahasa lain seolah-olah merupakan terjemahan.
- Jika data kategori untuk locale tertentu belum tersedia atau belum layak ditampilkan, perilaku halaman mengikuti aturan ketersediaan locale yang akan ditetapkan pada Feature Specification.
- Slug per locale mengikuti CAT-005 dan redirect mengikuti CAT-004.

## Status kategori
- Satu status Aktif/Nonaktif berlaku untuk seluruh kategori.
- Kategori Nonaktif tidak ditampilkan pada navigasi, listing kategori, atau filter publik.
- Kategori Nonaktif tidak dapat dipilih untuk konten baru.
- Penonaktifan tidak menghapus data ID/EN maupun konten terkait.

## Validasi CMS
- Nama kategori wajib diisi untuk bahasa yang akan ditampilkan secara publik.
- Slug wajib valid dan unik pada locale terkait.
- Deskripsi kategori dapat digunakan untuk pengantar halaman kategori; batas panjang dan apakah wajib akan ditetapkan dalam Feature Specification.
- CMS menampilkan indikator kelengkapan ID dan EN agar Admin mengetahui bahasa mana yang belum siap.
- Admin dapat menyimpan perubahan satu locale tanpa memaksa penyelesaian locale lainnya.

## Batasan MVP
- Bahasa yang didukung: ID dan EN.
- Tidak ada bahasa ketiga pada MVP.
- Tidak ada auto-translation yang langsung dipublikasikan.
- Tidak ada status aktif terpisah per bahasa; status aktif kategori berlaku global.

## Catatan teknis
Model data kategori harus memisahkan identitas/status kategori dari field terjemahan. Struktur tabel, aturan kelengkapan, dan perilaku publik ketika terjemahan belum tersedia ditetapkan pada Feature Specification dan Technical Specification.

---

## Materi sumber: `category-localized-slugs.md`

# CAT-005 — Slug Kategori Terpisah untuk Setiap Bahasa

**Status: Disepakati.**

## Keputusan
Setiap kategori memiliki slug yang dikelola secara terpisah untuk setiap bahasa yang didukung PidieKab, yaitu Bahasa Indonesia (`id`) dan Bahasa Inggris (`en`).

Contoh:
- Indonesia: `/id/wisata-destinasi`
- Inggris: `/en/tourism-destinations`

Slug ID dan EN dapat berbeda, dan perubahan slug pada satu bahasa tidak mengubah slug bahasa lainnya.

## Aturan pengelolaan
- Admin mengelola slug ID dan EN melalui form kategori dengan tab/bagian bahasa yang jelas.
- Setiap bahasa memiliki satu slug aktif untuk kategori tersebut.
- Slug mengikuti aturan URL-friendly dari CAT-004.
- Slug harus unik dalam ruang URL kategori pada locale yang sama.
- Kategori dapat memiliki slug untuk satu bahasa meskipun informasi kategori pada bahasa lainnya belum lengkap; perilaku publik mengikuti status publikasi kategori per bahasa yang akan ditetapkan.
- Perubahan slug untuk satu locale hanya memengaruhi URL locale tersebut.
- Perubahan slug kategori yang sudah dipublikasikan pada locale terkait wajib membuat permanent redirect HTTP 301 dari URL lama ke URL baru.
- Riwayat redirect disimpan per locale.
- Slug lama tidak boleh digunakan ulang selama masih menjadi target redirect.
- Redirect tidak boleh membentuk loop atau rantai redirect yang tidak perlu.

## Contoh
Kategori “Wisata & Destinasi”:
- ID slug: `wisata-destinasi`
- EN slug: `tourism-destinations`

Jika Admin mengubah slug ID menjadi `destinasi-wisata`, hanya URL `/id/wisata-destinasi` yang diarahkan ke `/id/destinasi-wisata`. URL bahasa Inggris tidak berubah.

## CMS
- Form kategori menyediakan pengaturan ID dan EN secara terpisah.
- Admin dapat menyimpan perubahan satu bahasa tanpa harus mengubah slug bahasa lainnya.
- CMS memberi peringatan sebelum menyimpan perubahan slug yang telah digunakan pada URL publik.
- Validasi keunikan dilakukan per locale.
- Perubahan dicatat pada audit log dengan locale, slug lama, slug baru, Admin, dan waktu perubahan.

## SEO dan URL
- URL kategori mengikuti prefix locale (`/id/` dan `/en/`).
- Canonical kategori mengarah ke URL kategori pada locale yang sama.
- Sitemap hanya mencantumkan halaman kategori yang memenuhi aturan publikasi/indexing locale terkait.
- Relasi `hreflang` hanya dipasang sesuai ketersediaan versi kategori yang benar-benar dapat diakses pada masing-masing bahasa. Detail final mengikuti strategi multilingual SEO.

## Batasan MVP
- Mendukung slug kategori ID dan EN.
- Tidak ada alias slug aktif tambahan.
- Tidak ada pengelolaan redirect manual umum oleh Admin.
- Tidak ada fallback yang menampilkan slug atau label bahasa lain seolah-olah diterjemahkan.

## Catatan teknis
Technical Specification harus menetapkan struktur penyimpanan slug per locale, constraint unik gabungan locale dan slug, tabel riwayat redirect, resolusi route, transaksi update slug, pencegahan loop/rantai, dan pengujian HTTP 301.

---

## Materi sumber: `category-specific-detail-fields.md`

# DET-001 — Informasi Khusus per Kategori

**Status: Disepakati**

Semua halaman detail menggunakan struktur bersama NAV-004. Setiap kategori memiliki informasi tambahan sesuai kebutuhan dan kelayakan data. Daftar di bawah adalah struktur awal; penentuan field wajib/opsional masih dalam Discovery.

## Wisata & Destinasi
- Nama destinasi dan jenis wisata.
- Deskripsi dan daya tarik.
- Alamat/lokasi; peta hanya bila relevan dan koordinat memadai (MAP-001).
- Jam operasional, biaya masuk, dan fasilitas bila tersedia, terverifikasi, serta dapat dipelihara.

## Kuliner & UMKM
- Nama usaha/tempat kuliner dan jenis usaha.
- Produk atau makanan unggulan dan deskripsi.
- Alamat atau area layanan bila relevan.
- Kontak eksternal dan jam operasional bila tersedia serta terverifikasi.

## Budaya & Sejarah
- Nama tradisi, budaya, atau objek sejarah dan jenisnya.
- Latar belakang, cerita, nilai/keunikan.
- Lokasi atau wilayah asal jika relevan dan didukung sumber.
- Sumber sejarah dan referensi.

## Aturan umum
- Tidak memaksakan field yang tidak relevan, kosong, atau belum dapat diverifikasi.
- Tidak mengarang koordinat, harga, jam buka, atau klaim sejarah.
- Mengikuti status publikasi per bahasa (LNG-002), tata kelola konten (CG-001), sumber informasi (SRC-001), dan satu foto utama (MED-001).
- Field wajib/opsional, format data, dan validasi final belum diputuskan.

---

## Materi sumber: `category-content-validation.md`

# DET-002 — Validasi Konten Fleksibel per Kategori

**Status: Disepakati**

PidieKab menggunakan pendekatan **Validasi Fleksibel** untuk pengisian dan publikasi konten.

## Field wajib untuk semua kategori

1. Judul konten.
2. Kategori.
3. Deskripsi atau isi utama.
4. Minimal satu sumber informasi yang dapat ditelusuri.

## Field opsional

Field khusus kategori seperti lokasi, koordinat, jam operasional, biaya masuk, fasilitas, kontak, produk unggulan, wilayah asal, dan informasi pendukung lainnya bersifat opsional sesuai relevansi dan ketersediaan data.

## Aturan publikasi dan kualitas

- Admin tidak diwajibkan mengisi field yang tidak relevan dengan konten.
- Field opsional yang kosong tidak ditampilkan pada halaman publik.
- Field yang berpotensi menyesatkan atau belum memiliki dasar yang memadai tidak boleh diterbitkan sebagai fakta.
- Sumber informasi wajib dicatat agar klaim dapat ditelusuri dan ditinjau.
- Persyaratan hak penggunaan foto tetap berlaku; foto hanya ditampilkan jika tersedia dan layak digunakan. Ketiadaan foto tidak otomatis menggagalkan publikasi, selama fallback visual disediakan.
- Status verifikasi editorial dan kelayakan publikasi tetap menjadi bagian dari proses governance; pendekatan fleksibel bukan berarti konten boleh diterbitkan tanpa pemeriksaan.
- Validasi teknis rinci, seperti batas panjang judul, format URL, dan aturan koordinat, ditentukan pada Feature Specification.

## Penerapan pada kategori

### Wisata & Destinasi
Field seperti jenis wisata, lokasi, jam operasional, biaya masuk, dan fasilitas diisi jika tersedia, relevan, serta dapat dipertanggungjawabkan.

### Kuliner & UMKM
Field seperti jenis usaha, produk unggulan, area layanan, kontak, dan jam operasional diisi jika tersedia, relevan, serta dapat dipertanggungjawabkan.

### Budaya & Sejarah
Field seperti jenis budaya/sejarah, latar belakang, keunikan, wilayah asal, dan referensi tambahan diisi sesuai konteks dan kekuatan sumber.

## Batasan

Validasi fleksibel tidak menghapus kewajiban editorial untuk memeriksa akurasi, sumber, hak penggunaan media, dan status publikasi bahasa.

---
