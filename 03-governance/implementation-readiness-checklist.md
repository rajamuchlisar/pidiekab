# PidieKab — Implementation Readiness Checklist

**Versi:** v78  
**Tujuan:** checklist gate sebelum memulai dan menyelesaikan tiap fase implementasi.

## A. Sebelum Phase 0 dimulai

- [ ] Repository Git dibuat dan default branch ditentukan.
- [ ] Struktur folder dan satu Go module mengikuti backend guide.
- [ ] Versi Go, Node.js, pnpm/npm, PostgreSQL ditetapkan dan dicatat.
- [ ] `.env.example` tersedia tanpa secret.
- [ ] `.gitignore` mencegah secret, build output, local media, dan database dump masuk Git.
- [ ] README memiliki instruksi setup lokal yang dapat diikuti dari mesin bersih.
- [ ] Format Go, lint, frontend typecheck/lint/build, dan test command disepakati.
- [ ] CI menjalankan checks dasar pada pull request.
- [ ] Tidak ada Redis, message broker, reverse proxy, atau service tambahan tanpa ADR baru.

## B. Phase 0 — Repository & development foundation

- [ ] `go test ./...` berjalan.
- [ ] Go server start/stop dengan graceful shutdown.
- [ ] `net/http` ServeMux dan middleware dasar berfungsi.
- [ ] Structured logging dan request ID tersedia.
- [ ] Config validation gagal cepat untuk konfigurasi wajib yang invalid.
- [ ] Vue public dan Admin dapat di-install/build.
- [ ] SSG proof-of-concept menghasilkan localized dynamic routes `/id/` dan `/en/`.
- [ ] POC menunjukkan SEO metadata, canonical/hreflang, sitemap, dan output static.
- [ ] Output static dapat disajikan oleh Go tanpa Node runtime.
- [ ] Hasil POC dan keputusan SSG dicatat dalam ADR.

## C. Database foundation

- [ ] PostgreSQL versi target ditetapkan.
- [ ] Goose migration format dibuktikan.
- [ ] Migration clean install berhasil pada database kosong.
- [ ] Migration rollback/forward diuji pada environment disposable.
- [ ] DDL constraints, FK, unique index, dan check constraints tervalidasi.
- [ ] Timestamp disimpan sebagai `timestamptz` UTC.
- [ ] Database role aplikasi bukan superuser.
- [ ] Connection pool dan shutdown lifecycle diuji.
- [ ] Tidak ada migration otomatis tak terkendali setiap service restart.

## D. Security foundation

- [ ] Password hashing menggunakan Argon2id atau bcrypt yang disetujui.
- [ ] Tidak ada password plaintext atau fast hash.
- [ ] Admin bootstrap dilakukan melalui one-time CLI.
- [ ] Session token opaque dan hash disimpan di database.
- [ ] Cookie Secure, HttpOnly, SameSite, Path, expiry ditetapkan.
- [ ] CSRF dan Origin validation diuji.
- [ ] Inactive Admin tidak dapat login.
- [ ] Logout, password reset, dan deactivation menginvalidasi session yang relevan.
- [ ] Preview hanya untuk Admin authenticated dan noindex.
- [ ] Draft tidak muncul di public API, SSG output, sitemap, atau search.
- [ ] Rich text disanitasi server-side.
- [ ] Upload divalidasi berdasarkan MIME/content, dire-encode bila perlu, dan disimpan non-executable.
- [ ] Logs tidak memuat password, session token, CSRF token, atau secret.

## E. Content & publication

- [ ] Satu content identity memiliki translation ID/EN terpisah.
- [ ] Status Published per locale independen.
- [ ] Source minimal satu dan valid sebelum Publish.
- [ ] Source tidak disebut sebagai verified badge atau approval status.
- [ ] Category hanya satu primary category per content.
- [ ] Kategori inactive tidak tampil di public dan tidak dapat dipilih untuk konten baru.
- [ ] Category deletion memerlukan pemindahan semua konten terkait.
- [ ] Slug changes menghasilkan 301 tanpa loop/chain.
- [ ] Publish + enqueue SSG job atomic dalam satu transaksi.
- [ ] Worker claim job menggunakan concurrency-safe locking.
- [ ] Build dilakukan pada temporary release.
- [ ] Release aktif hanya berubah setelah build sukses.
- [ ] Failure/restart/retry mempertahankan release sebelumnya.

## F. Public experience & search

- [ ] Public hanya menampilkan Published content pada locale aktif.
- [ ] Tidak ada silent fallback ID/EN.
- [ ] Search title, summary, body dengan bobot yang ditetapkan.
- [ ] Filter kategori mempertahankan query dan reset page ke 1.
- [ ] Pagination mempertahankan query, locale, dan filter.
- [ ] Search hanya mengembalikan locale aktif dan Published.
- [ ] Indonesian dan English tokenizer behavior telah diuji.
- [ ] Map hanya tampil bila koordinat tersedia.
- [ ] Tidak ada koordinat hasil tebakan.
- [ ] Responsive layout dan accessibility baseline diuji.

## G. Production readiness

- [ ] Ubuntu 24.04 dan systemd service tersedia.
- [ ] Service berjalan non-root.
- [ ] Go melayani HTTPS langsung; tidak ada reverse proxy.
- [ ] ACME issuance, renewal, reload, dan expiry alert diuji.
- [ ] Port database tidak terbuka ke internet.
- [ ] Backup database dan media terenkripsi serta off-host.
- [ ] Restore drill berhasil dan durasinya dicatat.
- [ ] `/health/live` dan `/health/ready` sesuai fungsi.
- [ ] Disk, DB, worker, backup, dan TLS alert dikonfigurasi.
- [ ] Runbook deployment, rollback, backup restore, incident, dan certificate renewal tersedia.
- [ ] QA acceptance dan security tests lulus.
- [ ] Tidak ada critical/high severity issue yang belum ditangani.

## H. Stop conditions untuk AI Coding Agent

Agent harus berhenti dan meminta keputusan apabila:
- instruksi bertentangan dengan ADR LOCKED;
- perubahan memerlukan library/infrastruktur baru yang belum disetujui;
- migration berpotensi menghapus atau mengubah data tanpa rencana;
- security behavior belum ditentukan;
- source data atau koordinat tidak dapat dibuktikan;
- scope Phase 2 diminta masuk MVP;
- test gagal dan penyebabnya belum dipahami;
- implementasi membutuhkan perubahan API/database lintas modul yang tidak tercakup task.

Agent tidak boleh menandai task selesai hanya karena kode berhasil dikompilasi. Definition of Done mencakup test, dokumentasi, migration/API updates, security review yang relevan, dan acceptance criteria.
