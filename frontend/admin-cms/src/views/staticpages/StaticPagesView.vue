<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const pages = ref<any[]>([])
const isLoading = ref(true)
const isSaving = ref(false)
const selectedPage = ref<any | null>(null)
const activeLocale = ref<'id' | 'en'>('id')

const form = reactive({
  pageKey: '',
  translations: {
    id: { title: '', slug: '', bodyHtml: '', status: 'draft' as 'draft' | 'published' | 'archived' },
    en: { title: '', slug: '', bodyHtml: '', status: 'draft' as 'draft' | 'published' | 'archived' },
  },
})

async function loadPages() {
  isLoading.value = true
  try {
    const res = await api.get<{ pages: any[] }>('/api/v1/admin/static-pages')
    if (res?.pages) {
      pages.value = res.pages
      if (pages.value.length > 0 && !selectedPage.value) {
        selectPage(pages.value[0])
      }
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Halaman', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function selectPage(page: any) {
  selectedPage.value = page
  form.pageKey = page.pageKey || page.page_key
  form.translations.id = {
    title: page.translations?.id?.title || '',
    slug: page.translations?.id?.slug || '',
    bodyHtml: page.translations?.id?.bodyHtml || page.translations?.id?.body_html || '',
    status: page.translations?.id?.status || 'draft',
  }
  form.translations.en = {
    title: page.translations?.en?.title || '',
    slug: page.translations?.en?.slug || '',
    bodyHtml: page.translations?.en?.bodyHtml || page.translations?.en?.body_html || '',
    status: page.translations?.en?.status || 'draft',
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
    notification.success('Halaman Disimpan', `Halaman "${form.pageKey}" berhasil disimpan.`)
    loadPages()
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSaving.value = false
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
      </div>

      <div v-if="isLoading" class="loading-state">Memuat halaman...</div>

      <div v-else class="layout-grid">
        <!-- List of Pages (Left) -->
        <div class="card page-selector-card">
          <h3 class="card-title">Daftar Halaman</h3>
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
              <h3 class="card-title">Edit Halaman: <code>{{ form.pageKey }}</code></h3>
              <p class="card-sub">Isi konten dan pilih status publikasi independen untuk ID dan EN.</p>
            </div>
            <button class="btn btn-primary" :disabled="isSaving" @click="savePage">
              <BladeIcon name="bookmark" size="16" />
              <span>{{ isSaving ? 'Menyimpan...' : 'Simpan Halaman' }}</span>
            </button>
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
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeLocale === 'en' }"
              @click="activeLocale = 'en'"
            >
              <BladeIcon name="globe-alt" size="16" />
              <span>English</span>
            </button>
          </div>

          <div class="form-body">
            <div class="form-row">
              <div class="form-group flex-1">
                <label>Judul Halaman ({{ activeLocale.toUpperCase() }}) *</label>
                <input v-model="form.translations[activeLocale].title" type="text" class="input-control" />
              </div>
              <div class="form-group flex-1">
                <label>Slug URL ({{ activeLocale.toUpperCase() }}) *</label>
                <input v-model="form.translations[activeLocale].slug" type="text" class="input-control" />
              </div>
              <div class="form-group" style="width: 140px;">
                <label>Status Publikasi</label>
                <select v-model="form.translations[activeLocale].status" class="select-control">
                  <option value="draft">Draft</option>
                  <option value="published">Published</option>
                  <option value="archived">Archived</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label>Konten Halaman (HTML / Markdown) *</label>
              <textarea
                v-model="form.translations[activeLocale].bodyHtml"
                rows="14"
                class="textarea-control font-mono"
                placeholder="<p>Tuliskan informasi halaman di sini...</p>"
              ></textarea>
            </div>
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
}

.page-title { margin: 0 0 0.25rem 0; font-size: 1.35rem; font-weight: 700; color: #0f172a; }
.page-desc { margin: 0; font-size: 0.85rem; color: #64748b; }

.layout-grid {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 1.5rem;
  align-items: start;
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
}

.card-title { margin: 0 0 0.25rem 0; font-size: 1.05rem; font-weight: 700; color: #0f172a; }
.card-sub { margin: 0 0 1rem 0; font-size: 0.8rem; color: #64748b; }

.page-list {
  list-style: none;
  padding: 0;
  margin: 1rem 0 0 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.page-item {
  padding: 0.75rem;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: all 0.15s;
}

.page-item:hover {
  background: #f8fafc;
}

.page-item.active {
  background: #ecfdf5;
  border-color: #059669;
}

.page-meta {
  display: flex;
  flex-direction: column;
}

.page-key {
  font-weight: 600;
  font-size: 0.9rem;
  color: #0f172a;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.page-label {
  font-size: 0.75rem;
  color: #64748b;
  margin-top: 0.2rem;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.25rem;
}

.lang-tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.tab-btn {
  padding: 0.5rem 1rem;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  color: #475569;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.tab-btn.active {
  background: #059669;
  color: #ffffff;
  border-color: #059669;
}

.form-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
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

.input-control, .select-control, .textarea-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
  font-family: inherit;
}

.font-mono { font-family: monospace; }

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

.loading-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
