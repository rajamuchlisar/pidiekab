# Repository Foundation & Development Standards

**Versi:** v79

## 1. Repository layout

```text
PidieKabProject/
├── backend/
│   ├── cmd/pidiekab/main.go
│   └── internal/
│       ├── app/
│       └── platform/
│           ├── config/
│           └── httpserver/
├── frontend/
│   ├── public-site/
│   └── admin-cms/
├── migrations/
├── deploy/
├── docs/
├── .github/workflows/
├── .env.example
├── .gitignore
└── README.md
```

Struktur ini adalah fondasi. Package bisnis (`content`, `identity`, dan lainnya) dibuat saat fase modul terkait dimulai, bukan dibuat kosong sejak awal.

## 2. Go conventions

- Gunakan Go standard library sebagai baseline.
- Satu Go module dan satu binary `pidiekab`.
- HTTP server menggunakan `net/http` dan `ServeMux`.
- Configuration melalui environment; validasi saat startup.
- Log memakai `log/slog` dalam format JSON pada production.
- Shutdown menggunakan `http.Server.Shutdown` dengan context timeout.
- Error tidak boleh membocorkan stack trace atau secret ke response.
- Gunakan `context.Context` untuk request lifecycle dan operasi I/O.
- Hindari global mutable state.
- Tidak menambahkan interface jika belum ada kebutuhan substitusi/testing yang nyata.

## 3. Frontend conventions

- Vue 3, TypeScript strict, Vite.
- Dua app terpisah: `public-site` dan `admin-cms`.
- Hindari duplikasi konfigurasi lint/typecheck/build yang tidak perlu; shared tooling boleh diletakkan di root jika benar-benar membantu.
- Public site memiliki build SSG terpisah dari Admin SPA.
- Pinia hanya untuk state lintas halaman yang memang diperlukan.
- API same-origin; credential session tidak disimpan di localStorage.

## 4. Environment

`.env.example` hanya memuat nama variable dan nilai contoh non-secret. Production secret disediakan melalui environment/service manager atau secret file dengan permission terbatas.

Variabel awal:
- `PIDIEKAB_ENV`
- `PIDIEKAB_ADDR`
- `PIDIEKAB_READ_HEADER_TIMEOUT`
- `PIDIEKAB_READ_TIMEOUT`
- `PIDIEKAB_WRITE_TIMEOUT`
- `PIDIEKAB_IDLE_TIMEOUT`
- `PIDIEKAB_SHUTDOWN_TIMEOUT`
- `PIDIEKAB_LOG_LEVEL`
- `PIDIEKAB_DATABASE_URL` (akan digunakan saat DB pool diintegrasikan)

Jangan commit `.env`, private key, database dump, generated media, atau build artifacts.

## 5. Development commands

Perintah final dicatat di root README setelah package manager dipilih. Minimal harus tersedia:
- run backend;
- test backend;
- vet backend;
- install frontend dependencies;
- run public dev server;
- run Admin dev server;
- typecheck/lint frontend;
- build public SSG;
- build Admin SPA;
- validate migrations.

## 6. Quality gates

Setiap perubahan harus:
1. lolos formatting;
2. lolos unit tests;
3. lolos static analysis/typecheck;
4. tidak menambahkan dependency tanpa alasan;
5. memperbarui docs bila kontrak/config berubah;
6. tidak memasukkan secret;
7. menjaga build output dan local state di luar Git.
