# SSG Proof-of-Concept Plan

**Versi:** v79  
**Status:** `vite-ssg` kandidat POC, belum dikunci.

## 1. Tujuan

Membuktikan bahwa Vue 3 + TypeScript dapat menghasilkan public website statis dengan localized dynamic routes dan SEO metadata, lalu hasilnya disajikan oleh Go `net/http`.

## 2. Kandidat

`vite-ssg` — https://www.npmjs.com/package/vite-ssg

Dokumentasi Vue tentang SSR/SSG: https://vuejs.org/guide/scaling-up/ssr  
Dokumentasi Vite SSR/pre-rendering: https://vite.dev/guide/ssr

Vite menyediakan mekanisme pre-rendering berbasis SSR build, tetapi dokumentasi Vite menyebutnya sebagai low-level API. Karena itu POC membandingkan kemudahan dan kontrol `vite-ssg` dengan pendekatan Vite native pre-render.

## 3. Test routes

Gunakan fixture data sintetis, bukan data destinasi faktual:
- `/id/`
- `/en/`
- `/id/wisata-destinasi/`
- `/en/tourism-destinations/`
- `/id/wisata-destinasi/contoh-konten/`
- `/en/tourism-destinations/sample-content/`

## 4. Acceptance criteria

| Kriteria | Lulus jika |
|---|---|
| Localized route | Setiap URL menghasilkan HTML tersendiri |
| Content isolation | ID tidak menampilkan body EN dan sebaliknya |
| Metadata | Title dan description sesuai locale |
| Canonical | Self-referential pada URL locale tersebut |
| hreflang | Hanya muncul bila pasangan ID/EN sama-sama Published |
| Sitemap | Hanya memuat route Published dan indexable |
| Draft privacy | Draft tidak dibuat sebagai HTML publik |
| Dynamic route | Daftar route detail berasal dari data snapshot build |
| Static output | Semua file output dapat dilayani Go |
| Runtime | Tidak memerlukan Node process di production |
| Redirect | Slug lama menghasilkan 301 melalui Go redirect mapping |
| Reproducibility | Build dengan input sama menghasilkan output konsisten |

## 5. Catatan implementasi

- Build SSG menerima content snapshot dari database/API saat proses build.
- Build harus menggunakan data Published saja.
- Jika salah satu locale tidak Published, route locale tersebut tidak dibuat.
- Simpan manifest route/redirect untuk Go server bila diperlukan.
- Asset menggunakan fingerprint/hash dan cache immutable.
- HTML dan sitemap menggunakan cache policy yang sesuai; jangan cache halaman Admin.
- Jangan mengakses `window` atau browser-only API saat render server.
- Jangan mengandalkan runtime server Node untuk fallback.

## 6. Keputusan setelah POC

Buat ADR yang mencatat:
- tool yang diuji dan versinya;
- struktur route dan jumlah halaman uji;
- waktu build dan ukuran output;
- hasil tiap acceptance criteria;
- batasan dynamic route, sitemap, metadata, redirect;
- cara Go menyajikan hasil;
- rekomendasi LOCKED atau alasan menolak kandidat.
