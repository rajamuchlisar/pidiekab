# Phase 0 — Task Breakdown

**Versi:** v79

## P0-01 Repository initialization
**Deliverables:** folder layout, README, `.gitignore`, `.env.example`, contribution conventions.  
**Acceptance:** clone pada mesin bersih dan setup lokal terdokumentasi.

## P0-02 Go application skeleton
**Deliverables:** Go module, `cmd/pidiekab/main.go`, config loader, HTTP server.  
**Acceptance:** server start, request handled, graceful shutdown.

## P0-03 Health and logging
**Deliverables:** `log/slog`, request ID middleware, `/health/live`, readiness integration seam.  
**Acceptance:** liveness returns 200; no secret in logs.

## P0-04 Frontend workspace
**Deliverables:** Vue 3 + TypeScript + Vite apps for public site and Admin CMS.  
**Acceptance:** each app installs, typechecks, and builds independently.

## P0-05 SSG proof-of-concept
**Deliverables:** localized route fixtures, metadata, sitemap, route manifest, Go static serving test.  
**Acceptance:** all criteria in `ssg-proof-of-concept-plan.md` pass.

## P0-06 Database migration proof
**Deliverables:** local PostgreSQL setup instructions, Goose baseline migration test.  
**Acceptance:** clean up/down migration on disposable DB; no production DB used.

## P0-07 CI baseline
**Deliverables:** GitHub Actions workflow or equivalent CI config.  
**Acceptance:** Go test/vet and frontend typecheck/build run on pull request.

## P0-08 Phase 0 review
**Deliverables:** report, updated ADR register, Phase 0 exit checklist.  
**Acceptance:** all failures recorded; unresolved decisions remain OPEN.

## Execution rules
- One task per pull request where practical.
- Do not implement future business modules early.
- Keep dependencies minimal.
- Do not mark complete based on compilation alone.
- Every task includes test evidence and updated documentation.
