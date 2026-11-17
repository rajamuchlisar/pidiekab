# 07 Location And Map

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `location-administrative-area-master-data.md`
- `map-location-experience.md`

---

## Materi sumber: `location-administrative-area-master-data.md`

# LOC-001 — Pengelolaan Data Wilayah dan Lokasi Konten

**Status: Disepakati berdasarkan rekomendasi yang dilanjutkan.**

## Keputusan
Data wilayah administratif menggunakan **master data tersendiri**, sedangkan alamat detail dan koordinat lokasi diisi pada form konten sesuai ketersediaan dan hasil verifikasi.

Master data wilayah MVP mencakup:
- Kabupaten: Kabupaten Pidie (konteks wilayah platform).
- Kecamatan.
- Gampong/desa.

## Prinsip master data wilayah
1. Daftar kecamatan dan gampong dikelola sebagai referensi terpusat, bukan diketik ulang bebas pada setiap konten.
2. Gampong terhubung ke kecamatan yang sesuai.
3. Admin memilih kecamatan dan gampong dari pilihan master data pada form konten.
4. Nama wilayah yang ditampilkan kepada publik mengikuti nama yang telah diverifikasi.
5. Perubahan master data harus memperhatikan konten yang sudah menggunakan wilayah tersebut.
6. Pengelolaan master data dan sumber validasi nama/wilayah ditetapkan pada spesifikasi CMS/teknis.

## Data lokasi pada konten
Untuk konten Wisata dan Kuliner & UMKM, form dapat menyimpan:
- Kecamatan (referensi master data).
- Gampong/desa (referensi master data, opsional jika tidak diketahui).
- Alamat atau deskripsi lokasi.
- Latitude dan longitude jika tersedia.
- Status ketepatan lokasi: akurat/terverifikasi atau perkiraan/belum terverifikasi.
- Catatan lokasi bila diperlukan.

Untuk Budaya & Sejarah, lokasi dapat diisi bila relevan; lokasi tidak diwajibkan untuk topik atau tradisi yang tidak terikat pada satu titik geografis.

## Prinsip koordinat dan peta
- Koordinat tidak boleh dibuat-buat.
- Pin peta hanya ditampilkan jika koordinat memadai.
- Lokasi perkiraan harus diberi penanda yang jelas dan tidak ditampilkan seolah-olah titik akurat.
- Jika koordinat tidak tersedia, halaman tetap dapat menampilkan informasi wilayah/alamat tanpa peta.
- Detail halaman mengikuti MAP-001: peta/lokasi hanya pada halaman detail; tidak ada peta gabungan atau pencarian peta pada MVP.

## Pengisian manual
- Alamat lengkap tetap dapat diisi manual karena format alamat tiap tempat berbeda.
- Catatan/petunjuk lokasi dapat diisi manual bila membantu pengunjung.
- Kecamatan dan gampong tidak menjadi teks bebas jika pilihan master data tersedia; opsi tidak diketahui dapat digunakan bila informasi belum dapat dipastikan.

## Batasan MVP
- Tidak ada master data seluruh kabupaten/kota di Indonesia.
- Tidak ada pencarian geospasial, radius, atau filter jarak.
- Tidak ada PostGIS hanya untuk kebutuhan menampilkan satu titik lokasi.
- Tidak ada geocoding otomatis yang menebak koordinat tanpa pemeriksaan Admin.
- Daftar resmi wilayah dan mekanisme impor/pembaruan master data ditentukan pada tahap persiapan data dan Technical Specification.

---

## Materi sumber: `map-location-experience.md`

# Map & Location Experience

## Decision MAP-001 --- Location Map on Detail Pages Only

**Status: Agreed**

For the MVP, PidieKab will show location information and a
map/directions link on each relevant detail page. A combined interactive
map for browsing all destinations and businesses is not included in the
MVP.

## MVP Behavior

For relevant destination and culinary/business detail pages: - Display
the address or location description when available. - Store geographic
coordinates (latitude and longitude) when available and verified. - Show
a map preview or embedded map on the detail page, subject to the final
provider and implementation decision. - Provide an external directions
link so visitors can open the location in a mapping/navigation
application. - If coordinates are not available, show the available
address/location description and do not fabricate a map pin. - Clearly
distinguish a verified location from an approximate or unverified
location.

## Out of Scope for MVP

-   One combined map showing all destinations, culinary places, and
    UMKM.
-   Map-based discovery/search and filtering.
-   Custom route planning or itinerary routing.
-   Turn-by-turn navigation within PidieKab.

These may be considered in a later phase based on user needs and
operational feasibility.

## Data Implications

The data model should allow location fields where relevant: -
`address` - `latitude` - `longitude` - Optional location
precision/status (for example exact, approximate, unverified) - Location
verification timestamp and source reference, where applicable

Coordinates should use appropriate PostgreSQL numeric types or PostGIS
only if spatial queries become necessary. Do not introduce PostGIS
solely for displaying a pin and linking to external directions; final
database choice belongs in the technical specification.

## Provider Decision --- Open

The map display/provider has not yet been selected. Candidates may
include OpenStreetMap-based tiles or an external maps link/provider.
Evaluate: - Licensing and attribution requirements. - Tile usage policy
and service reliability. - API keys, quotas, and operating costs. -
Privacy implications of sending visitor location/search requests to
third parties. - Mobile behavior and accessibility.

Do not assume the public OpenStreetMap tile servers are an unrestricted
production tile service.

## Acceptance Criteria (Initial)

-   A destination or business detail page can display its recorded
    location.
-   The external directions link uses the saved location and does not
    invent coordinates.
-   Pages with missing coordinates degrade gracefully.
-   The page communicates approximate/unverified location where
    applicable.
-   No combined map discovery screen is implemented in MVP.

---
