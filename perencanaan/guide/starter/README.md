# PidieKab Starter — Phase 0

Starter ini adalah skeleton development, bukan aplikasi production-ready.

## Backend
```bash
cd backend
go run ./cmd/pidiekab
```

Default address: `127.0.0.1:8080`

## Frontend
Public site dan Admin CMS adalah workspace terpisah. Scaffold package masing-masing pada Phase 0 sesuai `repository-foundation.md`.

## Important
- `go.mod` memakai module path placeholder `example.com/pidiekab`; ganti sebelum repository remote dibuat.
- Database belum diintegrasikan.
- `/health/ready` belum menyatakan database readiness; endpoint ini masih placeholder dan harus diubah saat DB pool tersedia.
- Belum ada authentication, CMS, SSG production pipeline, TLS, atau deployment configuration.
