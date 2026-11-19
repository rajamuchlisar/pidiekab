<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'
import RichTextEditor from '../../components/common/RichTextEditor.vue'

const route = useRoute()
const router = useRouter()
const notification = useNotificationStore()

const contentId = computed(() => {
  const id = route.params.id
  return id && id !== 'new' ? Number(id) : null
})

const isNew = computed(() => contentId.value === null)
const activeTab = ref<'id' | 'en'>('id')
const isLoading = ref(false)
const isSaving = ref(false)
const hasUnsavedChanges = ref(false)

// Master data
const categories = ref<any[]>([])
const locations = ref<any[]>([])
const mediaList = ref<any[]>([])
const showMediaPicker = ref(false)

// Form state
const form = reactive({
  categoryId: null as number | null,
  locationKecamatanId: null as number | null,
  address: '',
  latitude: null as number | null,
  longitude: null as number | null,
  mainPhotoId: null as number | null,
  mainPhotoUrl: '',
  sources: [] as Array<{ id?: number; sourceUrl: string; sourceName: string; sourceType: string; notes?: string }>,
  translations: {
    id: {
      title: '',
      slug: '',
      summary: '',
      bodyHtml: '',
      seoTitle: '',
      seoDescription: '',
      status: 'draft' as 'draft' | 'published' | 'archived',
    },
    en: {
      title: '',
      slug: '',
      summary: '',
      bodyHtml: '',
      seoTitle: '',
      seoDescription: '',
      status: 'draft' as 'draft' | 'published' | 'archived',
    },
  },
})

// New source temporary input
const newSource = reactive({
  sourceUrl: '',
  sourceName: '',
  sourceType: 'official_gov',
  notes: '',
})

async function loadMasterData() {
  try {
    const catRes = await api.get<any>('/api/v1/admin/categories').catch(() => null)
    const rawCats = catRes?.categories || catRes?.data || []
    if (Array.isArray(rawCats)) {
      categories.value = rawCats.map((c: any) => ({
        id: Number(c.id),
        name: c.translations?.id?.name || c.translations?.en?.name || c.name || c.code || `Kategori #${c.id}`,
        code: c.code || '',
      }))
    }
  } catch (err) {
    console.error('Failed to load categories', err)
  }

  try {
    const locRes = await api.get<any>('/api/v1/admin/locations').catch(() => null)
    const locs = locRes?.locations || locRes?.districts || locRes?.data || []
    if (Array.isArray(locs)) {
      locations.value = locs.map((l: any) => ({
        id: Number(l.id),
        name: l.name,
        type: l.type || 'kecamatan',
        kemendagriCode: l.kemendagriCode || l.code || '',
      }))
    }
  } catch (err) {
    console.error('Failed to load locations', err)
  }

  try {
    const mediaRes = await api.get<any>('/api/v1/admin/media?per_page=50').catch(() => null)
    const media = mediaRes?.items || mediaRes?.data || []
    if (Array.isArray(media)) {
      mediaList.value = media
    }
  } catch (err) {
    console.error('Failed to load media', err)
  }
}

async function loadContent() {
  if (isNew.value || !contentId.value) return

  isLoading.value = true
  try {
    const res = await api.get<any>(`/api/v1/admin/contents/${contentId.value}`)
    const data = res?.data || res
    if (data) {
      form.categoryId = data.categoryId ? Number(data.categoryId) : data.category_id ? Number(data.category_id) : data.CategoryID ? Number(data.CategoryID) : null
      form.locationKecamatanId = data.locationKecamatanId ? Number(data.locationKecamatanId) : data.location_kecamatan_id ? Number(data.location_kecamatan_id) : data.village_id ? Number(data.village_id) : null
      form.address = data.address || data.Address || ''
      form.latitude = data.latitude || data.Latitude || null
      form.longitude = data.longitude || data.Longitude || null
      form.mainPhotoId = data.mainPhotoId ? Number(data.mainPhotoId) : data.main_photo_id ? Number(data.main_photo_id) : data.main_media_id ? Number(data.main_media_id) : null
      form.mainPhotoUrl = data.mainPhotoUrl || data.main_photo_url || data.MainPhotoURL || ''

      const sources = data.sources || data.Sources
      if (Array.isArray(sources)) {
        form.sources = sources.map((s: any) => ({
          id: s.id || s.ID,
          sourceName: s.sourceName || s.source_name || s.SourceName || '',
          sourceType: s.sourceType || s.source_type || s.SourceType || 'official_gov',
          sourceUrl: s.sourceUrl || s.source_url || s.SourceURL || '',
          notes: s.notes || s.Notes || '',
        }))
      }

      const translations = data.translations || data.Translations
      if (translations) {
        ;(['id', 'en'] as const).forEach((loc) => {
          const t = translations[loc]
          if (t) {
            let bodyText = ''
            if (typeof t.body === 'string') {
              bodyText = t.body
            } else if (t.body && typeof t.body === 'object') {
              bodyText = t.body.html || t.body.markdown || t.body.raw || ''
            } else if (typeof t.bodyHtml === 'string') {
              bodyText = t.bodyHtml
            } else if (typeof t.body_html === 'string') {
              bodyText = t.body_html
            }

            form.translations[loc] = {
              title: t.title || t.Title || '',
              slug: t.slug || t.Slug || '',
              summary: t.summary || t.Summary || (typeof t.summary === 'string' ? t.summary : ''),
              bodyHtml: bodyText,
              seoTitle: t.seoTitle || t.seo_title || t.SEOTitle || '',
              seoDescription: t.seoDescription || t.seo_description || t.SEODescription || '',
              status: (t.status || t.Status || 'draft') as 'draft' | 'published' | 'archived',
            }
          }
        })
      }
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Konten', err.detail || 'Tidak dapat memuat detail artikel.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function generateSlug(locale: 'id' | 'en') {
  const title = form.translations[locale].title
  if (!title) return
  const slug = title
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
  form.translations[locale].slug = slug
  markUnsaved()
}

function markUnsaved() {
  hasUnsavedChanges.value = true
}

function addSource() {
  if (!newSource.sourceUrl || !newSource.sourceName) {
    alert('URL dan Nama Sumber wajib diisi.')
    return
  }
  form.sources.push({ ...newSource })
  newSource.sourceUrl = ''
  newSource.sourceName = ''
  newSource.sourceType = 'official_gov'
  newSource.notes = ''
  markUnsaved()
}

function removeSource(index: number) {
  form.sources.splice(index, 1)
  markUnsaved()
}

function selectMedia(media: any) {
  form.mainPhotoId = media.id
  form.mainPhotoUrl = media.url
  showMediaPicker.value = false
  markUnsaved()
}

async function saveDraft() {
  if (!form.categoryId) {
    alert('Silakan pilih Kategori terlebih dahulu.')
    return
  }

  isSaving.value = true
  try {
    const payload = {
      category_id: form.categoryId,
      village_id: form.locationKecamatanId,
      address: form.address || null,
      latitude: form.latitude ? Number(form.latitude) : null,
      longitude: form.longitude ? Number(form.longitude) : null,
      main_media_id: form.mainPhotoId,
      translations: {
        id: {
          title: form.translations.id.title,
          slug: form.translations.id.slug,
          summary: form.translations.id.summary || null,
          body_html: form.translations.id.bodyHtml,
          body: {
            html: form.translations.id.bodyHtml,
            markdown: form.translations.id.bodyHtml,
          },
          seo_title: form.translations.id.seoTitle || null,
          seo_description: form.translations.id.seoDescription || null,
        },
        en: {
          title: form.translations.en.title,
          slug: form.translations.en.slug,
          summary: form.translations.en.summary || null,
          body_html: form.translations.en.bodyHtml,
          body: {
            html: form.translations.en.bodyHtml,
            markdown: form.translations.en.bodyHtml,
          },
          seo_title: form.translations.en.seoTitle || null,
          seo_description: form.translations.en.seoDescription || null,
        },
      },
    }

    if (isNew.value) {
      const res = await api.post<any>('/api/v1/admin/contents', payload)
      const newId = res?.id || res?.data?.id
      notification.success('Draft Dibuat', 'Artikel baru berhasil disimpan sebagai draft.')
      hasUnsavedChanges.value = false
      router.push(`/contents/${newId}/edit`)
    } else {
      await api.put(`/api/v1/admin/contents/${contentId.value}`, payload)
      for (const loc of ['id', 'en'] as const) {
        const t = form.translations[loc]
        if (t.title) {
          await api.put(`/api/v1/admin/contents/${contentId.value}/translations/${loc}`, {
            title: t.title,
            slug: t.slug || t.title.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
            summary: t.summary || null,
            body: { html: t.bodyHtml, markdown: t.bodyHtml },
            seo_title: t.seoTitle || null,
            seo_description: t.seoDescription || null,
          }).catch(() => null)
        }
      }
      notification.success('Perubahan Disimpan', 'Draft artikel berhasil diperbarui.')
      hasUnsavedChanges.value = false
    }
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan saat menyimpan draft.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

async function publishLocale(locale: 'id' | 'en') {
  if (isNew.value || !contentId.value) {
    alert('Harap simpan draft artikel terlebih dahulu sebelum mempublikasikan.')
    return
  }

  const localeTitle = locale === 'id' ? 'Bahasa Indonesia' : 'English'
  const confirmed = confirm(
    `Apakah Anda yakin ingin mempublikasikan versi [${localeTitle}] ke situs publik?\n\nJob build SSG otomatis akan dijadwalkan.`
  )
  if (!confirmed) return

  isSaving.value = true
  try {
    await saveDraft()
    await api.post(`/api/v1/admin/contents/${contentId.value}/publish`, { locale }).catch(() => {
      return api.post(`/api/v1/admin/contents/${contentId.value}/translations/${locale}/publish`)
    })
    form.translations[locale].status = 'published'
    notification.success('Konten Dipublikasikan', `Versi ${localeTitle} berhasil dirilis ke publik.`)
  } catch (err: any) {
    notification.error('Gagal Publikasi', err.detail || 'Artikel tidak dapat dipublikasikan.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

async function archiveLocale(locale: 'id' | 'en') {
  if (isNew.value || !contentId.value) return

  const localeTitle = locale === 'id' ? 'Bahasa Indonesia' : 'English'
  const confirmed = confirm(
    `Apakah Anda yakin ingin mengarsipkan versi [${localeTitle}]? Artikel akan ditarik dari rilis publik.`
  )
  if (!confirmed) return

  isSaving.value = true
  try {
    await api.post(`/api/v1/admin/contents/${contentId.value}/archive`, { locale }).catch(() => {
      return api.post(`/api/v1/admin/contents/${contentId.value}/translations/${locale}/archive`)
    })
    form.translations[locale].status = 'archived'
    notification.success('Konten Diarsipkan', `Versi ${localeTitle} berhasil ditarik dari publik.`)
  } catch (err: any) {
    notification.error('Gagal Arsip', err.detail || 'Gagal mengubah status artikel.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

function openPrivatePreview(locale: 'id' | 'en') {
  if (isNew.value || !contentId.value) {
    alert('Simpan draft terlebih dahulu untuk melihat preview privat.')
    return
  }
  const previewUrl = `/api/v1/admin/preview/contents/${contentId.value}?locale=${locale}`
  window.open(previewUrl, '_blank')
}

onMounted(async () => {
  await loadMasterData()
  await loadContent()
})
</script>

<template>
  <AppLayout>
    <div class="content-editor-page">
      <!-- Top Action Bar -->
      <div class="editor-header">
        <div class="header-left">
          <router-link to="/contents" class="btn-back">
            <BladeIcon name="arrow-left" size="16" />
            <span>Kembali ke Daftar</span>
          </router-link>
          <div class="title-wrap">
            <h2 class="page-title">{{ isNew ? 'Buat Konten Baru' : 'Edit Konten' }}</h2>
            <span v-if="hasUnsavedChanges" class="badge-unsaved">Ada Perubahan Belum Disimpan</span>
          </div>
        </div>

        <div class="header-right">
          <button type="button" class="btn btn-secondary" @click="openPrivatePreview(activeTab)">
            <BladeIcon name="eye" size="16" />
            <span>Preview Privat ({{ activeTab.toUpperCase() }})</span>
          </button>
          <button type="button" class="btn btn-secondary" :disabled="isSaving" @click="saveDraft">
            <BladeIcon name="bookmark" size="16" />
            <span>Simpan Draft</span>
          </button>
          <button
            v-if="form.translations[activeTab].status !== 'published'"
            type="button"
            class="btn btn-primary"
            :disabled="isSaving"
            @click="publishLocale(activeTab)"
          >
            <BladeIcon name="rocket-launch" size="16" />
            <span>Publikasikan ({{ activeTab.toUpperCase() }})</span>
          </button>
          <button
            v-else
            type="button"
            class="btn btn-archive"
            :disabled="isSaving"
            @click="archiveLocale(activeTab)"
          >
            <BladeIcon name="archive-box" size="16" />
            <span>Tarik / Arsipkan ({{ activeTab.toUpperCase() }})</span>
          </button>
        </div>
      </div>

      <div v-if="isLoading" class="loading-panel">
        Memuat data artikel...
      </div>

      <div v-else class="editor-grid">
        <!-- Main Form (Left / Center) -->
        <div class="main-column">
          <!-- Language Tabs -->
          <div class="lang-tabs">
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'id' }"
              @click="activeTab = 'id'"
            >
              <BladeIcon name="language" size="16" />
              <span>Bahasa Indonesia</span>
              <span
                class="tab-status"
                :class="form.translations.id.status === 'published' ? 'status-pub' : 'status-draft'"
              >
                {{ form.translations.id.status }}
              </span>
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'en' }"
              @click="activeTab = 'en'"
            >
              <BladeIcon name="globe-alt" size="16" />
              <span>English</span>
              <span
                class="tab-status"
                :class="form.translations.en.status === 'published' ? 'status-pub' : 'status-draft'"
              >
                {{ form.translations.en.status }}
              </span>
            </button>
          </div>

          <!-- Active Language Fields -->
          <div class="card form-card">
            <div class="form-group">
              <label>Judul Artikel ({{ activeTab.toUpperCase() }}) *</label>
              <input
                v-model="form.translations[activeTab].title"
                type="text"
                class="input-control"
                placeholder="Contoh: Pesona Wisata Pantai Pelangi Sigli"
                @input="markUnsaved"
                @blur="!form.translations[activeTab].slug && generateSlug(activeTab)"
              />
            </div>

            <div class="form-group">
              <div class="label-with-action">
                <label>Slug URL ({{ activeTab.toUpperCase() }}) *</label>
                <button type="button" class="btn-text" @click="generateSlug(activeTab)">
                  <BladeIcon name="arrow-path" size="14" />
                  <span>Regenerasi Slug</span>
                </button>
              </div>
              <input
                v-model="form.translations[activeTab].slug"
                type="text"
                class="input-control"
                placeholder="pantai-pelangi-sigli"
                @input="markUnsaved"
              />
              <span class="field-hint">
                URL publik: /{{ activeTab }}/[kategori]/{{ form.translations[activeTab].slug || 'slug' }}/
              </span>
            </div>

            <div class="form-group">
              <label>Ringkasan Singkat (Summary) *</label>
              <textarea
                v-model="form.translations[activeTab].summary"
                rows="2"
                class="textarea-control"
                placeholder="Deskripsi singkat konten untuk kartu discovery dan preview..."
                @input="markUnsaved"
              ></textarea>
            </div>

            <div class="form-group">
              <label>Isi Konten Artikel (WYSIWYG Editor) *</label>
              <RichTextEditor
                v-model="form.translations[activeTab].bodyHtml"
                placeholder="Tuliskan narasi lengkap artikel di sini (format visual, judul sub-bab, poin butir, tautan, dan gambar)..."
                @change="markUnsaved"
              />
            </div>

            <div class="seo-section">
              <h4 class="seo-title">Metadata SEO & Search Engine</h4>
              <div class="form-group">
                <label>SEO Title</label>
                <input
                  v-model="form.translations[activeTab].seoTitle"
                  type="text"
                  class="input-control"
                  placeholder="Judul optimal di Google Search (maks 60 karakter)"
                  @input="markUnsaved"
                />
              </div>
              <div class="form-group">
                <label>SEO Meta Description</label>
                <textarea
                  v-model="form.translations[activeTab].seoDescription"
                  rows="2"
                  class="textarea-control"
                  placeholder="Ringkasan hasil pencarian search engine (maks 160 karakter)"
                  @input="markUnsaved"
                ></textarea>
              </div>
            </div>
          </div>

          <!-- Verified Sources / Citations Requirement -->
          <div class="card sources-card">
            <div class="sources-header">
              <div>
                <h3 class="card-title">Sitasi Sumber & Referensi Faktual *</h3>
                <p class="card-sub">
                  Kebijakan PidieKab: Wajib menyertakan minimal 1 sumber valid sebelum dapat dipublikasikan.
                </p>
              </div>
            </div>

            <div v-if="form.sources.length === 0" class="empty-sources">
              <BladeIcon name="exclamation-triangle" size="18" class="icon-warning" />
              <span>Belum ada sitasi sumber. Artikel tidak memenuhi syarat publikasi sampai sitasi ditambahkan.</span>
            </div>

            <ul v-else class="sources-list">
              <li v-for="(src, idx) in form.sources" :key="idx" class="source-item">
                <div class="source-info">
                  <span class="source-name">{{ src.sourceName }}</span>
                  <a :href="src.sourceUrl" target="_blank" class="source-url">{{ src.sourceUrl }}</a>
                  <span class="source-type">Tipe: {{ src.sourceType }}</span>
                </div>
                <button type="button" class="btn-remove" @click="removeSource(idx)">
                  <BladeIcon name="trash" size="14" />
                  <span>Hapus</span>
                </button>
              </li>
            </ul>

            <div class="add-source-form">
              <h4 class="form-subtitle">
                <BladeIcon name="plus" size="14" />
                <span>Tambah Sitasi Baru</span>
              </h4>
              <div class="source-inputs">
                <input
                  v-model="newSource.sourceName"
                  type="text"
                  placeholder="Nama Penerbit / Lembaga (misal: Disporapar Pidie)"
                  class="input-control"
                />
                <input
                  v-model="newSource.sourceUrl"
                  type="url"
                  placeholder="https://..."
                  class="input-control"
                />
                <select v-model="newSource.sourceType" class="select-control">
                  <option value="official_gov">Pemerintah Resmi</option>
                  <option value="news_media">Media Berita Kredibel</option>
                  <option value="academic">Kajian Akademik / Jurnal</option>
                  <option value="field_survey">Survei Lapangan Terverifikasi</option>
                </select>
                <button type="button" class="btn btn-secondary" @click="addSource">
                  <BladeIcon name="plus" size="16" />
                  <span>Tambah Sitasi</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Sidebar / Shared Metadata (Right) -->
        <div class="side-column">
          <!-- Category Card -->
          <div class="card meta-card">
            <h3 class="card-title">Kategori Utama *</h3>
            <select v-model="form.categoryId" class="select-control full-width" @change="markUnsaved">
              <option :value="null">-- Pilih Kategori --</option>
              <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                {{ cat.name || cat.names?.id || cat.code }}
              </option>
            </select>
          </div>

          <!-- Main Photo Picker -->
          <div class="card meta-card">
            <h3 class="card-title">Foto Utama (Main Photo)</h3>
            <div class="photo-preview-wrap">
              <img
                v-if="form.mainPhotoUrl"
                :src="form.mainPhotoUrl"
                alt="Foto Utama"
                class="photo-preview"
              />
              <div v-else class="photo-placeholder">
                Belum ada foto utama yang dipilih
              </div>
            </div>
            <button type="button" class="btn btn-secondary full-width mt-2" @click="showMediaPicker = true">
              <BladeIcon name="photo" size="18" />
              <span>Pilih dari Media Library</span>
            </button>
          </div>

          <!-- Location & Geographic Coordinates -->
          <div class="card meta-card">
            <h3 class="card-title">Lokasi & Geografis</h3>
            <div class="form-group">
              <label>Kecamatan (Kab. Pidie)</label>
              <select v-model="form.locationKecamatanId" class="select-control full-width" @change="markUnsaved">
                <option :value="null">-- Pilih Kecamatan --</option>
                <option v-for="loc in locations" :key="loc.id" :value="loc.id">
                  {{ loc.name }}
                </option>
              </select>
            </div>

            <div class="form-group">
              <label>Alamat / Gampong</label>
              <textarea
                v-model="form.address"
                rows="2"
                class="textarea-control"
                placeholder="Nama gampong, jalan, atau patokan lokasi..."
                @input="markUnsaved"
              ></textarea>
            </div>

            <div class="coords-row">
              <div class="form-group">
                <label>Latitude</label>
                <input
                  v-model="form.latitude"
                  type="number"
                  step="0.000001"
                  class="input-control"
                  placeholder="5.38..."
                  @input="markUnsaved"
                />
              </div>
              <div class="form-group">
                <label>Longitude</label>
                <input
                  v-model="form.longitude"
                  type="number"
                  step="0.000001"
                  class="input-control"
                  placeholder="95.95..."
                  @input="markUnsaved"
                />
              </div>
            </div>
            <span class="field-hint">Koordinat wilayah Kabupaten Pidie berkisar 4.8°–5.5° N, 95.7°–96.2° E.</span>
          </div>
        </div>
      </div>

      <!-- Media Picker Modal -->
      <div v-if="showMediaPicker" class="modal-backdrop" @click.self="showMediaPicker = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Pilih Foto dari Media Library</h3>
            <button class="modal-close" @click="showMediaPicker = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div v-if="mediaList.length === 0" class="empty-media">
              Belum ada foto di Media Library. Silakan unggah foto di menu Media Library terlebih dahulu.
            </div>
            <div v-else class="media-picker-grid">
              <div
                v-for="media in mediaList"
                :key="media.id"
                class="media-picker-item"
                @click="selectMedia(media)"
              >
                <img :src="media.url" :alt="media.title" class="media-thumb" />
                <span class="media-title">{{ media.title || media.fileName }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.content-editor-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.btn-back {
  font-size: 0.85rem;
  color: #059669;
  text-decoration: none;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.page-title {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 700;
  color: #0f172a;
}

.badge-unsaved {
  background: #fef3c7;
  color: #92400e;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  border: 1px solid #fde68a;
}

.header-right {
  display: flex;
  gap: 0.65rem;
  align-items: center;
}

.editor-grid {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 1.5rem;
  align-items: start;
}

@media (max-width: 960px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }
}

.main-column, .side-column {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.lang-tabs {
  display: flex;
  gap: 0.5rem;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1.25rem;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  border-radius: 8px 8px 0 0;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
  color: #475569;
  transition: all 0.15s;
}

.tab-btn.active {
  background: #059669;
  color: #ffffff;
  border-color: #059669;
}

.tab-status {
  font-size: 0.7rem;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  text-transform: capitalize;
}

.status-pub {
  background: #d1fae5;
  color: #065f46;
}

.status-draft {
  background: #f1f5f9;
  color: #64748b;
}

.tab-btn.active .status-pub {
  background: #ffffff;
  color: #065f46;
}

.tab-btn.active .status-draft {
  background: rgba(255, 255, 255, 0.2);
  color: #ffffff;
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-bottom: 1.25rem;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}

.label-with-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.btn-text {
  background: none;
  border: none;
  color: #059669;
  font-size: 0.775rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.input-control, .textarea-control, .select-control {
  padding: 0.6rem 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
  color: #1e293b;
  font-family: inherit;
}

.textarea-control {
  resize: vertical;
}

.font-mono {
  font-family: monospace;
}

.full-width {
  width: 100%;
}

.mt-2 {
  margin-top: 0.75rem;
}

.field-hint {
  font-size: 0.75rem;
  color: #64748b;
}

.seo-section {
  border-top: 1px solid #e2e8f0;
  padding-top: 1.25rem;
  margin-top: 0.5rem;
}

.seo-title {
  margin: 0 0 1rem 0;
  font-size: 0.95rem;
  font-weight: 700;
  color: #0f172a;
}

/* Sources */
.card-title {
  margin: 0 0 0.25rem 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
}

.card-sub {
  margin: 0 0 1rem 0;
  font-size: 0.8rem;
  color: #64748b;
}

.empty-sources {
  background: #fef2f2;
  border: 1px dashed #f87171;
  color: #991b1b;
  padding: 0.85rem;
  border-radius: 6px;
  font-size: 0.85rem;
  margin-bottom: 1.25rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.icon-warning {
  color: #dc2626;
  flex-shrink: 0;
}

.sources-list {
  list-style: none;
  padding: 0;
  margin: 0 0 1.25rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.source-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 0.65rem 0.85rem;
  border-radius: 6px;
}

.source-info {
  display: flex;
  flex-direction: column;
}

.source-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: #0f172a;
}

.source-url {
  font-size: 0.75rem;
  color: #059669;
  text-decoration: none;
}

.source-type {
  font-size: 0.7rem;
  color: #64748b;
}

.btn-remove {
  background: none;
  border: none;
  color: #ef4444;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.add-source-form {
  border-top: 1px solid #e2e8f0;
  padding-top: 1rem;
}

.form-subtitle {
  margin: 0 0 0.75rem 0;
  font-size: 0.85rem;
  color: #334155;
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.source-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
}

.source-inputs button {
  grid-column: span 2;
}

/* Sidebar Meta */
.photo-preview-wrap {
  width: 100%;
  height: 160px;
  background: #f1f5f9;
  border-radius: 6px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
}

.photo-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.photo-placeholder {
  font-size: 0.8rem;
  color: #94a3b8;
  text-align: center;
  padding: 1rem;
}

.coords-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.55rem 0.95rem;
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

.btn-secondary {
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.btn-secondary:hover:not(:disabled) {
  background: #e2e8f0;
}

.btn-archive {
  background: #fef3c7;
  color: #92400e;
  border: 1px solid #fde68a;
}

.btn-archive:hover:not(:disabled) {
  background: #fde68a;
}

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
  max-width: 640px;
  max-height: 80vh;
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

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
}

.modal-body {
  padding: 1.25rem;
  overflow-y: auto;
}

.media-picker-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 0.85rem;
}

.media-picker-item {
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.15s, border-color 0.15s;
}

.media-picker-item:hover {
  transform: translateY(-2px);
  border-color: #059669;
}

.media-thumb {
  width: 100%;
  height: 90px;
  object-fit: cover;
  display: block;
}

.media-title {
  display: block;
  font-size: 0.75rem;
  padding: 0.35rem 0.5rem;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
