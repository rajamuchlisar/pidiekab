# Ai Coding Agent Guide And Prompt

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `ai-coding-agent-execution-contract.md`
- `first-agent-prompt.md`

---

## Materi sumber: `ai-coding-agent-execution-contract.md`

# AI Coding Agent Execution Contract — PidieKab

**Purpose:** Operating rules for any coding agent working on the PidieKab repository.

## 1. Source of truth

Before coding, inspect the repository documentation. Primary sources:
- Product Discovery and PRD: product purpose, audience, scope, editorial rules.
- Architecture Decision Records: locked stack and deployment constraints.
- Database schema and migrations: data model and integrity rules.
- API conventions/OpenAPI: endpoint contracts.
- Security specification: authentication, CSRF, upload, secrets, database access.
- QA specification: required tests and acceptance scenarios.
- Observability specification: logs, health, metrics, alerting.
- Implementation roadmap: phase order and exit gates.

If documents conflict, do not silently choose. Report the exact conflict and propose the smallest resolution. Newer explicit decision notes in README supersede earlier drafts only when they clearly state the decision.

## 2. Non-negotiable architecture constraints

- Backend Go.
- PostgreSQL with `pgx/v5` and explicit SQL; no ORM.
- Go standard library `net/http`/`ServeMux` as the HTTP server.
- Vue 3 + TypeScript + Vite.
- Separate `frontend/public-site` and `frontend/admin-cms`.
- Public site uses SSG; CMS is dynamic SPA.
- One Go module, one binary, one deployable service.
- Pragmatic Modular Monolith.
- No Nginx/Caddy/Apache/LiteSpeed or other application reverse proxy.
- No Redis, message broker, microservices, or Kubernetes in MVP.
- Node.js is build/development tooling, not production static server.
- Do not add dependencies unless the benefit is concrete and documented.

## 3. Scope constraints

- MVP categories: Wisata & Destinasi, Kuliner & UMKM, Budaya & Sejarah.
- ID/EN locale versions are independent.
- Only one Admin role in MVP.
- No public registration or public listing submission.
- No visitor login, comments, ratings, favorites, or in-app messaging.
- No verification/approval workflow.
- Content statuses are Draft, Published, Archived.
- Every publishable content item needs at least one traceable source.
- No silent locale fallback.
- No automatic invented coordinates or factual claims.
- Do not implement Phase 2 features unless explicitly authorized.

## 4. Task execution protocol

For each task:
1. Read the relevant specification and inspect current code.
2. Summarize the intended change and files/modules affected.
3. Identify assumptions, risks, and dependencies.
4. Implement only the task scope.
5. Add or update tests.
6. Run formatting, static checks, relevant tests, and build.
7. Review diff for security, scope creep, generated artifacts, and architecture drift.
8. Update docs/API schema/migrations if applicable.
9. Report exact files changed, commands/tests run, outcomes, and unresolved issues.

Do not modify unrelated files to make a broad cleanup while executing a narrow task.

## 5. Database rules

- SQL migrations are versioned and reviewed.
- Never auto-run production migrations at every app startup.
- Use transactions for multi-step state changes.
- Respect foreign keys, unique constraints, and locale invariants.
- Use parameterized SQL.
- Add indexes based on query patterns and measured plans.
- Never drop/rename data destructively without explicit migration and backup plan.
- Do not edit an already-applied migration; create a new migration.

## 6. Security rules

- Never commit secrets or real credentials.
- Never log passwords, hashes, cookies, session tokens, CSRF tokens, or private keys.
- Never place session tokens in localStorage/sessionStorage.
- Enforce authentication, authorization, and CSRF server-side.
- Validate and sanitize rich text server-side.
- Validate uploaded file content and storage paths.
- Do not expose Draft, Archived, preview, or private media to public endpoints.
- Do not weaken security tests to make implementation pass.

## 7. Implementation quality

- Prefer readable, idiomatic Go and TypeScript.
- Keep modules cohesive and avoid speculative abstractions.
- Keep business rules out of HTTP handlers and Vue presentation components.
- Do not introduce generic frameworks or code generators without approval.
- Avoid giant files and giant functions; split by actual responsibility.
- Use explicit errors and stable API error contracts.
- Handle cancellation, timeouts, and graceful shutdown.
- Keep frontend accessibility and responsive behavior in scope.

## 8. Stop conditions

Stop and ask for a decision rather than guessing when:
- requirements conflict;
- a schema change would alter locked business rules;
- a new external service/provider is required;
- TLS/ACME provider credentials or production secrets are needed;
- destructive operation risks existing data;
- scope requires Phase 2 functionality;
- a dependency would materially change architecture;
- source documentation is insufficient to determine safe behavior.

For non-blocking ambiguity, state a conservative assumption and isolate it so it can be changed safely.

## 9. Definition of Done

A task is done only when:
- requested behavior is implemented;
- relevant tests are added and passing;
- lint/typecheck/build/static checks pass as applicable;
- API/schema/docs are updated where applicable;
- no known security or data-integrity regression is introduced;
- diff contains no unrelated changes or secrets;
- agent reports evidence and limitations.

---

## Materi sumber: `first-agent-prompt.md`

# First Coding Agent Prompt — PidieKab Repository Assessment

You are the coding agent assigned to PidieKab. This is an assessment/planning task only. Do not implement application features yet.

## Context

PidieKab is an independent regional promotion and visitor-discovery platform for Kabupaten Pidie, Aceh. MVP content categories are Wisata & Destinasi, Kuliner & UMKM, and Budaya & Sejarah. Indonesian and English locale versions are independent.

## Mandatory architecture

- Go backend, PostgreSQL, Vue 3 + TypeScript + Vite.
- Public website uses SSG; Admin CMS is a dynamic SPA.
- One Go module, one binary, one `net/http` server, one PostgreSQL database.
- Pragmatic Modular Monolith.
- `pgx/v5`, explicit SQL, Goose migrations, `log/slog`.
- No ORM, Redis, message broker, microservices, reverse proxy, or Kubernetes in MVP.
- Go serves API and static files; Node is build/dev only.
- Only one Admin role; no public registration or public listing submissions.

## Your task

1. Inspect the repository and all relevant project documentation.
2. Identify existing code, missing foundations, conflicting requirements, and risks.
3. Verify that the proposed repository structure is practical and consistent.
4. Produce a proposed Phase 0 implementation plan with exact file/folder changes.
5. Identify the smallest first implementation task and its tests.
6. Do not create large scaffolding or speculative abstractions.
7. Do not change architecture or product scope.

## Required output

Return:
- repository state summary;
- documentation references consulted;
- architecture constraint checklist;
- gap/risk list, ranked;
- proposed Phase 0 task sequence;
- exact proposed files/folders;
- first bounded implementation task;
- commands/tests that will prove it works;
- questions only where a decision is genuinely required.

Do not implement until the repository owner approves the assessment and first task. Never invent that a command or test passed; report only commands actually executed.

---
