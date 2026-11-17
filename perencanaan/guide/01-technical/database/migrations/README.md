# SQL Migration Baseline

File `001_initial_schema.sql` berisi baseline DDL PostgreSQL untuk domain utama PidieKab.

Catatan sebelum produksi:
1. Tinjau versi PostgreSQL yang akan digunakan.
2. Validasi seluruh relasi dan `ON DELETE` behavior terhadap keputusan CMS.
3. Implementasikan publish transaction yang memvalidasi minimal satu `content_sources`.
4. Implementasikan update `search_vector` per locale dan uji kualitas pencarian ID/EN.
5. Implementasikan claim/retry job worker dengan transaksi dan `FOR UPDATE SKIP LOCKED`.
6. Implementasikan validasi redirect loop/chain pada application service.
7. Tambahkan seed master wilayah Pidie dan bootstrap Admin melalui prosedur aman.
8. Migration ini adalah baseline awal; jangan jalankan pada database produksi yang sudah berisi data tanpa rencana migrasi dan backup.
