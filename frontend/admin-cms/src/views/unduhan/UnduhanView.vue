<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'
import BladeIcon from '../../components/common/BladeIcon.vue'

const notification = useNotificationStore()

interface DownloadItem {
  id: number
  title_id: string
  title_en: string
  category: string
  description_id: string
  description_en: string
  file_url: string
  file_name: string
  file_size: number
  file_type: string
  download_count: number
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

const items = ref<DownloadItem[]>([])
const isLoading = ref(true)
const isSubmitting = ref(false)
const isUploading = ref(false)

const selectedCategory = ref<string>('Semua')
const searchQuery = ref<string>('')

// Modal state
const showModal = ref(false)
const isEditing = ref(false)
const currentId = ref<number | null>(null)

// Delete confirm modal
const showDeleteModal = ref(false)
const itemToDelete = ref<DownloadItem | null>(null)

const standardCategories = [
  'Panduan Wisata',
  'Regulasi Daerah',
  'Formulir Publik',
  'Laporan & Statistik',
  'Peta & Brosur',
  'Informasi Publik',
  'Dokumen Lainnya',
]

const form = reactive({
  title_id: '',
  title_en: '',
  category: 'Panduan Wisata',
  custom_category: '',
  description_id: '',
  description_en: '',
  file_url: '',
  file_name: '',
  file_size: 0,
  file_type: 'PDF',
  is_active: true,
  sort_order: 1,
})

const fileInputRef = ref<HTMLInputElement | null>(null)

function formatFileSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const availableCategories = computed(() => {
  const cats = new Set<string>(standardCategories)
  items.value.forEach((it) => {
    if (it.category) cats.add(it.category)
  })
  return Array.from(cats)
})

const filteredItems = computed(() => {
  return items.value.filter((it) => {
    const matchCat =
      selectedCategory.value === 'Semua' || it.category === selectedCategory.value
    const q = searchQuery.value.toLowerCase().trim()
    const matchQuery =
      !q ||
      it.title_id.toLowerCase().includes(q) ||
      it.title_en.toLowerCase().includes(q) ||
      it.file_name.toLowerCase().includes(q) ||
      it.category.toLowerCase().includes(q)
    return matchCat && matchQuery
  })
})

const stats = computed(() => {
  const total = items.value.length
  const totalDownloads = items.value.reduce((acc, it) => acc + (it.download_count || 0), 0)
  const activeCount = items.value.filter((it) => it.is_active).length
  return { total, totalDownloads, activeCount }
})

async function loadData() {
  isLoading.value = true
  try {
    const res = await api.get<{ items: DownloadItem[]; total: number }>('/api/v1/admin/downloads?per_page=100')
    items.value = res.items || []
  } catch (err: any) {
    notification.error('Gagal Memuat Data', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function openCreateModal() {
  isEditing.value = false
  currentId.value = null
  form.title_id = ''
  form.title_en = ''
  form.category = 'Panduan Wisata'
  form.custom_category = ''
  form.description_id = ''
  form.description_en = ''
  form.file_url = ''
  form.file_name = ''
  form.file_size = 0
  form.file_type = 'PDF'
  form.is_active = true
  form.sort_order = items.value.length + 1
  showModal.value = true
}

function openEditModal(it: DownloadItem) {
  isEditing.value = true
  currentId.value = it.id
  form.title_id = it.title_id
  form.title_en = it.title_en
  if (standardCategories.includes(it.category)) {
    form.category = it.category
    form.custom_category = ''
  } else {
    form.category = 'custom'
    form.custom_category = it.category
  }
  form.description_id = it.description_id
  form.description_en = it.description_en
  form.file_url = it.file_url
  form.file_name = it.file_name
  form.file_size = it.file_size
  form.file_type = it.file_type || 'PDF'
  form.is_active = it.is_active
  form.sort_order = it.sort_order
  showModal.value = true
}

async function handleFileUpload(e: Event) {
  const target = e.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return

  const file = target.files[0]
  if (!file) return

  isUploading.value = true

  const formData = new FormData()
  formData.append('file', file)

  try {
    const res = await api.upload<{
      file_url: string
      file_name: string
      file_size: number
      file_type: string
    }>('/api/v1/admin/downloads/upload', formData)

    form.file_url = res.file_url
    form.file_name = res.file_name
    form.file_size = res.file_size
    form.file_type = res.file_type

    if (!form.title_id) {
      // Auto-populate title from filename if title is still empty
      const cleanName = file.name.replace(/\.[^/.]+$/, '').replace(/[-_]/g, ' ')
      form.title_id = cleanName.charAt(0).toUpperCase() + cleanName.slice(1)
    }

    notification.success('Berkas Terunggah', `File ${file.name} berhasil diunggah.`)
  } catch (err: any) {
    notification.error('Gagal Mengunggah Berkas', err.detail || 'Terjadi kesalahan saat mengunggah file.', err.requestId)
  } finally {
    isUploading.value = false
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  }
}

async function saveDownload() {
  if (!form.title_id.trim()) {
    notification.show('warning', 'Validasi Gagal', 'Judul Dokumen (Bahasa Indonesia) wajib diisi.')
    return
  }
  if (!form.file_url.trim()) {
    notification.show('warning', 'Validasi Gagal', 'Berkas atau URL unduhan wajib disediakan.')
    return
  }

  const category = form.category === 'custom' ? form.custom_category.trim() || 'Umum' : form.category

  isSubmitting.value = true
  try {
    const payload = {
      title_id: form.title_id.trim(),
      title_en: form.title_en.trim(),
      category: category,
      description_id: form.description_id.trim(),
      description_en: form.description_en.trim(),
      file_url: form.file_url.trim(),
      file_name: form.file_name.trim(),
      file_size: form.file_size,
      file_type: form.file_type.trim().toUpperCase(),
      is_active: form.is_active,
      sort_order: form.sort_order,
    }

    if (isEditing.value && currentId.value) {
      await api.put(`/api/v1/admin/downloads/${currentId.value}`, payload)
      notification.success('Dokumen Diperbarui', 'Data unduhan berhasil diperbarui.')
    } else {
      await api.post('/api/v1/admin/downloads', payload)
      notification.success('Dokumen Ditambahkan', 'Data unduhan baru berhasil disimpan.')
    }

    showModal.value = false
    await loadData()
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan saat menyimpan.', err.requestId)
  } finally {
    isSubmitting.value = false
  }
}

async function toggleActive(it: DownloadItem) {
  const previousState = it.is_active
  it.is_active = !previousState
  try {
    await api.put(`/api/v1/admin/downloads/${it.id}`, {
      title_id: it.title_id,
      title_en: it.title_en,
      category: it.category,
      description_id: it.description_id,
      description_en: it.description_en,
      file_url: it.file_url,
      file_name: it.file_name,
      file_size: it.file_size,
      file_type: it.file_type,
      is_active: it.is_active,
      sort_order: it.sort_order,
    })
    notification.success('Status Diperbarui', `Dokumen kini ${it.is_active ? 'aktif' : 'non-aktif'}.`)
  } catch (err: any) {
    it.is_active = previousState
    notification.error('Gagal Mengubah Status', err.detail || 'Terjadi kesalahan sistem.')
  }
}

function confirmDelete(it: DownloadItem) {
  itemToDelete.value = it
  showDeleteModal.value = true
}

async function executeDelete() {
  if (!itemToDelete.value) return
  try {
    await api.delete(`/api/v1/admin/downloads/${itemToDelete.value.id}`)
    notification.success('Dokumen Dihapus', `Dokumen "${itemToDelete.value.title_id}" berhasil dihapus.`)
    showDeleteModal.value = false
    itemToDelete.value = null
    await loadData()
  } catch (err: any) {
    notification.error('Gagal Menghapus', err.detail || 'Terjadi kesalahan sistem.')
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <AppLayout>
    <div class="unduhan-page">
      <!-- Page Header -->
      <div class="page-header">
        <div>
          <h2 class="page-title">Pusat Unduhan & Dokumen Publik</h2>
          <p class="page-desc">
            Kelola dokumen publik, panduan wisata, peta, regulasi, dan formulir untuk masyarakat & pengunjung.
          </p>
        </div>
        <button class="btn btn-primary" @click="openCreateModal">
          <BladeIcon name="plus" size="18" />
          <span>Tambah Dokumen Baru</span>
        </button>
      </div>

      <!-- Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon icon-primary">
            <BladeIcon name="arrow-down-tray" size="24" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">Total Dokumen</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon icon-success">
            <BladeIcon name="check-circle" size="24" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.activeCount }}</div>
            <div class="stat-label">Dokumen Aktif</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon icon-info">
            <BladeIcon name="chart-bar" size="24" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalDownloads }}</div>
            <div class="stat-label">Total Kali Diunduh</div>
          </div>
        </div>
      </div>

      <!-- Filters & Search Toolbar -->
      <div class="card toolbar-card">
        <div class="toolbar-inner">
          <div class="toolbar-left">
            <label class="filter-label">Kategori:</label>
            <select v-model="selectedCategory" class="form-select">
              <option value="Semua">Semua Kategori</option>
              <option v-for="cat in availableCategories" :key="cat" :value="cat">
                {{ cat }}
              </option>
            </select>
          </div>

          <div class="toolbar-right">
            <div class="search-input-wrapper">
              <BladeIcon name="magnifying-glass" size="18" class="search-icon" />
              <input
                v-model="searchQuery"
                type="text"
                class="form-input search-input"
                placeholder="Cari dokumen, judul, atau file..."
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Data Table -->
      <div class="card table-card">
        <div v-if="isLoading" class="loading-state">
          <BladeIcon name="arrow-path" size="24" class="spin-icon" />
          <span>Memuat data unduhan...</span>
        </div>

        <div v-else-if="filteredItems.length === 0" class="empty-state">
          <BladeIcon name="document-text" size="48" class="empty-icon" />
          <h3>Tidak ada dokumen ditemukan</h3>
          <p>Belum ada berkas unduhan yang cocok dengan kriteria pencarian Anda.</p>
        </div>

        <div v-else class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 50px;">Tipe</th>
                <th>Judul & Dokumen</th>
                <th>Kategori</th>
                <th>Ukuran</th>
                <th style="text-align: center;">Diunduh</th>
                <th style="text-align: center;">Status</th>
                <th style="width: 140px; text-align: right;">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="it in filteredItems" :key="it.id">
                <td>
                  <span class="file-badge" :class="'badge-' + it.file_type.toLowerCase()">
                    {{ it.file_type || 'FILE' }}
                  </span>
                </td>
                <td>
                  <div class="item-title-box">
                    <span class="item-main-title">{{ it.title_id }}</span>
                    <span v-if="it.title_en" class="item-sub-title">EN: {{ it.title_en }}</span>
                    <a :href="it.file_url" target="_blank" rel="noopener noreferrer" class="item-file-link">
                      <BladeIcon name="arrow-top-right-on-square" size="13" />
                      <span>{{ it.file_name || it.file_url }}</span>
                    </a>
                  </div>
                </td>
                <td>
                  <span class="category-pill">{{ it.category }}</span>
                </td>
                <td class="text-nowrap">
                  {{ formatFileSize(it.file_size) }}
                </td>
                <td style="text-align: center;">
                  <span class="download-counter">
                    <BladeIcon name="arrow-down-tray" size="14" />
                    {{ it.download_count }}
                  </span>
                </td>
                <td style="text-align: center;">
                  <button
                    class="status-toggle-btn"
                    :class="{ active: it.is_active }"
                    @click="toggleActive(it)"
                    :title="it.is_active ? 'Klik untuk non-aktifkan' : 'Klik untuk aktifkan'"
                  >
                    {{ it.is_active ? 'Aktif' : 'Non-aktif' }}
                  </button>
                </td>
                <td>
                  <div class="action-buttons">
                    <button class="btn-icon" @click="openEditModal(it)" title="Edit Dokumen">
                      <BladeIcon name="pencil-square" size="16" />
                    </button>
                    <button class="btn-icon btn-icon-danger" @click="confirmDelete(it)" title="Hapus Dokumen">
                      <BladeIcon name="trash" size="16" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- CREATE / EDIT MODAL -->
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal-box">
          <div class="modal-header">
            <h3 class="modal-title">{{ isEditing ? 'Edit Unduhan Dokumen' : 'Tambah Unduhan Dokumen Baru' }}</h3>
            <button class="modal-close" @click="showModal = false">&times;</button>
          </div>

          <div class="modal-body">
            <!-- Title ID -->
            <div class="form-group">
              <label class="form-label">Judul Dokumen (Bahasa Indonesia) *</label>
              <input
                v-model="form.title_id"
                type="text"
                class="form-input"
                placeholder="Misal: Panduan Wisata dan Budaya Kabupaten Pidie 2026"
              />
            </div>

            <!-- Title EN -->
            <div class="form-group">
              <label class="form-label">Judul Dokumen (English)</label>
              <input
                v-model="form.title_en"
                type="text"
                class="form-input"
                placeholder="e.g. Pidie Regency Official Travel and Culture Guide 2026"
              />
            </div>

            <!-- Category -->
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">Kategori Dokumen *</label>
                <select v-model="form.category" class="form-select">
                  <option v-for="cat in standardCategories" :key="cat" :value="cat">
                    {{ cat }}
                  </option>
                  <option value="custom">+ Kategori Kustom / Lainnya</option>
                </select>
              </div>

              <div v-if="form.category === 'custom'" class="form-group">
                <label class="form-label">Nama Kategori Baru *</label>
                <input
                  v-model="form.custom_category"
                  type="text"
                  class="form-input"
                  placeholder="Masukkan nama kategori"
                />
              </div>

              <div class="form-group">
                <label class="form-label">Urutan Tampil (Sort Order)</label>
                <input
                  v-model.number="form.sort_order"
                  type="number"
                  class="form-input"
                  min="0"
                />
              </div>
            </div>

            <!-- File Upload Box -->
            <div class="form-group">
              <label class="form-label">Berkas Dokumen *</label>
              <div class="upload-dropzone">
                <div v-if="isUploading" class="uploading-state">
                  <BladeIcon name="arrow-path" size="24" class="spin-icon" />
                  <span>Mengunggah berkas ke server...</span>
                </div>

                <div v-else class="upload-actions-row">
                  <input
                    ref="fileInputRef"
                    type="file"
                    class="hidden-file-input"
                    @change="handleFileUpload"
                  />
                  <button
                    type="button"
                    class="btn btn-outline"
                    @click="fileInputRef?.click()"
                  >
                    <BladeIcon name="arrow-up-tray" size="18" />
                    <span>Pilih Berkas dari Komputer</span>
                  </button>
                  <span class="upload-hint">Mendukung PDF, DOCX, XLSX, ZIP, gambar (Maks. 50MB)</span>
                </div>

                <!-- File Info Preview -->
                <div v-if="form.file_url" class="file-preview-card">
                  <div class="file-preview-info">
                    <span class="file-badge" :class="'badge-' + form.file_type.toLowerCase()">
                      {{ form.file_type }}
                    </span>
                    <div>
                      <div class="file-preview-name">{{ form.file_name || form.file_url }}</div>
                      <div class="file-preview-size">{{ formatFileSize(form.file_size) }}</div>
                    </div>
                  </div>
                  <a :href="form.file_url" target="_blank" class="btn-xs btn-outline">Buka Berkas</a>
                </div>
              </div>
            </div>

            <!-- Manual File URL Input (Optional alternative) -->
            <div class="form-group">
              <label class="form-label">Atau Masukkan URL Langsung (Google Drive, Cloud, dll.)</label>
              <input
                v-model="form.file_url"
                type="text"
                class="form-input"
                placeholder="https://... atau /media/..."
              />
            </div>

            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">Ekstensi / Tipe Berkas</label>
                <input
                  v-model="form.file_type"
                  type="text"
                  class="form-input"
                  placeholder="PDF, DOCX, ZIP, dll."
                />
              </div>
              <div class="form-group">
                <label class="form-label">Nama Berkas Asli (Opsional)</label>
                <input
                  v-model="form.file_name"
                  type="text"
                  class="form-input"
                  placeholder="dokumen.pdf"
                />
              </div>
            </div>

            <!-- Descriptions -->
            <div class="form-group">
              <label class="form-label">Deskripsi / Ringkasan Dokumen (Bahasa Indonesia)</label>
              <textarea
                v-model="form.description_id"
                class="form-textarea"
                rows="2"
                placeholder="Ringkasan singkat isi dokumen..."
              ></textarea>
            </div>

            <div class="form-group">
              <label class="form-label">Deskripsi / Ringkasan Dokumen (English)</label>
              <textarea
                v-model="form.description_en"
                class="form-textarea"
                rows="2"
                placeholder="Brief summary of document contents..."
              ></textarea>
            </div>

            <!-- Active Checkbox -->
            <div class="form-group">
              <label class="checkbox-label">
                <input v-model="form.is_active" type="checkbox" class="form-checkbox" />
                <span>Publikasikan & Tampilkan di Website (Aktif)</span>
              </label>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-outline" @click="showModal = false">Batal</button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="isSubmitting || isUploading"
              @click="saveDownload"
            >
              {{ isSubmitting ? 'Menyimpan...' : 'Simpan Dokumen' }}
            </button>
          </div>
        </div>
      </div>

      <!-- DELETE CONFIRMATION MODAL -->
      <div v-if="showDeleteModal" class="modal-backdrop" @click.self="showDeleteModal = false">
        <div class="modal-box modal-box-sm">
          <div class="modal-header">
            <h3 class="modal-title">Konfirmasi Hapus</h3>
            <button class="modal-close" @click="showDeleteModal = false">&times;</button>
          </div>
          <div class="modal-body">
            <p>
              Apakah Anda yakin ingin menghapus berkas unduhan
              <strong>"{{ itemToDelete?.title_id }}"</strong>?
            </p>
            <p class="modal-warning">Tindakan ini tidak dapat dibatalkan.</p>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline" @click="showDeleteModal = false">Batal</button>
            <button class="btn btn-danger" @click="executeDelete">Ya, Hapus</button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.unduhan-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-title {
  margin: 0 0 0.25rem 0;
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
}

.page-desc {
  margin: 0;
  font-size: 0.85rem;
  color: #64748b;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.15rem 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
}

.stat-icon {
  width: 46px;
  height: 46px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-primary {
  background: #ecfdf5;
  color: #059669;
}

.icon-success {
  background: #eff6ff;
  color: #2563eb;
}

.icon-info {
  background: #fdf4ff;
  color: #c026d3;
}

.stat-value {
  font-size: 1.4rem;
  font-weight: 800;
  color: #0f172a;
  line-height: 1.2;
}

.stat-label {
  font-size: 0.8rem;
  color: #64748b;
  margin-top: 0.15rem;
}

/* Toolbar */
.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.toolbar-card {
  padding: 0.85rem 1.25rem;
}

.toolbar-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #475569;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  color: #94a3b8;
  pointer-events: none;
}

.search-input {
  padding-left: 2.25rem !important;
  width: 260px;
}

/* Table */
.table-card {
  overflow: hidden;
}

.table-responsive {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  text-align: left;
}

.data-table th {
  background: #f8fafc;
  padding: 0.75rem 1rem;
  font-weight: 700;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.data-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  color: #334155;
  vertical-align: middle;
}

.data-table tr:hover td {
  background: #fafafa;
}

.item-title-box {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.item-main-title {
  font-weight: 700;
  color: #0f172a;
}

.item-sub-title {
  font-size: 0.775rem;
  color: #64748b;
}

.item-file-link {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  color: #059669;
  text-decoration: none;
  word-break: break-all;
}

.item-file-link:hover {
  text-decoration: underline;
}

/* File Badge */
.file-badge {
  display: inline-block;
  padding: 0.2rem 0.45rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: #64748b;
  color: #ffffff;
}

.badge-pdf { background: #dc2626; color: #ffffff; }
.badge-docx, .badge-doc { background: #2563eb; color: #ffffff; }
.badge-xlsx, .badge-xls { background: #059669; color: #ffffff; }
.badge-zip, .badge-rar { background: #d97706; color: #ffffff; }

.category-pill {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #475569;
}

.download-counter {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: #64748b;
  background: #f8fafc;
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}

.status-toggle-btn {
  padding: 0.25rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 700;
  border-radius: 9999px;
  cursor: pointer;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  color: #64748b;
  transition: all 0.15s ease;
}

.status-toggle-btn.active {
  background: #ecfdf5;
  color: #059669;
  border-color: #a7f3d0;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.35rem;
}

.btn-icon {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  padding: 0.35rem;
  color: #475569;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.btn-icon:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.btn-icon-danger:hover {
  background: #fee2e2;
  color: #dc2626;
  border-color: #fca5a5;
}

/* Modals */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-box {
  background: #ffffff;
  border-radius: 12px;
  width: 100%;
  max-width: 620px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.modal-box-sm {
  max-width: 420px;
}

.modal-header {
  padding: 1.15rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #94a3b8;
  cursor: pointer;
}

.modal-close:hover {
  color: #0f172a;
}

.modal-body {
  padding: 1.25rem 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.modal-warning {
  color: #dc2626;
  font-size: 0.85rem;
  margin: 0.5rem 0 0 0;
}

/* Form inputs */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-label {
  font-size: 0.825rem;
  font-weight: 600;
  color: #334155;
}

.form-input, .form-select, .form-textarea {
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: #ffffff;
  color: #1e293b;
  outline: none;
}

.form-input:focus, .form-select:focus, .form-textarea:focus {
  border-color: #059669;
  box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.15);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.85rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: #1e293b;
  cursor: pointer;
}

.form-checkbox {
  width: 16px;
  height: 16px;
  accent-color: #059669;
}

/* Upload dropzone */
.upload-dropzone {
  border: 2px dashed #cbd5e1;
  background: #f8fafc;
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.hidden-file-input {
  display: none;
}

.upload-actions-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  flex-wrap: wrap;
}

.upload-hint {
  font-size: 0.775rem;
  color: #64748b;
}

.uploading-state {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.85rem;
  color: #059669;
  font-weight: 600;
}

.file-preview-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.6rem 0.75rem;
  gap: 0.75rem;
}

.file-preview-info {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
}

.file-preview-name {
  font-size: 0.825rem;
  font-weight: 600;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-preview-size {
  font-size: 0.725rem;
  color: #64748b;
}

/* Common buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1rem;
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
}

.btn-primary {
  background: #059669;
  color: #ffffff;
}

.btn-primary:hover:not(:disabled) {
  background: #047857;
}

.btn-outline {
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #334155;
}

.btn-outline:hover {
  background: #f8fafc;
}

.btn-danger {
  background: #dc2626;
  color: #ffffff;
}

.btn-danger:hover {
  background: #b91c1c;
}

.btn-sm { padding: 0.35rem 0.65rem; font-size: 0.775rem; }
.btn-xs { padding: 0.2rem 0.45rem; font-size: 0.725rem; text-decoration: none; border-radius: 4px; }

.loading-state, .empty-state {
  padding: 3rem 1.5rem;
  text-align: center;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.empty-icon {
  color: #cbd5e1;
  margin-bottom: 0.5rem;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
