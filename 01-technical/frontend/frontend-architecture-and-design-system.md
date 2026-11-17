# Frontend Architecture And Design System

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `frontend-architecture.md`
- `frontend-design-system-and-quality-gates.md`

---

## Materi sumber: `frontend-architecture.md`

# Frontend Architecture — PidieKab

**Status:** Baseline implementation specification  
**Stack:** Vue 3 + TypeScript + Vite  
**Public delivery:** Static Site Generation (SSG)  
**CMS delivery:** Dynamic Vue SPA  
**API:** Go REST API (`/api/v1`)

## 1. Two frontend applications

Repository memiliki dua frontend:
1. `frontend/public-site`: Vue 3 + TypeScript + Vite dengan SSG.
2. `frontend/admin-cms`: Vue 3 + TypeScript + Vite sebagai SPA.

Node.js hanya digunakan untuk dependency installation, development, lint, test, dan build. Production menjalankan Go binary yang menyajikan hasil build statis.

Keduanya boleh berbagi design tokens/komponen dasar bila kebutuhan nyata muncul. Jangan membuat workspace atau shared-package architecture yang kompleks sejak awal.

## 2. Public rendering

- Homepage, category listing, content detail, dan static pages dihasilkan sebagai HTML statis.
- SSG mengambil data Published melalui build-time data provider yang terhubung ke Go application/query layer atau endpoint internal yang tidak diekspos publik.
- Public browser tidak melakukan fetch API untuk memuat isi halaman utama pada setiap page load.
- Global Search adalah fitur runtime dan memanggil Public Search API.
- Output dipisahkan berdasarkan locale `/id/` dan `/en/`.
- Root `/` mengarah ke `/id/`.
- Hanya translation Published dan category Active yang masuk public output.
- Preview Draft memakai jalur privat, tidak anonymous, dan tidak menulis ke public release.

Tool SSG spesifik belum dikunci. Lakukan proof-of-concept untuk localized dynamic routes, SEO metadata, sitemap, dan deployment output sebelum memilih tooling.

## 3. CMS rendering

CMS adalah SPA yang disajikan Go dari satu prefix tetap, misalnya `/cms/`, dan menggunakan same-origin Admin REST API.

- Tidak memakai SSR.
- Authentication status dibaca dari `/auth/me`.
- Session token tidak disimpan di localStorage/sessionStorage; gunakan secure HttpOnly cookie.
- Route guard hanya untuk UX; backend tetap memvalidasi setiap request.
- Draft editor state disimpan di memory dan dikirim ke API saat Save.
- Unsaved state ID/EN tetap tersimpan saat berpindah tab.
- Preview memakai renderer/layout publik melalui akses privat.

## 4. State management

Gunakan Pinia secara selektif untuk authentication/current Admin dan state lintas halaman yang memang diperlukan. Server data dikelola melalui composables/service layer dengan loading/error state. Hindari duplikasi cache server data.

Public site tidak memerlukan Pinia pada awal MVP kecuali interaksi lintas komponen membutuhkannya.

## 5. API client

Satu typed API client per aplikasi:
- base path same-origin `/api/v1`;
- JSON serialization/parsing terpusat;
- cookie credentials sesuai same-origin;
- mapping `application/problem+json` ke typed API error;
- request ID untuk troubleshooting;
- timeout/abort untuk request yang dapat dibatalkan;
- tidak auto-retry mutation kecuali idempotency sudah dipastikan.

Endpoint service dikelompokkan berdasarkan capability: auth, content, category, media, location, curation, static pages, settings, admin, audit, search. Hindari generic API framework berlebihan.

## 6. TypeScript

- `strict: true`.
- Hindari `any`; gunakan `unknown` untuk data belum tervalidasi.
- DTO mengikuti API/OpenAPI contract.
- ID database diperlakukan sebagai string.
- Locale dibatasi `'id' | 'en'`.
- Content status dibatasi `'draft' | 'published' | 'archived'`.
- Jangan menganggap network response selalu valid.

## 7. Routes

Public:
- `/id/`, `/en/`
- `/{locale}/{category_slug}`
- `/{locale}/{category_slug}/{content_slug}`
- localized static page routes.

CMS:
- `/cms/login`
- `/cms/dashboard`
- `/cms/contents`
- `/cms/contents/new`
- `/cms/contents/:id/edit`
- `/cms/categories`
- `/cms/media`
- `/cms/settings`

Tidak ada browser language auto-detection. Language switcher menuju pasangan translation jika Published; bila belum tersedia, tampilkan notice dan link ke locale yang tersedia. Tidak ada silent fallback.

## 8. Build outputs

Public SSG output dan CMS build output berada di direktori terpisah. Go menyajikan active public release dan CMS assets. Hashed assets dapat di-cache immutable. Preview tidak pernah dicampur ke public release. Release sebelumnya tetap tersedia sampai release baru berhasil diaktifkan.

## 9. Accessibility and responsive behavior

Gunakan semantic HTML; semua kontrol keyboard-accessible; focus state terlihat; form memiliki label dan error yang terhubung; kontras memadai; alt text per locale; responsive mobile/tablet/desktop; jangan memakai warna sebagai satu-satunya indikator; respect reduced-motion.

## 10. Performance and security

- SSG HTML menampilkan konten utama tanpa menunggu JavaScript.
- Lazy-load image di luar viewport; hero image diberi prioritas sesuai kebutuhan.
- Gunakan responsive image variants dan ukuran file wajar.
- Hindari library besar untuk interaksi sederhana.
- Hindari `v-html` untuk konten tidak tepercaya.
- Rich text dirender dari structured content yang telah divalidasi/disanitasi server-side.
- Jangan menyimpan credential/session token di browser storage.
- Jangan memasukkan secret ke frontend build variables.

---

## Materi sumber: `frontend-design-system-and-quality-gates.md`

# Frontend Design System & Quality Gates — PidieKab

**Status:** Baseline; exact visual tokens to be finalized

## 1. Design tokens

Centralize brand/neutral/semantic colors, typography family/scale/weight/line-height, spacing, radius, shadows, container widths, breakpoints, focus ring, dan z-index. Jangan menyebarkan nilai arbitrer per component. Uji kontras terutama teks di atas foto.

## 2. Component inventory

Public: `SiteHeader`, `MobileNavigation`, `LanguageSwitcher`, `HeroSearch`, `CategoryCard`, `ContentCard`, `ContentGrid`, `StoryLeadCard`, `StorySupportingCard`, `Breadcrumbs`, `Pagination`, `LocationMap`, `TranslationUnavailableNotice`, `SiteFooter`, `SourceList`.

CMS: `AdminLayout`, `AdminSidebar`, `AdminHeader`, `DataTable`, `StatusBadge`, `LocaleTabs`, `FormField`, `RichTextEditor`, `MediaPicker`, `MediaLibrary`, `SourceEditor`, `ConfirmDialog`, `Toast/InlineAlert`, `EmptyState`, `Pagination`, `PublishActionPanel`.

Komponen dipisah bila dipakai lintas halaman atau memiliki perilaku kompleks. Hindari component fragmentation.

## 3. Form standards

Label terlihat; format input dijelaskan; client validation untuk UX dan server validation otoritatif; input dipertahankan saat error; locale ditandai; destructive action memakai confirmation; loading mencegah double submit; error accessible untuk screen reader.

## 4. Content rendering safety

Rich text disimpan sebagai structured document JSON. Editor membatasi node/mark. Backend memvalidasi dan sanitasi. Renderer hanya mengeluarkan allowlist. Jangan mengeksekusi HTML, script, inline event handler, atau embed tidak dikenal. External links memakai rel yang aman.

## 5. Image handling

Main photo satu per content pada MVP; alt text per locale; rights/source metadata wajib; responsive variants bila disepakati; hero image memiliki loading priority; gambar lain lazy-load; set width/height/aspect ratio; fallback visual jika unavailable; jangan gunakan aset tanpa hak jelas.

## 6. Testing

Unit tests composables/utilities/locale routing; component tests editor tabs, publish action, filters, pagination; API client success/problem tests; ID/EN rendering tests; accessibility smoke tests; visual regression bila biaya maintenance wajar; E2E login → draft → save ID/EN → preview → publish → verify SSG.

## 7. Quality gates

Build/PR wajib melewati TypeScript typecheck, ESLint, unit/component tests, production build, SSG output validation, secret scan, basic accessibility checks, responsive smoke test. Jangan menetapkan target Lighthouse angka sebelum baseline staging diukur.

## 8. Definition of Done

Feature sesuai IA/PRD; ID dan EN tidak tercampur; loading/empty/error tersedia; responsive; keyboard-accessible; API errors ditangani; draft tidak bocor ke public; tests lulus; dokumentasi diperbarui bila kontrak berubah.

---
