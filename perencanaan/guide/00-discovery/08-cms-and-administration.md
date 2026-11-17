# 08 Cms And Administration

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `administrator-access-model.md`
- `cms-admin-dashboard.md`
- `cms-administrator-management.md`
- `cms-all-content-list.md`
- `cms-content-module-navigation.md`
- `cms-minimal-navigation.md`
- `cms-multilingual-content-form.md`
- `cms-central-media-library.md`
- `cms-rich-text-editor.md`
- `cms-static-pages-module.md`
- `cms-website-settings.md`
- `cms-featured-content-management.md`
- `cms-homepage-editorial-story-curation.md`
- `cms-public-site-preview.md`

---

## Materi sumber: `administrator-access-model.md`

# Administrator & Access Model

## Decision ADM-001 --- Single Admin Role

**Status: Agreed**

The PidieKab MVP uses one internal administrator role: `Admin`. All
admin accounts have the same permissions.

## MVP Rules

1.  All internal content managers authenticate using an Admin account.
2.  All Admin accounts have equal permissions for content management and
    system administration.
3.  There are no separate Super Admin, Editor, or Author roles in the
    MVP.
4.  Admin accounts are not created through public registration.
5.  Public visitors do not need an account to browse published content.
6.  Admin users can create, edit, review, publish, unpublish, and
    archive content.
7.  Admin users can manage other admin accounts and applicable system
    settings.
8.  The system records the actor and timestamps for important
    create/update/publish actions.
9.  Passwords must be stored using an appropriate password hashing
    algorithm. Plain-text password storage is prohibited.
10. Authentication, authorization, session security, and account
    recovery requirements will be specified in the technical
    specification.

## Auditability

Even with one role, record important administrative actions where
practical: - Actor/admin ID. - Action performed. - Target content or
resource. - Timestamp. - Relevant change metadata.

A full enterprise audit-log subsystem is not automatically required for
MVP; minimum auditability requirements will be defined in the PRD.

## Out of Scope for MVP

-   Role hierarchy.
-   Granular permission matrix.
-   Separate author/reviewer/publisher roles.
-   Approval workflow requiring two different accounts.
-   Public account registration.

## Future Evolution

The authorization design should avoid hard-coding role assumptions
throughout business logic. A simple role model can be implemented now
while keeping a reasonable path to introduce additional roles if
operational needs change.

## Open Decisions

-   Should admin accounts be created only by an existing Admin, or
    through a controlled initial setup command?
-   Is password reset by email required for MVP?
-   Is two-factor authentication required at launch or deferred?
-   Should admin sessions expire after a configurable idle period?

---

## Materi sumber: `cms-admin-dashboard.md`

# CMS-009 — Dashboard Administrator

**Status: Disepakati.**

## Keputusan
Dashboard CMS Admin menampilkan ringkasan operasional untuk membantu Administrator memahami kondisi konten dan aktivitas pengelolaan PidieKab. Dashboard bersifat informatif dan bukan analytics kompleks.

## Komponen dashboard MVP

1. **Total Konten** — jumlah seluruh identitas konten dari tiga kategori MVP.
2. **Konten berdasarkan Kategori** — ringkasan Wisata & Destinasi, Kuliner & UMKM, dan Budaya & Sejarah.
3. **Draft** — jumlah konten/versi bahasa yang masih Draft.
4. **Review** — jumlah konten/versi bahasa yang menunggu pemeriksaan.
5. **Published** — jumlah konten/versi bahasa yang telah diterbitkan.
6. **Status publikasi ID dan EN** — ringkasan status publikasi per bahasa secara terpisah.
7. **Konten terakhir diperbarui** — daftar ringkas konten yang paling baru diubah.
8. **Aktivitas administrator terakhir** — ringkasan tindakan administratif terbaru yang tercatat.

## Prinsip perhitungan dan tampilan
- Angka konten dan angka versi bahasa harus dibedakan dengan jelas. Satu identitas konten dapat memiliki dua versi bahasa dan status yang berbeda.
- Ringkasan ID dan EN tidak boleh digabung sehingga menyamarkan status publikasi independen.
- Konten Archived tidak dihitung sebagai Published.
- Komponen yang tidak memiliki data menampilkan keadaan kosong yang wajar, bukan angka atau aktivitas rekaan.
- Daftar konten terakhir diperbarui dan aktivitas admin dapat ditautkan ke halaman terkait jika sesuai.
- Semua data dashboard mengikuti hak akses Administrator yang sama.

## Batasan MVP
- Tidak ada grafik analitik kompleks.
- Tidak ada statistik pengunjung, page views, conversion, engagement, atau demografi.
- Tidak ada laporan performa konten atau tracking perilaku pengunjung.
- Tidak ada dashboard terpisah per role karena MVP hanya memiliki satu role Administrator.

## Hal yang ditetapkan pada spesifikasi UI/teknis
- Definisi hitungan yang tepat untuk kartu konten versus versi bahasa.
- Periode atau batas jumlah item pada daftar aktivitas terbaru.
- Refresh data (on-load atau berkala) dan strategi caching.
- Tindakan cepat yang tersedia dari dashboard.

---

## Materi sumber: `cms-administrator-management.md`

# CMS-008 — Pengelolaan Administrator

**Status: Disepakati.**

## Keputusan
CMS Admin menyediakan satu modul **Administrator** untuk mengelola akun yang dapat mengakses area administrasi PidieKab. Seluruh akun menggunakan satu role dengan hak akses yang sama, sesuai ADM-001.

## Fitur MVP
1. Daftar administrator.
2. Tambah administrator.
3. Edit data administrator.
4. Aktifkan/nonaktifkan akun.
5. Reset password administrator.
6. Profil administrator yang sedang login.
7. Ganti password sendiri.

## Prinsip akses dan keamanan
- Tidak tersedia pendaftaran administrator secara publik.
- Penambahan akun hanya dapat dilakukan oleh administrator yang sudah memiliki akses.
- Akun nonaktif tidak dapat melakukan login atau menggunakan sesi aktif.
- Reset password harus menggunakan mekanisme aman; password baru tidak ditampilkan atau dikirim dalam bentuk teks biasa.
- Password disimpan menggunakan password hashing yang kuat dan sesuai praktik keamanan.
- Perubahan akun dan tindakan administratif penting dicatat dalam audit log.
- Akun tidak boleh menghapus atau menonaktifkan dirinya sendiri jika tindakan tersebut dapat mengunci akses administrasi terakhir.
- Sistem perlu mencegah penguncian seluruh akses admin secara tidak sengaja.

## Data administrator
Data dasar yang diusulkan:
- Nama tampilan.
- Email (unik dan digunakan untuk login).
- Password (disimpan dalam bentuk hash, tidak dapat dibaca kembali).
- Status akun: aktif/nonaktif.
- Waktu terakhir login.
- Waktu dibuat dan diperbarui.

Field tambahan seperti nomor telepon tidak diwajibkan untuk MVP tanpa kebutuhan operasional yang jelas.

## Batasan MVP
- Hanya satu role: Administrator.
- Tidak ada role Super Admin, Editor, Author, atau permission matrix per akun.
- Tidak ada pendaftaran publik.
- Tidak ada manajemen tim atau undangan berbasis organisasi.

## Hal yang ditetapkan pada spesifikasi keamanan/teknis
- Cara bootstrap akun administrator pertama.
- Alur reset password (misalnya melalui email atau prosedur pemulihan operasional).
- Kebijakan password dan pembatasan percobaan login.
- Apakah 2FA diwajibkan atau ditunda.
- Batas idle session dan masa berlaku sesi.
- Mekanisme pencabutan sesi ketika akun dinonaktifkan atau password direset.

---

## Materi sumber: `cms-all-content-list.md`

# CMS-003 — Daftar Semua Konten

**Status: Disepakati.**

## Keputusan

Halaman **Semua Konten** pada CMS Admin menampilkan seluruh konten dari tiga kategori MVP dalam satu tabel.

## Kolom tabel

| Kolom | Isi |
|---|---|
| Judul | Judul utama konten |
| Kategori | Wisata & Destinasi, Kuliner & UMKM, atau Budaya & Sejarah |
| Bahasa | Status versi Bahasa Indonesia (ID) dan English (EN) |
| Status | Status publikasi/workflow konten |
| Unggulan | Penanda apakah konten dipilih sebagai konten unggulan |
| Terakhir diubah | Waktu perubahan terakhir |
| Aksi | Tindakan yang tersedia untuk konten |

## Bahasa dan status publikasi

- Kolom Bahasa harus memperlihatkan status ID dan EN secara terpisah.
- Status publikasi setiap bahasa bersifat independen sesuai keputusan multilingual.
- Tampilan daftar tidak boleh menyiratkan bahwa satu bahasa telah Published hanya karena bahasa lainnya Published.
- Status workflow yang digunakan mengikuti Draft, Published, dan Archived. Preview merupakan aksi pemeriksaan, bukan status.
- Detail representasi status gabungan pada baris daftar ditetapkan pada desain UI/technical specification.

## Pencarian dan kontrol daftar

Halaman menyediakan:
- Pencarian konten.
- Filter kategori.
- Filter status.
- Pagination.

Nilai pencarian dan filter dipertahankan saat berpindah halaman. Perubahan pencarian atau filter mengembalikan daftar ke halaman pertama.

## Aksi

Kolom Aksi menyediakan tindakan yang relevan untuk setiap konten, sekurang-kurangnya membuka/detail dan mengedit. Aksi Hapus menggunakan Permanent Delete dan wajib melalui dialog konfirmasi yang menjelaskan bahwa tindakan tidak dapat dipulihkan melalui CMS.

## Batasan dan hal yang belum ditetapkan

- Tidak menetapkan jumlah baris per halaman pada tahap Discovery ini.
- Tidak menetapkan daftar opsi pengurutan atau default sorting.
- Tidak menetapkan bulk actions.
- Tidak menetapkan apakah status ditampilkan sebagai dua badge bahasa, ringkasan, atau kombinasi visual lain.
- Tidak menetapkan field tambahan seperti penulis atau tanggal publikasi sebagai kolom wajib.

---

## Materi sumber: `cms-content-module-navigation.md`

# CMS-002 — Struktur Menu Modul Konten

**Status: Disepakati.**

## Keputusan

Modul **Konten** pada CMS Admin memiliki struktur submenu:

1. Semua Konten
2. Wisata & Destinasi
3. Kuliner & UMKM
4. Budaya & Sejarah

## Prinsip pengelolaan

- Semua Konten menampilkan seluruh konten dari tiga kategori MVP dalam satu daftar.
- Menu kategori menampilkan konten yang termasuk kategori tersebut.
- Setiap konten tetap memiliki satu identitas konten utama dengan versi Bahasa Indonesia dan English sesuai keputusan model konten multilingual.
- Cerita tentang Pidie bukan modul atau tipe konten terpisah pada MVP; konten editorial tersebut dikelola dalam modul Konten yang sama.
- Konten Unggulan bukan modul terpisah; status/atribut unggulan dikelola pada konten yang sama.
- Kategori MVP tetap terbatas pada Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.

## Batasan

- Tidak membuat CRUD terpisah untuk Cerita tentang Pidie.
- Tidak membuat CRUD terpisah untuk Konten Unggulan.
- Tidak menambah kategori konten baru hanya untuk memenuhi kebutuhan tampilan homepage.

## Hal yang ditetapkan pada spesifikasi CMS berikutnya

- Kolom yang ditampilkan pada daftar Semua Konten.
- Filter, pencarian, pengurutan, dan pagination di CMS.
- Aksi cepat dan bulk actions.
- Form pembuatan dan pengeditan konten.
- Cara Admin menetapkan konten unggulan dan memilih cerita editorial untuk homepage.

---

## Materi sumber: `cms-minimal-navigation.md`

# CMS-001 — Pendekatan Struktur CMS Minimalis

**Status: Disepakati.**

## Keputusan

CMS Admin PidieKab menggunakan pendekatan **minimalis**. Menu utama hanya menampilkan fungsi inti. Fitur pendukung seperti SEO, media, sumber informasi, dan pengaturan teknis ditempatkan di dalam modul terkait atau sebagai submenu jika diperlukan.

Tujuannya menjaga CMS tetap mudah dipahami dan digunakan, mengurangi keramaian navigasi, serta memungkinkan pengembangan bertahap.

## Usulan struktur menu utama

1. Dashboard
2. Konten
3. Halaman Informasi
4. Pengaturan
5. Administrator

Struktur ini adalah baseline navigasi MVP. Rincian submenu dan cakupan masing-masing modul ditetapkan pada pembahasan spesifikasi CMS berikutnya.

## Penempatan fitur pendukung

- **SEO:** menjadi bagian dari form pengelolaan konten atau halaman informasi; pengelolaan sitemap dan redirect ditempatkan pada pengaturan bila dibutuhkan.
- **Media:** perpustakaan media digunakan dari dalam form konten/halaman; tidak harus menjadi menu utama.
- **Sumber & Referensi:** dikelola dalam konteks konten yang didukung sumber tersebut, bukan sebagai modul utama terpisah.
- **Pengaturan website:** identitas, logo, favicon, kontak, media sosial, bahasa, dan pengaturan umum dikelompokkan dalam satu modul Pengaturan.
- **Audit log:** tersedia sebagai fitur sistem/administrasi; penempatannya ditentukan dalam spesifikasi teknis dan tidak harus menjadi menu utama.
- **Administrator:** pengelolaan akun admin dan profil akses ditempatkan dalam satu area.

## Prinsip navigasi

- Jangan membuat menu terpisah hanya karena ada satu jenis field atau fungsi kecil.
- Tampilkan submenu hanya jika membantu menemukan pekerjaan yang sering dilakukan.
- Pertahankan pemisahan antara konten kategori utama dan halaman informasi statis.
- Jangan menambah fitur di luar MVP tanpa kebutuhan yang jelas.

---

## Materi sumber: `cms-multilingual-content-form.md`

# CMS-004 — Form Tambah dan Edit Konten Multibahasa

**Status: Disepakati.**

## Keputusan utama
Form tambah dan edit konten menggunakan **satu form dengan tab Bahasa Indonesia (ID) dan English (EN)**. Admin mengelola kedua versi bahasa dalam satu identitas konten yang sama, bukan membuat dua konten terpisah.

Tab ID dan EN harus terlihat jelas dan dapat dipilih langsung. Masing-masing tab menampilkan field bahasa yang sesuai. Status publikasi kedua bahasa dikelola secara independen.

## Struktur form

### 1. Informasi Utama
Data bersama yang tidak perlu digandakan per bahasa, termasuk kategori dan identitas konten utama.

### 2. Tab Bahasa Indonesia (ID)
- Judul
- Slug
- Ringkasan
- Isi konten
- SEO title
- Meta description
- Status workflow/publikasi ID

### 3. Tab English (EN)
- Title
- Slug
- Summary
- Content body
- SEO title
- Meta description
- Status workflow/publikasi EN

Field ID dan EN dikelola terpisah. Tidak ada sinkronisasi atau publikasi otomatis antarbahasa. Machine translation, jika digunakan, harus ditinjau manusia sebelum publikasi.

### 4. Informasi Pendukung
Data bersama sesuai kategori dan relevansi, seperti alamat/area, koordinat bila tersedia dan terverifikasi, foto utama, serta field khusus kategori yang relevan. Validasi fleksibel DET-002 tetap berlaku.

### 5. Sumber & Referensi
Jenis dan nama sumber, URL/referensi, tanggal publikasi/akses bila diketahui, catatan klaim yang didukung, status verifikasi, tanggal pemeriksaan, serta asal/lisensi/izin/atribusi foto.

### 6. Publikasi
- Status ID dan EN dikelola secara independen.
- Penanda konten unggulan dikelola pada identitas konten yang sama.
- Aksi simpan draft, preview dan publish, dan archive mengikuti workflow yang berlaku serta aturan publikasi per bahasa.

## Prinsip interaksi
- Tab ID dan EN selalu tersedia, termasuk saat membuat konten baru.
- Admin dapat menyimpan konten meskipun salah satu bahasa belum lengkap atau belum siap diterbitkan.
- Validasi publikasi dilakukan terhadap versi bahasa yang akan diterbitkan; kekurangan pada bahasa lain tidak otomatis menghalangi publikasi bahasa yang siap.
- Indikator tab dapat menunjukkan status masing-masing bahasa.
- Perpindahan tab tidak boleh menghilangkan perubahan yang belum disimpan.
- Data bersama tidak digandakan di kedua tab.

## Batasan MVP
- Tidak ada dua record konten terpisah untuk ID dan EN.
- Tidak ada auto-publish atau auto-sync antarbahasa.
- Tidak ada page builder atau editor blok kompleks.
- Detail layout, autosave, dan pilihan rich-text/Markdown ditetapkan pada spesifikasi UI/teknis.

---

## Materi sumber: `cms-central-media-library.md`

# CMS-006 — Media Library Terpusat

**Status: Disepakati.**

## Keputusan
PidieKab menggunakan satu **Media Library** bersama untuk mengelola gambar yang digunakan pada konten utama maupun halaman informasi. Media Library tidak menjadi menu utama tersendiri; Admin mengaksesnya dari form konten/halaman ketika memilih atau mengunggah media.

## Fitur MVP
- Upload gambar.
- Preview gambar.
- Menampilkan nama file dan ukuran file.
- Mengelola alt text.
- Mencatat sumber gambar, lisensi/izin penggunaan, dan atribusi.
- Memilih gambar yang sudah tersimpan dari Media Library ketika mengisi konten.
- Menggunakan media yang sama pada lebih dari satu konten bila sesuai hak penggunaannya.

## Prinsip pengelolaan
1. Media disimpan sebagai aset terpisah dari konten yang menggunakannya.
2. Penghapusan media yang masih digunakan harus dicegah atau memerlukan penanganan relasi terlebih dahulu.
3. Hak penggunaan, sumber, dan atribusi dicatat agar dapat ditelusuri.
4. Alt text ditentukan sesuai konteks penggunaan; tidak sekadar menyalin nama file.
5. Media utama/detail dan gambar yang disisipkan dalam artikel menggunakan Media Library yang sama.
6. Media harus dioptimalkan untuk performa dan ukuran tampilan, tanpa mengubah atau menghilangkan atribusi yang diwajibkan.
7. File upload divalidasi berdasarkan tipe konten sebenarnya, ukuran, dan keamanan; jangan hanya mengandalkan ekstensi file.

## Batasan MVP
- Media Library tidak menjadi menu utama.
- Tidak ada editor gambar kompleks, crop studio, atau desain grafis di dalam CMS.
- Tidak ada galeri foto multi-gambar pada halaman detail; keputusan satu foto utama (MED-001) tetap berlaku. Penyisipan gambar di badan artikel tetap didukung.
- Format file yang diizinkan, batas ukuran, dimensi, strategi thumbnail/konversi, dan storage ditetapkan pada Technical Specification.

## Catatan teknis
Implementasi perlu mempertimbangkan nama file aman, akses upload terbatas untuk Admin, pemeriksaan MIME/signature, pencegahan file berbahaya, penyajian gambar yang efisien, backup, serta kebijakan penghapusan dan file yatim.

---

## Materi sumber: `cms-rich-text-editor.md`

# CMS-005 — Rich Text Editor untuk Isi Konten

**Status: Disepakati.**

## Keputusan
Field isi artikel pada tab Bahasa Indonesia (ID) dan English (EN) menggunakan **Rich Text Editor (RTE)** agar Admin dapat menyusun artikel terstruktur tanpa menulis HTML langsung.

## Fitur editor MVP
- Heading dan subheading.
- Bold dan italic.
- Bullet list dan numbered list.
- Hyperlink.
- Tabel sederhana.
- Penyisipan gambar di dalam artikel.

## Prinsip penggunaan
1. Setiap tab bahasa memiliki isi artikel dan formatnya sendiri.
2. Konten ID dan EN tetap dikelola secara independen.
3. Gambar yang disisipkan harus berasal dari media yang dikelola atau sumber yang dapat dipertanggungjawabkan.
4. Asal gambar, izin/lisensi, dan atribusi harus dapat dicatat.
5. Editor menghasilkan struktur konten semantik dan responsif.
6. Konten yang ditempel dari aplikasi lain dibersihkan dari format atau markup yang tidak aman.
7. Admin tidak dapat memasukkan HTML atau JavaScript bebas.

## Batasan MVP
- Tidak ada page builder atau drag-and-drop layout.
- Tidak ada kolaborasi real-time.
- Tidak ada embed bebas untuk kode/script.
- Tidak ada pengolahan gambar kompleks di dalam editor.
- Pemilihan library RTE, format penyimpanan, sanitasi HTML, upload media, dan batas ukuran gambar ditentukan pada Technical Specification.

## Aksesibilitas dan keamanan
- Toolbar dapat digunakan dengan keyboard.
- Konten hasil editor disanitasi di sisi server sebelum ditampilkan.
- Link dan gambar divalidasi; script, event handler, dan markup berbahaya tidak boleh dieksekusi.

---

## Materi sumber: `cms-static-pages-module.md`

# CMS-PAGE-001 — Pengelolaan Halaman Informasi Statis

**Status: Disepakati.**

## Keputusan

Halaman informasi umum PidieKab dikelola melalui **satu modul CRUD tersendiri** di CMS Admin, terpisah dari modul konten utama yang mengelola Wisata & Destinasi, Kuliner & UMKM, serta Budaya & Sejarah.

Modul ini digunakan untuk halaman seperti:

- Tentang PidieKab (About Us)
- Hubungi Kami (Contact)
- Privacy Policy
- Disclaimer
- Halaman informasi umum lainnya yang diperlukan

## Prinsip pengelolaan

1. Admin dapat membuat, melihat, mengubah, dan menghapus halaman melalui satu modul.
2. Setiap halaman memiliki identitas/judul dan URL/slug sendiri.
3. Konten mendukung Bahasa Indonesia dan English sesuai keputusan multilingual PidieKab.
4. Versi setiap bahasa memiliki konten dan status publikasi masing-masing; publikasi satu bahasa tidak otomatis menerbitkan bahasa lainnya.
5. Halaman yang belum berstatus Published tidak ditampilkan kepada pengunjung.
6. Modul ini tidak dicampur dengan CRUD konten kategori utama.
7. Halaman yang menjadi bagian dari kebutuhan legal atau operasional, seperti Privacy Policy dan Disclaimer, harus tersedia dan ditinjau sebelum peluncuran publik.

## Batasan MVP

- Tidak memerlukan page builder visual atau penyusunan layout bebas.
- Tidak memerlukan blok konten kompleks atau drag-and-drop.
- Menggunakan struktur konten sederhana yang konsisten.
- Tidak membuat modul CRUD terpisah untuk setiap jenis halaman.
- Halaman Kontak berisi informasi kontak dan tautan yang dikelola Admin; bukan sistem pesan atau live chat internal.

## Hal yang ditetapkan pada tahap spesifikasi teknis

- Daftar field setiap halaman.
- Apakah halaman dapat dihapus permanen atau hanya diarsipkan.
- Pengelolaan slug dan riwayat URL.
- Format editor konten (rich text atau Markdown).
- Penempatan menu dan tautan footer.
- Apakah halaman tertentu, seperti Privacy Policy dan Disclaimer, memiliki field atau aturan khusus.

---

## Materi sumber: `cms-website-settings.md`

# CMS-007 — Modul Pengaturan Website

**Status: Disepakati.**

## Keputusan
Pengaturan umum PidieKab dikelola melalui satu modul **Pengaturan Website** pada CMS Admin. Pengaturan ini terpisah dari form konten dan halaman informasi.

## Kelompok pengaturan MVP

### 1. Identitas Website
- Nama website.
- Tagline.
- Deskripsi singkat website.

### 2. Logo dan Favicon
- Logo website.
- Favicon.
- Pemilihan aset menggunakan Media Library.

### 3. Informasi Kontak
- Email kontak.
- Nomor telepon bila tersedia.
- Alamat atau informasi lokasi organisasi pengelola bila relevan.
- Hanya informasi yang telah dikonfirmasi yang boleh ditampilkan.

### 4. Media Sosial
- URL akun Instagram, Facebook, YouTube, dan platform lain yang digunakan.
- Hanya tautan akun resmi/terverifikasi milik PidieKab yang ditampilkan.

### 5. Pengaturan Bahasa
- Bahasa default: Bahasa Indonesia.
- Bahasa yang didukung: Bahasa Indonesia (ID) dan English (EN).
- Pengaturan ini tidak mengubah keputusan URL prefix per bahasa, root `/` menuju `/id/`, atau status publikasi independen tiap bahasa.

### 6. SEO Global
- Default SEO title atau pola judul website.
- Default meta description.
- Informasi website yang diperlukan untuk metadata global.
- Metadata spesifik konten/halaman tetap dikelola pada form masing-masing dan dapat mengesampingkan default global.

## Prinsip pengelolaan
- Semua kelompok pengaturan berada dalam satu modul dengan bagian/tab yang jelas.
- Pengaturan global tidak menggantikan field spesifik konten.
- Perubahan logo, kontak, atau tautan sosial berlaku pada bagian publik website yang menggunakan pengaturan tersebut.
- Jangan menampilkan nilai kontak atau tautan sosial yang kosong.
- Pengaturan harus membedakan identitas PidieKab sebagai platform independen dan tidak boleh memberi kesan sebagai situs resmi pemerintah.

## Batasan MVP
- Tidak ada theme builder atau pengaturan layout bebas.
- Tidak ada pengaturan teknis server melalui CMS.
- Tidak ada pengelolaan script tracking bebas oleh Admin pada MVP.
- Rincian field wajib/opsional, preview perubahan, dan mekanisme penyimpanan ditetapkan pada spesifikasi CMS/teknis.

---

## Materi sumber: `cms-featured-content-management.md`

# CMS-012 — Pengelolaan Konten Unggulan Homepage

**Status: Disepakati.**

## Keputusan
Admin dapat memilih beberapa konten sebagai **Konten Unggulan** untuk ditampilkan pada homepage, menghapus penanda unggulan kapan saja, dan mengatur urutan tampil secara manual.

Konten Unggulan merupakan atribut/pemilihan editorial pada konten yang sudah ada, bukan modul atau tipe konten terpisah.

## Kemampuan Admin
1. Menandai konten sebagai Unggulan.
2. Menghapus tanda Unggulan.
3. Memilih beberapa konten unggulan sekaligus.
4. Mengatur urutan tampil secara manual.
5. Mengubah pilihan dan urutan kapan saja melalui CMS.

## Aturan tampilan publik
- Hanya konten dengan status Published pada bahasa yang sedang ditampilkan yang boleh muncul.
- Konten Draft atau Archived tidak boleh ditampilkan sebagai konten unggulan.
- Konten yang tidak lagi Published pada suatu bahasa otomatis tidak ditampilkan pada versi bahasa tersebut.
- Jika konten tidak memiliki foto yang boleh digunakan, ikuti fallback visual yang sudah ditetapkan untuk kartu homepage.
- Penghapusan permanen konten otomatis melepaskan pilihan unggulan dan urutannya.
- Urutan yang diatur Admin menjadi urutan tampilan pada bagian grid homepage “Jelajahi Pidie”.
- Konten unggulan tidak boleh muncul ganda dalam satu daftar homepage apabila juga terpilih melalui sumber Latest atau Random, sesuai HOME-005.

## Hubungan dengan Cerita tentang Pidie
Konten Unggulan untuk grid “Jelajahi Pidie” tidak otomatis berarti menjadi cerita utama pada bagian “Cerita tentang Pidie”. Pemilihan lead story dan supporting stories untuk bagian editorial tersebut tetap mengikuti kebutuhan editorial HOME-011 dan akan dirinci terpisah bila diperlukan.

## Batasan MVP
- Tidak ada modul CRUD terpisah untuk Konten Unggulan.
- Tidak ada promosi berbayar atau penjadwalan unggulan.
- Tidak ada ranking otomatis berdasarkan popularitas atau engagement.
- Jumlah maksimum konten unggulan dan perilaku jika belum ada pilihan ditetapkan pada spesifikasi UI/teknis.

---

## Materi sumber: `cms-homepage-editorial-story-curation.md`

# CMS-013 — Kurasi Cerita tentang Pidie

**Status: Disepakati.**

## Keputusan
Bagian homepage **Cerita tentang Pidie** menggunakan konten yang sudah dikelola pada modul Konten. Admin memilih satu konten sebagai **Lead Story** dan beberapa konten sebagai **Supporting Stories**, lalu mengatur urutan Supporting Stories secara manual.

Tidak dibuat modul atau tipe konten terpisah khusus untuk Cerita tentang Pidie.

## Kemampuan Admin
1. Memilih satu konten sebagai Lead Story.
2. Memilih beberapa konten sebagai Supporting Stories.
3. Mengatur urutan Supporting Stories secara manual.
4. Mengganti atau menghapus pilihan editorial kapan saja.

## Aturan publikasi dan bahasa
- Hanya versi bahasa dengan status Published pada locale aktif yang boleh ditampilkan.
- Jika konten pilihan tidak Published pada locale aktif, konten tersebut tidak ditampilkan pada locale tersebut.
- ID dan EN dapat memiliki ketersediaan cerita yang berbeda karena status publikasi per bahasa bersifat independen.
- Sistem tidak boleh menampilkan konten Draft atau Archived kepada pengunjung.
- Admin perlu dapat melihat status bahasa pada saat memilih konten agar memahami ketersediaannya.

## Aturan kurasi
- Lead Story hanya satu pada satu waktu untuk homepage.
- Supporting Stories dapat berjumlah lebih dari satu; batas jumlah final ditetapkan pada spesifikasi UI.
- Urutan Supporting Stories mengikuti urutan manual Admin.
- Konten yang sama tidak boleh ditampilkan dua kali dalam bagian Cerita tentang Pidie.
- Konten yang dipilih sebagai Lead Story tidak boleh sekaligus menjadi Supporting Story.
- Konten yang dihapus permanen otomatis dilepas dari kurasi.
- Kurasi Cerita tentang Pidie terpisah dari Konten Unggulan pada grid “Jelajahi Pidie”; pemilihan pada satu bagian tidak otomatis memilih bagian lainnya.

## Batasan MVP
- Tidak ada modul CRUD artikel editorial terpisah.
- Tidak ada workflow publikasi khusus untuk Cerita tentang Pidie.
- Tidak ada penjadwalan pergantian cerita.
- Tidak ada pemilihan otomatis berdasarkan popularitas atau engagement.
- Fallback ketika Lead Story tidak tersedia dan batas jumlah Supporting Stories ditentukan pada spesifikasi UI/teknis.

---

## Materi sumber: `cms-public-site-preview.md`

# CMS-014 — Preview Menggunakan Tampilan Website Publik

**Status: Disepakati berdasarkan arahan untuk melanjutkan rekomendasi.**

## Keputusan
Fitur Preview pada CMS Admin menggunakan **tampilan website sebenarnya** (public-site rendering), bukan sekadar preview teks atau tampilan sederhana di dalam form CMS.

Preview memperlihatkan bagaimana halaman akan terlihat pada website PidieKab, termasuk layout, foto, isi artikel, informasi kategori/lokasi yang relevan, dan bahasa yang sedang diedit.

## Perilaku Preview
1. Admin dapat membuka Preview dari form tambah/edit konten.
2. Preview menampilkan versi bahasa yang sedang dipilih pada tab ID atau EN.
3. Preview menggunakan komponen dan template halaman publik yang sama dengan halaman yang diterbitkan.
4. Preview memuat perubahan Draft/unsaved yang sedang diperiksa, setelah perubahan tersebut diproses dengan aman.
5. Preview tidak mengubah status konten menjadi Published dan tidak membuat konten tampil bagi pengunjung umum.
6. Preview konten yang sudah Published dapat memperlihatkan perubahan terbaru yang belum diterbitkan, tanpa mengubah versi publik.
7. Admin dapat kembali ke form edit setelah memeriksa preview.
8. Jika versi bahasa lain belum Published, preview tetap dapat menampilkan versi Draft tersebut kepada Admin dalam konteks bahasa yang dipilih.

## Keamanan dan akses
- Preview hanya dapat diakses oleh Admin yang sudah terautentikasi dan berwenang.
- URL/token preview tidak boleh menjadi akses publik yang tidak terkontrol.
- Halaman preview tidak boleh diindeks mesin pencari; gunakan kebijakan noindex dan pembatasan akses.
- Preview tidak boleh masuk sitemap atau hasil pencarian publik.
- Konten Rich Text harus disanitasi dan diproses dengan aturan keamanan yang sama seperti konten publik.
- Preview tidak boleh membocorkan konten Draft melalui endpoint atau asset yang dapat diakses publik tanpa otorisasi.

## Konsistensi tampilan
- Template preview harus mengikuti template detail publik untuk kategori terkait.
- Tampilan responsif harus dapat diperiksa pada ukuran desktop dan mobile.
- Jika foto utama belum tersedia, preview menggunakan fallback visual yang sama dengan perilaku halaman publik.
- Data yang belum diisi dapat ditampilkan sesuai aturan halaman publik; preview tidak boleh mengarang data.

## Batasan MVP
- Tidak ada sistem approval/review.
- Tidak ada shareable preview publik tanpa login.
- Tidak ada preview multi-device otomatis atau visual regression testing.
- Mekanisme sesi, token, route preview, dan cara mengirim draft yang belum tersimpan ditetapkan pada Technical Specification.

---
