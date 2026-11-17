# Backend Frontend And Acceptance Tests

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `backend-testing-specification.md`
- `frontend-testing-specification.md`
- `e2e-and-acceptance-test-catalog.md`

---

## Materi sumber: `backend-testing-specification.md`

# Backend Testing Specification — PidieKab

## 1. Go baseline checks

Run at minimum:
- `gofmt` validation;
- `go vet ./...`;
- `go test ./...`;
- race detector for packages where practical (`go test -race ./...`);
- dependency and vulnerability review in CI.

Exact commands and CI implementation may evolve, but checks must remain reproducible.

## 2. Unit test requirements

Test business logic independently of HTTP and database:
- locale validation (`id`, `en`);
- slug normalization and validation;
- content status transitions;
- publish eligibility: active category and at least one source;
- independent publication status per locale;
- category cannot be deleted while content references it;
- inactive category cannot receive new content;
- featured/latest/random exclusion rules;
- Lead Story and Supporting Story uniqueness/order rules;
- redirect path normalization and loop prevention;
- pagination boundaries and search query limits;
- coordinate pair/range and accuracy rules;
- audit event redaction.

Use table-driven tests for input matrices and edge cases.

## 3. Repository and PostgreSQL integration tests

Use real PostgreSQL in an isolated test environment; do not rely exclusively on mocks for SQL correctness.

Cover:
- migrations from empty database;
- unique constraints for localized slugs;
- foreign-key restrictions and cascades as specified;
- transaction rollback on publish failure;
- publish update + audit + SSG job are atomic;
- job claim with `FOR UPDATE SKIP LOCKED` under concurrent workers/claims;
- session lookup, expiry, revocation;
- source and media references;
- query filtering by locale, Published status, and Active category;
- deterministic ordering and pagination;
- search ranking and Indonesian/English tokenization behavior.

Test migration upgrade from the previous supported schema version, not only fresh install.

## 4. HTTP/API integration tests

Use `net/http/httptest` with application wiring and test database.

Verify:
- request ID propagation;
- JSON content type and error format;
- body-size limits;
- authentication and authorization;
- CSRF enforcement on mutations;
- CORS policy;
- rate limiting;
- status codes and response schema;
- no stack traces or SQL details in public errors;
- `Cache-Control: no-store` for Admin and preview;
- public endpoints never expose Draft/Archived content.

For every endpoint, include at least one success case, validation failure, unauthenticated/unauthorized case where applicable, and not-found case.

## 5. SSG and publication worker tests

- Published ID and EN routes generated independently.
- Unpublished locale is not generated and does not silently fall back.
- Draft and Archived content absent from pages, sitemap, search index, and public assets where restricted.
- Canonical, hreflang, title, description, and sitemap entries match locale publication state.
- Slug changes create correct permanent redirects.
- Category slug change/content move preserves redirects for previous paths.
- Failed generation does not replace active release.
- Successful generation activates complete release atomically.
- Duplicate/retried jobs are idempotent.
- Worker crash/restart does not leave permanent locked jobs.
- Retry count and failed state are observable.

## 6. Concurrency and resilience

Test concurrent edits/publish operations, simultaneous job claims, DB connection interruption, graceful shutdown, and bounded retry. Define optimistic concurrency/version checks if lost updates are possible; do not silently overwrite a newer edit.

---

## Materi sumber: `frontend-testing-specification.md`

# Frontend Testing Specification — PidieKab

## 1. Public site component tests

Cover:
- header navigation and mobile menu;
- ID/EN switcher and locale-specific route mapping;
- category cards and content cards;
- Featured/Latest/Random sections and duplicate exclusion;
- Lead Story and Supporting Stories ordering;
- breadcrumb and optional detail fields;
- location/map display only when valid coordinates exist;
- source/reference presentation;
- footer independent-platform disclaimer;
- empty, loading, and error states.

## 2. CMS component and workflow tests

Cover:
- login validation and generic error display;
- inactive Admin handling;
- content list filters, pagination, and independent ID/EN status;
- language-tab switching preserves unsaved edits;
- structured rich-text editor only emits allowed schema;
- preview does not publish;
- publish/archive per locale;
- source requirement before publish;
- category active/inactive constraints;
- media upload validation and in-use delete prevention;
- static page translations;
- curation selection, ordering, and no duplicates;
- Admin create/deactivate/reset flows and self-lockout prevention.

## 3. State and API behavior

- Test loading, success, empty, validation, unauthorized, forbidden, and server-error states.
- Ensure Admin session credentials are not stored in localStorage/sessionStorage.
- Ensure API errors are handled without exposing internal details.
- Verify filters/query/locale persist through pagination as specified.
- Avoid global Pinia state for server data that should be page/request scoped.

## 4. Accessibility and responsive quality

Test keyboard navigation, visible focus, form labels, accessible names, contrast, heading hierarchy, and semantic landmarks. Validate at representative mobile, tablet, and desktop widths. Do not rely on screenshots alone; assert behavior and semantics.

## 5. Build checks

- TypeScript strict typecheck.
- Lint and unit/component tests.
- Production build for both frontend apps.
- SSG output validation for expected routes and metadata.
- Check no development server, source map policy violation, test fixture, or secret is accidentally shipped.

---

## Materi sumber: `e2e-and-acceptance-test-catalog.md`

# E2E & Acceptance Test Catalog — PidieKab

**Purpose:** Trace MVP requirements to observable end-to-end outcomes. These scenarios form the initial acceptance baseline.

## A. Public visitor journeys

| ID | Scenario | Expected result |
|---|---|---|
| PUB-001 | Open `/` | Redirects to `/id/` |
| PUB-002 | Browse Indonesian homepage | Indonesian published content only |
| PUB-003 | Switch to English | Opens corresponding published EN context where available |
| PUB-004 | EN translation unavailable | Shows unavailable notice and link to available locale; no silent ID substitution |
| PUB-005 | Browse each MVP category | Only Published content in active locale and Active category |
| PUB-006 | Open content detail | Correct category path, title, body, main photo, optional location, sources |
| PUB-007 | Search by title | Relevant title matches rank above summary/body matches |
| PUB-008 | Filter search by category | Query retained; filter applied; page resets to 1 |
| PUB-009 | Paginate search | 12 results default; query/category/locale retained |
| PUB-010 | Open old slug | Permanent redirect to current canonical URL |
| PUB-011 | View homepage curation | Featured/Latest/Random follow rules; no duplicates |
| PUB-012 | View Stories section | One Lead and ordered Supporting Stories; no duplicate content |
| PUB-013 | View footer | Independent platform notice and configured contact links |

## B. Admin content workflows

| ID | Scenario | Expected result |
|---|---|---|
| ADM-001 | Anonymous opens CMS | Login required |
| ADM-002 | Inactive Admin logs in | Access denied with generic response |
| ADM-003 | Admin creates content | One content identity created with locale versions |
| ADM-004 | Admin edits ID then switches EN tab | Unsaved ID edits preserved |
| ADM-005 | Publish ID with source and active category | ID becomes public; SSG job enqueued atomically |
| ADM-006 | Publish without source | Publish blocked with actionable validation |
| ADM-007 | Publish in inactive category | Publish blocked |
| ADM-008 | Publish ID only | EN remains unpublished and is not generated |
| ADM-009 | Edit published locale without republish | Public release remains unchanged |
| ADM-010 | Archive one locale | Only that locale removed from public output |
| ADM-011 | Preview Draft | Authenticated preview; no public indexing or publication |
| ADM-012 | Permanently delete content | Confirmation required; content/locale versions removed; shared media retained |
| ADM-013 | Delete media in use | Operation rejected |
| ADM-014 | Deactivate category with linked content | Allowed only after all linked content moved |
| ADM-015 | Change published slug | 301 redirect created; no loop/chain |
| ADM-016 | Admin deactivates self/last active Admin | Prevented according to account safety rule |

## C. Reliability and security journeys

| ID | Scenario | Expected result |
|---|---|---|
| SEC-001 | Mutation without CSRF token | Rejected |
| SEC-002 | Expired/revoked session | Admin API returns 401 |
| SEC-003 | Stored XSS payload in rich text | Rejected or rendered inert by allowlist |
| SEC-004 | SQL injection payload in search/filter | Treated as data |
| SEC-005 | Upload oversized or spoofed file | Rejected |
| SEC-006 | Request path traversal in media URL | Rejected; no filesystem escape |
| SEC-007 | Request Draft by public API/URL | Not exposed |
| REL-001 | SSG build fails | Existing active release remains online |
| REL-002 | Worker restarts during processing | Job recoverable/retried without partial release |
| REL-003 | Database unavailable | Controlled error; no secret/stack trace leakage |
| REL-004 | Restore backup in isolated environment | Data and media recovery validated |

## D. Acceptance rules

A scenario passes only when:
- setup/preconditions are documented;
- steps can be repeated;
- expected result is asserted, not visually assumed;
- relevant logs contain no secrets;
- test data is cleaned or isolated;
- defects are linked to the scenario ID.

Not every scenario must be automated in the first iteration. Security, publication privacy, locale isolation, and data integrity scenarios are release-critical and should be automated wherever practical.

---
