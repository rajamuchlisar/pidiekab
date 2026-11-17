# Public Site And Admin Cms Ui Specification

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `public-site-ui-specification.md`
- `admin-cms-ui-specification.md`

---

## Materi sumber: `public-site-ui-specification.md`

# Public Site UI Specification — PidieKab

**Status:** MVP UI baseline

## 1. Visual direction

Gunakan Discover Tasmania sebagai referensi pengalaman visitor discovery, bukan menyalin layout atau aset. Identitas PidieKab harus modern, editorial, hangat, mudah dibaca, dan menonjolkan lanskap, budaya, kuliner, serta potensi lokal.

Prinsip: fotografi kuat tetapi tidak memenuhi seluruh layar; whitespace cukup; tipografi jelas; card sederhana; navigasi mudah dipahami wisatawan domestik/internasional; konten lebih penting daripada dekorasi. Design tokens final ditetapkan terpusat.

## 2. Header

Desktop: logo kiri; navigasi Beranda, Jelajahi Pidie, Wisata, Kuliner & UMKM, Budaya & Sejarah, Tentang PidieKab; language switcher ID/EN; sticky header.

Mobile: logo, hamburger accessible, collapsible menu, language switcher mudah ditemukan. Search tidak berada di header; global search berada di homepage hero dan search page.

## 3. Homepage

Urutan:
1. Header.
2. Hero dengan satu foto statis rights-cleared, headline, supporting text, global search.
3. Tiga category visual cards.
4. “Jelajahi Pidie” content grid dengan Featured, Latest, Random; tanpa carousel.
5. “Cerita tentang Pidie”: satu Lead Story dan Supporting Stories.
6. Footer.

Hero tidak fullscreen/oversized dan tidak memakai carousel/video. Card kategori menampilkan image dan title. Content card menampilkan main photo, title, category, dan location bila ada; tanpa summary. Featured/Latest/Random tidak boleh menampilkan duplikasi sesuai aturan curation.

Story section memakai existing content, bukan content type terpisah. Satu Lead Story dan Supporting Stories diurutkan manual per locale.

## 4. Category listing

Breadcrumb, category title/description, responsive content grid, numbered pagination, empty state. Hanya Published pada locale aktif. Tidak ada multi-category/tag filters. Category inactive tidak muncul di publik.

## 5. Content detail

Urutan: breadcrumb, category label, title, main photo, article body, location/map bila relevan dan tersedia, sources/references yang boleh ditampilkan. Hide field kosong. Tidak ada gallery/carousel. Jika koordinat approximate, tampilkan label approximate. Jangan membuat koordinat.

## 6. Search

- Search entry dari homepage hero.
- Locale route, misalnya `/id/search?q=pantai`.
- Query tetap terlihat.
- Filter: Semua, Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah.
- Filter mempertahankan query dan reset pagination ke page 1.
- Hasil: title, category, location bila tersedia; tanpa image/summary.
- Numbered pagination, default 12.
- Empty state memberi arahan mengubah kata kunci.
- Loading/error tidak menghapus query.

## 7. Language behavior

Default `/id/`. Language switcher menuju pasangan translation bila Published. Jika belum tersedia, tetap dalam konteks konten, tampilkan translation unavailable notice dan link ke locale tersedia. Tidak ada silent fallback. Isi dan metadata harus cocok dengan locale URL.

## 8. Footer

Kelompok About PidieKab; Explore Pidie (tiga kategori); Information (About Us, Privacy Policy, Disclaimer); Contact & Social (hanya informasi terkonfirmasi). Cantumkan pernyataan eksplisit bahwa PidieKab adalah platform independen, bukan situs resmi/representasi Pemerintah Kabupaten Pidie.

## 9. Responsive

Validasi minimal 320–375 px, 390–430 px, 768–1024 px, dan 1280 px ke atas. Grid berubah sesuai ruang. Pastikan tidak ada horizontal overflow, clipping, atau kontrol terlalu kecil.

## 10. UI states

Setiap halaman/component mendefinisikan loading, content, empty, error, unavailable translation, dan disabled/inactive state bila relevan. Jangan menampilkan dummy data di production bila SSG/API gagal. Gunakan fallback visual untuk gambar unavailable.

---

## Materi sumber: `admin-cms-ui-specification.md`

# Admin CMS UI Specification — PidieKab

**Status:** MVP UI baseline

## 1. Principles

CMS digunakan tim internal PidieKab. Prioritaskan kemudahan operasional, keterbacaan, konsistensi form, dan pencegahan kesalahan publikasi. Tidak ada approval workflow, role hierarchy, atau public operator portal. Satu Admin role pada MVP; backend tetap menjadi otoritas akses.

## 2. Layout and navigation

Desktop: sidebar, header dengan page title/status sesi/profile, main content. Mobile: collapsible navigation.

Menu:
1. Dashboard
2. Konten: Semua Konten, Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah
3. Halaman Informasi
4. Pengaturan
5. Administrator

## 3. Login

Email, password, accessible show/hide control, submit, generic login error. Tidak ada public registration. Jangan menyimpan password/session token pada localStorage. Setelah login fetch `/auth/me`.

## 4. Dashboard

Ringkasan operasional: jumlah konten per kategori, Draft/Published/Archived, status ID/EN terpisah, konten terakhir diubah, aktivitas Admin, status generation terakhir bila tersedia. Tidak ada page views, visitor analytics, popularity, atau revenue metrics.

## 5. Content list

Kolom: Judul, Kategori, Bahasa ID, Bahasa EN, Status, Unggulan, Terakhir diubah, Aksi. Search, filter kategori/status, pagination, edit, permanent delete confirmation. Status ID/EN harus terlihat terpisah. Table mobile memakai horizontal scroll terkontrol atau row/card adaptation.

## 6. Content editor

Satu content identity dengan tab ID dan EN. Shared fields: kategori utama, kecamatan/gampong, alamat, koordinat/accuracy, main photo, sources.

Locale fields: title, slug, summary, rich text body, SEO title/description, publication status, metadata perubahan.

Perilaku:
- unsaved state tiap locale dipertahankan saat pindah tab;
- indikator unsaved changes;
- Save menyimpan draft;
- Preview privat;
- Publish/Archive hanya locale aktif;
- tidak ada approval/verification state;
- edit Published tidak mengubah public version sampai Publish kembali.

Rich Text Editor mendukung heading, bold/italic, list, link, simple table, image. Raw HTML/JS tidak diperbolehkan; server-side sanitization wajib.

## 7. Delete

Permanent delete menjelaskan bahwa semua translations dan sources terkait dihapus. Shared media tidak ikut terhapus. Konfirmasi eksplisit. Tidak ada trash/recycle bin/restore CMS.

## 8. Media Library

List/grid, upload, preview, filename/size, alt ID/EN, source/origin, license/permission, attribution, search/filter sederhana. Media in-use tidak dapat dihapus. Jangan tampilkan storage path internal.

## 9. Categories and curation

Category list/create/edit, translation tabs, active/inactive, delete hanya jika tidak ada content reference. Inactive tidak dapat dipilih untuk konten baru.

Featured Content dan Story curation terpisah. Hanya locale Published yang dapat dipilih. Story memiliki satu Lead dan Supporting Stories berurutan; Lead tidak boleh menjadi Supporting.

## 10. Static pages and settings

Static pages dikelola per locale dengan publication status independen. Website settings mengelola identity, logo/favicon, contact, social URLs, language, default SEO. SEO konten dapat override default.

## 11. Administrator

List, add, edit profile, activate/deactivate, reset password, manage own profile/password. Cegah penonaktifan Admin aktif terakhir. Admin inactive tidak dapat login.

## 12. Form feedback

Label selalu terlihat; error dekat field; API error global menampilkan request ID bila tersedia; confirmation untuk destructive actions; loading mencegah double submit; input dipertahankan saat validation error; status sukses/gagal tidak hanya mengandalkan warna.

---
