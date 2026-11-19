<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const isLoading = ref(true)
const isSaving = ref(false)

const form = reactive({
  siteName: 'PidieKab',
  taglineId: 'Eksplorasi Pesona Wisata, Kuliner & Budaya Kabupaten Pidie',
  taglineEn: 'Explore the Beauty, Culinary & Culture of Pidie Regency',
  logoUrl: '',
  logoMediaId: null as number | null,
  faviconUrl: '',
  faviconMediaId: null as number | null,
  contactEmail: 'kontak@pidiekab.id',
  contactPhone: '',
  address: 'Sigli, Kabupaten Pidie, Aceh, Indonesia',
  socialInstagram: '',
  socialFacebook: '',
  socialYoutube: '',
  socialTwitter: '',
  defaultSeoTitleId: 'PidieKab — Direktori Wisata, Kuliner & Budaya Kabupaten Pidie',
  defaultSeoTitleEn: 'PidieKab — Tourism, Culinary & Heritage of Pidie Regency',
  defaultSeoDescId: 'Portal informasi independen potensi wisata, warisan budaya, dan kekayaan kuliner khas Kabupaten Pidie.',
  defaultSeoDescEn: 'Independent digital portal exploring tourism, culinary delights, and rich cultural heritage of Pidie, Aceh.',
})

// Media Library Picker state
const showMediaPicker = ref(false)
const pickerTarget = ref<'logo' | 'favicon'>('logo')
const mediaList = ref<any[]>([])
const mediaSearch = ref('')
const isUploading = ref(false)
const uploadProgress = ref('')

const filteredMedia = computed(() => {
  if (!mediaSearch.value.trim()) return mediaList.value
  const q = mediaSearch.value.toLowerCase()
  return mediaList.value.filter(
    (m) =>
      (m.title && m.title.toLowerCase().includes(q)) ||
      (m.fileName && m.fileName.toLowerCase().includes(q))
  )
})

async function loadMediaLibrary() {
  try {
    const res = await api.get<any>('/api/v1/admin/media?per_page=100').catch(() => null)
    const items = res?.items || res?.data || []
    if (Array.isArray(items)) {
      mediaList.value = items.map((m: any) => ({
        id: m.id,
        storageKey: m.storage_key || m.storageKey,
        url: m.url || (m.storage_key ? `/media/${m.storage_key}` : (m.storageKey ? `/media/${m.storageKey}` : '')),
        title: m.title || m.original_filename || m.originalFilename || `Foto #${m.id}`,
        fileName: m.original_filename || m.originalFilename || `media-${m.id}`,
      }))
    }
  } catch (err) {
    console.error('Failed to load media items', err)
  }
}

function openMediaPicker(target: 'logo' | 'favicon') {
  pickerTarget.value = target
  showMediaPicker.value = true
  if (mediaList.value.length === 0) {
    loadMediaLibrary()
  }
}

function selectMedia(media: any) {
  if (pickerTarget.value === 'logo') {
    form.logoUrl = media.url
    form.logoMediaId = media.id
  } else {
    form.faviconUrl = media.url
    form.faviconMediaId = media.id
  }
  showMediaPicker.value = false
}

function clearLogo() {
  form.logoUrl = ''
  form.logoMediaId = null
}

function clearFavicon() {
  form.faviconUrl = ''
  form.faviconMediaId = null
}

// Client-side smart image compression prior to upload
async function compressImage(file: File, maxDim = 1920, quality = 0.82): Promise<Blob> {
  return new Promise((resolve) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        let width = img.width
        let height = img.height
        if (width > maxDim || height > maxDim) {
          if (width > height) {
            height = Math.round((height * maxDim) / width)
            width = maxDim
          } else {
            width = Math.round((width * maxDim) / height)
            height = maxDim
          }
        }
        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          return resolve(file)
        }
        ctx.drawImage(img, 0, 0, width, height)
        canvas.toBlob(
          (blob) => {
            if (blob && blob.size < file.size) {
              resolve(blob)
            } else {
              resolve(file)
            }
          },
          'image/webp',
          quality
        )
      }
      img.onerror = () => resolve(file)
      img.src = e.target?.result as string
    }
    reader.onerror = () => resolve(file)
    reader.readAsDataURL(file)
  })
}

async function handleModalFileUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  isUploading.value = true
  uploadProgress.value = 'Mengompresi file...'

  try {
    const compressedBlob = await compressImage(file)
    uploadProgress.value = 'Mengunggah ke server...'

    const ext = file.name.substring(file.name.lastIndexOf('.')) || '.jpg'
    const fileName = file.name.replace(/\.[^/.]+$/, '') + (compressedBlob.type === 'image/webp' ? '.webp' : ext)
    const uploadFile = new File([compressedBlob], fileName, { type: compressedBlob.type || file.type })

    const formData = new FormData()
    formData.append('file', uploadFile)
    formData.append('title', file.name.replace(/\.[^/.]+$/, ''))

    const uploaded = await api.upload<any>('/api/v1/admin/media/upload', formData)
    const mediaId = uploaded?.data?.id || uploaded?.id
    const storageKey = uploaded?.data?.storage_key || uploaded?.storage_key || uploaded?.storageKey
    const mediaUrl = uploaded?.url || uploaded?.data?.url || (storageKey ? `/media/${storageKey}` : '')

    const newMedia = {
      id: mediaId,
      storageKey,
      url: mediaUrl,
      title: file.name,
      fileName,
    }

    mediaList.value.unshift(newMedia)
    selectMedia(newMedia)
    notification.success('Upload Berhasil', 'Gambar baru berhasil diunggah dan dipilih.')
  } catch (err: any) {
    notification.error('Gagal Mengunggah', err.detail || err.message || 'Gagal mengunggah media.', err.requestId)
  } finally {
    isUploading.value = false
    uploadProgress.value = ''
    target.value = ''
  }
}

async function loadSettings() {
  isLoading.value = true
  try {
    const res = await api.get<any>('/api/v1/admin/settings')
    const s = res?.data || res?.settings || res
    if (s) {
      const transId = s.translations?.id || {}
      const transEn = s.translations?.en || {}

      form.siteName = transId.site_name || transEn.site_name || s.site_name || s.siteName || form.siteName
      form.taglineId = transId.tagline || s.tagline_id || s.taglineId || form.taglineId
      form.taglineEn = transEn.tagline || s.tagline_en || s.taglineEn || form.taglineEn

      form.logoUrl = s.logo_url || s.logoUrl || ''
      form.logoMediaId = s.logo_media_id || s.logoMediaId || null
      form.faviconUrl = s.favicon_url || s.faviconUrl || ''
      form.faviconMediaId = s.favicon_media_id || s.faviconMediaId || null

      form.contactEmail = s.contact_email || s.contactEmail || form.contactEmail
      form.contactPhone = s.contact_phone || s.contactPhone || ''
      form.address = s.contact_address || s.address || form.address

      let social: Record<string, string> = {}
      if (s.social_links) {
        if (typeof s.social_links === 'string') {
          try {
            social = JSON.parse(s.social_links)
          } catch {}
        } else if (typeof s.social_links === 'object') {
          social = s.social_links
        }
      }
      form.socialInstagram = social.instagram || s.social_instagram || s.socialInstagram || ''
      form.socialFacebook = social.facebook || s.social_facebook || s.socialFacebook || ''
      form.socialYoutube = social.youtube || s.social_youtube || s.socialYoutube || ''
      form.socialTwitter = social.twitter || s.social_twitter || s.socialTwitter || ''

      form.defaultSeoTitleId = transId.default_seo_title || s.default_seo_title_id || s.defaultSeoTitleId || form.defaultSeoTitleId
      form.defaultSeoDescId = transId.default_seo_description || s.default_seo_desc_id || s.defaultSeoDescId || form.defaultSeoDescId
      form.defaultSeoTitleEn = transEn.default_seo_title || s.default_seo_title_en || s.defaultSeoTitleEn || form.defaultSeoTitleEn
      form.defaultSeoDescEn = transEn.default_seo_description || s.default_seo_desc_en || s.defaultSeoDescEn || form.defaultSeoDescEn
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Pengaturan', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

async function saveSettings() {
  isSaving.value = true
  try {
    const payload = {
      site_name: form.siteName,
      tagline_id: form.taglineId,
      tagline_en: form.taglineEn,
      logo_url: form.logoUrl,
      logo_media_id: form.logoMediaId,
      favicon_url: form.faviconUrl,
      favicon_media_id: form.faviconMediaId,
      contact_email: form.contactEmail,
      contact_phone: form.contactPhone,
      contact_address: form.address,
      address: form.address,
      social_links: {
        instagram: form.socialInstagram,
        facebook: form.socialFacebook,
        youtube: form.socialYoutube,
        twitter: form.socialTwitter,
      },
      social_instagram: form.socialInstagram,
      social_facebook: form.socialFacebook,
      social_youtube: form.socialYoutube,
      social_twitter: form.socialTwitter,
      default_seo_title_id: form.defaultSeoTitleId,
      default_seo_title_en: form.defaultSeoTitleEn,
      default_seo_desc_id: form.defaultSeoDescId,
      default_seo_desc_en: form.defaultSeoDescEn,
    }

    const res = await api.put<any>('/api/v1/admin/settings', payload)
    if (res?.data || res?.settings) {
      const s = res.data || res.settings
      form.logoUrl = s.logo_url || form.logoUrl
      form.faviconUrl = s.favicon_url || form.faviconUrl
      form.logoMediaId = s.logo_media_id ?? form.logoMediaId
      form.faviconMediaId = s.favicon_media_id ?? form.faviconMediaId
    }

    notification.success('Pengaturan Disimpan', 'Konfigurasi website berhasil diperbarui dan dipublikasikan.')
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || err.message || 'Terjadi kesalahan saat menyimpan pengaturan.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadSettings()
  loadMediaLibrary()
})
</script>

<template>
  <AppLayout>
    <div class="settings-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Pengaturan Website</h2>
          <p class="page-desc">Konfigurasi identitas situs, logo, favicon, informasi kontak, media sosial, dan SEO default.</p>
        </div>
        <button class="btn btn-primary" :disabled="isSaving" @click="saveSettings">
          <BladeIcon name="check" size="18" />
          <span>{{ isSaving ? 'Menyimpan...' : 'Simpan Perubahan' }}</span>
        </button>
      </div>

      <div v-if="isLoading" class="loading-state">
        <BladeIcon name="arrow-path" size="24" class="spin-icon" />
        <p>Memuat pengaturan website...</p>
      </div>

      <div v-else class="settings-form">
        <!-- Site Identity Card -->
        <div class="card">
          <h3 class="card-title">Identitas Situs & Brand</h3>

          <div class="form-group">
            <label>Nama Situs / Portal *</label>
            <input v-model="form.siteName" type="text" class="input-control" placeholder="PidieKab" />
          </div>

          <!-- Logo & Favicon Picker Grid -->
          <div class="media-pickers-row">
            <!-- Logo Picker -->
            <div class="media-box flex-1">
              <label class="media-box-label">Logo Website</label>
              <div class="preview-area logo-preview-area">
                <img v-if="form.logoUrl" :src="form.logoUrl" alt="Logo Website" class="logo-preview-img" />
                <div v-else class="preview-placeholder">
                  <BladeIcon name="photo" size="32" class="placeholder-icon" />
                  <span>Belum ada logo dipilih</span>
                </div>
              </div>
              <div class="media-actions">
                <button type="button" class="btn btn-outline btn-sm" @click="openMediaPicker('logo')">
                  <BladeIcon name="photo" size="16" />
                  <span>{{ form.logoUrl ? 'Ganti Logo' : 'Pilih dari Media' }}</span>
                </button>
                <button v-if="form.logoUrl" type="button" class="btn btn-danger-outline btn-sm" @click="clearLogo" title="Hapus Logo">
                  <BladeIcon name="trash" size="15" />
                </button>
              </div>
              <div class="url-input-wrap">
                <input
                  v-model="form.logoUrl"
                  type="text"
                  class="input-control input-sm"
                  placeholder="Atau masukkan URL /media/..."
                />
              </div>
            </div>

            <!-- Favicon Picker -->
            <div class="media-box flex-1">
              <label class="media-box-label">Favicon Website</label>
              <div class="preview-area favicon-preview-area">
                <img v-if="form.faviconUrl" :src="form.faviconUrl" alt="Favicon" class="favicon-preview-img" />
                <div v-else class="preview-placeholder">
                  <BladeIcon name="globe-alt" size="28" class="placeholder-icon" />
                  <span>Belum ada favicon</span>
                </div>
              </div>
              <div class="media-actions">
                <button type="button" class="btn btn-outline btn-sm" @click="openMediaPicker('favicon')">
                  <BladeIcon name="photo" size="16" />
                  <span>{{ form.faviconUrl ? 'Ganti Favicon' : 'Pilih dari Media' }}</span>
                </button>
                <button v-if="form.faviconUrl" type="button" class="btn btn-danger-outline btn-sm" @click="clearFavicon" title="Hapus Favicon">
                  <BladeIcon name="trash" size="15" />
                </button>
              </div>
              <div class="url-input-wrap">
                <input
                  v-model="form.faviconUrl"
                  type="text"
                  class="input-control input-sm"
                  placeholder="Atau masukkan URL /media/..."
                />
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>Tagline (Bahasa Indonesia)</label>
            <input v-model="form.taglineId" type="text" class="input-control" />
          </div>

          <div class="form-group">
            <label>Tagline (English)</label>
            <input v-model="form.taglineEn" type="text" class="input-control" />
          </div>
        </div>

        <!-- Contact & Location Card -->
        <div class="card">
          <h3 class="card-title">Kontak & Lokasi Kantor</h3>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>Email Kontak Resmi *</label>
              <input v-model="form.contactEmail" type="email" class="input-control" />
            </div>
            <div class="form-group flex-1">
              <label>Nomor Telepon / WhatsApp</label>
              <input v-model="form.contactPhone" type="text" class="input-control" placeholder="+62 8..." />
            </div>
          </div>

          <div class="form-group">
            <label>Alamat Sekretariat</label>
            <textarea v-model="form.address" rows="2" class="textarea-control"></textarea>
          </div>
        </div>

        <!-- Social Media Links Card -->
        <div class="card">
          <h3 class="card-title">Akun Media Sosial Resmi</h3>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>Instagram URL</label>
              <input v-model="form.socialInstagram" type="text" class="input-control" placeholder="https://instagram.com/..." />
            </div>
            <div class="form-group flex-1">
              <label>Facebook URL</label>
              <input v-model="form.socialFacebook" type="text" class="input-control" placeholder="https://facebook.com/..." />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>YouTube URL</label>
              <input v-model="form.socialYoutube" type="text" class="input-control" placeholder="https://youtube.com/..." />
            </div>
            <div class="form-group flex-1">
              <label>X / Twitter URL</label>
              <input v-model="form.socialTwitter" type="text" class="input-control" placeholder="https://x.com/..." />
            </div>
          </div>
        </div>

        <!-- Default SEO Card -->
        <div class="card">
          <h3 class="card-title">Default SEO Metadata (Homepage & Fallback)</h3>
          <div class="form-group">
            <label>Default Title (ID)</label>
            <input v-model="form.defaultSeoTitleId" type="text" class="input-control" />
          </div>
          <div class="form-group">
            <label>Default Meta Description (ID)</label>
            <textarea v-model="form.defaultSeoDescId" rows="2" class="textarea-control"></textarea>
          </div>

          <div class="form-group">
            <label>Default Title (EN)</label>
            <input v-model="form.defaultSeoTitleEn" type="text" class="input-control" />
          </div>
          <div class="form-group">
            <label>Default Meta Description (EN)</label>
            <textarea v-model="form.defaultSeoDescEn" rows="2" class="textarea-control"></textarea>
          </div>
        </div>

        <div class="form-footer">
          <button class="btn btn-primary btn-lg" :disabled="isSaving" @click="saveSettings">
            <BladeIcon name="check" size="20" />
            <span>{{ isSaving ? 'Menyimpan Perubahan...' : 'Simpan Semua Perubahan' }}</span>
          </button>
        </div>
      </div>

      <!-- Media Picker Modal -->
      <div v-if="showMediaPicker" class="modal-backdrop" @click.self="showMediaPicker = false">
        <div class="modal-card">
          <div class="modal-header">
            <div class="modal-title-group">
              <BladeIcon name="photo" size="20" class="text-emerald" />
              <h3>Pilih {{ pickerTarget === 'logo' ? 'Logo' : 'Favicon' }} dari Media Library</h3>
            </div>
            <button class="modal-close" @click="showMediaPicker = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>

          <div class="modal-toolbar">
            <div class="modal-search">
              <BladeIcon name="magnifying-glass" size="16" class="search-icon" />
              <input
                v-model="mediaSearch"
                type="text"
                class="search-input"
                placeholder="Cari gambar berdasarkan nama..."
              />
            </div>
            <label class="btn btn-primary btn-sm upload-btn" :class="{ disabled: isUploading }">
              <BladeIcon name="cloud-arrow-up" size="16" />
              <span>{{ isUploading ? uploadProgress || 'Mengunggah...' : 'Unggah Foto Baru' }}</span>
              <input
                type="file"
                accept="image/*"
                class="sr-only"
                :disabled="isUploading"
                @change="handleModalFileUpload"
              />
            </label>
          </div>

          <div class="modal-body">
            <div v-if="mediaList.length === 0" class="empty-media">
              <BladeIcon name="photo" size="48" class="empty-icon" />
              <p>Belum ada gambar di Media Library.</p>
              <label class="btn btn-primary btn-sm upload-btn mt-2">
                <BladeIcon name="cloud-arrow-up" size="16" />
                <span>Unggah Foto Sekarang</span>
                <input type="file" accept="image/*" class="sr-only" @change="handleModalFileUpload" />
              </label>
            </div>
            <div v-else-if="filteredMedia.length === 0" class="empty-media">
              <p>Tidak ada gambar yang cocok dengan pencarian "{{ mediaSearch }}".</p>
            </div>
            <div v-else class="media-picker-grid">
              <div
                v-for="media in filteredMedia"
                :key="media.id"
                class="media-picker-item"
                @click="selectMedia(media)"
              >
                <div class="media-thumb-wrap">
                  <img :src="media.url" :alt="media.title" class="media-thumb" loading="lazy" />
                </div>
                <span class="media-title" :title="media.title || media.fileName">
                  {{ media.title || media.fileName }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.settings-page {
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

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 860px;
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.card-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 0.75rem;
}

.form-row {
  display: flex;
  gap: 1rem;
}

.flex-1 {
  flex: 1;
}

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

.input-control,
.textarea-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
  font-family: inherit;
  transition: border-color 0.15s;
}

.input-control:focus,
.textarea-control:focus {
  border-color: #059669;
  outline: none;
}

.input-sm {
  padding: 0.4rem 0.6rem;
  font-size: 0.8rem;
}

.media-pickers-row {
  display: flex;
  gap: 1.25rem;
  margin: 0.5rem 0;
}

@media (max-width: 640px) {
  .media-pickers-row {
    flex-direction: column;
  }
}

.media-box {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 1rem;
  background: #f8fafc;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.media-box-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #1e293b;
}

.preview-area {
  background: #ffffff;
  border: 1px dashed #cbd5e1;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80px;
  padding: 0.75rem;
}

.logo-preview-area {
  min-height: 90px;
}

.favicon-preview-area {
  min-height: 90px;
}

.logo-preview-img {
  max-height: 70px;
  max-width: 100%;
  object-fit: contain;
}

.favicon-preview-img {
  width: 48px;
  height: 48px;
  object-fit: contain;
}

.preview-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  color: #94a3b8;
  font-size: 0.8rem;
}

.placeholder-icon {
  color: #cbd5e1;
}

.media-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.url-input-wrap {
  width: 100%;
}

.url-input-wrap input {
  width: 100%;
  box-sizing: border-box;
}

.form-footer {
  display: flex;
  justify-content: flex-end;
  padding: 0.5rem 0 2rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.btn-sm {
  padding: 0.35rem 0.65rem;
  font-size: 0.8rem;
}

.btn-lg {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
}

.btn-primary {
  background: #059669;
  color: #ffffff;
}

.btn-primary:hover:not(:disabled) {
  background: #047857;
}

.btn-outline {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #334155;
}

.btn-outline:hover:not(:disabled) {
  border-color: #059669;
  color: #059669;
}

.btn-danger-outline {
  background: #ffffff;
  border-color: #fecaca;
  color: #dc2626;
}

.btn-danger-outline:hover:not(:disabled) {
  background: #fef2f2;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

/* Modal Styles */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1.5rem;
}

.modal-card {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  width: 100%;
  max-width: 680px;
  max-height: 82vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.text-emerald {
  color: #059669;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
}

.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  border-radius: 4px;
}

.modal-close:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.modal-toolbar {
  padding: 0.75rem 1.25rem;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.modal-search {
  position: relative;
  flex: 1;
}

.search-icon {
  position: absolute;
  left: 0.65rem;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
}

.search-input {
  width: 100%;
  padding: 0.45rem 0.65rem 0.45rem 2.2rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.85rem;
  box-sizing: border-box;
}

.search-input:focus {
  border-color: #059669;
  outline: none;
}

.upload-btn {
  white-space: nowrap;
}

.modal-body {
  padding: 1.25rem;
  overflow-y: auto;
  max-height: 52vh;
}

.empty-media {
  text-align: center;
  padding: 2.5rem 1rem;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.empty-icon {
  color: #cbd5e1;
}

.media-picker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 0.85rem;
}

.media-picker-item {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  background: #ffffff;
  transition: transform 0.15s, border-color 0.15s, box-shadow 0.15s;
  display: flex;
  flex-direction: column;
}

.media-picker-item:hover {
  transform: translateY(-2px);
  border-color: #059669;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
}

.media-thumb-wrap {
  width: 100%;
  height: 96px;
  background: #f1f5f9;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.media-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.media-title {
  padding: 0.45rem 0.5rem;
  font-size: 0.75rem;
  color: #334155;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
  background: #ffffff;
  border-top: 1px solid #f1f5f9;
}
</style>
