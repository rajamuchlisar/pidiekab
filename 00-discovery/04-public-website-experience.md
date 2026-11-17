# 04 Public Website Experience

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `public-information-architecture.md`
- `public-header-navigation.md`
- `footer-information-architecture.md`
- `homepage-section-order.md`
- `homepage-hero-strategy.md`
- `homepage-hero-image.md`
- `homepage-hero-search-placement.md`
- `homepage-category-section.md`
- `homepage-content-sections.md`
- `pidie-editorial-storytelling.md`
- `public-interaction-model.md`
- `content-detail-page-strategy.md`
- `error-page-handling.md`

---

## Materi sumber: `public-information-architecture.md`

# Public Website Information Architecture

## Decision IA-001 --- Add a Dedicated "Jelajahi Pidie" Menu

**Status: Agreed**

The public website navigation will include a dedicated top-level menu
named **Jelajahi Pidie**.

The menu is intended to be the main discovery entry point for visitors,
especially people who are not familiar with Kabupaten Pidie. It brings
together the MVP's three content categories: - Wisata & Destinasi. -
Kuliner & UMKM. - Budaya & Sejarah.

## Proposed MVP Navigation

1.  **Beranda** --- introduction to PidieKab and selected highlights.
2.  **Jelajahi Pidie** --- discovery landing page bringing together
    destinations, culinary/UMKM, and culture/history.
3.  **Wisata** --- browse tourism destinations.
4.  **Kuliner & UMKM** --- browse culinary places and local businesses.
5.  **Budaya & Sejarah** --- browse cultural and historical content.
6.  **Tentang PidieKab** --- platform purpose, editorial approach, and
    independence disclaimer.

The exact relationship between "Jelajahi Pidie" and category pages (for
example, dropdown navigation versus a landing page with category links)
remains a UX design detail to resolve.

## Jelajahi Pidie --- MVP Intent

The landing page should help visitors discover Pidie through curated
content and clear paths into the three MVP categories. It may include: -
A short introduction to exploring Pidie. - Featured or editor-curated
content. - Entry points to Wisata, Kuliner & UMKM, and Budaya &
Sejarah. - A search entry point if search is confirmed in the MVP
requirements.

It must not imply that all locations, businesses, or attractions in
Pidie are listed comprehensively.

## Scope Boundaries

-   No combined interactive map in MVP; map/location display remains on
    relevant detail pages under MAP-001.
-   No route planner or turn-by-turn navigation.
-   Phase 2 modules such as Akomodasi & Fasilitas, Panduan Wisata, and
    Berita & Agenda Kegiatan are not part of MVP implementation.
-   The platform remains independent and must not imply official
    government affiliation.

## Open Decisions

-   Should "Jelajahi Pidie" be a curated landing page only, or also the
    parent navigation item for the three category pages?
-   Should public search be included in MVP, and if so, which content
    fields and categories should it search?
-   Should the homepage and Jelajahi Pidie page have different content
    layouts and purposes?

---

## Materi sumber: `public-header-navigation.md`

# NAV-001 — Header dan Navigasi Publik

**Status: Disepakati**

PidieKab menggunakan header sederhana sebagai struktur dasar, dilengkapi navigasi mobile yang responsif.

## Struktur dasar

- Logo PidieKab berada di sisi kiri.
- Menu navigasi utama berada di sisi kanan/area tengah sesuai breakpoint dan ruang yang tersedia.
- Pilihan bahasa Indonesia (`id`) dan Inggris (`en`) tersedia pada header.
- Tidak ada kolom pencarian di header.
- Global Search tetap ditempatkan di dalam Hero homepage sesuai HOME-003.

## Navigasi mobile

- Pada layar kecil, navigasi desktop berubah menjadi menu mobile responsif, misalnya hamburger menu.
- Menu harus mudah dibuka, ditutup, dan digunakan dengan sentuhan maupun keyboard.
- Status menu terbuka/tertutup harus jelas bagi teknologi bantu.
- Navigasi tidak boleh menyebabkan horizontal scrolling.
- Tautan dan pilihan bahasa tetap dapat diakses pada perangkat mobile.

## Menu utama

Mengikuti IA-001:
- Beranda
- Jelajahi Pidie
- Wisata
- Kuliner & UMKM
- Budaya & Sejarah
- Tentang PidieKab

Relasi visual antara “Jelajahi Pidie” dan kategori masih dapat ditentukan pada tahap UX, tanpa mengubah kategori MVP yang sudah disepakati.

## Prinsip

- Header konsisten di seluruh halaman publik.
- Navigasi dan label mengikuti bahasa aktif.
- Tidak menampilkan menu fase 2 sebagai fitur aktif sebelum dirilis.
- Tidak menduplikasi global search pada header.
- Detail sticky/non-sticky header, ukuran, spacing, dan interaksi dropdown (jika diperlukan) belum ditetapkan.


## NAV-002 — Perilaku Header saat Scroll

**Status: Disepakati**

Header publik menggunakan pola **Sticky Header**.

- Header tetap terlihat di bagian atas viewport ketika pengunjung menggulir halaman.
- Tinggi header dibuat ringkas agar tidak mengambil terlalu banyak ruang konten.
- Header harus tetap terbaca dan tidak menutupi judul atau elemen penting halaman.
- Perilaku harus konsisten pada desktop, tablet, dan mobile.
- Pada mobile, sticky header tetap mendukung hamburger menu dan pilihan bahasa.
- Detail tinggi, shadow/border, latar, dan perilaku saat menu mobile terbuka ditetapkan pada tahap UX/visual specification.


## LNG-003 — Language Switcher pada Header

**Status: Disepakati**

Pilihan bahasa menggunakan **toggle ID / EN** yang terlihat langsung pada header.

- Kedua pilihan bahasa ditampilkan secara jelas tanpa perlu membuka dropdown.
- Bahasa aktif memiliki penanda visual yang mudah dikenali.
- Switcher tersedia pada desktop dan mobile.
- Perubahan bahasa mengikuti kebijakan LNG-002: status publikasi setiap bahasa berdiri sendiri.
- Jika halaman padanan pada bahasa tujuan belum dipublikasikan, jangan menampilkan konten bahasa asal seolah-olah sudah diterjemahkan. Tampilkan pengalaman bahasa tidak tersedia sesuai aturan fallback yang akan ditetapkan.
- Perpindahan bahasa sebaiknya menuju halaman padanan yang sama jika tersedia; jika tidak, gunakan perilaku fallback yang konsisten.
- Detail tampilan toggle dan perilaku fallback URL masih perlu ditetapkan pada UX/technical specification.

---

## Materi sumber: `footer-information-architecture.md`

# FOOT-001 — Struktur Informasi Footer Website

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

Footer website PidieKab menggunakan empat kelompok informasi:

1. **Tentang PidieKab**
   - Penjelasan singkat identitas dan tujuan platform.
2. **Jelajahi Pidie**
   - Tautan menuju Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.
3. **Informasi**
   - Tentang Kami.
   - Kebijakan Privasi.
   - Disclaimer.
4. **Kontak & Media Sosial**
   - Alamat email kontak PidieKab jika tersedia.
   - Tautan akun media sosial resmi PidieKab jika sudah ditetapkan.

## Pernyataan independensi

Footer harus menjelaskan secara jelas bahwa PidieKab adalah platform independen dan bukan situs resmi, bagian, atau perwakilan Pemerintah Kabupaten Pidie maupun instansi pemerintah lainnya.

Informasi resmi yang dikutip atau dirujuk harus tetap menyebutkan sumber dan, jika relevan, menyediakan tautan ke instansi yang bertanggung jawab. Identitas visual, logo, dan bahasa footer tidak boleh menimbulkan kesan afiliasi pemerintah.

## Prinsip implementasi

- Footer tersedia pada halaman publik dan konsisten di seluruh halaman.
- Tautan kategori mengikuti bahasa aktif (`id` atau `en`).
- Kontak dan media sosial hanya ditampilkan jika alamat/akun telah ditetapkan dan diverifikasi.
- Jangan menampilkan alamat, nomor telepon, atau akun yang belum dikonfirmasi.
- Kebijakan Privasi dan Disclaimer harus tersedia sebelum peluncuran publik.
- Detail susunan visual, copyright notice, dan konten legal final akan dirinci pada tahap UX/Feature Specification dan legal/content review.

---

## Materi sumber: `homepage-section-order.md`

# HOME-010 — Urutan Section Homepage

**Status: Disepakati berdasarkan arahan untuk melanjutkan rancangan.**

Urutan section homepage PidieKab:

1. **Hero**
   - Satu gambar statis.
   - Headline singkat.
   - Global Search berada di dalam hero.
   - Tinggi hero proporsional, tidak memenuhi seluruh layar.
2. **Kategori**
   - Tiga kartu visual bergambar untuk Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.
3. **Jelajahi Pidie**
   - Satu section gabungan yang memuat Konten Unggulan, Konten Terbaru, dan Konten Acak.
   - Menggunakan Grid Card.
   - Kartu menampilkan foto utama, judul, kategori, dan lokasi jika tersedia serta relevan.
4. **Cerita tentang Pidie**
   - Area editorial untuk memperkenalkan karakter, kekayaan, dan potensi Kabupaten Pidie.

## Prinsip

- Urutan ini menjadi struktur awal homepage MVP.
- Konten fase 2 tidak dimasukkan sebagai section aktif pada MVP.
- Homepage harus responsif dan menjaga agar konten penting terlihat tanpa hero yang terlalu tinggi.
- Detail jumlah item, dimensi, tipografi, jarak, dan komponen visual akan ditentukan pada tahap UX/Feature Specification.
- “Cerita tentang Pidie” belum memiliki aturan editorial dan jumlah konten final; hal tersebut masih perlu didalami.

---

## Materi sumber: `homepage-hero-strategy.md`

# Homepage Hero Strategy

## Decision HOME-001 --- Combined Emotional Hero and Search

**Status: Agreed**

The PidieKab homepage hero will combine: - A prominent, inspiring
destination-led visual and headline. - A clearly visible global search
entry point.

The hero should create an emotional first impression while also allowing
visitors with a specific information need to search immediately.

## Experience Requirements

1.  Use a strong, locally relevant photograph or visual representing
    Pidie.
2.  Present a concise headline that communicates discovery and the
    region's potential.
3.  Include a short supporting statement where useful.
4.  Provide a prominent global search input that searches the three MVP
    content categories.
5.  Preserve the approved global-search behaviour: category filters,
    relevance ordering, pagination, and compact result entries
    containing title, category, and location only.
6.  Ensure the hero remains usable on mobile, with search easy to find
    and operate.
7.  Use image assets only when publication rights and attribution
    requirements are satisfied.
8.  Avoid unsupported claims or imagery that misrepresents Pidie.

## Proposed Content Direction

Possible headline direction: - "Discover the Beauty, Culture & Potential
of Pidie." - Supporting message: "Explore destinations, local flavors,
culture, and stories from Pidie."

These are copy directions for later content/design work, not final
approved marketing copy.

## Design Notes

-   The hero may use a full-bleed image or a large editorial image area.
-   Text contrast and search control visibility must be maintained
    across image variations.
-   Avoid heavy animation and ensure responsive image loading.
-   The hero should not dominate the page so much that visitors cannot
    discover the next sections.

## Related Decisions

See `homepage-hero-image.md` for the approved image behaviour (HOME-002) and `homepage-hero-search-placement.md` for search placement and hero proportion (HOME-003).

## Open Decisions

-   What content should appear directly after the hero?
-   What content should appear directly after the hero?

---

## Materi sumber: `homepage-hero-image.md`

# Homepage Hero Image Behaviour

## Decision HOME-002 --- Single Static Image

**Status: Agreed**

The PidieKab homepage hero will use one static image. It will not use an
automatically rotating image carousel or a video background.

## Rationale

1. Keeps the homepage lightweight and supports faster initial loading.
2. Provides a stable visual focus instead of moving content.
3. Keeps attention on the hero message and global search.
4. Reduces implementation and content-maintenance complexity for MVP.
5. Supports a predictable, accessible mobile experience.

## Requirements

- Use one locally relevant, representative image that genuinely reflects
  Kabupaten Pidie.
- Confirm usage rights, license, permission, and attribution requirements
  before publication.
- Optimize image dimensions, format, and file size for responsive delivery.
- Provide an appropriate crop or focal position for desktop and mobile.
- Maintain sufficient contrast and legibility for the headline and search
  interface over the image.
- Avoid autoplay, automatic image transitions, and video backgrounds.
- Do not use imagery that implies a destination, facility, event, or
  experience exists in Pidie unless the image and its context are verified.

## Out of Scope for MVP

- Hero image carousel or slideshow.
- Video background.
- Automatic seasonal or campaign-based hero rotation.

## Related Decisions

- `homepage-hero-strategy.md` (HOME-001): emotional hero combined with
  global search.
- `visual-direction-decision.md` (VIS-001): Discover Tasmania as a
  visual and visitor-experience reference, adapted to PidieKab.
- `homepage-hero-search-placement.md` (HOME-003): global search is
  embedded in a moderate-height hero.

## Remaining Homepage Questions

- What content should appear directly after the hero?
- What content should appear directly after the hero?

---

## Materi sumber: `homepage-hero-search-placement.md`

# Homepage Hero Search Placement and Proportion

## Decision HOME-003 --- Search Embedded in a Compact Hero

**Status: Agreed**

The global search input will be placed directly inside the homepage hero. The hero must remain proportionate and moderate in height; it must not become a full-screen or oversized section.

## Experience Requirements

1. Keep the single static image approved in HOME-002.
2. Place the headline and global search input within the hero area.
3. Keep the hero visually strong but compact enough that visitors can see the beginning of the next homepage section without excessive scrolling.
4. Do not design the hero as a full-viewport or full-screen banner.
5. Set responsive height, spacing, text sizing, and image crop independently where needed for desktop and mobile.
6. Ensure the search input remains prominent, legible, and easy to operate on small screens.
7. Preserve image optimization, accessibility, and image-rights requirements from HOME-001 and HOME-002.

## Design Direction

Use a concise destination-led headline, optional short supporting text, and a clearly visible search field. The visual should introduce PidieKab without pushing the actual discovery content too far below the fold. Exact dimensions, breakpoints, copy, and image crop will be established during UX/UI design and implementation planning.

## Rationale

- Visitors can search immediately without leaving the hero.
- A moderate hero height balances emotional impact and practical content discovery.
- The next section remains discoverable early, particularly on mobile.

## Related Decisions

- `homepage-hero-strategy.md` (HOME-001): combined emotional hero and global search.
- `homepage-hero-image.md` (HOME-002): single static image.
- `public-global-search.md` (SRCH-001): global search across the three MVP categories.

## Open Decision

- What content and section should appear immediately after the homepage hero?

---

## Materi sumber: `homepage-category-section.md`

# HOME-009 — Tampilan Kategori pada Homepage

**Status: Disepakati**

Tiga kategori MVP ditampilkan sebagai **tiga kartu visual bergambar** pada homepage:

1. **Wisata & Destinasi**
2. **Kuliner & UMKM**
3. **Budaya & Sejarah**

## Prinsip tampilan

- Setiap kategori memiliki gambar representatif, nama kategori, dan tautan menuju halaman kategori.
- Gambar harus relevan dengan kategori dan memiliki hak penggunaan yang jelas.
- Tampilan mengikuti identitas visual PidieKab dan arah destination-led yang telah disepakati.
- Layout responsif: kartu tersusun proporsional di desktop, tablet, dan mobile.
- Teks kategori harus terbaca jelas di atas gambar atau pada area kartu.
- Jangan menggunakan gambar generik yang berpotensi memberi kesan keliru tentang Pidie.

## Batasan MVP

- Hanya tiga kategori yang telah disepakati untuk MVP.
- Kategori fase berikutnya tidak ditampilkan sebagai kategori aktif sebelum dirilis.
- Jumlah kartu kategori tetap tiga; bukan daftar konten destinasi.
- Detail ukuran, rasio gambar, dan perlakuan hover/focus ditentukan pada tahap UX/feature specification.

---

## Materi sumber: `homepage-content-sections.md`

# HOME-004 & HOME-005 — Struktur Konten Homepage Setelah Hero

## Status

**Disepakati**

## Keputusan

Bagian homepage setelah hero menggunakan kombinasi kategori utama dan konten unggulan. Struktur ini memberi pengunjung dua jalur: menjelajahi Pidie berdasarkan kategori atau langsung membuka konten pilihan yang dikurasi tim PidieKab.

## Urutan bagian

1. **Jelajahi Kategori**
   - Wisata & Destinasi
   - Kuliner & UMKM
   - Budaya & Sejarah
   - Setiap kategori menjadi pintu masuk ke halaman kategori terkait.

2. **Pilihan untuk Dijelajahi**
   - Menampilkan sejumlah konten pilihan yang dikurasi Admin.
   - Konten harus berasal dari kategori MVP dan sudah dipublikasikan dalam bahasa aktif.
   - Penetapan jumlah item, aturan kurasi, dan bentuk kartu visual ditentukan pada tahap UX/UI atau feature specification.

3. **Cerita tentang Pidie**
   - Bagian editorial untuk memperkenalkan kekayaan, karakter, dan potensi Kabupaten Pidie.
   - Konten tetap mengikuti kebijakan sumber, verifikasi, hak penggunaan media, dan publikasi multibahasa.
   - Bentuk dan jumlah artikel/cerita yang ditampilkan masih perlu dirinci.

## Prinsip tampilan

- Bagian setelah hero harus mulai terlihat relatif cepat karena ukuran hero dibatasi dan tidak full-screen (HOME-003).
- Hierarki visual harus jelas; kategori menjadi navigasi eksplorasi, konten pilihan menjadi rekomendasi editorial, dan cerita menjadi konteks yang lebih mendalam.
- Tampilan responsif, ringan, mudah dipindai, dan konsisten dengan arah visual Discover Tasmania tanpa menyalin desainnya.
- Jangan menampilkan konten yang belum dipublikasikan dalam bahasa yang sedang dipilih.
- Hindari membuat bagian homepage terlalu panjang atau mengulang seluruh daftar konten yang tersedia.
- Phase 2 seperti akomodasi, panduan wisata, berita, dan agenda tidak ditampilkan sebagai fitur aktif MVP.

## HOME-005 — Strategi Konten Homepage

**Status: Disepakati**

Bagian konten rekomendasi homepage menggunakan tiga sumber tampilan:

1. **Konten Unggulan**
   - Dipilih dan diatur secara manual oleh Admin.
   - Digunakan untuk menonjolkan konten yang dianggap penting secara editorial atau strategis.
2. **Konten Terbaru**
   - Menampilkan konten yang paling baru dipublikasikan.
   - Urutan berdasarkan waktu publikasi yang valid pada bahasa aktif.
3. **Konten Acak**
   - Menampilkan variasi konten yang dipilih secara acak dari konten MVP yang memenuhi syarat publikasi.
   - Tujuannya memberi kesempatan penemuan konten yang lebih beragam.

### Aturan konsistensi

- Ketiga kelompok hanya mengambil konten yang sudah dipublikasikan dalam bahasa aktif.
- Konten yang belum tersedia dalam bahasa aktif tidak boleh ditampilkan seolah-olah sudah diterjemahkan.
- Konten acak tidak boleh menduplikasi item yang sudah tampil pada bagian unggulan atau terbaru dalam satu halaman homepage.
- Semua item harus berasal dari kategori MVP: Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.
- Konten yang tidak lagi layak tayang atau membutuhkan pembaruan harus mengikuti status editorial dan kebijakan publikasi.
- Mekanisme pengacakan, frekuensi rotasi, jumlah item, dan strategi cache ditentukan pada tahap technical/feature specification. Pengacakan tidak harus berubah pada setiap refresh apabila hal itu berdampak buruk pada performa atau pengalaman pengguna.

## Hal yang belum diputuskan

- Jumlah konten pada bagian Pilihan untuk Dijelajahi.
- Apakah bagian Cerita tentang Pidie menggunakan artikel editorial tersendiri atau kurasi dari konten yang sudah ada.
- Apakah seluruh bagian ditampilkan pada homepage sejak MVP atau sebagian dapat disembunyikan jika kontennya belum memadai.
- Detail layout, kartu, gambar, dan perilaku pada mobile.

## Dampak terhadap scope

Keputusan ini mengatur susunan homepage, bukan menambah kategori atau fitur baru. Semua konten tetap dikelola Admin internal. Tidak ada submission publik atau interaksi pengunjung baru yang diperkenalkan.


## HOME-006 — Penyajian Konten Homepage dalam Satu Section

**Status: Disepakati**

Konten Unggulan, Konten Terbaru, dan Konten Acak ditampilkan dalam **satu section gabungan**, bukan sebagai tiga section terpisah.

- Section menggunakan satu judul utama, misalnya **“Jelajahi Pidie”**.
- Ketiga sumber konten dapat berkontribusi dalam satu area tampilan.
- Pengunjung tidak perlu memahami label atau mekanisme internal pemilihan konten.
- Konten Unggulan tetap mendapat prioritas editorial, sementara konten terbaru dan acak menambah variasi penemuan.
- Aturan bahasa aktif, status publikasi, kategori MVP, dan pencegahan duplikasi dari HOME-005 tetap berlaku.

**Belum diputuskan:** pola visual kartu, urutan campuran yang tepat, jumlah item yang ditampilkan, serta apakah ada label kecil untuk membedakan konten unggulan. Detail tersebut dibahas pada tahap UX/feature specification.


## HOME-007 — Pola Tampilan Konten Jelajahi Pidie

**Status: Disepakati**

Konten dalam section gabungan **“Jelajahi Pidie”** menggunakan pola **Grid Card**.

Prinsip tampilan:
- Desktop menampilkan beberapa kartu dalam satu baris sesuai lebar layar.
- Tablet dan mobile menggunakan grid responsif dengan jumlah kolom yang menyesuaikan ruang.
- Kartu memiliki hierarki visual yang konsisten dan area klik yang jelas menuju halaman detail.
- Grid tidak boleh menyebabkan konten terpotong atau memaksa horizontal scrolling pada layar kecil.
- Jumlah kolom, rasio gambar, metadata yang ditampilkan, dan jumlah item awal akan ditetapkan pada tahap UX/feature specification.
- Aturan sumber konten HOME-005 dan satu section gabungan HOME-006 tetap berlaku.

Carousel tidak digunakan sebagai pola utama untuk section ini.


## HOME-008 — Informasi pada Grid Card

**Status: Disepakati**

Setiap kartu pada section **“Jelajahi Pidie”** menampilkan informasi ringkas berikut:

1. **Foto utama** — gambar representatif yang hak penggunaannya jelas.
2. **Judul konten** — nama destinasi, kuliner/UMKM, atau topik budaya/sejarah.
3. **Kategori** — penanda kategori konten.
4. **Lokasi** — ditampilkan jika tersedia dan relevan.

Deskripsi atau ringkasan tidak ditampilkan pada kartu. Informasi lebih lengkap tersedia pada halaman detail.

Ketentuan:
- Foto tidak boleh dibuat-buat atau menyesatkan.
- Jika foto tidak tersedia atau tidak memiliki hak penggunaan yang jelas, gunakan fallback visual yang konsisten atau jangan tampilkan gambar.
- Lokasi hanya ditampilkan jika datanya tersedia dan cukup dapat dipercaya; jangan mengarang lokasi.
- Kartu mengarah ke halaman detail konten.

---

## Materi sumber: `pidie-editorial-storytelling.md`

# HOME-011 — Format Editorial Storytelling “Cerita tentang Pidie”

**Status: Disepakati**

Section **“Cerita tentang Pidie”** menggunakan format kombinasi:

- **Satu cerita utama** dengan area visual lebih besar, foto utama, judul, dan pengantar singkat.
- **Beberapa cerita pendukung** dengan tampilan lebih ringkas sebagai pelengkap.

## Tujuan editorial

Bagian ini bukan feed berita umum. Fokusnya adalah storytelling yang memperkenalkan karakter, sejarah, budaya, kehidupan, dan potensi Kabupaten Pidie kepada pembaca.

Contoh tema:
- Sejarah dan asal-usul Pidie.
- Tradisi dan kebudayaan masyarakat.
- Potensi ekonomi dan UMKM lokal.
- Keunikan alam dan kehidupan masyarakat.

## Prinsip konten dan tampilan

- Cerita utama dan cerita pendukung dikelola serta dipublikasikan oleh Admin sesuai kebijakan CMS.
- Materi harus memiliki sumber yang dapat ditelusuri dan memperhatikan hak penggunaan gambar.
- Hindari klaim sejarah atau budaya yang tidak memiliki dasar sumber yang memadai.
- Gunakan bahasa Indonesia dan Inggris sesuai kebijakan publikasi per bahasa.
- Tautan mengarah ke halaman detail cerita.
- Jumlah cerita pendukung, aturan pemilihan cerita utama, serta desain responsif final ditentukan pada tahap UX/Feature Specification.

## Batasan MVP

Section ini merupakan bagian editorial homepage, bukan modul berita dan agenda fase 2. Tidak mencakup komentar, rating, atau interaksi publik.

---

## Materi sumber: `public-interaction-model.md`

# Public Visitor Interaction Model

## Decision UX-001 --- Information-Only Public Website

**Status: Agreed**

The PidieKab MVP is an information and promotion website. Public
visitors can browse published content without creating an account. The
MVP will not include public visitor accounts or user-generated ratings,
reviews, comments, or favorites.

## MVP Rules

1.  Public content browsing does not require login.
2.  No public visitor registration or login flow.
3.  No favorites/bookmark feature in MVP.
4.  No public rating, review, comment, or discussion feature in MVP.
5.  No public user profile or social interaction.
6.  Admin authentication remains required for internal content
    management.
7.  Contact details or external links published by admins may be shown
    as informational content; this does not create an in-platform
    messaging feature.

## Product Rationale

The initial product goal is to promote Pidie's potential and provide
useful information to visitors. The first release prioritizes content
quality, discoverability, source transparency, and ease of maintenance
rather than community interaction features.

## Out of Scope for MVP

-   Visitor accounts and authentication.
-   User-generated reviews and ratings.
-   Comments and discussion threads.
-   User favorites/bookmarks.
-   Public user profiles.
-   In-platform chat or direct messaging.

These features may be reconsidered in a later phase if user needs and
moderation capacity justify them.

## Technical Implications

-   No public user identity/authentication subsystem is required for
    MVP.
-   Admin authentication and authorization remain in scope.
-   No rating, review, comment, or favorite tables are required for MVP.
-   Public read endpoints should expose only published content for the
    requested locale.
-   Avoid collecting personal information from visitors unless required
    for a separately approved feature.

## Initial Acceptance Criteria

-   A visitor can browse public pages without registration or login.
-   No public registration/login entry point is shown.
-   No rating, review, comment, or favorite controls are shown.
-   Admin users can still authenticate to manage content.
-   Unpublished content remains inaccessible through public pages and
    APIs.

## Open Decisions

-   Is a general contact form needed, or are published external contact
    links sufficient?
-   Should analytics be used, and what privacy-conscious configuration
    is acceptable?
-   Should a public correction/report-information channel be introduced
    later without user accounts?

---

## Materi sumber: `content-detail-page-strategy.md`

# Content Detail Page Strategy

## Analysis --- Shared Detail Template with Category-Specific Sections

**Status: Proposed for discussion; not yet approved**

PidieKab has three MVP content categories with different information
needs. The recommended approach is to use one shared detail-page
framework while allowing category-specific sections and fields.

This avoids two extremes: - One rigid page layout that forces unrelated
information onto every content type. - Three completely independent page
systems that increase implementation and maintenance complexity.

## Shared Page Structure

The following elements are candidates for a common detail-page
framework:

1.  Breadcrumb/navigation context.
2.  Content title.
3.  Category label.
4.  Main content/media area, when media exists and is permitted for
    publication.
5.  Main editorial content.
6.  Relevant location information, when applicable and available.
7.  Source/attribution information where appropriate.
8.  Related content, if the editorial team has suitable published items.
9.  Last reviewed/updated information, if approved as a public trust
    feature.

These are recommendations, not confirmed requirements.

## Category-Specific Information Candidates

### Wisata & Destinasi

Potential fields/sections: - Destination name. - Destination
type/category. - Description and visitor-relevant overview. - Address or
area. - Coordinates/map preview and external directions link, following
MAP-001. - Visitor information such as operating hours, entrance fee,
facilities, and contact details only when verified and maintained.

### Kuliner & UMKM

Potential fields/sections: - Business or culinary name. - Business
type/category. - Products or specialties. - Description. - Address or
service area. - Contact or external website/social links when verified
and approved for publication. - Operating hours or price range only when
reliable and periodically reviewed.

### Budaya & Sejarah

Potential fields/sections: - Content title. - Cultural/historical topic
or type. - Historical or cultural narrative. - Relevant people, period,
place, or community where substantiated. - Location/map preview only
when applicable and sufficiently verified. - Source references and
attribution.

## Content Integrity and Editorial Rules

-   Do not show empty sections or placeholder values on public pages.
-   Do not invent details to fill a template.
-   Time-sensitive facts such as hours, prices, contact details, and
    access conditions require a review/update process before
    publication.
-   Distinguish exact verified coordinates from approximate or
    unverified locations.
-   Use language-specific published versions according to LNG-002.
-   Preserve source and image-rights metadata in the CMS even if not all
    metadata is displayed publicly.

## Recommendation

Adopt a shared detail-page framework with category-specific data
sections. Keep the common structure small and avoid making optional
fields mandatory across all categories.

## Decisions Still Required

1.  Confirm whether to use the shared framework with category-specific
    sections.
2.  Decide which information fields are mandatory for each category.
3.  Decide whether related content should appear on detail pages in MVP.
4.  Decide which source/verification information should be visible to
    public visitors.
5.  Decide whether a last-reviewed date should be displayed publicly.


## NAV-003 — Breadcrumb Navigation pada Halaman Detail

**Status: Disepakati**

Halaman detail konten menggunakan **Breadcrumb Navigation** sebagai navigasi kontekstual.

Contoh:
`Beranda > Jelajahi Pidie > Wisata > Nama Destinasi`

Ketentuan:
- Breadcrumb menunjukkan posisi halaman dalam struktur informasi website.
- Setiap tingkat yang memiliki halaman tujuan dapat diklik, kecuali item terakhir yang menandai halaman aktif.
- Breadcrumb harus mengikuti kategori dan bahasa aktif.
- Label breadcrumb menggunakan nama kategori/halaman yang sesuai dengan locale aktif.
- Breadcrumb tetap berfungsi ketika pengunjung membuka halaman detail langsung dari mesin pencari atau tautan eksternal.
- Jangan bergantung pada browser history untuk membangun breadcrumb.
- Struktur breadcrumb harus konsisten untuk halaman Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.
- Detail markup, structured data, dan perilaku pada layar mobile ditetapkan pada tahap UX/technical specification.

Tombol kembali khusus tidak menjadi pola navigasi utama halaman detail pada MVP.


## MED-001 — Foto pada Halaman Detail

**Status: Disepakati**

Halaman detail menggunakan **satu foto utama saja** sebagai media visual utama.

Ketentuan:
- Satu foto utama dipilih untuk mewakili konten.
- Tidak ada galeri thumbnail atau carousel foto pada MVP.
- Foto harus relevan dengan konten, memiliki hak penggunaan yang jelas, dan tidak menyesatkan.
- Jika foto tidak tersedia atau tidak layak digunakan, halaman tetap berfungsi dengan fallback visual yang konsisten.
- Foto dioptimalkan untuk performa dan responsivitas, termasuk perangkat mobile.
- Pengelolaan banyak foto dan galeri dapat dievaluasi setelah MVP, berdasarkan kebutuhan nyata.


## NAV-004 — Struktur Informasi Halaman Detail

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

Struktur dasar halaman detail konten:

1. **Breadcrumb** — menunjukkan posisi halaman dalam struktur website.
2. **Judul konten** — nama destinasi, kuliner/UMKM, atau topik budaya/sejarah.
3. **Kategori** — identitas kategori konten.
4. **Foto utama** — satu foto utama sesuai MED-001.
5. **Isi artikel/deskripsi lengkap** — informasi utama yang dikelola editor.
6. **Informasi lokasi dan peta** — hanya jika relevan dan data lokasi memadai sesuai MAP-001.
7. **Sumber informasi dan referensi** — ditampilkan secara proporsional untuk transparansi informasi.

### Prinsip penerapan

- Bagian yang tidak relevan atau tidak memiliki data yang layak tidak ditampilkan sebagai bagian kosong.
- Peta tidak boleh menampilkan koordinat yang dikarang; lokasi perkiraan harus dibedakan dari lokasi terverifikasi.
- Sumber dan referensi harus merujuk pada catatan sumber yang mendukung informasi tersebut.
- Konten mengikuti status publikasi locale aktif sesuai LNG-002.
- Struktur ini merupakan template bersama; detail khusus tiap kategori tetap dapat ditambahkan sesuai strategi halaman detail.
- Klaim sejarah, jam operasional, harga, dan informasi yang mudah berubah harus ditinjau sebelum publikasi.

---

## Materi sumber: `error-page-handling.md`

# ERR-001 — Penanganan Halaman Error

**Status: Disepakati.**

## Keputusan

PidieKab tidak membuat desain atau halaman error kustom pada MVP. Kode HTTP error menggunakan penanganan default umum dari framework, web server, atau mekanisme standar yang digunakan aplikasi.

Keputusan ini berlaku untuk error seperti:

- `400` — Bad Request
- `401` — Unauthorized
- `403` — Forbidden
- `404` — Not Found
- `405` — Method Not Allowed
- `419` — Page Expired (jika relevan dengan framework)
- `422` — Unprocessable Entity
- `429` — Too Many Requests
- `500` — Internal Server Error
- `502` / `503` — Bad Gateway / Service Unavailable

Daftar tersebut merupakan contoh, bukan daftar kode yang harus dibuatkan halaman tersendiri.

## Batasan

- Tidak ada custom error page dengan desain khusus pada MVP.
- Tidak ada kebutuhan untuk membuat ilustrasi, copywriting, atau layout khusus untuk setiap kode error.
- Penanganan autentikasi dan otorisasi Admin tetap mengikuti mekanisme keamanan standar aplikasi.
- Error internal tetap harus dicatat melalui logging yang sesuai dan tidak boleh membocorkan stack trace, kredensial, atau detail sensitif kepada pengunjung.
- Kebutuhan UX dan halaman error kustom dapat dievaluasi kembali setelah MVP stabil.

## Catatan teknis

Implementasi harus memastikan status HTTP yang benar tetap dikirim meskipun tampilan error menggunakan default umum. Detail konfigurasi mengikuti framework dan web server yang dipilih.

---
