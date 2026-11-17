# 05 Multilingual Url And Seo

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `language-strategy.md`
- `multilingual-content-management.md`
- `multilingual-publication.md`
- `multilingual-seo-strategy.md`
- `root-domain-language-behavior.md`
- `public-url-structure.md`
- `url-slug-rules.md`
- `content-detail-url-structure.md`
- `category-slug-and-redirects.md`

---

## Materi sumber: `language-strategy.md`

# Language & Localization Strategy

## Decision LNG-001 --- Indonesian and English at MVP

**Status: Agreed**

PidieKab will support two languages from the MVP launch:

  Language           Locale code   Role
  ------------------ ------------- -------------------------------------------------
  Bahasa Indonesia   `id`          Primary and default language
  English            `en`          Alternative language for international visitors

## Product Rules

1.  Indonesian (`id`) is the default language for first-time visitors
    unless a supported locale preference is available.
2.  Visitors can switch language through a visible language selector.
3.  Public content should support Indonesian and English versions.
4.  Admin/editor users manage both language versions through the
    internal CMS.
5.  A content translation may have its own editorial/publication state.
    Do not assume that publishing one language automatically publishes
    the other.
6.  If the requested language version is unavailable, clearly
    communicate this and provide a route to the Indonesian version. Do
    not silently display Indonesian text while labeling the page as
    English.
7.  Preserve proper names, local place names, and culturally significant
    terms where translation would be misleading.
8.  Machine-assisted translation may be used as an editorial aid, but
    must not be published without human review.
9.  Historical, cultural, geographic, and practical visitor information
    require careful translation review.
10. Language support must be reflected in navigation, content URLs, page
    metadata, search, and SEO metadata.

## Content and Data Implications

The content model should support language-specific fields or translation
records, including: - Localized title. - Localized slug where needed. -
Localized summary and body. - SEO title and description per language. -
Translation status. - Published status and timestamps per language. -
Editor/reviewer information.

Shared facts such as coordinates, source references, and certain contact
details should not be duplicated unnecessarily across translations.

## URL and SEO Direction

Proposed URL pattern for evaluation: - Indonesian: `/id/...` - English:
`/en/...`

The final routing strategy, canonical URLs, `hreflang`, sitemap
behavior, and fallback behavior will be defined in the PRD and technical
specification.

## Out of Scope at MVP

-   Acehnese language interface/content.
-   Automatic public translation without editorial approval.
-   User-contributed translations.

## Open Decisions

-   Should the root URL `/` redirect to `/id/` or render Indonesian
    without a locale prefix?
-   Must every MVP content item have both Indonesian and English before
    publication?
-   Can an Indonesian version be published while its English translation
    remains draft?
-   Who reviews English translations and what is the expected editorial
    process?
-   Should the admin CMS itself be bilingual, or is Indonesian-only
    admin UI sufficient initially?

---

## Materi sumber: `multilingual-content-management.md`

# LNG-004 — Model Pengelolaan Konten Multibahasa

**Status: Disepakati**

PidieKab menggunakan model **Satu Konten dengan Dua Versi Bahasa** untuk Bahasa Indonesia (`id`) dan English (`en`).

## Struktur pengelolaan

- Satu entri konten menjadi identitas utama konten.
- Di dalam entri tersebut, Admin mengelola versi Bahasa Indonesia dan English.
- Setiap versi bahasa memiliki field editorialnya sendiri, termasuk judul, slug, ringkasan/isi, SEO title, SEO description, dan status publikasi.
- Status publikasi setiap bahasa berdiri sendiri sesuai LNG-002.
- Sumber informasi, koordinat, data lokasi, dan metadata faktual bersama disimpan pada tingkat konten utama jika memang sama untuk kedua bahasa.
- Informasi yang memerlukan variasi bahasa atau konteks lokal dapat memiliki nilai khusus per bahasa bila diperlukan.

## Aturan publikasi

- Versi Indonesia dapat diterbitkan meskipun versi English masih Draft atau Review, dan sebaliknya jika telah melalui proses editorial.
- Status Published pada satu bahasa tidak otomatis menerbitkan bahasa lainnya.
- Pengunjung hanya melihat versi yang telah dipublikasikan dalam bahasa yang sedang dipilih.
- Sistem tidak boleh menampilkan isi bahasa Indonesia ketika halaman diberi label English, atau sebaliknya.
- Perubahan pada satu versi bahasa tidak otomatis mengubah status versi bahasa lainnya.

## Hubungan antarversi

- Kedua versi terhubung ke satu identitas konten yang sama.
- Jika tersedia, language switcher pada halaman detail mengarah ke versi bahasa lain dari konten yang sama.
- Jika versi bahasa tujuan belum Published, sistem mengikuti perilaku fallback yang akan ditentukan pada keputusan LNG-005.
- Slug dapat berbeda untuk setiap bahasa dan harus unik sesuai aturan URL yang akan ditetapkan pada Technical Specification.

## Batasan dan konsistensi data

- Hindari menggandakan data bersama seperti koordinat dan sumber yang sama pada kedua versi bahasa.
- Data faktual yang berbeda menurut bahasa atau kebutuhan editorial harus diberi struktur yang jelas, bukan ditumpuk pada satu field teks.
- Terjemahan mesin boleh membantu proses penulisan, tetapi harus ditinjau manusia sebelum diterbitkan sesuai LNG-001.
- Pembuatan konten dalam bahasa kedua tidak wajib dilakukan sebelum versi pertama dapat diterbitkan.

## Dampak pada CMS

CMS menampilkan satu entri konten dengan pengelolaan locale/tab bahasa `ID` dan `EN`. Admin dapat melihat status masing-masing bahasa secara terpisah dan mengeditnya tanpa membuat entri konten duplikat.

## Batasan MVP

- Bahasa yang didukung hanya Indonesia dan English.
- Tidak ada Acehnese pada MVP.
- Tidak ada sinkronisasi otomatis isi antarbahasa.
- Detail struktur database, URL, dan fallback locale ditetapkan pada tahap Feature Specification dan Technical Specification.

---

## Materi sumber: `multilingual-publication.md`

# Multilingual Publication Workflow

## Decision LNG-002 --- Independent Publication Status per Language

**Status: Agreed**

Each language version of a content item has its own publication status.
Indonesian content may be published while its English translation
remains a draft or is still being reviewed. The reverse is also
technically possible if an authorized editor has approved that language
version.

## Product Rules

1.  Publication status belongs to a specific language version, not only
    to the parent content item.
2.  Publishing Indonesian (`id`) does not automatically publish English
    (`en`).
3.  Publishing English (`en`) does not automatically publish Indonesian
    (`id`).
4.  Each translation has its own editorial readiness and publication
    timestamps.
5.  Only authorized internal admin/editor accounts may publish a
    language version.
6.  A public page must never show one language's content while labeling
    it as another language.
7.  When the requested language version is not published, show a clear
    unavailable-translation state and provide a link to the available
    Indonesian version, if published.
8.  Search results, category listings, sitemap entries, and SEO metadata
    must respect the publication state of the selected language.
9.  Unpublishing one language version must not automatically unpublish
    another language version.
10. Shared factual data (for example coordinates) may be stored once
    where appropriate, while localized editorial fields remain separate.

## Suggested Data Model Implication

Use a parent content entity with one or more localized
content/translation records. Each localized record should be able to
store: - Locale (`id`, `en`). - Localized title, slug, summary, and
body. - Localized SEO title and description. - Editorial state and
publication state. - Published timestamp. - Last editor/reviewer
metadata.

The exact schema and state machine will be defined in the technical
specification.

## Example

A destination may have: - Indonesian: `published` - English: `draft`

Public behavior: - `/id/destinasi/...` is available. -
`/en/destinations/...` displays a clear translation-unavailable state or
a not-found response according to the final routing/SEO specification. -
The visitor can navigate to the Indonesian version.

## Acceptance Criteria (Initial)

-   An editor can save Indonesian and English versions independently.
-   Publishing one locale does not change the other locale's status.
-   Public pages only render content for the requested locale when that
    locale is published.
-   Locale-specific search and sitemap do not expose unpublished
    translations.
-   Admin UI clearly shows translation status per locale.

## Related Open Decisions

-   Should a missing translation return a 404, a dedicated unavailable
    page, or redirect to Indonesian?
-   Can English be published before Indonesian?
-   Should the admin CMS interface itself be available in English, or
    Indonesian only?


## LNG-005 — Perilaku Language Switcher jika Terjemahan Belum Tersedia

**Status: Disepakati**

Jika pengunjung sedang membaca halaman dalam satu bahasa lalu memilih bahasa lain yang versi terjemahannya belum berstatus Published:

- Pengunjung tetap berada pada konteks halaman/konten yang sama.
- Sistem tidak menampilkan konten bahasa asal seolah-olah merupakan terjemahan.
- Sistem menampilkan pemberitahuan yang jelas bahwa versi bahasa yang dipilih belum tersedia.
- Pemberitahuan menyediakan pilihan untuk kembali ke versi bahasa yang tersedia (misalnya Bahasa Indonesia).
- Sistem tidak mengarahkan pengunjung secara otomatis ke homepage dan tidak menampilkan 404 hanya karena terjemahan belum diterbitkan.
- Pilihan bahasa aktif dan status halaman harus tetap konsisten; UI tidak boleh memberi kesan bahwa konten sudah tampil dalam bahasa yang belum tersedia.

Contoh pesan:
“Versi bahasa Inggris untuk halaman ini belum tersedia. Anda dapat membaca versi Bahasa Indonesia.”

Teks pemberitahuan harus dilokalkan sesuai bahasa antarmuka yang sedang digunakan. Detail apakah pemberitahuan tampil sebagai halaman ringan, inline notice, atau dialog ditentukan pada tahap UX.

---

## Materi sumber: `multilingual-seo-strategy.md`

# SEO-001 — Strategi SEO Multibahasa

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

PidieKab menerapkan SEO terpisah untuk setiap bahasa yang didukung, yaitu Bahasa Indonesia (`id`) dan English (`en`).

## Aturan utama

1. Setiap versi bahasa memiliki URL/slug sendiri.
2. Setiap versi bahasa memiliki SEO Title dan Meta Description sendiri.
3. Halaman yang memiliki padanan terjemahan Published saling dihubungkan menggunakan anotasi `hreflang` yang sesuai.
4. Setiap halaman menggunakan canonical URL yang menunjuk ke URL versi bahasa itu sendiri, bukan canonical lintas bahasa.
5. Sitemap hanya memuat URL halaman yang telah Published pada locale terkait dan memenuhi syarat indeksasi.
6. Halaman Draft, Review, atau Archived tidak dimasukkan ke sitemap dan tidak boleh tersedia sebagai konten publik yang dapat diindeks.
7. Halaman bahasa yang belum diterbitkan tidak boleh dibuat seolah-olah tersedia hanya untuk memenuhi kebutuhan SEO.
8. Metadata, judul, deskripsi, heading, dan isi halaman harus sesuai dengan bahasa halaman tersebut.

## Hubungan antarbahasa

- `hreflang` hanya menghubungkan versi bahasa yang benar-benar tersedia dan Published.
- Jangan membuat tautan `hreflang` ke versi bahasa yang belum diterbitkan.
- Jika versi bahasa lain belum tersedia, halaman tetap dapat diindeks pada bahasa yang tersedia tanpa pasangan `hreflang` yang belum ada.
- Language switcher tetap mengikuti LNG-005 ketika versi bahasa tujuan belum tersedia.

## URL dan slug

- Prefix URL `/id/` dan `/en/` merupakan kandidat yang direkomendasikan, tetapi format final masih menunggu keputusan URL strategy.
- Slug dapat berbeda per bahasa dan harus unik pada locale masing-masing.
- Perubahan slug halaman Published harus ditangani dengan redirect permanen yang tepat; detail teknis ditentukan pada Feature/Technical Specification.

## Kualitas indeksasi

- Hanya konten Published yang memenuhi kualitas editorial dan memiliki informasi cukup yang dapat diindeks.
- Metadata sosial (Open Graph), bila digunakan, harus sesuai bahasa.
- Metadata terjemahan harus ditinjau manusia.
- Jangan menerapkan canonical ID ke EN atau EN ke ID karena keduanya merupakan versi bahasa yang berbeda.
- Structured data, bila digunakan, harus konsisten dengan isi dan bahasa halaman.

## Batasan MVP

- SEO Bahasa Indonesia dan English.
- SEO Title dan Meta Description per versi bahasa.
- Canonical dan sitemap per bahasa.
- `hreflang` untuk pasangan halaman yang sama-sama Published.
- Detail robots, sitemap index, redirect, Open Graph, dan structured data ditetapkan pada Technical Specification.

---

## Materi sumber: `root-domain-language-behavior.md`

# URL-002 — Perilaku Root Domain tanpa Prefix Bahasa

**Status: Disepakati**

Ketika pengunjung membuka root domain PidieKab (`/`) tanpa prefix bahasa, sistem mengarahkan pengunjung ke halaman beranda Bahasa Indonesia (`/id/`).

## Ketentuan

- Bahasa Indonesia (`id`) menjadi bahasa default.
- Root domain melakukan redirect permanen atau sementara sesuai keputusan implementasi SEO/teknis; jenis HTTP redirect final ditetapkan pada Technical Specification.
- Pengunjung tetap dapat mengganti bahasa melalui toggle ID/EN di header setelah berada di website.
- Preferensi bahasa browser tidak mengubah perilaku default root domain pada MVP.
- URL yang sudah memiliki prefix bahasa (`/id/` atau `/en/`) harus mempertahankan bahasa yang dinyatakan oleh URL.
- Redirect root tidak boleh menyebabkan loop atau menghapus path/query pada URL non-root tanpa aturan yang jelas.

## Batasan MVP

- Tidak ada halaman pemilihan bahasa sebelum homepage.
- Tidak ada auto-detection bahasa browser untuk mengalihkan root domain.
- Tidak menyimpan atau menggunakan preferensi bahasa pengunjung untuk menentukan bahasa root pada MVP.

---

## Materi sumber: `public-url-structure.md`

# URL-001 — Struktur URL dengan Prefix Bahasa

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

Semua halaman publik PidieKab menggunakan prefix bahasa `/id/` atau `/en/`. Bahasa Indonesia (`id`) adalah bahasa default produk; cara penanganan permintaan ke root `/` belum diputuskan.

## Contoh pola URL (path usulan, belum dikunci per kategori)

| Halaman | Indonesia | English |
|---|---|---|
| Beranda | `/id/` | `/en/` |
| Wisata | `/id/wisata` | `/en/tourism` |
| Kuliner & UMKM | `/id/kuliner-umkm` | `/en/culinary-msme` |
| Budaya & Sejarah | `/id/budaya-sejarah` | `/en/culture-history` |
| Detail wisata | `/id/wisata/nama-destinasi` | `/en/tourism/destination-name` |

Contoh tidak menetapkan nama domain atau slug aktual konten.

## Ketentuan

- Bahasa halaman ditentukan secara eksplisit dari prefix URL.
- Slug halaman detail dapat berbeda per bahasa dan harus unik sesuai ruang lingkup routing yang ditentukan kemudian.
- Hanya versi konten Published pada locale tersebut yang tersedia sebagai halaman konten publik.
- Canonical, sitemap, dan `hreflang` mengikuti SEO-001; `hreflang` tidak mengarah ke versi yang belum terbit.
- Perubahan URL/slug setelah publikasi memerlukan redirect permanen dari URL lama ke URL baru yang benar, tanpa rantai redirect bila memungkinkan.
- Jika pengunjung berpindah bahasa saat terjemahan belum tersedia, pengalaman tetap mempertahankan konteks konten dan memberi pemberitahuan sesuai LNG-005, tanpa mengindeks halaman terjemahan semu.

## Belum diputuskan

- Perilaku root `/`: redirect ke `/id/` atau pemilihan bahasa lain.
- Nama path kategori final dalam setiap bahasa.
- Aturan pembuatan slug, penanganan konflik, dan riwayat perubahan slug.
- Format URL untuk halaman pemberitahuan terjemahan yang belum tersedia.

---

## Materi sumber: `url-slug-rules.md`

# URL-003 — Aturan Slug URL

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

Slug URL publik PidieKab menggunakan format yang sederhana, konsisten, mudah dibaca, dan ramah SEO.

## Aturan slug

1. Menggunakan huruf kecil (`lowercase`).
2. Menggunakan tanda hubung (`-`) sebagai pemisah kata.
3. Tidak menggunakan spasi.
4. Tidak menggunakan karakter khusus atau simbol.
5. Aksen dan karakter yang tidak didukung dinormalisasi ke bentuk URL yang aman.
6. Slug harus unik pada ruang lingkup routing yang ditetapkan untuk bahasa dan kategori terkait.
7. Slug Bahasa Indonesia dan English boleh berbeda.
8. Admin dapat menentukan atau mengubah slug per versi bahasa.
9. Perubahan slug pada konten yang sudah dipublikasikan harus membuat permanent redirect dari URL lama ke URL baru.
10. Redirect tidak boleh membentuk rantai atau loop; target harus merupakan URL kanonis terbaru yang valid.

## Contoh

| Bahasa Indonesia | English |
|---|---|
| `/id/wisata/pantai-pasir-putih` | `/en/tourism/white-sand-beach` |
| `/id/kuliner/kuah-beulangong` | `/en/culinary/kuah-beulangong` |
| `/id/budaya/tari-tradisional` | `/en/culture/traditional-dance` |

Contoh tersebut bersifat ilustratif. Slug final ditentukan berdasarkan nama dan istilah yang benar-benar relevan dengan konten.

## Prinsip editorial

- Slug tidak harus diterjemahkan secara harfiah.
- Prioritaskan nama yang alami, ringkas, mudah dibaca, dan sesuai dengan istilah yang digunakan pengunjung.
- Nama lokal, nama tempat, dan istilah budaya dapat dipertahankan jika memang merupakan nama yang dikenal.
- Hindari perubahan slug tanpa alasan yang jelas setelah URL dipublikasikan dan telah dibagikan atau diindeks.

## Detail yang ditetapkan pada tahap teknis

- Algoritma normalisasi dan transliterasi.
- Batas panjang slug.
- Ruang lingkup keunikan slug dan penanganan benturan.
- Penyimpanan riwayat slug dan tabel redirect.
- Jenis HTTP redirect permanen yang digunakan.

---

## Materi sumber: `content-detail-url-structure.md`

# URL-004 — Struktur URL Detail Konten Berbasis Kategori

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

## Keputusan
URL detail konten publik menggunakan struktur kategori:

`/{locale}/{category-slug}/{content-slug}`

Contoh:
- Indonesia: `/id/wisata-destinasi/pantai-mantak-tari`
- Inggris: `/en/tourism-destinations/mantak-tari-beach`

Kategori dan konten memiliki slug terpisah untuk setiap locale. Struktur URL mengikuti bahasa yang sedang dibuka.

## Aturan
- Prefix locale menggunakan `/id/` atau `/en/`.
- Segmen kategori menggunakan slug kategori pada locale tersebut.
- Segmen konten menggunakan slug versi bahasa konten yang sedang dibuka.
- Slug ID dan EN dapat berbeda.
- Satu konten tetap memiliki satu kategori utama sesuai CNT-001.
- URL detail hanya tersedia jika versi bahasa konten terkait berstatus Published dan kategori terkait Aktif.
- Draft dan Archived tidak tersedia pada website publik.
- URL kategori dan konten harus tervalidasi sebagai pasangan yang benar; slug konten tidak boleh ditampilkan di bawah kategori yang salah.
- Slug kategori dan konten harus unik dalam ruang URL masing-masing locale dan konteks routing yang ditetapkan pada Technical Specification.

## Perubahan kategori atau slug
- Perubahan slug konten yang sudah dipublikasikan harus mempertahankan URL lama melalui HTTP 301 ke URL baru.
- Perubahan slug kategori dapat mengubah segmen kategori pada URL seluruh detail konten yang berada di kategori tersebut. Sistem harus mempertahankan URL lama untuk halaman kategori dan seluruh halaman detail terdampak melalui redirect yang benar.
- Jika konten dipindahkan ke kategori lain, URL detail lama harus diarahkan dengan HTTP 301 ke URL detail baru pada locale terkait.
- Perubahan pada satu locale tidak boleh mengubah URL locale lainnya.
- Redirect harus langsung menuju URL final, menghindari loop dan rantai redirect yang tidak perlu.
- URL lama tidak boleh digunakan ulang selama masih menjadi sumber redirect aktif.
- Perubahan yang memengaruhi URL publik harus dicatat dalam audit log.

## Multilingual dan SEO
- Setiap locale memiliki canonical self-referencing.
- `hreflang` hanya menghubungkan versi bahasa yang benar-benar Published dan dapat diakses.
- Sitemap hanya memuat URL detail yang Published, berada dalam kategori Aktif, dan dapat diindeks.
- URL tidak boleh menyajikan konten ID dengan label atau metadata EN, atau sebaliknya.
- Jika terjemahan locale yang diminta tidak Published, ikuti LNG-005: tampilkan pemberitahuan bahwa terjemahan belum tersedia dan tautan ke bahasa yang tersedia; jangan diam-diam menampilkan konten bahasa lain.

## Batasan MVP
- Tidak ada URL detail tanpa segmen kategori.
- Tidak ada multi-category.
- Tidak ada alias URL publik tambahan.
- Tidak ada pengelolaan redirect manual umum oleh Admin.
- Redirect otomatis dibuat oleh sistem saat perubahan slug atau kategori menyebabkan URL detail berubah.

## Catatan teknis
Technical Specification harus menetapkan:
- Route resolution dan validasi relasi category/content.
- Penyimpanan slug per locale.
- Strategi redirect historis untuk perubahan slug kategori, slug konten, dan pemindahan kategori.
- Penanganan beberapa perubahan berturut-turut agar tidak membentuk redirect chain.
- Unique constraints dan transaksi perubahan kategori/slug.
- Dampak terhadap cache, pencarian, sitemap, canonical, dan `hreflang`.
- Pengujian HTTP 301, locale, URL lama, dan akses ke konten non-Published.

---

## Materi sumber: `category-slug-and-redirects.md`

# CAT-004 — Slug Kategori dan Permanent Redirect

**Status: Disepakati; dilengkapi CAT-005.**

Setiap kategori memiliki slug yang dapat diatur Admin. Slug menggunakan huruf kecil,
kata dipisahkan tanda hubung, URL-friendly, dan unik dalam ruang URL kategori pada
locale terkait.

Perubahan slug kategori yang sudah dipublikasikan harus menyimpan pemetaan URL lama
ke URL baru dan menggunakan HTTP 301 Permanent Redirect. Redirect tidak boleh
membentuk loop atau rantai yang tidak perlu. Slug lama tidak boleh digunakan ulang
selama masih menjadi target redirect.

**Penyempurnaan CAT-005:** Slug dikelola secara terpisah untuk Bahasa Indonesia (`id`)
dan Bahasa Inggris (`en`). Perubahan slug satu locale tidak memengaruhi locale lain.
Lihat `category-localized-slugs.md` untuk aturan lengkap.

---
