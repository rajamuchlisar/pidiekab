# Implementation Roadmap And Backlog

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `implementation-roadmap.md`
- `implementation-backlog-and-task-template.md`
- `backend-implementation-backlog.md`

---

## Materi sumber: `implementation-roadmap.md`

# PidieKab Implementation Roadmap

**Status:** Baseline execution plan  
**Product:** PidieKab — Independent Regional Promotion & Visitor Discovery Platform  
**Implementation style:** Sequential, test-gated delivery with a coding agent  
**Architecture baseline:** Go Modular Monolith, PostgreSQL, Vue 3 + TypeScript, public SSG, dynamic Admin CMS

## 1. Delivery principles

1. Build the smallest complete vertical slice before expanding modules.
2. Follow locked product and architecture decisions in the existing Discovery, PRD, Technical, API, Security, QA, and Observability documents.
3. One Go codebase, one binary, one `net/http` server, one PostgreSQL database.
4. No microservices, Redis, message broker, ORM, reverse proxy, or heavyweight backend framework in MVP.
5. Implement only Phase 1 content categories: Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah.
6. ID and EN are independent locale versions from MVP.
7. Every publish operation must preserve source requirements, locale isolation, audit, and SSG job atomicity.
8. Each phase ends with demonstrable acceptance evidence; do not move forward with unresolved release-blocking defects.

## 2. Delivery phases

| Phase | Focus | Main outcome | Exit gate |
|---|---|---|---|
| 0 | Repository & environment | Reproducible local development and CI | Clean clone builds and tests |
| 1 | Database & platform foundation | PostgreSQL migrations, config, logging, HTTP server | Migrations and health endpoints pass |
| 2 | Identity & Admin access | Bootstrap Admin, login, sessions, CSRF | Security tests pass |
| 3 | Content & category core | Bilingual content, sources, category and location data | CRUD and validation pass |
| 4 | Media & editorial CMS | Secure media, editor, preview, static pages/settings | CMS workflows pass |
| 5 | Publication & SSG | Publish/archive, job queue, atomic static release | Locale/privacy/rollback tests pass |
| 6 | Public discovery | Homepage, category, detail, search, SEO | Public acceptance tests pass |
| 7 | Curation & content operations | Featured/Latest/Random and Stories | Curation rules pass |
| 8 | Hardening & production readiness | Security, QA, monitoring, backup, deployment | Go-live checklist approved |
| 9 | Launch & stabilization | Production release and defect stabilization | Stable operations and documented handover |

Phase 7 may overlap with Phase 6 only when the public page contracts and publication pipeline are stable. Do not begin Phase 8 as a substitute for missing functional acceptance.

## 3. Phase details

### Phase 0 — Repository & development foundation
- Create repository and Go module.
- Create public-site and admin-cms Vue/Vite applications.
- Pin Go, Node, package manager, and tool versions.
- Add formatting, lint, typecheck, test, and build commands.
- Add `.gitignore`, `.env.example` with placeholders, README developer setup.
- Add CI baseline and secret scanning.

**Exit:** a clean clone can install dependencies, run checks, and build without manual undocumented steps.

### Phase 1 — Database & platform foundation
- Implement config loading and validation.
- Implement `pgxpool` lifecycle and readiness check.
- Add Goose migration execution as deployment operation.
- Implement `slog` JSON logging, request ID, recovery, body limits, security headers.
- Add `net/http` routing and graceful shutdown.
- Implement liveness/readiness endpoints.
- Apply initial schema only after reviewing migration baseline.

**Exit:** migrations work on empty and upgrade test databases; service starts, reports health, shuts down cleanly.

### Phase 2 — Identity & Admin access
- Implement Admin repository/application logic.
- Implement secure password hashing and bootstrap command.
- Implement opaque random session token, hash-at-rest, cookie handling.
- Implement login/logout/session validation and CSRF.
- Implement rate limiting and session revocation on deactivation/password reset.
- Add audit events for security-relevant actions.

**Exit:** authentication, CSRF, session expiry/revocation, and inactive Admin tests pass. No public registration.

### Phase 3 — Content & category core
- Implement category and category translation CRUD.
- Implement content identity and ID/EN translation CRUD.
- Implement source/reference CRUD and publish eligibility validation.
- Implement location master-data read/select and optional coordinates.
- Implement slug validation and uniqueness.
- Implement independent locale statuses.
- Implement audit trail for relevant mutations.

**Exit:** Admin can manage bilingual Draft content and sources; all locale and category invariants pass.

### Phase 4 — Media & editorial CMS
- Implement upload validation and safe storage.
- Implement shared Media Library metadata and rights fields.
- Implement structured rich-text editor and server renderer/sanitizer.
- Implement content preview without publishing or public indexing.
- Implement static pages and website settings.
- Implement Admin management and self-lockout protections.
- Implement CMS list/detail/edit UX and form state preservation.

**Exit:** complete Admin content workflow works in staging; security upload/XSS tests pass.

### Phase 5 — Publication & SSG
- Implement transactional publish/archive operations.
- In the same DB transaction, update locale state, add audit event, and enqueue generation job.
- Implement PostgreSQL job claiming with `FOR UPDATE SKIP LOCKED`.
- Implement idempotent worker, retry, failure state, and bounded recovery.
- Generate into temporary release directory.
- Validate output, then atomically activate release.
- Preserve current active release on build failure.
- Implement redirects, sitemap, canonical, hreflang, and noindex preview behavior.

**Exit:** ID/EN publication independence, Draft privacy, redirect behavior, and atomic release tests pass.

### Phase 6 — Public discovery
- Implement public SSG routes for homepage, category, detail, and static pages.
- Implement responsive header, navigation, language switcher, hero, category cards, footer.
- Implement detail page, source presentation, optional map and location.
- Implement public search API and results page.
- Implement locale-unavailable behavior without silent fallback.
- Implement SEO metadata and structured route output.

**Exit:** public visitor journeys and SEO/locale acceptance tests pass.

### Phase 7 — Editorial curation
- Implement manual Featured selection.
- Implement Lead Story and Supporting Stories per locale with manual order.
- Implement Latest and Random selection rules.
- Ensure exclusions and duplicate prevention.
- Expose curation controls in CMS and render on homepage.

**Exit:** curation rules, ordering, and locale publication constraints pass.

### Phase 8 — Hardening & production readiness
- Complete security review and threat-based tests.
- Run backend/frontend tests and E2E acceptance suite.
- Run representative load tests and record baseline.
- Configure TLS provisioning/renewal and test renewal/reload.
- Configure systemd service, firewall, private PostgreSQL access, permissions.
- Configure encrypted off-host DB/media backups and perform restore drill.
- Configure logs, health checks, metrics, alerts, disk/backup monitoring.
- Verify rollback and incident runbooks.

**Exit:** all go-live blockers cleared and accountable owner approves production readiness.

### Phase 9 — Launch & stabilization
- Deploy immutable release and run smoke tests.
- Verify TLS, DNS, public pages, Admin login, search, publishing, backup.
- Monitor errors, queue age, disk, DB pool, and certificate status.
- Triage defects; avoid adding Phase 2 scope during stabilization.
- Document operational handover and known limitations.

**Exit:** agreed stabilization window completed with no unresolved blocker/critical defects.

## 4. Cross-phase requirements

- Every API endpoint updates OpenAPI and contract tests.
- Every schema change includes forward migration and upgrade test.
- Every security-sensitive change includes negative tests.
- Every public rendering change checks locale and publication visibility.
- Every publish-related change checks audit + generation job atomicity.
- Every frontend workflow has loading, empty, validation, unauthorized, and error states.
- Documentation changes accompany implementation changes.

## 5. Explicitly out of MVP

Do not implement without a new approved scope decision:
- Akomodasi & Fasilitas;
- Panduan Wisata;
- Berita & Agenda Kegiatan;
- public operator submissions or self-service listings;
- visitor accounts, favorites, ratings, comments, or messaging;
- visitor analytics/personalization;
- AI search or semantic ranking;
- PostGIS/geospatial search;
- microservices, Redis, message broker, Kubernetes, or reverse proxy.

## 6. First executable milestone

Start with Phase 0 and Phase 1 only. Do not ask an AI coding agent to generate the entire application in one prompt. First request repository inspection and a concise implementation plan, then execute one bounded task at a time with tests and review.

---

## Materi sumber: `implementation-backlog-and-task-template.md`

# Implementation Backlog & Task Template — PidieKab

## 1. Backlog hierarchy

Use this hierarchy:
- **Phase:** delivery stage from implementation roadmap.
- **Epic:** cohesive capability (e.g. Admin Authentication).
- **Task:** one bounded implementation unit.
- **Subtask:** optional breakdown when a task is still too large.
- **Acceptance scenario:** observable behavior linked to QA catalog.

A task should generally be small enough to implement, test, and review in one focused agent session.

## 2. Initial prioritized backlog

### Phase 0 — Repository foundation

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P0-001 | Inspect docs and produce repository implementation plan | None | Plan lists constraints, structure, risks |
| P0-002 | Initialize Go module and command entrypoint | P0-001 | `go test ./...` succeeds |
| P0-003 | Initialize public-site Vue/Vite/TS app | P0-001 | Typecheck and production build pass |
| P0-004 | Initialize admin-cms Vue/Vite/TS app | P0-001 | Typecheck and production build pass |
| P0-005 | Add formatting/lint/test scripts and pinned tool versions | P0-002–004 | Reproducible checks documented |
| P0-006 | Add CI baseline and secret scanning | P0-005 | CI passes on clean checkout |

### Phase 1 — Platform and database

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P1-001 | Implement config loading and validation | P0-002 | Missing/invalid config fails safely |
| P1-002 | Implement slog JSON logger | P1-001 | Structured logs and redaction tests |
| P1-003 | Implement pgxpool lifecycle | P1-001 | Connect/ping/close tests |
| P1-004 | Add Goose migration runner command | P1-003 | Fresh and upgrade migration tests |
| P1-005 | Implement HTTP server bootstrap and graceful shutdown | P1-001–003 | Startup/shutdown integration test |
| P1-006 | Implement request ID/recovery/body-limit/security middleware | P1-002, P1-005 | Middleware tests pass |
| P1-007 | Implement liveness/readiness endpoints | P1-003, P1-005 | Correct dependency semantics |
| P1-008 | Apply and validate initial schema baseline | P1-004 | Constraints and migration tests pass |

### Phase 2 — Identity

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P2-001 | Admin repository and account model | P1-008 | Repository integration tests |
| P2-002 | Password hashing and bootstrap command | P2-001 | No secret in logs; hash verify works |
| P2-003 | Login/logout/session repository | P2-002 | Token hash-at-rest and revoke tests |
| P2-004 | Session authentication middleware | P2-003, P1-006 | 401/active/inactive behavior |
| P2-005 | CSRF middleware and token lifecycle | P2-004 | Mutation without token rejected |
| P2-006 | Login throttling and security audit events | P2-004 | Abuse tests and audit assertions |
| P2-007 | Admin login UI and session UX | P2-004–005 | E2E login/logout |

### Phase 3 — Category/content/location

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P3-001 | Category repository and translations | P2-004 | ID/EN CRUD and uniqueness |
| P3-002 | Category Admin API/UI | P3-001 | Active/inactive constraints |
| P3-003 | Content identity and translation repository | P3-001 | One identity, locale rows |
| P3-004 | Content CRUD API and validation | P3-003 | Locale statuses independent |
| P3-005 | Source/reference CRUD and publish eligibility | P3-003 | At least one source required |
| P3-006 | Location master data read/select | P1-008 | Valid hierarchy and optional coordinates |
| P3-007 | Content list/editor CMS | P3-004–006 | Unsaved locale edits preserved |

### Phase 4 — Media and CMS

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P4-001 | Media storage abstraction and safe path handling | P1-001 | Traversal tests pass |
| P4-002 | Image upload validation pipeline | P4-001 | MIME spoof/oversize rejected |
| P4-003 | Media metadata/rights API and UI | P4-002 | Rights fields persisted |
| P4-004 | Structured rich-text schema and renderer | P3-004 | Allowlist/XSS tests pass |
| P4-005 | Draft preview route | P4-004, P2-004 | Authenticated, noindex, no publish |
| P4-006 | Static pages and translations | P3-001 | Independent locale status |
| P4-007 | Website settings and Admin management | P2-004 | Authorized CRUD and audit |

### Phase 5 — Publication and SSG

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P5-001 | Publication application service | P3-004–005 | Publish validation rules |
| P5-002 | Transactional audit + SSG job enqueue | P5-001 | Atomic DB transaction test |
| P5-003 | Job claim/retry/recovery worker | P5-002 | Concurrent claim and retry tests |
| P5-004 | SSG output generator | P4-004, P5-003 | Correct localized output |
| P5-005 | Atomic release activation/rollback safety | P5-004 | Failed build retains active release |
| P5-006 | Sitemap/canonical/hreflang/redirect generation | P5-004 | SEO and redirect acceptance |
| P5-007 | Publish/archive CMS workflow | P5-001–006 | E2E per-locale publication |

### Phase 6 — Public discovery

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P6-001 | Public site shell and locale routing | P0-003, P5-004 | `/` redirects to `/id/` |
| P6-002 | Homepage hero/category cards | P6-001 | Responsive and accessible |
| P6-003 | Category listing pages | P6-001, P5-004 | Active locale content only |
| P6-004 | Content detail pages | P6-001, P5-004 | Correct localized detail |
| P6-005 | Public search API | P1-008, P3-003 | Ranking/filter/pagination tests |
| P6-006 | Search results UI | P6-005 | Query/filter retained |
| P6-007 | Footer/static page rendering | P4-006, P6-001 | Disclaimer and localized pages |

### Phase 7 — Curation

| ID | Task | Dependency | Acceptance |
|---|---|---|---|
| P7-001 | Featured curation API/UI | P3-004, P5-001 | Manual ordering |
| P7-002 | Lead/Supporting Stories API/UI | P3-004, P5-001 | One Lead per locale; no duplicates |
| P7-003 | Latest/Random selection logic | P6-001 | Exclusion and determinism rules |
| P7-004 | Homepage curation rendering | P7-001–003 | Published locale only |

### Phase 8–9 — Production readiness and launch

Create focused tasks for TLS/ACME, systemd hardening, firewall, backup/restore, metrics/alerts, load test, E2E release suite, deployment rehearsal, smoke tests, rollback, and stabilization. Do not combine all production work into one oversized task.

## 3. Task template

Use this template for each agent task:

### Task ID — Short title

**Phase/Epic:**  
**Priority:**  
**Dependencies:**  
**Specification references:**  
**Files/modules likely affected:**

**Objective**  
One clear outcome.

**In scope**
- Explicit behavior to implement.

**Out of scope**
- Adjacent features that must not be added.

**Constraints**
- Relevant architecture/security/business rules.

**Acceptance criteria**
- Observable, testable outcomes.

**Required tests**
- Unit/integration/API/E2E/security as applicable.

**Documentation updates**
- API, schema, ADR, runbook, or README changes.

**Completion report**
- Files changed.
- Commands run and results.
- Acceptance evidence.
- Known limitations and follow-up.

## 4. Priority rule

Prioritize dependency blockers and critical invariants over visible UI polish. Authentication, data integrity, locale isolation, publication privacy, and atomic SSG release are foundational. Do not parallelize tasks that modify the same schema or shared interfaces without explicit coordination.

---

## Materi sumber: `backend-implementation-backlog.md`

# Backend Implementation Backlog — PidieKab

**Status:** Proposed execution order  
**Principle:** Stabilize foundation before feature breadth

## Phase A — Repository foundation

1. Initialize Git repository and Go module.
2. Set minimum supported Go version.
3. Add `.gitignore`, `.env.example`, README, Makefile.
4. Add `cmd/pidiekab/main.go` with config/logger only.
5. Add config parsing and startup validation.
6. Add PostgreSQL pool with `pgx/v5`.
7. Add Goose migration workflow; do not auto-migrate at application startup.
8. Add health/readiness endpoints.
9. Add structured logging and request ID middleware.
10. Add graceful shutdown.

**Acceptance:** Application starts locally, connects PostgreSQL, exposes health/readiness, and shuts down cleanly.

## Phase B — HTTP/API foundation

1. Add response and problem-details encoder.
2. Add standard `net/http` router and route registration.
3. Add security headers and body size limits.
4. Add API convention tests.
5. Add public/admin route groups.
6. Keep CORS disabled unless required.
7. Add basic login/search rate limiting when endpoints exist.

**Acceptance:** API payloads follow conventions; request ID appears in logs and errors.

## Phase C — Identity and Admin session

1. Implement password hashing and verification.
2. Implement login/logout/current Admin.
3. Implement opaque session token and hash-at-rest.
4. Implement session expiry/revocation and inactive Admin rejection.
5. Implement CSRF for cookie-authenticated state changes.
6. Implement Admin management safeguards and audit logging.

**Acceptance:** Admin login/logout works; inactive/expired/revoked sessions fail; secrets are never exposed.

## Phase D — Core content management

1. Implement category read and management.
2. Implement content identity and shared fields.
3. Implement ID/EN translation CRUD.
4. Implement content sources.
5. Implement media metadata and storage upload.
6. Implement permanent delete constraints and audit.
7. Implement private preview.

**Acceptance:** Admin can manage bilingual draft content; locales remain independent; media/source rules enforced.

## Phase E — Publish and SSG

1. Implement transactional publish/archive use cases.
2. Implement `site_generation_jobs` claim/lease/retry.
3. Implement static renderer and release directory.
4. Implement atomic activation and release retention.
5. Implement redirects, sitemap, hreflang, canonical.
6. Implement worker recovery and operational status.
7. Implement end-to-end publish pipeline tests.

**Acceptance:** Publish enqueues atomically; successful job updates public release; failed job leaves previous release online; only published locale pages appear.

## Phase F — Public discovery

1. Implement PostgreSQL search vector/index strategy.
2. Implement localized search endpoint/category filter.
3. Test Indonesian/English tokenization and relevance.
4. Implement homepage/category/detail SSG pages.
5. Implement Featured and Story curation.
6. Implement static pages and website settings.
7. Validate SEO output.

**Acceptance:** Search returns only published active-locale results; generated pages match IA/SEO requirements.

## Phase G — Production hardening

1. Add upload abuse protections and media rights checks.
2. Add database backup/restore drill.
3. Add migration forward-fix procedure.
4. Add TLS certificate provisioning/renewal and HTTPS enforcement.
5. Add resource limits and service restart policy.
6. Add log rotation/retention.
7. Run security review and dependency audit.
8. Load test search, static delivery, generation.
9. Document deployment and recovery runbooks.

**Acceptance:** Repeatable deployment, restore test, TLS renewal test, no secret leakage, stable MVP load.

## Explicitly deferred

Redis/message broker, microservices, Elasticsearch/AI semantic search, public accounts/user-generated content, multi-role RBAC, analytics/recommendations, complex incremental SSG invalidation, distributed rate limiting until multi-instance deployment is justified.

---
