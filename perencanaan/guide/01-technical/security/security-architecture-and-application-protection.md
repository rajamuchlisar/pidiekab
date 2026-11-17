# Security Architecture And Application Protection

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `security-architecture.md`
- `web-application-security.md`

---

## Materi sumber: `security-architecture.md`

# Security Architecture — PidieKab

**Status:** Baseline security specification  
**Scope:** Public website, Admin CMS, Go API, PostgreSQL, media storage, deployment  
**Principle:** Secure defaults, minimal attack surface, no unnecessary infrastructure

## 1. Security objectives

1. Draft, archived, private, and unpublished locale content tidak boleh dapat diakses publik.
2. Hanya Admin aktif yang dapat mengakses CMS dan Admin API.
3. Session dan password tidak boleh terekspos ke browser storage, logs, atau API response.
4. Konten rich text dan file upload tidak boleh menjadi jalur stored XSS atau malware delivery.
5. Mutasi penting harus tervalidasi, transactional, dan tercatat pada audit log.
6. Database, media, backup, TLS private key, dan application secrets harus memiliki akses minimum.
7. Kegagalan generation tidak boleh menghapus public release terakhir yang valid.

## 2. Threat model ringkas

| Asset | Threat | Baseline mitigation |
|---|---|---|
| Admin account | Credential theft, brute force | Password hashing, rate limit, generic login errors, session controls |
| Admin session | Theft, fixation, CSRF | Opaque random token, hash-at-rest, HttpOnly/Secure cookie, CSRF token |
| Published content | Unauthorized modification | Authentication, server-side validation, audit, transaction |
| Draft content | Data leakage | No public query/SSG output, authenticated preview |
| Rich text | Stored XSS | Structured JSON schema, allowlist, server validation/sanitization |
| Media files | Malicious upload, path traversal | MIME/content validation, generated storage key, size limits |
| PostgreSQL | Unauthorized access, injection | Private binding, least privilege, parameterized SQL, backups |
| Server | Privilege escalation, secret exposure | Non-root service, permissions, firewall, updates |
| Backups | Theft, ransomware, accidental loss | Off-host encrypted backup, restricted access, restore drill |

Threat model ini adalah baseline MVP, bukan klaim bahwa semua risiko telah dieliminasi. Lakukan review ulang sebelum production dan ketika scope berubah.

## 3. Security boundaries

- Public browser: tidak dipercaya.
- Admin browser: tetap tidak dipercaya; semua authorization dan validation dilakukan server-side.
- Go application: trusted execution boundary, tetapi input eksternal tetap divalidasi.
- PostgreSQL: hanya dapat diakses application host dan administrator database yang berwenang.
- Media storage: tidak boleh mengizinkan executable upload atau path yang dikendalikan pengguna.
- SSG release: hanya output Published; preview terpisah dan privat.
- Backup destination: terpisah dari host utama dan memiliki credential sendiri.

## 4. Secure defaults

- HTTPS wajib untuk production.
- HTTP hanya redirect ke HTTPS dan/atau melayani ACME challenge yang dibutuhkan.
- HSTS aktif setelah HTTPS/domain policy tervalidasi.
- Same-origin untuk CMS dan API sebagai default; CORS tidak dibuka secara permissive.
- Admin session cookie: `HttpOnly`, `Secure` pada production, `SameSite` yang sesuai, path terbatas bila memungkinkan.
- CSRF protection untuk seluruh state-changing cookie-authenticated requests.
- Password disimpan dengan adaptive password hashing.
- SQL menggunakan parameterized queries.
- Rich text memakai structured schema dan allowlist.
- File upload menggunakan storage key acak dan validasi konten.
- Error publik tidak mengungkap stack trace, SQL, path, atau secret.
- Admin API memakai `Cache-Control: no-store`.
- Draft/preview tidak masuk public sitemap/cache/release.

## 5. Security ownership

- Identity module: Admin, password, session lifecycle.
- Content module: publication permissions, locale isolation, content validation.
- Media module: upload validation, rights metadata, storage safety.
- HTTP platform: middleware, CSRF, security headers, request limits.
- Database platform: connection security and least privilege.
- Deployment: OS, TLS, filesystem permissions, backup secrets.
- Audit module: security-relevant event recording.

Security tidak hanya menjadi tanggung jawab middleware; setiap module harus menjaga invariant bisnisnya sendiri.

---

## Materi sumber: `web-application-security.md`

# Web Application Security — PidieKab

**Status:** Baseline implementation specification

## 1. Input validation

Semua input divalidasi di server:
- JSON schema/field types;
- locale hanya `id` atau `en`;
- slug normalization dan uniqueness;
- content status transitions;
- URL schemes/host allowlist sesuai field;
- coordinates range dan pair completeness;
- pagination/query length;
- upload metadata dan file content.

Frontend validation hanya untuk UX. Jangan mengandalkan browser validation sebagai security boundary.

## 2. SQL injection

- Gunakan parameterized SQL melalui `pgx`.
- Jangan menggabungkan user input ke SQL string.
- Dynamic sorting/filter columns harus dipilih dari allowlist.
- Query search menggunakan parameter dan fungsi PostgreSQL yang aman.
- Database errors tidak dikembalikan mentah ke client.

## 3. XSS and rich text

Rich text disimpan sebagai structured JSON document, bukan arbitrary HTML.

Allowed nodes/marks harus ditentukan, misalnya paragraph, heading levels yang didukung, bold, italic, ordered/unordered list, link, simple table, dan image reference. Backend:
- menolak node/attribute yang tidak dikenal;
- memvalidasi link protocol;
- menolak `javascript:`, `data:` untuk link navigation, inline event handlers, script, iframe, arbitrary embed;
- memastikan image reference mengarah ke media library yang sah;
- menghasilkan HTML hanya dari allowlist renderer;
- sanitasi output sebagai defense-in-depth.

Frontend tidak boleh memakai `v-html` untuk data yang belum disanitasi. CSP membantu membatasi dampak XSS, bukan menggantikan sanitization.

## 4. CSRF and CORS

- Cookie-authenticated mutation wajib CSRF-protected.
- Same-origin adalah deployment default.
- Jangan mengaktifkan wildcard CORS bersama credentials.
- Jika origin terpisah dibutuhkan, gunakan allowlist eksplisit dan credential policy yang sempit.
- Validasi Origin/Referer sebagai lapisan tambahan.

## 5. Clickjacking and browser headers

Baseline headers:
- `X-Content-Type-Options: nosniff`;
- `Referrer-Policy` yang sesuai;
- `Content-Security-Policy` dengan `frame-ancestors 'none'` kecuali ada kebutuhan embed yang disetujui;
- `X-Frame-Options: DENY` sebagai kompatibilitas;
- HSTS hanya pada HTTPS production setelah konfigurasi domain tervalidasi.

CSP harus disusun berdasarkan asset yang benar-benar digunakan. Hindari `unsafe-eval`; hindari `unsafe-inline` jika build dapat menggunakan nonce/hash/external assets yang aman.

## 6. Open redirect and URL safety

- Redirect slug hanya menuju path internal canonical.
- External URLs pada content source/contact/social divalidasi scheme dan ditampilkan dengan safe link attributes.
- Jangan menerima arbitrary redirect target dari query parameter login.
- Redirect 301 table harus mencegah loop, chain berlebihan, dan path traversal.

## 7. Enumeration and abuse

- Login response generic.
- Search query length dan pagination dibatasi.
- Rate limit login/search/upload.
- Upload body dibatasi sebelum file dibaca penuh.
- Public endpoints tidak mengungkap draft existence atau private metadata.
- Error response konsisten dan tidak membocorkan resource internal.

## 8. Caching and privacy

- Admin API: `Cache-Control: no-store`.
- Preview: no-store dan noindex.
- Public static pages boleh cached sesuai cache policy.
- Jangan cache response draft/private pada shared cache.
- Jangan mengumpulkan visitor personal data/analytics pada MVP tanpa kebutuhan dan privacy policy yang jelas.

---
