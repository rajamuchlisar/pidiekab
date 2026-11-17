# PidieKab — Final MVP Scope & Release Boundary

**Versi:** v78  
**Status:** Baseline ruang lingkup MVP

## 1. Sasaran MVP

Meluncurkan platform independen untuk membantu masyarakat dan pengunjung domestik maupun internasional menemukan potensi Kabupaten Pidie melalui konten destinasi, kuliner/UMKM, serta budaya/sejarah dalam bahasa Indonesia dan Inggris.

MVP berfokus pada pengalaman discovery dan pengelolaan editorial internal. MVP bukan sistem reservasi, marketplace, social network, portal pemerintahan, atau sistem transaksi.

## 2. Fitur yang termasuk MVP

### Public website
- Homepage dengan hero statis, pencarian global, kategori, konten pilihan/terbaru/acak, dan editorial story.
- Jelajahi Pidie dan halaman kategori.
- Detail konten dengan breadcrumb, foto utama, artikel, lokasi/peta jika data tersedia, serta sumber/referensi.
- Tiga kategori MVP: Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah.
- Bahasa ID dan EN dengan URL prefix `/id/` dan `/en/`.
- Search dengan filter kategori, relevance ordering, dan pagination.
- Static information pages: About, Contact, Privacy Policy, Disclaimer.
- SEO per locale, canonical, hreflang bila pasangan bahasa published, sitemap, dan 301 redirects.
- Responsive layout dan aksesibilitas dasar.

### Admin CMS
- Login Admin internal; satu role Admin.
- Dashboard operasional.
- CRUD konten dan translation ID/EN.
- Draft, Preview, Publish, Archive; status per bahasa independen.
- Pengelolaan kategori dan status aktif/nonaktif.
- Master lokasi baca/pilih; detail dan koordinat opsional.
- Media Library dengan metadata alt text, source, license/permission/attribution.
- Featured content dan homepage story curation.
- Static pages bilingual.
- Website settings: identity, logo/favicon, contact/social, language, SEO defaults.
- Administrator management sesuai keputusan satu-role.
- Audit log untuk tindakan penting.
- Permanent delete dengan confirmation.

### Backend & platform
- Go Modular Monolith dan `net/http`.
- PostgreSQL satu database.
- REST API `/api/v1`.
- Server-side Admin sessions, CSRF protection.
- PostgreSQL-backed SSG job queue dan in-process worker.
- Atomic release activation.
- Structured logs, health endpoints, metrics dasar.
- Deployment Ubuntu 24.04 + systemd; Go melayani HTTP/HTTPS dan static files.

## 3. Tidak termasuk MVP

- Akomodasi & fasilitas sebagai kategori/fitur tersendiri.
- Panduan wisata, itinerary builder, route planner, turn-by-turn navigation.
- Berita & agenda kegiatan sebagai modul terpisah.
- Akun pengunjung, operator listing, public submission, user-generated content.
- Rating, review, komentar, favorit, chat, forum, social feed.
- Booking, reservasi, pembayaran, e-commerce, marketplace.
- Visitor analytics/dashboard page views.
- AI search, semantic search, chatbot, recommendation personalization.
- Elasticsearch, Redis, message broker, microservices.
- PostGIS, spatial search, nationwide location data.
- Multi-role RBAC atau approval workflow.
- Gallery foto, video hero, homepage carousel.
- Custom error-page design untuk semua status HTTP.
- Multi-instance/high-availability deployment kecuali kebutuhan aktual mengharuskan.

## 4. Definition of MVP Complete

MVP dapat dinyatakan siap rilis jika:
1. Admin dapat login aman, mengelola konten ID/EN, preview, publish, archive, dan mengelola sumber/media.
2. Konten Draft tidak dapat diakses publik, termasuk melalui file statis, sitemap, search, dan API publik.
3. Pengunjung dapat menemukan dan membaca konten Published dalam locale yang dipilih.
4. Bahasa yang belum dipublikasikan tidak diam-diam menampilkan konten bahasa lain.
5. Search, filter, pagination, URL, canonical, hreflang, sitemap, dan redirects berfungsi sesuai spesifikasi.
6. Kegagalan SSG tidak mengganti release publik yang masih sehat.
7. Upload, session, CSRF, rich text, dan database access lulus security tests.
8. Backup database dan media berhasil dipulihkan dalam restore drill.
9. TLS renewal dan certificate reload berhasil diuji.
10. QA acceptance, responsive, accessibility baseline, logging, health checks, dan operational runbook tersedia.

## 5. Aturan perubahan scope

- Fitur Phase 2 tidak masuk MVP hanya karena mudah ditambahkan.
- Perubahan scope harus mencantumkan alasan pengguna/bisnis, dampak database/API/UX/security/QA, dan perubahan estimasi.
- Prioritaskan stabilitas, kualitas konten, kecepatan halaman, keamanan CMS, dan operasional sederhana.
