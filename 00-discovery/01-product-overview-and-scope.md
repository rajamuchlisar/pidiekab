# 01 Product Overview And Scope

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `product-vision.md`
- `release-scope.md`
- `scope-candidates.md`
- `information-sources.md`
- `visual-direction-decision.md`
- `visual-experience-reference-discover-tasmania.md`

---

## Materi sumber: `product-vision.md`

# Product Vision & Discovery Decisions

## Identitas

-   Nama kerja: PidieKab
-   Area: Kabupaten Pidie, Aceh
-   Jenis: Independent Regional Promotion & Visitor Discovery Platform
-   Status: Product Discovery

## Visi

Memperkenalkan potensi Kabupaten Pidie kepada masyarakat luas serta
memudahkan wisatawan dari luar daerah menemukan destinasi, pengalaman
lokal, dan informasi perjalanan yang bermanfaat.

## Positioning

**Promoting Pidie, Discovering Its Potential.**

PidieKab bukan sekadar direktori wisata. Produk diarahkan menjadi
*digital regional discovery platform* yang menghubungkan potensi daerah
dengan orang yang ingin mengenal atau mengunjungi Pidie.

## Keputusan Disepakati

  -----------------------------------------------------------------------
  ID                      Keputusan               Status
  ----------------------- ----------------------- -----------------------
  PD-001                  Fokus utama adalah      Disepakati
                          promosi potensi         
                          Kabupaten Pidie         

  PD-002                  Platform membantu       Disepakati
                          wisatawan luar daerah   

  PD-003                  Platform independen,    Disepakati
                          bukan website resmi     
                          pemerintah              

  PD-004                  Backend menggunakan Go  Arah teknologi
                                                  disepakati

  PD-005                  Database menggunakan    Arah teknologi
                          PostgreSQL              disepakati
  -----------------------------------------------------------------------

## Target Pengguna

**Utama:** Orang yang ingin mengenal potensi Kabupaten Pidie, termasuk
calon pengunjung dan wisatawan luar daerah.

**Pendukung:** Pelaku UMKM/usaha lokal, pengelola destinasi, komunitas
budaya, penyelenggara kegiatan, masyarakat lokal, dan pengunjung umum.

## Kebutuhan Pengguna (Hipotesis)

-   Menemukan tempat dan pengalaman menarik di Pidie.
-   Memahami daya tarik, lokasi, akses, jam operasional, biaya, dan
    fasilitas.
-   Menemukan kuliner, produk lokal, dan usaha masyarakat.
-   Mengenal budaya, sejarah, dan agenda kegiatan.
-   Menyusun rencana kunjungan berdasarkan lokasi dan minat.
-   Menilai kebaruan dan sumber informasi.
-   Menghubungi pengelola atau pelaku usaha.

Hipotesis harus divalidasi sebelum menjadi requirement final.

## Kandidat Domain Konten

1.  Wisata dan destinasi.
2.  Kuliner dan UMKM.
3.  Budaya dan sejarah.
4.  Akomodasi dan fasilitas.
5.  Panduan wisata.
6.  Berita, artikel, dan agenda.

Daftar ini belum merupakan scope MVP.

## Prinsip Produk

-   Transparansi independensi dan non-afiliasi.
-   Sumber informasi, status verifikasi, dan tanggal pembaruan yang
    jelas.
-   Mobile-first dan aksesibel.
-   Konten mudah ditemukan melalui pencarian dan kategori.
-   Pengelolaan konten tidak bergantung pada developer.
-   Koreksi dan pembaruan informasi memiliki alur yang jelas.
-   Implementasi bertahap berdasarkan prioritas MVP.

## Hal yang Belum Otomatis Masuk MVP

-   Pemesanan hotel atau tiket.
-   Payment gateway atau marketplace.
-   Jejaring sosial atau pesan langsung.
-   Navigasi turn-by-turn.
-   Portal administrasi pemerintahan.
-   Klaim sebagai sumber resmi pemerintah.

## Open Decisions

-   Konten apa yang wajib tersedia saat peluncuran?
-   Apakah pemilik usaha/destinasi dapat mengajukan atau mengelola
    listing?
-   Siapa yang memverifikasi dan memperbarui konten?
-   Apakah pengguna boleh memberi rating, ulasan, atau koreksi?
-   Bahasa apa yang tersedia saat peluncuran?
-   Apakah peta interaktif masuk MVP?
-   Apakah artikel dikelola hanya oleh tim internal?
-   Apa sumber data awal dan bagaimana izin foto/konten diperoleh?

## Discovery Berikutnya

Tentukan prioritas konten untuk peluncuran pertama. Keputusan ini
menjadi dasar modul MVP, model data awal, dan task backlog.

---

## Materi sumber: `release-scope.md`

# Product Scope & Release Phases

## 1. Release Strategy

PidieKab will be developed using phased feature delivery. The first
release focuses on three core content domains. Additional domains are
planned for a later phase.

## 2. Phase 1 --- MVP

**Status: Agreed**

### 2.1 Wisata & Destinasi

Purpose: Promote tourist attractions and places of interest in Kabupaten
Pidie.

Candidate content: - Destination name and description. - Main image and
photo gallery. - Destination category and tags. - Address and geographic
coordinates. - Access information and visitor facilities, when
available. - Opening hours and entrance fees, when applicable and
verified. - Contact or official information source, when available. -
Publication status, verification status, and last-updated date.

### 2.2 Kuliner & UMKM

Purpose: Promote local food, culinary places, products, and community
businesses.

Candidate content: - Business or product name and description. -
Business type/category. - Main image and gallery. - Address and
geographic coordinates. - Contact and available social/contact links. -
Opening hours and price range, when available. - Product/menu
highlights, where relevant. - Publication status, verification status,
and last-updated date.

### 2.3 Budaya & Sejarah

Purpose: Introduce Pidie's history, traditions, arts, cultural heritage,
and historical stories.

Candidate content: - Article/story title and body. - Content type
(history, tradition, art, heritage, historical figure, etc.). - Cover
image and gallery, where relevant. - Related location, where relevant. -
Author/editor and source references. - Publication status, verification
status, and last-updated date.

## 3. Phase 2 --- Expansion

**Status: Planned; not part of the initial MVP implementation**

1.  Akomodasi & Fasilitas --- hotels, lodging, visitor facilities, and
    travel support services.
2.  Panduan Wisata --- itineraries, routes, travel tips, and practical
    visitor guides.
3.  Berita & Agenda Kegiatan --- local news, festivals, events, and
    activity calendars.

Phase 2 should be supported by the product's modular design, but its
detailed features and implementation tasks will be specified later.

## 4. Shared Content Management Foundation

The three Phase 1 domains should reuse a common content-management
foundation where appropriate:

-   Categories and subcategories.
-   Title and description/body.
-   Featured image and media gallery.
-   Location/address and optional geographic coordinates.
-   Contact details where relevant.
-   Draft/published/archived publication lifecycle.
-   Source/provenance references.
-   Verification status.
-   Created, updated, and published timestamps.
-   SEO metadata where relevant.

Do not force all content types into an identical form. Use shared
primitives while allowing domain-specific fields and validation.

## 5. Cross-Cutting Capabilities to Evaluate for MVP

These capabilities are candidates and still require prioritization: -
Public browsing and detail pages. - Search, category filtering, and
basic sorting. - Responsive/mobile-first interface. - Admin
authentication and role permissions. - Content
create/edit/review/publish workflow. - Media upload and image
optimization. - Source and verification metadata. - SEO-friendly URLs
and metadata. - User-submitted correction or listing proposal.

## 6. Scope Guardrails

-   Phase 1 must not silently expand to all Phase 2 domains.
-   Do not implement booking, marketplace checkout, payment processing,
    social networking, or turn-by-turn navigation unless separately
    approved.
-   Phase 2 modules must not be built merely because their placeholder
    appears in the architecture.
-   Data accuracy, source attribution, and image usage rights are
    product requirements, not optional polish.

---

## Materi sumber: `scope-candidates.md`

# Product Scope Candidates

Dokumen ini mencatat kandidat domain untuk dinilai. Tidak ada kandidat
yang otomatis menjadi fitur MVP sebelum disepakati.

  -----------------------------------------------------------------------
  Domain                  Tujuan                  Contoh informasi
  ----------------------- ----------------------- -----------------------
  Wisata & Destinasi      Menemukan tempat        Deskripsi, foto,
                          menarik                 lokasi, akses,
                                                  fasilitas, biaya, jam
                                                  operasional

  Kuliner & UMKM          Menemukan produk/usaha  Menu/produk, alamat,
                          lokal                   kontak, jam buka,
                                                  kisaran harga

  Budaya & Sejarah        Mengenal identitas      Tradisi, seni, sejarah,
                          daerah                  warisan lokal, sumber

  Akomodasi & Fasilitas   Mendukung perjalanan    Penginapan, fasilitas
                                                  pengunjung, kontak

  Panduan Wisata          Merencanakan kunjungan  Itinerary, rute, tips,
                                                  transportasi

  Berita & Agenda         Mengetahui kegiatan     Artikel, event,
                          terbaru                 tanggal, lokasi,
                                                  penyelenggara
  -----------------------------------------------------------------------

## Kapabilitas lintas domain untuk dievaluasi

-   Pencarian dan filter.
-   Kategori dan tag.
-   Peta dan koordinat lokasi.
-   Galeri foto dan pengelolaan hak penggunaan.
-   Sumber, status verifikasi, dan tanggal pembaruan.
-   Formulir koreksi atau usulan perubahan.
-   CMS/admin untuk mengelola konten.
-   SEO dan metadata halaman.
-   Bahasa Indonesia dan bahasa tambahan.
-   Analitik penggunaan.

## Kriteria Prioritas

1.  Kesesuaian dengan visi promosi potensi daerah.
2.  Manfaat langsung bagi wisatawan luar daerah.
3.  Ketersediaan dan keandalan data.
4.  Beban operasional untuk menjaga informasi akurat.
5.  Kompleksitas implementasi dan dependensi.
6.  Risiko hukum, privasi, dan hak cipta.

MVP harus memiliki alur inti yang didukung data yang dapat
dipertanggungjawabkan, bukan sekadar sebanyak mungkin modul.

---

## Materi sumber: `information-sources.md`

# Information Sources & Verification Policy

## Decision SRC-001 --- Approved Information Source Types

**Status: Agreed**

PidieKab will gather information from four source categories:

  -----------------------------------------------------------------------
  Code              Source type       Description       Typical use
  ----------------- ----------------- ----------------- -----------------
  SRC-01            Research          Independent       Discover facts,
                                      research using    background,
                                      the internet,     leads, and
                                      articles, and     references
                                      open sources      

  SRC-02            Visit             Direct            Confirm location,
                                      observation or    access,
                                      on-site visits    facilities,
                                                        photos, and
                                                        current
                                                        conditions

  SRC-03            Official          Publications,     Support official
                                      documents, and    facts, public
                                      channels from     information,
                                      relevant          event details,
                                      institutions      and historical
                                                        references

  SRC-04            Social            Social media and  Discover local
                                      publications by   businesses,
                                      businesses,       products,
                                      destination       announcements,
                                      operators, or     and contact
                                      relevant parties  channels
  -----------------------------------------------------------------------

These source types describe how information was obtained. They do not by
themselves determine whether a content item is verified.

## 2. Verification Principles

1.  Record the source for factual claims whenever possible.
2.  Distinguish source type from verification status.
3.  A social-media post is not automatically verified.
4.  An official source can still be outdated; check publication/update
    dates and context.
5.  Direct observation confirms only what was observed at that time. It
    does not prove permanent operating hours, prices, ownership, or
    future availability.
6.  For material claims, seek corroboration where practical.
7.  Record when the information was last checked and who performed the
    check.
8.  Do not publish content that cannot be responsibly substantiated;
    keep it in draft or mark it for further review.
9.  Use images only when the team has permission, a valid license, or
    another lawful basis for use. Record attribution and usage
    conditions.
10. Do not present PidieKab content as officially endorsed by a
    government institution.

## 3. Suggested Source Metadata

The content system should be able to record, where applicable:

-   Source type (`research`, `visit`, `official`, `social`).
-   Source name or publisher.
-   Source URL or reference.
-   Source publication date, if known.
-   Date accessed or observed.
-   Notes describing which claim or content field the source supports.
-   Verification status.
-   Last verified date.
-   Verifier/editor account.
-   Image origin, license/permission, and attribution.

A content item may have multiple sources. Avoid storing all provenance
in one unstructured text field if a relational source record can be
maintained.

## 4. Proposed Verification States

These are proposed states for product/technical design and need final
confirmation: - `unverified`: not yet checked by an editor. -
`in_review`: being checked. - `verified`: checked against adequate
supporting evidence. - `needs_update`: potentially outdated or
contradicted. - `rejected`: evidence is insufficient or content is
unsuitable.

Verification status is separate from publication status. For example, a
verified item may remain a draft until an authorized editor publishes
it.

## 5. Content Freshness

Different content types need different review intervals. Do not assume
one global expiry period.

Potentially time-sensitive fields: - Opening hours. - Prices and
entrance fees. - Contact details. - Business operating status. - Event
dates. - Access conditions and facilities.

The product should record `last_verified_at` and support an editorial
review process. Specific review cadence remains an open decision.

## 6. Open Decisions

-   Should all public content require at least one recorded source?
-   Which claims require multiple corroborating sources?
-   What review interval should apply to destinations, businesses, and
    historical/cultural articles?
-   Should the admin dashboard show content that has not been reviewed
    recently?
-   What is the process for correcting a published factual error?
-   Which image licensing/permission evidence must be recorded?

---

## Materi sumber: `visual-direction-decision.md`

# Decision --- Primary Visual Reference

## Decision VIS-001 --- Discover Tasmania as the Primary Visual Reference

**Status: Agreed**

Discover Tasmania (https://www.discovertasmania.com.au/) is the primary
visual and visitor-experience reference for PidieKab.

## Application to PidieKab

PidieKab will adapt relevant design principles from the reference: -
Destination-led hero imagery. - Editorial storytelling and curated
content. - Visitor-oriented navigation and discovery paths. - Clear
regional/place discovery. - Strong visual hierarchy and readable content
layouts. - Responsive, mobile-friendly presentation.

## Adaptation Requirements

1.  The reference is inspiration, not a template to copy
    pixel-for-pixel.
2.  PidieKab must have its own identity, content, visual assets, and
    brand expression.
3.  The design must represent Kabupaten Pidie and its local context
    authentically.
4.  The website must remain clearly independent and must not imply
    government affiliation.
5.  All previously agreed MVP scope and product decisions remain in
    force.
6.  Search results continue to display only title, category, and
    location; imagery and descriptions are not added to search result
    entries by this visual decision.
7.  The visual direction must support Indonesian and English content.
8.  Design must prioritize responsive performance, accessibility, and
    image rights.

## Scope

This decision establishes the primary visual reference. It does not
finalize: - PidieKab logo or complete brand identity. - Exact color
palette and typography. - Final homepage section order. - Final category
detail-page templates. - Component library or frontend framework.

These will be defined in subsequent design and technical work.

---

## Materi sumber: `visual-experience-reference-discover-tasmania.md`

# Visual & Experience Reference --- Discover Tasmania

## Reference

Website: https://www.discovertasmania.com.au/

**Status: User-selected design inspiration; not a final UI
specification**

The user likes the Discover Tasmania website template as a visual and
experience reference for PidieKab.

## Observed Design Patterns

The reference website combines: - Strong destination-led hero imagery
and introductory messaging. - Visitor-oriented navigation groupings such
as places, things to do, accommodation, and trip planning. - Editorial
feature stories and curated inspiration content. - Search entry point
and popular search suggestions. - Region discovery sections that
introduce places through imagery and short narratives. - Useful visitor
information and FAQ content.

## Recommended Adaptation for PidieKab

Use the reference as inspiration for information presentation and visual
hierarchy, not as a design to copy.

### Homepage

Potential sections: 1. Hero area introducing Kabupaten Pidie with a
clear value proposition. 2. Global search entry point. 3. Curated
highlights from the three MVP categories. 4. Jelajahi Pidie entry
points. 5. Editorial storytelling about local potential, culture, and
destinations. 6. About PidieKab and its independent status.

### Navigation

Adapt visitor-oriented groupings to the approved PidieKab MVP: -
Beranda. - Jelajahi Pidie. - Wisata & Destinasi. - Kuliner & UMKM. -
Budaya & Sejarah. - Tentang PidieKab. - Language switcher: Indonesian /
English.

Do not add Phase 2 modules to MVP navigation as implemented features.

### Detail Pages

Use a strong editorial layout with prominent content title, relevant
imagery where available and rights-cleared, readable narrative, and
category-specific information. This is compatible with the current
proposal for a shared detail-page framework with category-specific
sections, which remains unapproved.

### Search Results

Preserve the already approved PidieKab search-result decision: search
result entries show only title, category, and location. Do not copy
image-heavy cards from the reference into search results. Rich imagery
may be used on the homepage, curated editorial sections, and detail
pages.

## Design Principles

-   Destination-led and visually engaging.
-   Editorial rather than directory-like.
-   Clear navigation for first-time visitors.
-   Mobile-first and responsive.
-   Accessible typography, contrast, and controls.
-   Fast-loading images with responsive sizes and modern formats.
-   Avoid excessive animation or heavy visual effects.
-   Do not imply government affiliation.
-   Respect image rights, source attribution, and content verification.

## Decisions Still Open

-   Confirm the overall design direction inspired by Discover Tasmania.
-   Finalize homepage section order and content hierarchy.
-   Decide whether category detail pages share a common template with
    category-specific sections.
-   Select typography, color palette, image treatment, and responsive
    navigation behaviour.

---
