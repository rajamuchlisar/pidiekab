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

// Rename page key modal
const showRenameModal = ref(false)
const renamePageKey = ref('')
const renameTargetPage = ref<any | null>(null)
const isRenaming = ref(false)

// Delete confirmation modal
const showDeleteModal = ref(false)
const deleteTargetPage = ref<any | null>(null)
const isDeleting = ref(false)

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

function cleanSlug(str: string): string {
  return (str || '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

function humanize(key: string): string {
  return (key || '')
    .split('-')
    .filter(Boolean)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

function selectPage(page: any) {
  selectedPage.value = page
  const key = page.pageKey || page.page_key || ''
  form.pageKey = key

  const transId = page.translations?.id || {}
  const transEn = page.translations?.en || {}

  const idTitle = transId.title || (key ? humanize(key) : '')
  const idSlug = cleanSlug(transId.slug || key)

  let defaultEnTitle = transEn.title || ''
  if (!defaultEnTitle && key) {
    if (key === 'hubungi') defaultEnTitle = 'Contact Us'
    else if (key === 'tentang-kami') defaultEnTitle = 'About Us'
    else if (key === 'syarat-ketentuan') defaultEnTitle = 'Terms & Conditions'
    else if (key === 'kebijakan-privasi') defaultEnTitle = 'Privacy Policy'
    else defaultEnTitle = humanize(key)
  }

  let defaultEnSlug = cleanSlug(transEn.slug || '')
  if (!defaultEnSlug && key) {
    if (key === 'hubungi') defaultEnSlug = 'contact-us'
    else if (key === 'tentang-kami') defaultEnSlug = 'about-us'
    else if (key === 'syarat-ketentuan') defaultEnSlug = 'terms-and-conditions'
    else if (key === 'kebijakan-privasi') defaultEnSlug = 'privacy-policy'
    else defaultEnSlug = cleanSlug(key)
  }

  form.translations.id = {
    title: transId.title || idTitle,
    slug: transId.slug || idSlug,
    bodyHtml: extractBodyHtml(transId),
    status: transId.status || 'draft',
  }
  form.translations.en = {
    title: transEn.title || defaultEnTitle,
    slug: transEn.slug || defaultEnSlug,
    bodyHtml: extractBodyHtml(transEn),
    status: transEn.status || 'draft',
  }
}

function onTitleInput(locale: 'id' | 'en') {
  const t = form.translations[locale]
  if (!t.slug || t.slug.trim() === '') {
    t.slug = cleanSlug(t.title)
  }
}

function onSlugBlur(locale: 'id' | 'en') {
  const t = form.translations[locale]
  t.slug = cleanSlug(t.slug || t.title || form.pageKey)
}

async function savePage() {
  if (!selectedPage.value) return

  // Validate active tab
  const curLocale = activeLocale.value
  const curTrans = form.translations[curLocale]
  if (!curTrans.title.trim()) {
    notification.error('Validasi Gagal', `Judul Halaman (${curLocale.toUpperCase()}) wajib diisi.`)
    return
  }

  // Ensure slugs are clean
  form.translations.id.slug = cleanSlug(form.translations.id.slug || form.translations.id.title || form.pageKey)
  form.translations.en.slug = cleanSlug(form.translations.en.slug || form.translations.en.title || form.pageKey)

  // Ensure title is present if body is provided
  if (!form.translations.en.title.trim()) {
    form.translations.en.title = humanize(form.pageKey) || 'Page'
  }
  if (!form.translations.id.title.trim()) {
    form.translations.id.title = humanize(form.pageKey) || 'Halaman'
  }

  isSaving.value = true
  try {
    const payload = {
      page_key: form.pageKey,
      translations: {
        id: {
          title: form.translations.id.title.trim(),
          slug: form.translations.id.slug.trim(),
          body_html: form.translations.id.bodyHtml,
          status: form.translations.id.status,
        },
        en: {
          title: form.translations.en.title.trim(),
          slug: form.translations.en.slug.trim(),
          body_html: form.translations.en.bodyHtml,
          status: form.translations.en.status,
        },
      },
    }

    await api.put(`/api/v1/admin/static-pages/${selectedPage.value.id}`, payload)
    notification.success('Halaman Disimpan', `Halaman "${form.pageKey}" berhasil disimpan.`)
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

function openRenameModal(page?: any) {
  const target = page || selectedPage.value
  if (!target) return
  renameTargetPage.value = target
  renamePageKey.value = target.pageKey || target.page_key || ''
  showRenameModal.value = true
}

async function submitRename() {
  if (!renameTargetPage.value) return
  const cleanKey = renamePageKey.value.trim().toLowerCase().replace(/[^a-z0-9-]/g, '-')
  if (!cleanKey) {
    notification.error('Validasi Gagal', 'Nama kunci halaman tidak boleh kosong.')
    return
  }

  isRenaming.value = true
  try {
    await api.put(`/api/v1/admin/static-pages/${renameTargetPage.value.id}/key`, { page_key: cleanKey })
    notification.success('Nama Halaman Diperbarui', `Nama kunci halaman berhasil diubah menjadi "${cleanKey}".`)
    showRenameModal.value = false
    await loadPages(cleanKey)
  } catch (err: any) {
    notification.error('Gagal Mengubah Nama', err.detail || err.message || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isRenaming.value = false
  }
}

function openDeleteModal(page?: any) {
  const target = page || selectedPage.value
  if (!target) return
  deleteTargetPage.value = target
  showDeleteModal.value = true
}

async function submitDelete() {
  if (!deleteTargetPage.value) return
  const targetId = deleteTargetPage.value.id
  const targetKey = deleteTargetPage.value.pageKey || deleteTargetPage.value.page_key

  isDeleting.value = true
  try {
    await api.delete(`/api/v1/admin/static-pages/${targetId}`)
    notification.success('Halaman Dihapus', `Halaman "${targetKey}" berhasil dihapus.`)
    showDeleteModal.value = false
    if (selectedPage.value?.id === targetId) {
      selectedPage.value = null
    }
    deleteTargetPage.value = null
    await loadPages()
  } catch (err: any) {
    notification.error('Gagal Menghapus Halaman', err.detail || err.message || 'Terjadi kesalahan saat menghapus.', err.requestId)
  } finally {
    isDeleting.value = false
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
              <div class="page-item-content">
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
                <div class="page-item-actions" @click.stop>
                  <button
                    type="button"
                    class="action-icon-btn edit-key-btn"
                    title="Ubah Nama / Key Halaman"
                    @click="openRenameModal(p)"
                  >
                    <BladeIcon name="pencil-square" size="15" />
                  </button>
                  <button
                    type="button"
                    class="action-icon-btn delete-btn"
                    title="Hapus Halaman"
                    @click="openDeleteModal(p)"
                  >
                    <BladeIcon name="trash" size="15" />
                  </button>
                </div>
              </div>
            </li>
          </ul>
        </div>

        <!-- Editor (Right) -->
        <div v-if="selectedPage" class="card editor-card">
          <div class="editor-header">
            <div>
              <div class="header-title-row">
                <h3 class="card-title">
                  Edit Halaman: <code class="key-tag">{{ form.pageKey }}</code>
                </h3>
                <button
                  type="button"
                  class="btn-inline-action"
                  title="Ubah Nama / Key Halaman"
                  @click="openRenameModal(selectedPage)"
                >
                  <BladeIcon name="pencil" size="14" />
                  <span>Ubah Nama</span>
                </button>
              </div>
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
              <button
                type="button"
                class="btn btn-danger-outline"
                title="Hapus Halaman Ini"
                @click="openDeleteModal(selectedPage)"
              >
                <BladeIcon name="trash" size="16" />
                <span>Hapus Halaman</span>
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
                  @input="onTitleInput(activeLocale)"
                />
              </div>
              <div class="form-group flex-1">
                <label>Slug URL ({{ activeLocale.toUpperCase() }}) *</label>
                <input
                  v-model="form.translations[activeLocale].slug"
                  type="text"
                  class="input-control"
                  placeholder="contoh: tentang-kami"
                  @blur="onSlugBlur(activeLocale)"
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
                :key="selectedPage.id + '-' + activeLocale"
                v-model="form.translations[activeLocale].bodyHtml"
                placeholder="Tuliskan narasi dan informasi halaman lengkap di sini..."
                min-height="400px"
              />
            </div>
          </div>
        </div>

        <!-- Empty selection / no pages state (Right) -->
        <div v-else class="card empty-editor-card">
          <BladeIcon name="document-duplicate" size="48" class="empty-icon" />
          <h3>Tidak Ada Halaman yang Dipilih</h3>
          <p>Pilih halaman dari daftar di sebelah kiri atau buat halaman informasi baru.</p>
          <button class="btn btn-primary" @click="showCreateModal = true">
            <BladeIcon name="plus" size="18" />
            <span>Tambah Halaman Baru</span>
          </button>
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

      <!-- Rename Page Modal -->
      <div v-if="showRenameModal" class="modal-backdrop" @click.self="showRenameModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Ubah Nama / Key Halaman</h3>
            <button class="modal-close" @click="showRenameModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>Nama / Key Identifikasi Baru *</label>
              <input
                v-model="renamePageKey"
                type="text"
                class="input-control"
                placeholder="misal: tentang-kami, kontak, faq"
                @keyup.enter="submitRename"
              />
              <span class="field-hint">Hanya huruf kecil, angka, dan tanda hubung (-). Contoh: kontak-kami</span>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline" :disabled="isRenaming" @click="showRenameModal = false">Batal</button>
            <button class="btn btn-primary" :disabled="!renamePageKey.trim() || isRenaming" @click="submitRename">
              <BladeIcon name="check" size="16" />
              <span>{{ isRenaming ? 'Menyimpan...' : 'Simpan Nama' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Delete Confirmation Modal -->
      <div v-if="showDeleteModal" class="modal-backdrop" @click.self="showDeleteModal = false">
        <div class="modal-card">
          <div class="modal-header modal-header-danger">
            <h3>Hapus Halaman Statis</h3>
            <button class="modal-close" @click="showDeleteModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <div class="modal-body">
            <div class="delete-warning">
              <div class="warning-icon-wrapper">
                <BladeIcon name="exclamation-triangle" size="28" class="warning-icon" />
              </div>
              <div class="warning-text">
                <p class="warning-title">
                  Apakah Anda yakin ingin menghapus halaman
                  <strong>"{{ deleteTargetPage?.pageKey || deleteTargetPage?.page_key }}"</strong>?
                </p>
                <p class="warning-sub">
                  Semua terjemahan (ID & EN), konten, serta riwayat halaman ini akan dihapus secara permanen. Tindakan ini tidak dapat dibatalkan.
                </p>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline" :disabled="isDeleting" @click="showDeleteModal = false">Batal</button>
            <button class="btn btn-danger" :disabled="isDeleting" @click="submitDelete">
              <BladeIcon name="trash" size="16" />
              <span>{{ isDeleting ? 'Menghapus...' : 'Ya, Hapus Halaman' }}</span>
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

.page-item-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
}

.page-item-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  opacity: 0.4;
  transition: opacity 0.15s ease;
}

.page-item:hover .page-item-actions,
.page-item.active .page-item-actions {
  opacity: 1;
}

.action-icon-btn {
  background: none;
  border: 1px solid transparent;
  border-radius: 4px;
  padding: 0.3rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  transition: all 0.15s ease;
}

.action-icon-btn:hover {
  background: #ffffff;
  border-color: #cbd5e1;
}

.edit-key-btn:hover {
  color: #059669;
  border-color: #a7f3d0;
  background: #f0fdf4;
}

.delete-btn:hover {
  color: #dc2626;
  border-color: #fecaca;
  background: #fef2f2;
}

.header-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.btn-inline-action {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  border-radius: 4px;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #475569;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-inline-action:hover {
  border-color: #059669;
  color: #059669;
  background: #f0fdf4;
}

.btn-danger-outline {
  background: #ffffff;
  border: 1px solid #fca5a5;
  color: #dc2626;
}

.btn-danger-outline:hover:not(:disabled) {
  background: #fef2f2;
  border-color: #dc2626;
}

.btn-danger {
  background: #dc2626;
  color: #ffffff;
}

.btn-danger:hover:not(:disabled) {
  background: #b91c1c;
}

.empty-editor-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  color: #64748b;
  gap: 1rem;
}

.empty-icon {
  color: #94a3b8;
}

.empty-editor-card h3 {
  margin: 0;
  color: #1e293b;
  font-size: 1.15rem;
  font-weight: 700;
}

.empty-editor-card p {
  margin: 0;
  max-width: 400px;
  font-size: 0.875rem;
}

.modal-header-danger {
  background: #fef2f2;
  border-bottom: 1px solid #fecaca;
}

.modal-header-danger h3 {
  color: #991b1b;
}

.delete-warning {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.warning-icon-wrapper {
  background: #fee2e2;
  color: #dc2626;
  padding: 0.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.warning-title {
  margin: 0 0 0.5rem 0;
  font-weight: 600;
  color: #1e293b;
  font-size: 0.95rem;
}

.warning-sub {
  margin: 0;
  font-size: 0.825rem;
  color: #64748b;
  line-height: 1.45;
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
