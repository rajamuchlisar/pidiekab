# PidieKab — Phase 0 Execution Plan

**Versi:** v79  
**Tujuan:** membangun fondasi repository dan membuktikan keputusan teknis berisiko sebelum implementasi fitur.

## 1. Hasil yang harus dicapai

Phase 0 menghasilkan repository yang dapat dijalankan dari mesin bersih, standar development yang konsisten, Go HTTP skeleton, dua frontend workspace, CI baseline, dan proof-of-concept SSG.

Phase 0 tidak membangun fitur bisnis CMS atau public discovery.

## 2. Urutan kerja

| Urutan | Workstream | Hasil |
|---|---|---|
| 0.1 | Repository foundation | Struktur folder, README, `.gitignore`, environment example |
| 0.2 | Go foundation | Go module, config validation, `net/http` server, graceful shutdown |
| 0.3 | Health & logging | `/health/live`, readiness seam, `log/slog`, request ID |
| 0.4 | Frontend workspace | Public site dan Admin CMS Vue 3 + TypeScript + Vite |
| 0.5 | SSG proof-of-concept | Dynamic localized routes menjadi static HTML |
| 0.6 | Database tooling proof | PostgreSQL lokal dan Goose migration proof |
| 0.7 | CI baseline | Go test/vet, frontend typecheck/build, migration validation |
| 0.8 | Phase 0 review | Catatan hasil, ADR SSG, keputusan migration, checklist |

## 3. SSG candidate

`vite-ssg` menjadi kandidat pertama untuk POC, bukan keputusan final. Package ini menyediakan static-site generation untuk Vue 3 di atas Vite dan membutuhkan Vue Router serta `@unhead/vue` menurut dokumentasi package. Referensi: https://www.npmjs.com/package/vite-ssg

Alasan diuji:
- tetap menggunakan Vue 3 + TypeScript + Vite;
- dapat menghasilkan HTML statis saat build;
- tidak membutuhkan Node.js untuk melayani hasil build di production.

Risiko yang harus dibuktikan:
- daftar dynamic routes dapat dibangun dari snapshot data konten;
- route `/id/...` dan `/en/...` menghasilkan HTML yang benar;
- SEO title, description, canonical, hreflang, sitemap dan 301 dapat dihasilkan;
- output dapat disajikan langsung oleh Go;
- tidak ada Node SSR server yang berjalan pada production.

Dokumentasi resmi Vue menjelaskan SSG sebagai pre-render HTML yang dapat disajikan sebagai static files, dan cocok untuk content-based websites: https://vuejs.org/guide/scaling-up/ssr

## 4. Definition of Done

- [ ] Clone repository pada mesin bersih dan ikuti README tanpa langkah tersembunyi.
- [ ] Go server berjalan dengan graceful shutdown.
- [ ] Liveness endpoint mengembalikan HTTP 200.
- [ ] Config wajib divalidasi saat startup.
- [ ] Kedua frontend dapat di-install dan build.
- [ ] SSG POC membuat minimal route homepage ID/EN, satu category ID/EN, dan satu detail ID/EN.
- [ ] Generated HTML berisi title, description, canonical, serta locale content yang sesuai.
- [ ] Draft/unpublished route tidak ikut dihasilkan.
- [ ] Sitemap hanya memuat halaman Published yang indexable.
- [ ] Build output dapat dilayani oleh Go static file handler.
- [ ] PostgreSQL migration proof dapat dijalankan pada database disposable.
- [ ] CI menjalankan baseline checks.
- [ ] Keputusan yang belum terbukti tetap OPEN; tidak ditandai LOCKED.

## 5. Batasan

Tidak menambahkan Redis, message broker, Docker/Kubernetes sebagai kewajiban, ORM, framework backend besar, reverse proxy, analytics, atau fitur Phase 2.

## 6. Output Phase 0

1. Repository foundation.
2. Go application skeleton.
3. Frontend workspace.
4. SSG POC report dan ADR.
5. Migration tooling proof.
6. CI pipeline.
7. Updated readiness checklist.
