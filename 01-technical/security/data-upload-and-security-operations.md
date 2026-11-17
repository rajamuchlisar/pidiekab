# Data Upload And Security Operations

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `database-and-secrets-hardening.md`
- `media-upload-and-storage-security.md`
- `security-testing-and-incident-response.md`

---

## Materi sumber: `database-and-secrets-hardening.md`

# Database & Secrets Hardening — PidieKab

**Status:** Baseline implementation specification

## 1. PostgreSQL access

- PostgreSQL tidak listen pada public interface.
- Bila colocated, bind ke loopback/private socket.
- Bila remote, gunakan private network dan TLS.
- Firewall hanya mengizinkan application host dan DBA access yang dibutuhkan.
- Gunakan dedicated application database role, bukan PostgreSQL superuser.
- Role aplikasi hanya mendapat privilege yang dibutuhkan pada database/schema.
- Migration role dipisahkan dari runtime role bila operasional memungkinkan.
- Nonaktifkan/default-account access yang tidak dibutuhkan dan gunakan strong credential.

## 2. SQL safety

- Semua SQL parameterized.
- Runtime role tidak memiliki superuser/CREATEDB/CREATEROLE.
- DDL dilakukan oleh migration deployment step, bukan request runtime.
- Database errors dicatat server-side dengan sanitasi.
- Gunakan transaction boundaries untuk publish, audit, dan enqueue SSG job.

## 3. Transport and encryption

- TLS untuk koneksi PostgreSQL lintas host.
- Certificate validation aktif; jangan menggunakan `sslmode=disable` untuk koneksi remote production.
- Backup dienkripsi sebelum dikirim ke off-host destination.
- Secrets tidak dikirim lewat URL/query string.

## 4. Secret management

Secrets minimum:
- database credential;
- session/CSRF signing secret bila digunakan;
- TLS private key;
- DNS provider/ACME API token;
- backup encryption key;
- email provider credential bila fitur email ditambahkan.

Simpan pada root-owned configuration atau systemd credentials dengan permission ketat. Jangan:
- commit ke Git;
- menaruh secret di `.env.example`;
- memasukkan ke frontend build;
- menulis ke logs;
- menaruh pada screenshot/documentation publik;
- memasukkan ke shell history atau process arguments.

## 5. Rotation

Dokumentasikan prosedur rotasi:
- database password: update role/secret dengan overlap terkontrol;
- session/CSRF secret: rencanakan invalidation/rotation tanpa membuat CSRF bypass;
- TLS certificate: automated renewal/reload;
- DNS API token: revoke old token setelah new token diuji;
- backup encryption key: key versioning dan restore test dengan key lama.

## 6. Database backups

Ikuti backup/recovery spec. Pastikan backup credential berbeda dari runtime database role dan backup destination tidak dapat dihapus oleh application process.

---

## Materi sumber: `media-upload-and-storage-security.md`

# Media Upload & Storage Security — PidieKab

**Status:** Baseline implementation specification

## 1. Upload policy

MVP hanya menerima image format yang diperlukan website, misalnya JPEG, PNG, dan WebP. SVG tidak diterima pada baseline karena dapat memuat active content; jika kelak diperlukan, harus ada sanitization pipeline khusus.

Batas ukuran file dan dimensi maksimum ditetapkan melalui konfigurasi dan diuji pada server. Batas awal yang diusulkan: 10 MB per image, dengan batas pixel/dimensi untuk mencegah decompression/resource exhaustion. Angka final dikonfirmasi setelah kebutuhan editorial diketahui.

## 2. Validation pipeline

1. Authenticate Admin dan validate CSRF.
2. Enforce request body limit.
3. Validate extension allowlist.
4. Detect MIME dari file content, bukan hanya browser header.
5. Decode image menggunakan library yang terpelihara.
6. Validasi dimensi, pixel count, dan file integrity.
7. Re-encode image ke format yang diizinkan bila memungkinkan untuk menghapus metadata/active payload yang tidak diperlukan.
8. Generate random storage key; jangan memakai user filename sebagai path.
9. Simpan file pada storage non-executable.
10. Simpan metadata/rights/alt text pada database.
11. Audit upload event.

Jangan menjalankan uploaded file sebagai program. Jangan menaruh upload di direktori executable/public application root.

## 3. Storage paths

- Pisahkan media originals/derivatives dari app release.
- Gunakan path yang dibangun dari generated storage key.
- Resolve path dan pastikan tetap berada di storage root.
- Reject traversal, absolute path, encoded separators, dan symlink escape.
- Directory listing dinonaktifkan.
- Permission file minimal yang dibutuhkan Go service.
- Backup media dilakukan terpisah dan konsisten dengan DB.

## 4. Public media delivery

Media yang digunakan pada Published content dapat disajikan melalui controlled static handler atau path media public yang tidak mengeksekusi file. Gunakan content type yang tervalidasi, `X-Content-Type-Options: nosniff`, dan cache policy berdasarkan versioned URL.

Draft-only media tidak boleh bocor melalui public URL yang mudah ditebak. Jika media preview diperlukan, gunakan authenticated handler atau token terbatas yang tidak dapat ditebak dan memiliki expiry.

## 5. Rights and attribution

Media metadata mencakup source/origin, license/permission, attribution, dan alt text ID/EN. Publish validation memastikan informasi hak penggunaan yang diwajibkan tersedia. Informasi sumber internal yang ditandai private tidak ikut ditampilkan di public page.

## 6. Delete and retention

- Media in-use tidak dapat dihapus.
- Menghapus content tidak otomatis menghapus shared media.
- File orphan dapat dilaporkan untuk pemeriksaan Admin; jangan auto-delete pada MVP.
- Permanent delete media harus memeriksa referensi database dan mencatat audit.

---

## Materi sumber: `security-testing-and-incident-response.md`

# Security Testing & Incident Response — PidieKab

**Status:** Baseline pre-production checklist

## 1. Security test cases

### Authentication/session
- invalid credentials receive generic response;
- inactive Admin cannot login;
- expired/revoked session rejected;
- logout invalidates session;
- password reset/deactivation revokes sessions;
- session token is absent from JSON/localStorage/logs;
- login throttling works without permanent account lockout.

### CSRF/CORS
- mutation without CSRF token rejected;
- invalid Origin rejected where policy requires;
- cross-origin credentialed request not allowed;
- GET request cannot mutate state.

### Authorization/content privacy
- anonymous cannot access Admin API;
- anonymous cannot access preview;
- Draft/Archived locale absent from SSG, sitemap, search, and public routes;
- unpublished EN does not fall back to ID content;
- category inactive content is not listed publicly.

### XSS/input
- script tags, event handlers, javascript URLs, unknown rich-text nodes rejected;
- reflected values safely escaped;
- slug and redirect edge cases cannot create loops/open redirects;
- SQL injection payloads are treated as data.

### Upload
- oversized file rejected;
- MIME spoofing rejected;
- SVG/HTML/executable rejected;
- malformed image rejected;
- path traversal rejected;
- media in use cannot be deleted;
- draft-only media is not publicly accessible.

### Infrastructure
- PostgreSQL not reachable from public internet;
- Go service runs non-root;
- secrets have restrictive permissions;
- TLS certificate valid and renewal tested;
- backup restore drill passes;
- logs do not contain passwords, cookies, or secrets.

## 2. Automated checks

- Go unit/integration tests.
- Frontend typecheck/lint/tests.
- Dependency vulnerability review.
- Static analysis (`go vet` and selected security linting).
- Secret scanning in repository and CI.
- DAST smoke test against staging.
- Manual review of authentication, rich-text renderer, upload, and publish flows.

Use automated scanners as assistance, not proof of security.

## 3. Incident response baseline

1. Identify affected asset and time window.
2. Preserve relevant logs and evidence.
3. Contain access: revoke sessions/tokens, disable compromised account, block abusive source where appropriate.
4. Rotate affected secrets/credentials.
5. Restore trusted application state or patch vulnerability.
6. Verify database/media integrity and public output.
7. Restore from backup only after assessing scope and recovery point.
8. Document root cause, impact, actions, and prevention.

Do not delete logs or overwrite evidence during initial investigation.

## 4. Security release gate

Production release is blocked if any of the following remain:
- known critical/high exploitable vulnerability without mitigation;
- Admin authentication or CSRF bypass;
- draft/private content exposure;
- stored XSS in published content;
- arbitrary file upload/path traversal;
- public database exposure;
- unprotected secret/private key;
- untested backup restore procedure.

## 5. Deferred controls

2FA, WAF, centralized SIEM, external secrets manager, distributed rate limiter, malware scanning service, and advanced anomaly detection may be evaluated later based on operational risk. They are not prerequisites to introduce unnecessary complexity in the MVP, but high-risk exposure may make them necessary before launch.

---
