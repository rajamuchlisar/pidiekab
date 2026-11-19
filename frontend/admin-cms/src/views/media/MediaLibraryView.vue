<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const mediaList = ref<any[]>([])
const isLoading = ref(true)
const showUploadModal = ref(false)
const isUploading = ref(false)

const selectedFile = ref<File | null>(null)
const previewUrl = ref('')

const uploadForm = reactive({
  title: '',
  altTextId: '',
  altTextEn: '',
  attribution: '',
  sourceOrigin: '',
})

async function loadMedia() {
  isLoading.value = true
  try {
    const res = await api.get<{ items: any[]; total: number }>('/api/v1/admin/media?per_page=60')
    if (res?.items) {
      mediaList.value = res.items
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Media', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function onFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    if (file.size > 10 * 1024 * 1024) {
      alert('Ukuran file maksimal 10MB.')
      return
    }
    selectedFile.value = file
    previewUrl.value = URL.createObjectURL(file)
    if (!uploadForm.title) {
      uploadForm.title = file.name.replace(/\.[^/.]+$/, '').replace(/[-_]/g, ' ')
    }
  }
}

async function handleUpload() {
  if (!selectedFile.value) {
    alert('Silakan pilih berkas gambar.')
    return
  }

  isUploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', selectedFile.value)
    fd.append('title', uploadForm.title)
    fd.append('alt_text_id', uploadForm.altTextId)
    fd.append('alt_text_en', uploadForm.altTextEn)
    fd.append('attribution', uploadForm.attribution)
    fd.append('source_origin', uploadForm.sourceOrigin)

    await api.upload('/api/v1/admin/media/upload', fd)
    notification.success('Gambar Berhasil Diunggah', 'Aset media baru telah disimpan.')
    showUploadModal.value = false
    selectedFile.value = null
    previewUrl.value = ''
    uploadForm.title = ''
    uploadForm.altTextId = ''
    uploadForm.altTextEn = ''
    uploadForm.attribution = ''
    uploadForm.sourceOrigin = ''
    loadMedia()
  } catch (err: any) {
    notification.error('Gagal Mengunggah', err.detail || 'Format file ditolak atau ukuran melebihi batas.', err.requestId)
  } finally {
    isUploading.value = false
  }
}

async function deleteMedia(item: any) {
  const confirmed = confirm(
    `Apakah Anda yakin ingin menghapus media "${item.title || item.fileName}"?\n\nMedia yang sedang digunakan sebagai foto utama artikel tidak dapat dihapus.`
  )
  if (!confirmed) return

  try {
    await api.del(`/api/v1/admin/media/${item.id}`)
    notification.success('Media Dihapus', 'File media berhasil dihapus.')
    loadMedia()
  } catch (err: any) {
    notification.error('Gagal Menghapus Media', err.detail || 'Gambar sedang digunakan dalam artikel.', err.requestId)
  }
}

function copyMediaUrl(url: string) {
  navigator.clipboard.writeText(url)
  notification.success('URL Disalin', 'Tautan media telah disalin ke clipboard.')
}

onMounted(() => {
  loadMedia()
})
</script>

<template>
  <AppLayout>
    <div class="media-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Media Library</h2>
          <p class="page-desc">Kelola repositori foto rights-cleared untuk artikel dan beranda (JPEG, PNG, WebP).</p>
        </div>
        <button class="btn btn-primary" @click="showUploadModal = true">
          <BladeIcon name="cloud-arrow-up" size="18" />
          <span>Unggah Foto Baru</span>
        </button>
      </div>

      <div class="media-container">
        <div v-if="isLoading" class="loading-state">Memuat galeri media...</div>

        <div v-else-if="mediaList.length === 0" class="empty-state">
          <span class="empty-icon"><BladeIcon name="photo" size="48" /></span>
          <p>Belum ada media foto yang diunggah. Mulai dengan mengunggah foto rights-cleared.</p>
          <button class="btn btn-primary btn-sm" @click="showUploadModal = true">
            <BladeIcon name="cloud-arrow-up" size="16" />
            <span>Unggah Foto</span>
          </button>
        </div>

        <div v-else class="media-grid">
          <div v-for="item in mediaList" :key="item.id" class="media-card">
            <div class="thumb-wrap">
              <img :src="item.url" :alt="item.altTextId || item.title" class="media-img" loading="lazy" />
              <span v-if="item.inUse" class="badge-in-use">Digunakan</span>
            </div>
            <div class="media-meta">
              <span class="media-name" :title="item.title || item.fileName">
                {{ item.title || item.fileName }}
              </span>
              <span class="media-specs">
                {{ item.mimeType?.replace('image/', '').toUpperCase() }} &bull;
                {{ (item.fileSize / 1024).toFixed(1) }} KB
              </span>
              <div class="media-actions">
                <button class="btn-xs btn-outline" @click="copyMediaUrl(item.url)">
                  <BladeIcon name="link" size="12" />
                  <span>Salin URL</span>
                </button>
                <button
                  class="btn-xs btn-danger"
                  :disabled="item.inUse"
                  @click="deleteMedia(item)"
                  :title="item.inUse ? 'Media sedang digunakan dalam artikel' : 'Hapus media'"
                >
                  <BladeIcon name="trash" size="12" />
                  <span>Hapus</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Upload Modal -->
      <div v-if="showUploadModal" class="modal-backdrop" @click.self="showUploadModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Unggah Foto Media</h3>
            <button class="modal-close" @click="showUploadModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <form @submit.prevent="handleUpload" class="modal-form">
            <!-- File Drop / Select -->
            <div class="upload-dropzone">
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp"
                @change="onFileSelect"
                class="file-input"
                id="media-file-input"
              />
              <label for="media-file-input" class="dropzone-label">
                <img v-if="previewUrl" :src="previewUrl" alt="Preview" class="preview-img" />
                <div v-else class="dropzone-placeholder">
                  <span class="upload-icon"><BladeIcon name="folder" size="36" /></span>
                  <strong>Pilih Berkas Foto</strong>
                  <span class="drop-hint">Maksimal 10MB (JPEG, PNG, WebP). SVG/Executable ditolak.</span>
                </div>
              </label>
            </div>

            <div class="form-group">
              <label>Judul Media / Foto *</label>
              <input v-model="uploadForm.title" type="text" class="input-control" placeholder="Pantai Pelangi Sigli" required />
            </div>

            <div class="form-row">
              <div class="form-group flex-1">
                <label>Alt Text (Bahasa Indonesia)</label>
                <input v-model="uploadForm.altTextId" type="text" class="input-control" placeholder="Pemandangan matahari terbit di Pantai Pelangi" />
              </div>
              <div class="form-group flex-1">
                <label>Alt Text (English)</label>
                <input v-model="uploadForm.altTextEn" type="text" class="input-control" placeholder="Sunrise view at Pelangi Beach Sigli" />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group flex-1">
                <label>Atribusi / Fotografer / Hak Cipta</label>
                <input v-model="uploadForm.attribution" type="text" class="input-control" placeholder="Foto: Humas Pemkab Pidie / Lisensi Bebas" />
              </div>
              <div class="form-group flex-1">
                <label>Sumber / Asal Berkas</label>
                <input v-model="uploadForm.sourceOrigin" type="text" class="input-control" placeholder="Dokumentasi Lapangan / Disporapar" />
              </div>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showUploadModal = false">Batal</button>
              <button type="submit" class="btn btn-primary" :disabled="isUploading || !selectedFile">
                {{ isUploading ? 'Mengunggah...' : 'Unggah Foto' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.media-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.media-container {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
}

.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 1.25rem;
}

.media-card {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  transition: transform 0.15s, box-shadow 0.15s;
}

.media-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px -2px rgba(0, 0, 0, 0.08);
}

.thumb-wrap {
  position: relative;
  width: 100%;
  height: 130px;
  background: #f1f5f9;
}

.media-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.badge-in-use {
  position: absolute;
  top: 6px;
  right: 6px;
  background: rgba(6, 78, 59, 0.85);
  color: #ffffff;
  font-size: 0.65rem;
  padding: 0.15rem 0.35rem;
  border-radius: 4px;
  font-weight: 600;
}

.media-meta {
  padding: 0.65rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.media-name {
  font-size: 0.825rem;
  font-weight: 600;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.media-specs {
  font-size: 0.7rem;
  color: #64748b;
}

.media-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid #f1f5f9;
}

.btn-xs {
  padding: 0.2rem 0.45rem;
  font-size: 0.7rem;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-outline { border: 1px solid #cbd5e1; background: #ffffff; color: #334155; }
.btn-outline:hover { background: #f8fafc; }
.btn-danger { border: 1px solid #fecaca; background: #fee2e2; color: #991b1b; }
.btn-danger:hover:not(:disabled) { background: #fca5a5; }
.btn-danger:disabled { opacity: 0.4; cursor: not-allowed; }

.btn {
  display: inline-flex;
  align-items: center;
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  border: none;
}

.btn-primary { background: #059669; color: #ffffff; }
.btn-primary:hover:not(:disabled) { background: #047857; }
.btn-secondary { background: #f1f5f9; color: #334155; border: 1px solid #cbd5e1; }
.btn-secondary:hover { background: #e2e8f0; }

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-card {
  background: #ffffff;
  border-radius: 12px;
  width: 100%;
  max-width: 580px;
  overflow: hidden;
}

.modal-header {
  padding: 1.1rem 1.25rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 { margin: 0; font-size: 1.15rem; }
.modal-close { background: none; border: none; font-size: 1.4rem; cursor: pointer; color: #94a3b8; }

.modal-form {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.upload-dropzone {
  border: 2px dashed #cbd5e1;
  border-radius: 8px;
  text-align: center;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  background: #f8fafc;
}

.file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  z-index: 2;
}

.dropzone-label {
  display: block;
  padding: 1.5rem;
  cursor: pointer;
}

.dropzone-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
}

.upload-icon { font-size: 2rem; }
.drop-hint { font-size: 0.75rem; color: #64748b; }

.preview-img {
  max-height: 180px;
  margin: 0 auto;
  border-radius: 6px;
  display: block;
}

.form-row { display: flex; gap: 0.75rem; }
.flex-1 { flex: 1; }

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-group label {
  font-size: 0.825rem;
  font-weight: 600;
  color: #334155;
}

.input-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid #f1f5f9;
}

.loading-state, .empty-state {
  padding: 4rem;
  text-align: center;
  color: #64748b;
}

.empty-icon { font-size: 2.5rem; display: block; margin-bottom: 0.5rem; }
</style>
