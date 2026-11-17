# PidieKab — Documentation Guide

Dokumentasi dikelompokkan berdasarkan topik besar agar lebih mudah dibaca dan dipelihara.

## 00 — Product & Discovery
1. `00-discovery/01-product-overview-and-scope.md` — visi, audience, MVP, scope, sumber, referensi visual.
2. `00-discovery/02-content-model-and-governance.md` — kategori, relasi, bilingual fields, dan content rules.
3. `00-discovery/03-content-publication-and-lifecycle.md` — Draft/Published/Archived, preview, publish, delete.
4. `00-discovery/04-public-website-experience.md` — public IA, homepage, navigation, detail, footer.
5. `00-discovery/05-multilingual-url-and-seo.md` — ID/EN, locale behavior, URL, slug, SEO.
6. `00-discovery/06-search-and-discovery.md` — global search, filters, ranking, pagination.
7. `00-discovery/07-location-and-map.md` — wilayah administratif, koordinat, map.
8. `00-discovery/08-cms-and-administration.md` — Admin, CMS navigation, dashboard, media, settings, curation.

## 01 — Technical
- `01-technical/architecture-baseline.md` — keputusan stack dan arsitektur lintas platform.
- `01-technical/database/database-design-and-migration-rules.md` — schema, relasi, migration, integrity.
- `01-technical/database/migrations/001_initial_schema.sql` — baseline SQL DDL.
- `01-technical/api/api-specification.md` — API conventions dan endpoint contracts.
- `01-technical/api/openapi.yaml` — machine-readable API contract.
- Folder `backend`, `frontend`, `deployment`, `security`, `quality-assurance`, dan `observability` berisi spesifikasi gabungan per area.

## 02 — Implementation
- `02-implementation/implementation-roadmap-and-backlog.md` — phases, task backlog, dependencies, task template.
- `02-implementation/ai-coding-agent-guide-and-prompt.md` — execution contract dan prompt assessment awal.

## Cara membaca
Mulai dari README dan Product Overview. Gunakan Discovery untuk keputusan produk, Technical untuk implementasi, dan Implementation untuk urutan kerja. File SQL dan OpenAPI tetap terpisah karena merupakan artefak yang dibaca tool. Dokumen gabungan mencantumkan nama sumber agar keputusan lama tetap dapat ditelusuri.


## 03-governance — Final Review & Architecture Lock (v78)

Dokumen governance untuk memastikan kesiapan implementasi dan menjaga keputusan tetap konsisten.

| Dokumen | Kegunaan |
|---|---|
| `final-documentation-review.md` | Hasil audit konsistensi v77, risiko, dan rekomendasi Conditional Go |
| `architecture-decision-register.md` | Register keputusan LOCKED, PROPOSED, dan OPEN |
| `open-decisions-and-recommendations.md` | Daftar keputusan yang belum final, rekomendasi, dan deadline fase |
| `final-mvp-scope.md` | Batas fitur MVP dan Definition of MVP Complete |
| `implementation-readiness-checklist.md` | Checklist Phase 0, database, security, publication, public site, dan production |

**Aturan:** keputusan PROPOSED/OPEN tidak boleh dianggap final tanpa persetujuan. Gunakan register ini saat memperbarui PRD, schema, API, atau implementation backlog.


## Phase 0 — Repository & Development Foundation (v79)

| Dokumen | Kegunaan |
|---|---|
| `03-governance/phase-0-execution-plan.md` | Urutan kerja dan Definition of Done Phase 0 |
| `01-technical/development/repository-foundation.md` | Struktur repository dan standar development |
| `01-technical/frontend/ssg-poc/ssg-proof-of-concept-plan.md` | Rencana pembuktian kandidat SSG |
| `02-implementation/phase-0/phase-0-task-breakdown.md` | Task breakdown P0-01 sampai P0-08 |
| `starter/` | Skeleton Go awal dan catatan frontend workspace |

`vite-ssg` hanya kandidat POC pada v79; belum menjadi keputusan arsitektur LOCKED.
