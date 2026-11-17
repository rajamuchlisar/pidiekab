# Admin Identity And Session Security

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `admin-authentication-and-session-security.md`

---

## Materi sumber: `admin-authentication-and-session-security.md`

# Admin Authentication & Session Security — PidieKab

**Status:** Baseline implementation specification

## 1. Account model

- MVP hanya memiliki satu role: Admin.
- Tidak ada public registration.
- Admin harus dibuat melalui bootstrap procedure atau Admin yang berwenang.
- Admin inactive tidak dapat login.
- Sistem mencegah penonaktifan Admin aktif terakhir.
- Email dinormalisasi konsisten sebelum lookup dan memiliki unique constraint.
- Login error tidak membocorkan apakah email terdaftar, inactive, atau password salah.

## 2. Password storage

- Gunakan Argon2id dengan parameter yang ditetapkan melalui benchmark pada server target; bcrypt dapat menjadi fallback bila implementasi Argon2id yang aman tidak tersedia.
- Jangan gunakan SHA-256, MD5, plaintext, reversible encryption, atau custom password algorithm.
- Gunakan salt unik per password; library password hashing menangani format/parameter.
- Password tidak pernah dikirim kembali melalui API, ditulis ke log, atau disimpan pada audit metadata.
- Rehash password saat login bila parameter hash lama sudah di bawah policy.
- Password reset menghasilkan credential/token sekali pakai atau alur reset yang aman; jangan mengirim password permanen dalam plaintext.

Policy panjang minimum, maximum input size, breached-password check, dan 2FA akan difinalisasi pada security review sebelum production. Jangan memotong password secara diam-diam.

## 3. Login protection

- Rate limit berdasarkan kombinasi IP dan identifier yang dinormalisasi; hindari lockout permanen yang dapat dipakai untuk denial of service.
- Tambahkan progressive delay atau temporary throttling setelah kegagalan berulang.
- Pesan kegagalan bersifat generic.
- Catat login success/failure secara aman tanpa menyimpan password atau session token.
- Pertimbangkan notification untuk perubahan credential atau aktivitas anomali pada tahap berikutnya.
- Gunakan TLS untuk semua login.

## 4. Session token

Baseline:
1. Generate opaque cryptographically random token dengan entropy memadai (minimum 256-bit).
2. Kirim token hanya melalui cookie HTTPS.
3. Simpan hash token, bukan token asli, pada `admin_sessions`.
4. Ikat session ke Admin, created time, expiry, last activity, dan revoked state.
5. Bandingkan token hash secara aman.
6. Rotate session identifier setelah login dan perubahan privilege/security-sensitive events.
7. Logout merevoke session server-side.
8. Password reset dan deactivation merevoke semua session Admin terkait.

Cookie baseline:
- `HttpOnly`;
- `Secure` pada production;
- `SameSite=Lax` sebagai default awal, ditinjau terhadap alur CMS;
- `Path=/api/v1/admin` bila kompatibel dengan kebutuhan routing; jika cookie perlu dipakai pada preview path, tentukan path yang paling sempit tanpa memecahkan flow;
- tidak menetapkan `Domain` kecuali ada kebutuhan subdomain yang jelas.

Cookie name final ditetapkan konfigurasi deployment. Session token tidak disimpan pada localStorage/sessionStorage.

## 5. Session lifetime

Tetapkan absolute lifetime dan idle timeout yang eksplisit. Nilai awal yang diusulkan untuk MVP:
- idle timeout: 30 menit;
- absolute lifetime: 12 jam.

Nilai tersebut harus dapat dikonfigurasi dan ditinjau terhadap kebutuhan operasional. Activity refresh tidak boleh membuat session hidup tanpa batas; absolute expiry tetap berlaku.

## 6. CSRF

Karena autentikasi menggunakan cookie, semua state-changing endpoints memerlukan CSRF defense:
- synchronizer token atau signed double-submit pattern yang dirancang benar;
- token tidak boleh berasal dari cookie session saja;
- validasi `Origin`/`Referer` sebagai defense-in-depth;
- reject cross-site requests yang tidak sesuai;
- GET/HEAD/OPTIONS tidak boleh mengubah state.

`SameSite` cookie membantu tetapi bukan pengganti CSRF token. Login endpoint juga harus mempertimbangkan login CSRF.

## 7. Session validation

Pada setiap Admin API request:
- parse cookie;
- hash token;
- lookup session aktif;
- validasi expiry/revocation;
- validasi Admin masih active;
- update last activity secara efisien dan tidak pada setiap query bila menimbulkan write amplification;
- inject Admin identity ke request context.

Session invalid/expired mengembalikan `401`. Admin authenticated tetapi tidak diizinkan mengembalikan `403`.

## 8. Admin account recovery

- Bootstrap Admin dilakukan melalui command/one-time setup procedure yang tidak tersedia sebagai public HTTP endpoint.
- Reset password harus menggunakan token sekali pakai dengan expiry dan hash-at-rest, atau prosedur operator yang aman.
- Jangan membuat endpoint yang dapat reset password Admin tanpa autentikasi/authorization yang memadai.
- Deactivation/password reset merevoke sessions.
- Perubahan email/password dicatat ke audit log tanpa secret.

---
