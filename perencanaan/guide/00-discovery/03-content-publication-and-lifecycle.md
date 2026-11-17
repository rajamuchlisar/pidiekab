# 03 Content Publication And Lifecycle

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `content-publication-workflow.md`
- `content-preview-publication-workflow.md`
- `no-publication-verification-gate.md`
- `cms-permanent-content-deletion.md`

---

## Materi sumber: `content-publication-workflow.md`

# CG-002 — Alur Publikasi Konten (Diperbarui oleh CMS-010)

**Status: Disepakati. Dokumen ini memperbarui dan menggantikan alur Review pada CG-002 untuk MVP.**

## Keputusan

PidieKab tidak menggunakan tahap atau status Review dalam workflow publikasi MVP. Admin menyiapkan konten sebagai Draft, melakukan Preview untuk memeriksa tampilan, lalu dapat langsung menerbitkan versi bahasa yang siap.

**Preview adalah aksi pemeriksaan, bukan status workflow.**

## Status konten per bahasa

1. **Draft** — versi bahasa sedang dibuat atau diperbarui dan belum dipublikasikan.
2. **Published** — versi bahasa sudah diterbitkan dan dapat diakses publik.
3. **Archived** — versi bahasa ditarik dari publikasi dan tidak ditampilkan pada website publik.

Status ID dan EN tetap independen sesuai keputusan LNG-002. Admin dapat menerbitkan ID tanpa menerbitkan EN, atau sebaliknya.

## Alur kerja

1. Admin membuat konten; versi bahasa dimulai sebagai Draft.
2. Admin mengisi atau memperbarui field pada tab ID dan/atau EN.
3. Admin menggunakan **Preview** untuk melihat tampilan halaman sebelum publikasi.
4. Admin melakukan pemeriksaan mandiri terhadap isi, tautan, sumber, kelengkapan informasi, terjemahan, dan hak penggunaan gambar.
5. Admin memilih **Publish** untuk menerbitkan versi bahasa yang dipilih.
6. Versi bahasa yang Published tampil pada website sesuai aturan locale, URL, pencarian, homepage, dan sitemap.
7. Jika konten perlu diperbarui, Admin mengubah versi terkait. Perubahan pada versi yang sudah Published tidak langsung mengganti konten publik sebelum Admin melakukan Publish kembali.
8. Admin dapat mengarsipkan versi bahasa untuk menariknya dari website publik.

## Perilaku Preview

- Preview tersedia sebelum publikasi dan saat Admin mengedit konten yang sudah Published.
- Preview harus memperlihatkan perubahan Draft yang belum dipublikasikan.
- Preview hanya dapat diakses oleh Admin yang berwenang; tidak boleh terindeks mesin pencari atau terlihat oleh pengunjung umum.
- Preview menampilkan bahasa yang sedang dipilih, tanpa mengganti atau menerbitkan bahasa lain.
- Jika data belum lengkap, Preview tetap dapat membantu pemeriksaan; validasi wajib dilakukan saat Publish sesuai aturan DET-002.
- Mekanisme token/URL preview dan pengamanan akses ditentukan pada Technical Specification.

## Validasi saat Publish

- Validasi dilakukan untuk versi bahasa yang akan diterbitkan.
- Versi bahasa lain yang masih Draft tidak menghalangi publikasi versi yang siap.
- Persyaratan konten wajib, sumber yang dapat ditelusuri, dan pemeriksaan hak gambar tetap berlaku.
- Field kategori yang opsional mengikuti DET-002; field opsional yang kosong tidak ditampilkan.
- Kesalahan validasi harus ditampilkan kepada Admin dan tidak boleh mengubah status menjadi Published.

## Batasan MVP

- Tidak ada tahap Review.
- Tidak ada status Review atau antrean persetujuan.
- Tidak ada pemisahan role pembuat, reviewer, dan publisher.
- Satu role Administrator tetap berlaku untuk seluruh akun.
- Tidak ada penjadwalan publikasi pada MVP.
- Preview tidak mengubah status publikasi dan tidak membuat konten terlihat publik.

---

## Materi sumber: `content-preview-publication-workflow.md`

# CMS-010 — Alur Preview dan Publikasi Konten

**Status: Disepakati. Dokumen ini memperbarui dan menggantikan alur Review pada CG-002 untuk MVP.**

## Keputusan

PidieKab tidak menggunakan tahap atau status Review dalam workflow publikasi MVP. Admin menyiapkan konten sebagai Draft, melakukan Preview untuk memeriksa tampilan, lalu dapat langsung menerbitkan versi bahasa yang siap.

**Preview adalah aksi pemeriksaan, bukan status workflow.**

## Status konten per bahasa

1. **Draft** — versi bahasa sedang dibuat atau diperbarui dan belum dipublikasikan.
2. **Published** — versi bahasa sudah diterbitkan dan dapat diakses publik.
3. **Archived** — versi bahasa ditarik dari publikasi dan tidak ditampilkan pada website publik.

Status ID dan EN tetap independen sesuai keputusan LNG-002. Admin dapat menerbitkan ID tanpa menerbitkan EN, atau sebaliknya.

## Alur kerja

1. Admin membuat konten; versi bahasa dimulai sebagai Draft.
2. Admin mengisi atau memperbarui field pada tab ID dan/atau EN.
3. Admin menggunakan **Preview** untuk melihat tampilan halaman sebelum publikasi.
4. Admin melakukan pemeriksaan mandiri terhadap isi, tautan, sumber, kelengkapan informasi, terjemahan, dan hak penggunaan gambar.
5. Admin memilih **Publish** untuk menerbitkan versi bahasa yang dipilih.
6. Versi bahasa yang Published tampil pada website sesuai aturan locale, URL, pencarian, homepage, dan sitemap.
7. Jika konten perlu diperbarui, Admin mengubah versi terkait. Perubahan pada versi yang sudah Published tidak langsung mengganti konten publik sebelum Admin melakukan Publish kembali.
8. Admin dapat mengarsipkan versi bahasa untuk menariknya dari website publik.

## Perilaku Preview

- Preview tersedia sebelum publikasi dan saat Admin mengedit konten yang sudah Published.
- Preview harus memperlihatkan perubahan Draft yang belum dipublikasikan.
- Preview hanya dapat diakses oleh Admin yang berwenang; tidak boleh terindeks mesin pencari atau terlihat oleh pengunjung umum.
- Preview menampilkan bahasa yang sedang dipilih, tanpa mengganti atau menerbitkan bahasa lain.
- Jika data belum lengkap, Preview tetap dapat membantu pemeriksaan; validasi wajib dilakukan saat Publish sesuai aturan DET-002.
- Mekanisme token/URL preview dan pengamanan akses ditentukan pada Technical Specification.

## Validasi saat Publish

- Validasi dilakukan untuk versi bahasa yang akan diterbitkan.
- Versi bahasa lain yang masih Draft tidak menghalangi publikasi versi yang siap.
- Persyaratan konten wajib, sumber yang dapat ditelusuri, dan pemeriksaan hak gambar tetap berlaku.
- Field kategori yang opsional mengikuti DET-002; field opsional yang kosong tidak ditampilkan.
- Kesalahan validasi harus ditampilkan kepada Admin dan tidak boleh mengubah status menjadi Published.

## Batasan MVP

- Tidak ada tahap Review.
- Tidak ada status Review atau antrean persetujuan.
- Tidak ada pemisahan role pembuat, reviewer, dan publisher.
- Satu role Administrator tetap berlaku untuk seluruh akun.
- Tidak ada penjadwalan publikasi pada MVP.
- Preview tidak mengubah status publikasi dan tidak membuat konten terlihat publik.

---

## Materi sumber: `no-publication-verification-gate.md`

# GOV-001 — Tidak Ada Gerbang Verifikasi Sebelum Publikasi

**Status: Disepakati.**

## Keputusan
PidieKab tidak mewajibkan proses atau status verifikasi khusus sebagai prasyarat publikasi konten. Admin dapat memublikasikan konten setelah menyiapkan dan menyimpan konten melalui workflow CMS yang telah disepakati.

## Dampak pada workflow
- Status publikasi tetap Draft, Published, dan Archived sesuai CMS-010.
- Tidak ada status workflow `Verified`, `In Review`, atau `Needs Verification` sebagai syarat publikasi.
- Tidak ada approval/verifikasi oleh Admin lain sebelum Publish.
- Preview dan pemeriksaan mandiri oleh Admin tetap tersedia sesuai workflow CMS-010 dan CMS-014.
- Admin bertanggung jawab atas konten yang dipublikasikan.

## Sumber informasi
Keputusan SRC-001 mengenai pencatatan sumber informasi tetap berlaku sebagai metadata/referensi konten. Pencatatan sumber tidak berarti konten harus melewati status verifikasi atau approval sebelum dipublikasikan.

## Batasan
- Tidak membangun verification workflow pada MVP.
- Tidak ada verifikasi pihak ketiga atau klaim bahwa semua informasi telah diverifikasi.
- Tidak menampilkan badge “Terverifikasi” pada konten hanya karena konten diterbitkan.
- Fitur peninjauan dan pembaruan konten di masa depan dapat dipertimbangkan terpisah jika kebutuhan operasional muncul.

## Catatan
Tidak adanya gerbang verifikasi bukan berarti PidieKab menyatakan semua informasi selalu akurat. Admin tetap dapat memperbarui atau mengarsipkan informasi yang keliru atau sudah tidak relevan melalui CMS.

---

## Materi sumber: `cms-permanent-content-deletion.md`

# CMS-011 — Penghapusan Konten Permanen

**Status: Disepakati.**

## Keputusan
PidieKab menggunakan **Permanent Delete** untuk penghapusan konten dari CMS. Tidak ada Trash/Recycle Bin atau mekanisme pemulihan konten melalui CMS setelah penghapusan dikonfirmasi.

## Perilaku penghapusan
1. Admin dapat memilih aksi Hapus pada konten.
2. Sistem menampilkan dialog konfirmasi yang jelas sebelum penghapusan.
3. Konfirmasi harus menjelaskan bahwa penghapusan bersifat permanen dan tidak dapat dipulihkan melalui CMS.
4. Setelah dikonfirmasi, record konten dan seluruh versi bahasa yang terkait dihapus permanen sesuai relasi data.
5. Konten yang sedang Published harus terlebih dahulu dihilangkan dari website publik; URL terkait tidak boleh tetap menyajikan konten yang sudah dihapus.
6. Jika konten terpilih sebagai konten unggulan atau cerita editorial homepage, penempatan tersebut harus dilepas saat konten dihapus.
7. Media yang digunakan bersama oleh konten lain tidak ikut dihapus.
8. Penghapusan dicatat dalam audit log dengan identitas administrator, waktu, dan identitas konten yang dihapus. Audit log tidak menyimpan salinan isi konten yang telah dihapus.

## Batasan dan perlindungan
- Tidak ada Soft Delete.
- Tidak ada Trash/Recycle Bin.
- Tidak ada pemulihan konten melalui CMS.
- Penghapusan harus dibatasi pada relasi yang memang dimiliki konten tersebut; data media bersama dan sumber yang masih direferensikan konten lain harus tetap aman.
- Backup operasional dapat tetap tersedia sesuai kebijakan backup sistem, tetapi bukan fitur pemulihan mandiri dari CMS.

## Hal yang ditetapkan pada Technical Specification
- Strategi transaksi dan urutan penghapusan relasi.
- Kebijakan foreign key (`CASCADE`, `RESTRICT`, atau penanganan eksplisit) per relasi.
- Penanganan URL lama dan HTTP response setelah konten dihapus.
- Kebijakan retensi audit log dan backup.

---
