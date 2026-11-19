<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'
import RichTextEditor from '../../components/common/RichTextEditor.vue'

const notification = useNotificationStore()

const pages = ref<any[]>([])
const isLoading = ref(true)
const isSaving = ref(false)
const selectedPage = ref<any | null>(null)
const activeLocale = ref<'id' | 'en'>('id')

// Create new page modal
const showCreateModal = ref(false)
const newPageKey = ref('')
const isCreating = ref(false)

const form = reactive({
  pageKey: '',
  translations: {
    id: { title: '', slug: '', bodyHtml: '', status: 'draft' as 'draft' | 'published' | 'archived' },
    en: { title: '', slug: '', bodyHtml: '', status: 'draft' as 'draft' | 'published' | 'archived' },
  },
})

function extractBodyHtml(trans: any): string {
  if (!trans) return ''
  if (trans.body_html) return trans.body_html
  if (trans.bodyHtml) return trans.bodyHtml
  if (typeof trans.body === 'string') return trans.body
  if (typeof trans.body === 'object' && trans.body) {
    return trans.body.html || trans.body.markdown || ''
  }
  return ''
}

async function loadPages(selectKey?: string) {
  isLoading.value = true
  try {
    const res = await api.get<any>('/api/v1/admin/static-pages')
    const items = res?.data || res?.pages || []
    if (Array.isArray(items)) {
      pages.value = items
      if (items.length > 0) {
        if (selectKey) {
          const match = items.find((p: any) => (p.page_key || p.pageKey) === selectKey)
          selectPage(match || items[0])
        } else if (!selectedPage.value) {
          selectPage(items[0])
        } else {
          const currentId = selectedPage.value.id
          const updated = items.find((p: any) => p.id === currentId)
          if (updated) selectPage(updated)
        }
      }
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Halaman', err.detail || err.message || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function selectPage(page: any) {
  selectedPage.value = page
  form.pageKey = page.pageKey || page.page_key

  const transId = page.translations?.id || {}
  const transEn = page.translations?.en || {}

  form.translations.id = {
    title: transId.title || '',
    slug: transId.slug || '',
    bodyHtml: extractBodyHtml(transId),
    status: transId.status || 'draft',
  }
  form.translations.en = {
    title: transEn.title || '',
    slug: transEn.slug || '',
    bodyHtml: extractBodyHtml(transEn),
    status: transEn.status || 'draft',
  }
}

async function savePage() {
  if (!selectedPage.value) return

  isSaving.value = true
  try {
    const payload = {
      page_key: form.pageKey,
      translations: {
        id: {
          title: form.translations.id.title,
          slug: form.translations.id.slug,
          body_html: form.translations.id.bodyHtml,
          status: form.translations.id.status,
        },
        en: {
          title: form.translations.en.title,
          slug: form.translations.en.slug,
          body_html: form.translations.en.bodyHtml,
          status: form.translations.en.status,
        },
      },
    }

    await api.put(`/api/v1/admin/static-pages/${selectedPage.value.id}`, payload)
    notification.success('Halaman Disimpan', `Halaman "${form.pageKey}" berhasil disimpan dan dipublikasikan.`)
    await loadPages(form.pageKey)
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || err.message || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

async function createNewPage() {
  const key = newPageKey.value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '-')
  if (!key) return

  isCreating.value = true
  try {
    await api.post<any>('/api/v1/admin/static-pages', { page_key: key })
    notification.success('Halaman Dibuat', `Halaman statis "${key}" berhasil dibuat.`)
    showCreateModal.value = false
    newPageKey.value = ''
    await loadPages(key)
  } catch (err: any) {
    notification.error('Gagal Membuat Halaman', err.detail || err.message || 'Gagal membuat halaman baru.', err.requestId)
  } finally {
    isCreating.value = false
  }
}

function openPublicPage() {
  const slug = form.translations[activeLocale.value].slug
  if (slug) {
    window.open(`/${activeLocale.value}/${slug}/`, '_blank')
  }
}

onMounted(() => {
  loadPages()
})
</script>

<template>
  <AppLayout>
    <div class="static-pages-view">
      <div class="page-header">
        <div>
          <h2 class="page-title">Halaman Informasi Statis</h2>
          <p class="page-desc">Kelola halaman Tentang Kami, Kebijakan Privasi, dan Syarat Ketentuan daerah.</p>
        </div>
        <button class="btn btn-primary" @click="showCreateModal = true">
          <BladeIcon name="plus" size="18" />
          <span>Tambah Halaman Baru</span>
        </button>
      </div>

      <div v-if="isLoading" class="loading-state">
        <BladeIcon name="arrow-path" size="24" class="spin-icon" />
        <p>Memuat halaman informasi statis...</p>
      </div>

      <div v-else class="layout-grid">
        <!-- List of Pages (Left) -->
        <div class="card page-selector-card">
          <div class="selector-header">
            <h3 class="card-title">Daftar Halaman</h3>
            <span class="badge-count">{{ pages.length }}</span>
          </div>
          <ul class="page-list">
            <li
              v-for="p in pages"
              :key="p.id"
              class="page-item"
              :class="{ active: selectedPage?.id === p.id }"
              @click="selectPage(p)"
            >
              <div class="page-meta">
                <span class="page-key">
                  <BladeIcon name="document-text" size="16" />
                  <span>{{ p.pageKey || p.page_key }}</span>
                </span>
                <span class="page-label">
                  <span
                    class="status-dot"
                    :class="p.translations?.id?.status === 'published' ? 'dot-published' : 'dot-draft'"
                  ></span>
                  ID: {{ p.translations?.id?.status || 'draft' }} &bull; EN: {{ p.translations?.en?.status || 'draft' }}
                </span>
              </div>
            </li>
          </ul>
        </div>

        <!-- Editor (Right) -->
        <div v-if="selectedPage" class="card editor-card">
          <div class="editor-header">
            <div>
              <h3 class="card-title">
                Edit Halaman: <code class="key-tag">{{ form.pageKey }}</code>
              </h3>
              <p class="card-sub">Isi konten dan kelola publikasi independen untuk Bahasa Indonesia dan English.</p>
            </div>
            <div class="header-actions">
              <button
                v-if="form.translations[activeLocale].status === 'published' && form.translations[activeLocale].slug"
                type="button"
                class="btn btn-outline"
                @click="openPublicPage"
                title="Buka Halaman di Tab Baru"
              >
                <BladeIcon name="arrow-top-right-on-square" size="16" />
                <span>Lihat Publik</span>
              </button>
              <button class="btn btn-primary" :disabled="isSaving" @click="savePage">
                <BladeIcon name="bookmark" size="16" />
                <span>{{ isSaving ? 'Menyimpan...' : 'Simpan Halaman' }}</span>
              </button>
            </div>
          </div>

          <!-- Locale Tab -->
          <div class="lang-tabs">
            <button
              class="tab-btn"
              :class="{ active: activeLocale === 'id' }"
              @click="activeLocale = 'id'"
            >
              <BladeIcon name="language" size="16" />
              <span>Bahasa Indonesia</span>
              <span class="tab-status-pill" :class="'pill-' + form.translations.id.status">
                {{ form.translations.id.status }}
              </span>
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeLocale === 'en' }"
              @click="activeLocale = 'en'"
            >
              <BladeIcon name="globe-alt" size="16" />
              <span>English</span>
              <span class="tab-status-pill" :class="'pill-' + form.translations.en.status">
                {{ form.translations.en.status }}
              </span>
            </button>
          </div>

          <div class="form-body">
            <div class="form-row">
              <div class="form-group flex-1">
                <label>Judul Halaman ({{ activeLocale.toUpperCase() }}) *</label>
                <input
                  v-model="form.translations[activeLocale].title"
                  type="text"
                  class="input-control"
                  placeholder="Contoh: Tentang PidieKab"
                />
              </div>
              <div class="form-group flex-1">
                <label>Slug URL ({{ activeLocale.toUpperCase() }}) *</label>
                <input
                  v-model="form.translations[activeLocale].slug"
                  type="text"
                  class="input-control"
                  placeholder="contoh: tentang-kami"
                />
              </div>
              <div class="form-group status-group">
                <label>Status Publikasi</label>
                <select v-model="form.translations[activeLocale].status" class="select-control">
                  <option value="draft">Draft</option>
                  <option value="published">Published</option>
                  <option value="archived">Archived</option>
                </select>
              </div>
            </div>

            <!-- WYSIWYG Editor -->
            <div class="form-group">
              <label>Konten Halaman (WYSIWYG Rich Editor) *</label>
              <RichTextEditor
                v-model="form.translations[activeLocale].bodyHtml"
                placeholder="Tuliskan narasi dan informasi halaman lengkap di sini..."
                min-height="400px"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Create New Page Modal -->
      <div v-if="showCreateModal" class="modal-backdrop" @click.self="showCreateModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Tambah Halaman Informasi Statis Baru</h3>
            <button class="modal-close" @click="showCreateModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>Key Identifikasi Halaman *</label>
              <input
                v-model="newPageKey"
                type="text"
                class="input-control"
                placeholder="misal: pedoman-media, kontak, struktur-organisasi"
                @keyup.enter="createNewPage"
              />
              <span class="field-hint">Hanya huruf kecil, angka, dan tanda hubung (-). Contoh: tentang-kami</span>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline" @click="showCreateModal = false">Batal</button>
            <button class="btn btn-primary" :disabled="!newPageKey.trim() || isCreating" @click="createNewPage">
              <BladeIcon name="check" size="16" />
              <span>{{ isCreating ? 'Membuat...' : 'Buat Halaman' }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.static-pages-view {
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

.layout-grid {
  display: grid;
  grid-template-columns: 290px 1fr;
  gap: 1.5rem;
  align-items: start;
}

@media (max-width: 900px) {
  .layout-grid {
    grid-template-columns: 1fr;
  }
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
}

.selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #f1f5f9;
}

.badge-count {
  background: #f1f5f9;
  color: #475569;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
}

.card-title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
}

.card-sub {
  margin: 0.25rem 0 1rem 0;
  font-size: 0.8rem;
  color: #64748b;
}

.key-tag {
  background: #f1f5f9;
  color: #047857;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.9rem;
}

.page-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.page-item {
  padding: 0.75rem 0.85rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  background: #ffffff;
}

.page-item:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.page-item.active {
  border-color: #059669;
  background: #ecfdf5;
}

.page-meta {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.page-key {
  font-weight: 600;
  font-size: 0.9rem;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.page-label {
  font-size: 0.75rem;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  display: inline-block;
}

.dot-published {
  background: #10b981;
}

.dot-draft {
  background: #f59e0b;
}

.editor-card {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.lang-tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 0.5rem;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 6px;
  background: none;
  font-size: 0.85rem;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  background: #f1f5f9;
  color: #1e293b;
}

.tab-btn.active {
  background: #059669;
  color: #ffffff;
}

.tab-status-pill {
  font-size: 0.7rem;
  text-transform: uppercase;
  padding: 0.1rem 0.4rem;
  border-radius: 999px;
  font-weight: 700;
}

.tab-btn.active .tab-status-pill {
  background: rgba(255, 255, 255, 0.25);
  color: #ffffff;
}

.tab-btn:not(.active) .pill-published {
  background: #d1fae5;
  color: #065f46;
}

.tab-btn:not(.active) .pill-draft {
  background: #fef3c7;
  color: #92400e;
}

.tab-btn:not(.active) .pill-archived {
  background: #f1f5f9;
  color: #64748b;
}

.form-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.flex-1 {
  flex: 1;
  min-width: 200px;
}

.status-group {
  width: 150px;
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

.field-hint {
  font-size: 0.75rem;
  color: #64748b;
}

.input-control,
.select-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
  font-family: inherit;
  transition: border-color 0.15s;
}

.input-control:focus,
.select-control:focus {
  border-color: #059669;
  outline: none;
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

/* Modal */
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
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 480px;
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
  font-size: 1.05rem;
  font-weight: 700;
}

.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  color: #64748b;
  padding: 0.25rem;
  border-radius: 4px;
}

.modal-close:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.modal-body {
  padding: 1.25rem;
}

.modal-footer {
  padding: 0.85rem 1.25rem;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
