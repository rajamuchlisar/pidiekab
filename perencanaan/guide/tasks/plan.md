# Phase 0 Implementation Plan — PidieKab

## Overview
Phase 0 establishes the repository foundation, Go backend skeleton, Vue 3 frontend workspaces, SSG proof-of-concept, PostgreSQL Goose migration verification, and CI quality gates.

## Tasks Breakdown

### P0-01: Repository Foundation
- Scaffold directory layout matching `01-technical/development/repository-foundation.md`.
- Move and initialize `backend/`, `frontend/public-site`, `frontend/admin-cms`, `migrations/`, `deploy/`.
- Provide root `.env.example`, `.gitignore`, and updated `README.md`.

### P0-02: Go Application Skeleton
- Go module at `backend/go.mod` (Go 1.24).
- Entrypoint `backend/cmd/pidiekab/main.go`.
- Config loader `backend/internal/platform/config/config.go` with strict startup validation.
- HTTP server `backend/internal/platform/httpserver/server.go` with timeouts.
- Middleware: request ID, structured logging, basic security headers, panic recovery.
- Unit tests for config, middleware, and server.

### P0-03: Health Check Endpoints
- `GET /health/live` returning 200 OK.
- `GET /health/ready` returning 503 Service Unavailable (truthful database unready status).
- Unit tests for both endpoints.

### P0-04: Frontend Workspace
- `frontend/public-site`: Vue 3 + TypeScript strict + Vite + routing.
- `frontend/admin-cms`: Vue 3 + TypeScript strict + Vite + Vue Router + Pinia preparation.
- NPM dependency installation and verified lockfiles.
- Typecheck, lint, and build scripts.

### P0-05: SSG Proof of Concept
- Localized synthetic fixtures: `/id/`, `/en/`, `/id/wisata-destinasi/`, `/en/tourism-destinations/`, `/id/wisata-destinasi/contoh-konten/`, `/en/tourism-destinations/sample-content/`.
- Test `vite-ssg` candidate for static generation.
- Verify SEO title, description, canonical, hreflang, sitemap.xml, and draft exclusion.
- Serve generated output via Go `net/http` static file handler without Node.js runtime.
- Produce SSG evaluation and recommendation.

### P0-06: Database Migration Proof
- Adapt `001_initial_schema.sql` into Goose format with complete `-- +goose Up` and `-- +goose Down`.
- Test on clean disposable PostgreSQL database (`pidiekab_test`).
- Validate Up, Down, Up cycles, constraints, and timestamp behavior.

### P0-07: CI Baseline
- GitHub Actions workflow `.github/workflows/ci.yml`.
- Gates: Go format, vet, test, build; Frontend npm ci, typecheck, lint, build; Migration check.

### P0-08: Final Validation & Implementation Report
- Execute all tests.
- Compile comprehensive report at `docs/reports/PHASE-0-IMPLEMENTATION-REPORT.md`.
