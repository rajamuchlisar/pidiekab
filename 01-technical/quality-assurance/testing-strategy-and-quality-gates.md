# Testing Strategy And Quality Gates

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `testing-strategy.md`
- `test-data-and-ci-quality-gates.md`

---

## Materi sumber: `testing-strategy.md`

# Testing Strategy — PidieKab

**Status:** MVP testing baseline  
**Applies to:** Go backend, PostgreSQL, Vue public SSG, Vue Admin CMS, deployment and security controls

## 1. Objectives

Testing must establish that:
- published public information is correct and locale-specific;
- Draft and Archived content never leak into public output;
- Admin workflows preserve content, source, media, and publication rules;
- API and UI contracts remain compatible;
- database constraints and transactions preserve data integrity;
- SSG generation is safe, repeatable, and atomic;
- security controls are enforced server-side;
- the application behaves correctly on supported desktop and mobile viewports.

## 2. Test layers

| Layer | Purpose | Typical tools |
|---|---|---|
| Unit | Pure business rules and validation | Go `testing`, Vue test utilities |
| Component | Vue component behavior and rendering | Vitest, Vue Test Utils |
| Repository | SQL behavior and mapping | Go tests + PostgreSQL test DB |
| Integration | Multiple modules, API, DB, worker | `httptest`, PostgreSQL |
| Contract | API request/response compatibility | OpenAPI validation and HTTP tests |
| SSG | Route/output/SEO/publication correctness | Build output assertions |
| E2E | Complete Admin and visitor journeys | Playwright or equivalent |
| Security | Abuse, access control, injection, upload | Automated + manual tests |
| Acceptance | Product requirement validation | Requirement-linked scenarios |
| Operational | Deployment, backup, restore, rollback | Staging runbooks |

Testing libraries are recommendations; keep the toolchain small and avoid introducing overlapping frameworks without need.

## 3. Test environments

- Unit tests must not require production credentials or external services.
- Integration tests use an isolated PostgreSQL database with the same supported major version as production.
- Never run destructive tests against production.
- E2E tests use a dedicated staging/test deployment and seeded synthetic data.
- External links and third-party websites are not dependencies for core test success.
- Test data must not contain real visitor or Admin credentials.
- CI must provision dependencies deterministically and clean up test state.

## 4. Test data strategy

Maintain deterministic fixtures for:
- one active Admin and one inactive Admin;
- all three MVP categories with ID/EN translations;
- content with ID and EN both published;
- content with only ID published;
- Draft, Published, and Archived translations;
- active and inactive categories;
- content with and without optional location;
- source references and media with rights metadata;
- featured/latest/random and Lead/Supporting curation;
- old/new slugs and redirect history.

Use synthetic Pidie-related examples, not copied personal data. Tests must explicitly cover empty, boundary, invalid, and duplicate values.

## 5. Quality gates

A change is ready to merge only when:
1. formatting, lint, typecheck, and static analysis pass;
2. relevant unit and integration tests pass;
3. database migrations apply cleanly to an empty test database;
4. API contract tests pass for changed endpoints;
5. security-sensitive changes have negative tests;
6. frontend build and SSG output checks pass when relevant;
7. no secrets or unintended generated artifacts are committed.

A production release additionally requires E2E smoke tests, migration/rollback plan, backup readiness, and operational health checks.

## 6. Test pyramid and scope control

Prefer many fast unit tests, a focused set of PostgreSQL integration tests, and a smaller number of stable E2E tests. Do not duplicate every business rule at every layer. Critical user journeys and security boundaries deserve cross-layer coverage.

## 7. Failure handling

- Tests must fail loudly; do not silently skip required assertions.
- Flaky tests are defects: isolate root cause and fix or quarantine with an owner and expiry.
- Do not make a test pass by weakening a product requirement without an approved decision.
- Record failing test, reproduction steps, environment, expected behavior, and actual behavior.

---

## Materi sumber: `test-data-and-ci-quality-gates.md`

# Test Data & CI Quality Gates — PidieKab

## 1. Test fixture conventions

- Use stable fixture identifiers, not auto-increment assumptions across test runs.
- Use synthetic email domains such as `example.test`.
- Keep ID and EN fixtures intentionally asymmetric to expose accidental fallback.
- Include boundary values: empty strings, maximum length, invalid locale, duplicate slug, missing source, inactive category, invalid coordinates, and oversized upload.
- Fixtures must be created by migrations/seeds dedicated to test environment.
- Production data must never be copied into developer or CI environments without approved anonymization.

## 2. Suggested CI stages

1. **Repository hygiene:** formatting, whitespace, secret scan.
2. **Backend static checks:** `go vet`, dependency checks.
3. **Backend unit tests:** `go test ./...`.
4. **Database integration:** start isolated PostgreSQL, migrate, run repository/API integration tests.
5. **Frontend checks:** install from lockfile, typecheck, lint, component tests.
6. **Build:** Go binary and both frontend production builds.
7. **SSG validation:** expected routes, locale metadata, sitemap and redirects.
8. **E2E smoke:** core public and Admin flows against ephemeral/staging deployment.
9. **Artifact review:** ensure no secrets, test fixtures, or unintended files included.

CI should use pinned tool versions and lockfiles. Do not make production deployment depend on a developer workstation.

## 3. Merge and release policy

- Pull requests include changed requirement IDs and tests added/updated.
- Bug fixes include a regression test when feasible.
- Database changes include migration and upgrade test.
- API changes update OpenAPI and contract tests.
- Security changes include negative/abuse tests.
- Release candidate must pass all required gates on the same commit/artifact intended for deployment.

## 4. Coverage

Do not set a single arbitrary global coverage target as a substitute for meaningful tests. Track coverage by critical package and business rule. Increase coverage for authentication, publication, locale visibility, SQL repositories, redirects, and media handling. Coverage reports identify gaps; they do not prove correctness.

## 5. Defect severity

- **Blocker:** data loss, public exposure of Draft/private content, authentication bypass, production unavailable.
- **Critical:** exploitable security flaw, broken publishing for all locales/categories, unrecoverable data integrity issue.
- **Major:** core journey materially broken with no reasonable workaround.
- **Minor:** limited UI/edge issue with workaround.
- **Trivial:** cosmetic issue without functional impact.

Release blockers and critical defects must be resolved or formally risk-accepted by the accountable owner; security/privacy exposure should not be waived casually.

---
