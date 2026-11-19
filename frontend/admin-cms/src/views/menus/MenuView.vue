<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'
import BladeIcon from '../../components/common/BladeIcon.vue'

const notification = useNotificationStore()

interface MenuItem {
  id: number
  title_id: string
  title_en: string
  item_type: 'static_page' | 'content' | 'category' | 'external'
  target_id?: number | null
  external_url?: string
  open_in_new_tab: boolean
  sort_order: number
  is_active: boolean
  target_title?: string
  resolved_url?: string
}

const menus = ref<MenuItem[]>([])
const staticPages = ref<any[]>([])
const contents = ref<any[]>([])
const categories = ref<any[]>([])

const isLoading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const currentId = ref<number | null>(null)
const isSubmitting = ref(false)

const form = reactive({
  title_id: '',
  title_en: '',
  item_type: 'static_page' as 'static_page' | 'content' | 'category' | 'external',
  target_id: null as number | null,
  external_url: '',
  open_in_new_tab: false,
  sort_order: 1,
  is_active: true,
})

async function loadData() {
  isLoading.value = true
  try {
    const [menusRes, pagesRes, contentsRes, catsRes] = await Promise.all([
      api.get<any>('/api/v1/admin/menus').catch(() => ({ data: [] })),
      api.get<any>('/api/v1/admin/static-pages').catch(() => ({ data: [] })),
      api.get<any>('/api/v1/admin/contents').catch(() => ({ data: [] })),
      api.get<any>('/api/v1/admin/categories').catch(() => ({ categories: [] })),
    ])

    menus.value = Array.isArray(menusRes?.data) ? menusRes.data : []
    staticPages.value = Array.isArray(pagesRes?.data) ? pagesRes.data : (Array.isArray(pagesRes) ? pagesRes : [])
    contents.value = Array.isArray(contentsRes?.data) ? contentsRes.data : (Array.isArray(contentsRes) ? contentsRes : [])

    const rawCats = catsRes?.categories || catsRes?.data || []
    categories.value = Array.isArray(rawCats) ? rawCats : []
  } catch (err: any) {
    notification.error('Gagal Memuat Data Menu', err.detail || 'Terjadi kesalahan saat memuat menu.')
  } finally {
    isLoading.value = false
  }
}

function openCreateModal() {
  isEditing.value = false
  currentId.value = null
  form.title_id = ''
  form.title_en = ''
  form.item_type = 'static_page'
  form.target_id = staticPages.value.length > 0 ? staticPages.value[0].id : null
  form.external_url = ''
  form.open_in_new_tab = false
  form.sort_order = menus.value.length + 1
  form.is_active = true
  showModal.value = true
}

function openEditModal(item: MenuItem) {
  isEditing.value = true
  currentId.value = item.id
  form.title_id = item.title_id
  form.title_en = item.title_en
  form.item_type = item.item_type
  form.target_id = item.target_id ?? null
  form.external_url = item.external_url || ''
  form.open_in_new_tab = item.open_in_new_tab
  form.sort_order = item.sort_order
  form.is_active = item.is_active
  showModal.value = true
}

function handleTypeChange() {
  if (form.item_type === 'static_page') {
    form.target_id = staticPages.value.length > 0 ? staticPages.value[0].id : null
    onTargetSelected()
  } else if (form.item_type === 'content') {
    form.target_id = contents.value.length > 0 ? contents.value[0].id : null
    onTargetSelected()
  } else if (form.item_type === 'category') {
    form.target_id = categories.value.length > 0 ? categories.value[0].id : null
    onTargetSelected()
  } else {
    form.target_id = null
  }
}

function onTargetSelected() {
  if (!form.target_id) return

  if (form.item_type === 'static_page') {
    const sp = staticPages.value.find((p) => p.id === form.target_id)
    if (sp) {
      if (!form.title_id) {
        form.title_id = sp.translations?.id?.title || sp.translations?.en?.title || sp.page_key || ''
      }
      if (!form.title_en && sp.translations?.en?.title) {
        form.title_en = sp.translations.en.title
      }
    }
  } else if (form.item_type === 'content') {
    const c = contents.value.find((item) => item.id === form.target_id)
    if (c) {
      if (!form.title_id) {
        form.title_id = c.title || c.translations?.id?.title || ''
      }
      if (!form.title_en && c.translations?.en?.title) {
        form.title_en = c.translations.en.title
      }
    }
  } else if (form.item_type === 'category') {
    const cat = categories.value.find((c) => c.id === form.target_id)
    if (cat) {
      if (!form.title_id) {
        form.title_id = cat.names?.id || cat.translations?.id?.name || cat.name || ''
      }
      if (!form.title_en && (cat.names?.en || cat.translations?.en?.name)) {
        form.title_en = cat.names?.en || cat.translations?.en?.name
      }
    }
  }
}

async function handleSubmit() {
  if (!form.title_id.trim()) {
    alert('Nama menu dalam Bahasa Indonesia wajib diisi.')
    return
  }

  if (form.item_type === 'external' && !form.external_url.trim()) {
    alert('URL eksternal wajib diisi.')
    return
  }

  if (form.item_type !== 'external' && !form.target_id) {
    alert('Target halaman / konten / kategori wajib dipilih.')
    return
  }

  isSubmitting.value = true
  try {
    const payload = {
      title_id: form.title_id.trim(),
      title_en: form.title_en.trim(),
      item_type: form.item_type,
      target_id: form.item_type === 'external' ? null : form.target_id,
      external_url: form.item_type === 'external' ? form.external_url.trim() : '',
      open_in_new_tab: form.open_in_new_tab,
      sort_order: Number(form.sort_order) || 0,
      is_active: form.is_active,
    }

    if (isEditing.value && currentId.value) {
      await api.put(`/api/v1/admin/menus/${currentId.value}`, payload)
      notification.success('Menu Diperbarui', 'Perubahan menu berhasil disimpan dan disinkronkan ke SSG.')
    } else {
      await api.post('/api/v1/admin/menus', payload)
      notification.success('Menu Dibuat', 'Menu baru berhasil ditambahkan dan disinkronkan ke SSG.')
    }

    showModal.value = false
    await loadData()
  } catch (err: any) {
    notification.error('Gagal Menyimpan Menu', err.detail || err.message || 'Terjadi kesalahan sistem.')
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(item: MenuItem) {
  if (!confirm(`Apakah Anda yakin ingin menghapus menu "${item.title_id}"?`)) {
    return
  }

  try {
    await api.delete(`/api/v1/admin/menus/${item.id}`)
    notification.success('Menu Dihapus', `Menu "${item.title_id}" berhasil dihapus.`)
    await loadData()
  } catch (err: any) {
    notification.error('Gagal Menghapus Menu', err.detail || 'Terjadi kesalahan sistem.')
  }
}

async function moveOrder(index: number, direction: 'up' | 'down') {
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= menus.value.length) return

  const list = [...menus.value]
  const current = list[index]
  const target = list[targetIndex]
  if (!current || !target) return
  list[index] = target
  list[targetIndex] = current

  const ids = list.map((m) => m.id)
  try {
    await api.post('/api/v1/admin/menus/reorder', { ids })
    menus.value = list
    notification.success('Urutan Diperbarui', 'Urutan menu navigasi berhasil diubah.')
    await loadData()
  } catch (err: any) {
    notification.error('Gagal Mengubah Urutan', err.detail || 'Terjadi kesalahan sistem.')
  }
}

async function toggleActive(item: MenuItem) {
  try {
    await api.put(`/api/v1/admin/menus/${item.id}`, {
      title_id: item.title_id,
      title_en: item.title_en,
      item_type: item.item_type,
      target_id: item.target_id,
      external_url: item.external_url,
      open_in_new_tab: item.open_in_new_tab,
      sort_order: item.sort_order,
      is_active: !item.is_active,
    })
    item.is_active = !item.is_active
    notification.success('Status Menu', `Menu "${item.title_id}" ${item.is_active ? 'diaktifkan' : 'dinonaktifkan'}.`)
  } catch (err: any) {
    notification.error('Gagal Mengubah Status', err.detail || 'Terjadi kesalahan sistem.')
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <AppLayout>
    <div class="menu-management-view">
      <!-- Header -->
      <div class="view-header">
        <div>
          <h2>Navigasi Menu Header</h2>
          <p class="subtitle">Kelola struktur menu navigasi header pada situs publik PidieKab (mendukung Halaman Statis, Konten/Artikel, Kategori, atau Link Eksternal).</p>
        </div>
        <button class="btn btn-primary" @click="openCreateModal">
          <BladeIcon name="plus" size="18" />
          <span>Tambah Menu Baru</span>
        </button>
      </div>

      <!-- Table Section -->
      <div class="card table-card">
        <div v-if="isLoading" class="loading-state">
          <BladeIcon name="arrow-path" size="24" class="spin" />
          <p>Memuat konfigurasi menu navigasi...</p>
        </div>

        <div v-else-if="menus.length === 0" class="empty-state">
          <BladeIcon name="bars-3" size="48" class="text-muted" />
          <h3>Belum Ada Menu Navigasi</h3>
          <p>Tambahkan menu navigasi pertama Anda untuk ditampilkan pada header website.</p>
          <button class="btn btn-primary btn-sm" @click="openCreateModal">
            <BladeIcon name="plus" size="16" />
            <span>Tambah Menu</span>
          </button>
        </div>

        <div v-else class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 70px;">Urutan</th>
                <th>Label Menu (ID & EN)</th>
                <th>Tipe Menu</th>
                <th>Target / URL</th>
                <th style="width: 90px; text-align: center;">Tab Baru</th>
                <th style="width: 100px; text-align: center;">Status</th>
                <th style="width: 140px; text-align: right;">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in menus" :key="item.id">
                <td class="order-cell">
                  <div class="order-controls">
                    <button
                      class="btn-icon btn-tiny"
                      :disabled="index === 0"
                      title="Naikkan Urutan"
                      @click="moveOrder(index, 'up')"
                    >
                      ▲
                    </button>
                    <span class="order-badge">{{ index + 1 }}</span>
                    <button
                      class="btn-icon btn-tiny"
                      :disabled="index === menus.length - 1"
                      title="Turunkan Urutan"
                      @click="moveOrder(index, 'down')"
                    >
                      ▼
                    </button>
                  </div>
                </td>
                <td>
                  <div class="menu-titles">
                    <span class="title-id font-bold">{{ item.title_id }}</span>
                    <span v-if="item.title_en" class="title-en text-muted">EN: {{ item.title_en }}</span>
                  </div>
                </td>
                <td>
                  <span v-if="item.item_type === 'static_page'" class="badge badge-indigo">Halaman Statis</span>
                  <span v-else-if="item.item_type === 'content'" class="badge badge-emerald">Konten / Artikel</span>
                  <span v-else-if="item.item_type === 'category'" class="badge badge-sky">Kategori</span>
                  <span v-else class="badge badge-slate">Link Eksternal</span>
                </td>
                <td>
                  <div class="target-info">
                    <span v-if="item.target_title" class="target-title">{{ item.target_title }}</span>
                    <code class="target-url">{{ item.resolved_url || item.external_url || '-' }}</code>
                  </div>
                </td>
                <td style="text-align: center;">
                  <span v-if="item.open_in_new_tab" class="badge badge-gray" title="Membuka di Tab Baru">Ya</span>
                  <span v-else class="text-muted">—</span>
                </td>
                <td style="text-align: center;">
                  <button
                    class="btn-badge"
                    :class="item.is_active ? 'badge-success' : 'badge-danger'"
                    @click="toggleActive(item)"
                  >
                    {{ item.is_active ? 'Aktif' : 'Nonaktif' }}
                  </button>
                </td>
                <td style="text-align: right;">
                  <div class="action-buttons">
                    <button class="btn-icon text-primary" title="Edit Menu" @click="openEditModal(item)">
                      <BladeIcon name="pencil" size="16" />
                    </button>
                    <button class="btn-icon text-danger" title="Hapus Menu" @click="handleDelete(item)">
                      <BladeIcon name="trash" size="16" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Modal Form Tambah / Edit -->
    <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Menu Navigasi' : 'Tambah Menu Navigasi' }}</h3>
          <button class="btn-icon" @click="showModal = false">✕</button>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="modal-body">
            <!-- Label ID -->
            <div class="form-group">
              <label class="form-label required">Label Menu (Bahasa Indonesia)</label>
              <input
                v-model="form.title_id"
                type="text"
                class="form-input"
                placeholder="Misal: Beranda, Wisata Alam, Tentang Pidie"
                required
              />
            </div>

            <!-- Label EN -->
            <div class="form-group">
              <label class="form-label">Label Menu (Bahasa Inggris - Opsional)</label>
              <input
                v-model="form.title_en"
                type="text"
                class="form-input"
                placeholder="Misal: Home, Nature Tourism, About Pidie"
              />
            </div>

            <!-- Item Type -->
            <div class="form-group">
              <label class="form-label required">Tipe Menu</label>
              <select v-model="form.item_type" class="form-select" @change="handleTypeChange">
                <option value="static_page">Halaman Statis (Tentang Kami, dll.)</option>
                <option value="content">Konten / Artikel Spesifik</option>
                <option value="category">Kategori Wilayah / Budaya</option>
                <option value="external">Tautan Eksternal / Kustom (URL)</option>
              </select>
            </div>

            <!-- Dynamic Target Selectors -->
            <!-- 1. Static Page Selector -->
            <div v-if="form.item_type === 'static_page'" class="form-group">
              <label class="form-label required">Pilih Halaman Statis</label>
              <select v-model="form.target_id" class="form-select" @change="onTargetSelected">
                <option :value="null" disabled>-- Pilih Halaman Statis --</option>
                <option v-for="p in staticPages" :key="p.id" :value="p.id">
                  {{ p.translations?.id?.title || p.translations?.en?.title || p.page_key }} (Slug: {{ p.translations?.id?.slug || p.page_key }})
                </option>
              </select>
            </div>

            <!-- 2. Content Selector -->
            <div v-else-if="form.item_type === 'content'" class="form-group">
              <label class="form-label required">Pilih Artikel Konten</label>
              <select v-model="form.target_id" class="form-select" @change="onTargetSelected">
                <option :value="null" disabled>-- Pilih Konten Artikel --</option>
                <option v-for="c in contents" :key="c.id" :value="c.id">
                  {{ c.title || c.translations?.id?.title }} ({{ c.category_name || c.category?.name || 'Artikel' }})
                </option>
              </select>
            </div>

            <!-- 3. Category Selector -->
            <div v-else-if="form.item_type === 'category'" class="form-group">
              <label class="form-label required">Pilih Kategori</label>
              <select v-model="form.target_id" class="form-select" @change="onTargetSelected">
                <option :value="null" disabled>-- Pilih Kategori --</option>
                <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                  {{ cat.names?.id || cat.translations?.id?.name || cat.name }} (Slug: {{ cat.slugs?.id || cat.translations?.id?.slug || cat.code }})
                </option>
              </select>
            </div>

            <!-- 4. External URL Input -->
            <div v-else-if="form.item_type === 'external'" class="form-group">
              <label class="form-label required">URL Tujuan / Eksternal</label>
              <input
                v-model="form.external_url"
                type="text"
                class="form-input"
                placeholder="Contoh: https://disbudpar.acehprov.go.id atau /"
                required
              />
              <span class="help-text">Gunakan '/' untuk beranda, atau awali dengan 'https://' untuk website luar.</span>
            </div>

            <!-- Options Row -->
            <div class="form-row">
              <div class="form-group" style="flex: 1;">
                <label class="form-label">Urutan</label>
                <input v-model.number="form.sort_order" type="number" min="0" class="form-input" />
              </div>
              <div class="form-group checkbox-group" style="flex: 1; justify-content: flex-end;">
                <label class="checkbox-label">
                  <input v-model="form.open_in_new_tab" type="checkbox" />
                  <span>Buka di Tab Baru</span>
                </label>
              </div>
              <div class="form-group checkbox-group" style="flex: 1; justify-content: flex-end;">
                <label class="checkbox-label">
                  <input v-model="form.is_active" type="checkbox" />
                  <span>Status Aktif</span>
                </label>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
              <BladeIcon v-if="isSubmitting" name="arrow-path" size="16" class="spin" />
              <span>{{ isSubmitting ? 'Menyimpan...' : 'Simpan Menu' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.menu-management-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.view-header h2 {
  font-size: 1.35rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

.subtitle {
  color: #64748b;
  font-size: 0.875rem;
  margin: 0;
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
}

.table-responsive {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.875rem;
}

.data-table th {
  background: #f8fafc;
  padding: 0.75rem 1rem;
  font-weight: 700;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.data-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}

.data-table tr:hover {
  background: #fafaf9;
}

.order-cell {
  padding: 0.5rem !important;
}

.order-controls {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.order-badge {
  font-weight: 700;
  font-size: 0.8rem;
  min-width: 20px;
  text-align: center;
  color: #475569;
}

.btn-tiny {
  padding: 0.15rem 0.35rem;
  font-size: 0.65rem;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  background: #f8fafc;
  cursor: pointer;
  color: #64748b;
}

.btn-tiny:hover:not(:disabled) {
  background: #e2e8f0;
  color: #0f172a;
}

.btn-tiny:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.menu-titles {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.title-id {
  color: #0f172a;
  font-weight: 600;
}

.title-en {
  font-size: 0.75rem;
}

.target-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.target-title {
  font-size: 0.8rem;
  color: #334155;
  font-weight: 600;
}

.target-url {
  font-size: 0.75rem;
  color: #059669;
  background: #ecfdf5;
  padding: 0.15rem 0.35rem;
  border-radius: 4px;
  display: inline-block;
  width: fit-content;
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  font-size: 0.725rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.badge-indigo { background: #e0e7ff; color: #4338ca; }
.badge-emerald { background: #d1fae5; color: #065f46; }
.badge-sky { background: #e0f2fe; color: #0369a1; }
.badge-slate { background: #f1f5f9; color: #475569; }
.badge-gray { background: #f3f4f6; color: #6b7280; }

.btn-badge {
  border: none;
  cursor: pointer;
  padding: 0.25rem 0.55rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  transition: opacity 0.15s;
}

.btn-badge:hover {
  opacity: 0.8;
}

.badge-success { background: #dcfce7; color: #15803d; }
.badge-danger { background: #fee2e2; color: #b91c1c; }

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.35rem;
}

.btn-icon {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0.4rem;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}

.btn-icon:hover {
  background: #f1f5f9;
}

.text-primary { color: #047857; }
.text-danger { color: #dc2626; }
.text-muted { color: #94a3b8; }
.font-bold { font-weight: 600; }

.loading-state, .empty-state {
  padding: 3rem 1.5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  color: #64748b;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
}

.btn-sm {
  padding: 0.4rem 0.75rem;
  font-size: 0.8rem;
}

.btn-primary {
  background: #047857;
  color: #ffffff;
}

.btn-primary:hover:not(:disabled) {
  background: #065f46;
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

/* Modal styles */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 1rem;
}

.modal-card {
  background: #ffffff;
  border-radius: 14px;
  width: 100%;
  max-width: 540px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-label {
  font-size: 0.825rem;
  font-weight: 700;
  color: #334155;
}

.form-label.required::after {
  content: " *";
  color: #dc2626;
}

.form-input, .form-select {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.875rem;
  color: #1e293b;
  outline: none;
  background: #ffffff;
  transition: border-color 0.15s;
}

.form-input:focus, .form-select:focus {
  border-color: #047857;
  box-shadow: 0 0 0 2px rgba(4, 120, 87, 0.15);
}

.form-row {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.checkbox-group {
  padding-top: 1rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.825rem;
  font-weight: 600;
  color: #334155;
  cursor: pointer;
}

.help-text {
  font-size: 0.75rem;
  color: #64748b;
}
</style>
