<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const items = ref<any[]>([])
const categories = ref<any[]>([])
const isLoading = ref(true)

const searchQuery = ref('')
const selectedCategory = ref<number | ''>('')
const selectedStatus = ref('')
const currentPage = ref(1)
const perPage = ref(12)
const totalItems = ref(0)

async function loadCategories() {
  try {
    const res = await api.get<{ categories: any[] }>('/api/v1/admin/categories')
    if (res && res.categories) {
      categories.value = res.categories
    }
  } catch (err: any) {
    console.error('Failed to load categories', err)
  }
}

function getCategoryName(id: number | null | undefined): string {
  if (!id) return 'Umum'
  const cat = categories.value.find((c) => c.id === id)
  return cat?.names?.id || cat?.names?.en || cat?.code || 'Umum'
}

async function loadContents() {
  isLoading.value = true
  try {
    const params = new URLSearchParams()
    params.set('page', String(currentPage.value))
    params.set('per_page', String(perPage.value))
    if (selectedCategory.value) {
      params.set('category_id', String(selectedCategory.value))
    }
    if (selectedStatus.value) {
      params.set('status', selectedStatus.value)
    }
    if (searchQuery.value) {
      params.set('q', searchQuery.value)
    }

    const res = await api.get<any>(`/api/v1/admin/contents?${params.toString()}`)
    const list = res?.items || res?.data || []
    if (Array.isArray(list)) {
      items.value = list
      totalItems.value = res?.total || res?.meta?.total_items || list.length
    } else {
      items.value = []
      totalItems.value = 0
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Konten', err.detail || 'Terjadi kesalahan saat memuat daftar artikel.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

async function deleteContent(item: any) {
  const title = item.translations?.id?.title || item.translations?.en?.title || `ID #${item.id}`
  const confirmed = confirm(
    `PERINGATAN: Menghapus konten "${title}" akan menghapus permanen seluruh translasi (ID/EN) dan sitasi sumber terkait.\n\nApakah Anda yakin ingin menghapus permanen?`
  )
  if (!confirmed) return

  try {
    await api.del(`/api/v1/admin/contents/${item.id}`)
    notification.success('Konten Dihapus', `Konten "${title}" berhasil dihapus permanen.`)
    loadContents()
  } catch (err: any) {
    notification.error('Gagal Menghapus', err.detail || 'Konten tidak dapat dihapus.', err.requestId)
  }
}

watch([selectedCategory, selectedStatus], () => {
  currentPage.value = 1
  loadContents()
})

function onSearchSubmit() {
  currentPage.value = 1
  loadContents()
}

function getStatusBadgeClass(status?: string) {
  if (status === 'published') return 'badge-published'
  if (status === 'scheduled') return 'badge-scheduled'
  if (status === 'archived') return 'badge-archived'
  return 'badge-draft'
}

function formatStatus(status?: string) {
  if (status === 'published') return 'Published'
  if (status === 'scheduled') return 'Dijadwalkan'
  if (status === 'archived') return 'Diarsipkan'
  if (status === 'draft') return 'Draft'
  return 'Draft'
}

onMounted(() => {
  loadCategories()
  loadContents()
})
</script>

<template>
  <AppLayout>
    <div class="content-list-page">
      <!-- Header Actions -->
      <div class="page-header">
        <div>
          <h2 class="page-title">Manajemen Konten & Artikel</h2>
          <p class="page-desc">Kelola seluruh artikel wisata, kuliner, dan budaya Kabupaten Pidie.</p>
        </div>
        <router-link to="/contents/new" class="btn btn-primary">
          <BladeIcon name="plus" size="18" />
          <span>Buat Artikel Baru</span>
        </router-link>
      </div>

      <!-- Filters & Search Toolbar -->
      <div class="filters-card">
        <form @submit.prevent="onSearchSubmit" class="search-form">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari judul artikel..."
            class="input-control search-input"
          />
          <button type="submit" class="btn btn-secondary">
            <BladeIcon name="magnifying-glass" size="16" />
            <span>Cari</span>
          </button>
        </form>

        <div class="filter-group">
          <select v-model="selectedCategory" class="select-control">
            <option :value="''">Semua Kategori</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.names?.id || cat.code }}
            </option>
          </select>

          <select v-model="selectedStatus" class="select-control">
            <option value="">Semua Status</option>
            <option value="draft">Draft</option>
            <option value="scheduled">Dijadwalkan</option>
            <option value="published">Published</option>
            <option value="archived">Archived</option>
          </select>
        </div>
      </div>

      <!-- Contents Table -->
      <div class="table-card">
        <div v-if="isLoading" class="loading-state">
          Memuat daftar konten...
        </div>

        <div v-else-if="items.length === 0" class="empty-state">
          <span class="empty-icon"><BladeIcon name="inbox" size="48" /></span>
          <p>Tidak ada konten yang sesuai dengan filter atau kata kunci pencarian.</p>
          <router-link to="/contents/new" class="btn btn-primary btn-sm">
            <BladeIcon name="plus" size="16" />
            <span>Buat Konten Baru</span>
          </router-link>
        </div>

        <div v-else class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 35%;">Judul Artikel</th>
                <th>Kategori</th>
                <th>Status (ID)</th>
                <th>Status (EN)</th>
                <th>Terakhir Diubah</th>
                <th class="text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in items" :key="item.id">
                <td>
                  <div class="title-cell">
                    <span class="content-title">
                      {{ item.translations?.id?.title || item.translations?.en?.title || 'Tanpa Judul' }}
                    </span>
                    <span v-if="item.kecamatanName" class="content-sub">
                      <BladeIcon name="map-pin" size="14" />
                      <span>{{ item.kecamatanName }}</span>
                    </span>
                  </div>
                </td>
                <td>
                  <span class="badge badge-category">{{ item.categoryName || getCategoryName(item.category_id || item.categoryId) || 'Umum' }}</span>
                </td>
                <td>
                  <div class="status-cell">
                    <span
                      class="badge"
                      :class="getStatusBadgeClass(item.translations?.id?.status)"
                      :title="item.translations?.id?.scheduled_at ? `Dijadwalkan: ${new Date(item.translations?.id?.scheduled_at).toLocaleString('id-ID')}` : ''"
                    >
                      <BladeIcon v-if="item.translations?.id?.status === 'scheduled'" name="clock" size="12" />
                      <span>{{ formatStatus(item.translations?.id?.status) }}</span>
                    </span>
                    <span v-if="item.translations?.id?.status === 'scheduled' && item.translations?.id?.scheduled_at" class="scheduled-time">
                      {{ new Date(item.translations.id.scheduled_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) }}
                    </span>
                  </div>
                </td>
                <td>
                  <div class="status-cell">
                    <span
                      class="badge"
                      :class="getStatusBadgeClass(item.translations?.en?.status)"
                      :title="item.translations?.en?.scheduled_at ? `Dijadwalkan: ${new Date(item.translations?.en?.scheduled_at).toLocaleString('id-ID')}` : ''"
                    >
                      <BladeIcon v-if="item.translations?.en?.status === 'scheduled'" name="clock" size="12" />
                      <span>{{ formatStatus(item.translations?.en?.status) }}</span>
                    </span>
                    <span v-if="item.translations?.en?.status === 'scheduled' && item.translations?.en?.scheduled_at" class="scheduled-time">
                      {{ new Date(item.translations.en.scheduled_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) }}
                    </span>
                  </div>
                </td>
                <td class="text-muted">
                  {{ new Date(item.updatedAt || item.updated_at || item.createdAt || item.created_at).toLocaleDateString('id-ID') }}
                </td>
                <td class="text-right">
                  <div class="action-buttons">
                    <router-link :to="`/contents/${item.id}/edit`" class="btn-sm btn-outline">
                      <BladeIcon name="pencil-square" size="14" />
                      <span>Edit</span>
                    </router-link>
                    <button @click="deleteContent(item)" class="btn-sm btn-danger">
                      <BladeIcon name="trash" size="14" />
                      <span>Hapus</span>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalItems > perPage" class="pagination-bar">
          <span class="pagination-info">Menampilkan {{ items.length }} dari {{ totalItems }} artikel</span>
          <div class="pagination-controls">
            <button
              class="btn-page"
              :disabled="currentPage <= 1"
              @click="currentPage--; loadContents()"
            >
              <BladeIcon name="chevron-left" size="14" />
              <span>Sebelumnya</span>
            </button>
            <span class="page-current">Halaman {{ currentPage }}</span>
            <button
              class="btn-page"
              :disabled="currentPage * perPage >= totalItems"
              @click="currentPage++; loadContents()"
            >
              <span>Selanjutnya</span>
              <BladeIcon name="chevron-right" size="14" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.content-list-page {
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

.filters-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.search-form {
  display: flex;
  gap: 0.5rem;
  flex: 1;
  max-width: 400px;
}

.filter-group {
  display: flex;
  gap: 0.75rem;
}

.input-control, .select-control {
  padding: 0.5rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #1e293b;
}

.search-input {
  width: 100%;
}

.btn {
  display: inline-flex;
  align-items: center;
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 6px;
  text-decoration: none;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
}

.btn-primary {
  background: #059669;
  color: #ffffff;
}

.btn-primary:hover {
  background: #047857;
}

.btn-secondary {
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

.table-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
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
  color: #475569;
  font-weight: 600;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  color: #1e293b;
  vertical-align: middle;
}

.title-cell {
  display: flex;
  flex-direction: column;
}

.content-title {
  font-weight: 600;
  color: #0f172a;
}

.content-sub {
  font-size: 0.75rem;
  color: #64748b;
  margin-top: 0.2rem;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.status-cell {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  align-items: flex-start;
}

.scheduled-time {
  font-size: 0.72rem;
  color: #7c3aed;
  font-weight: 500;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: capitalize;
}

.badge-published {
  background: #d1fae5;
  color: #065f46;
}

.badge-scheduled {
  background: #ede9fe;
  color: #6d28d9;
  border: 1px solid #c4b5fd;
}

.badge-archived {
  background: #fee2e2;
  color: #991b1b;
}

.badge-draft {
  background: #f1f5f9;
  color: #64748b;
}

.badge-category {
  background: #e2e8f0;
  color: #334155;
}

.text-muted {
  color: #64748b;
}

.text-right {
  text-align: right;
}

.action-buttons {
  display: inline-flex;
  gap: 0.4rem;
}

.btn-sm {
  padding: 0.35rem 0.65rem;
  font-size: 0.775rem;
  font-weight: 600;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
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
  border: 1px solid #fecaca;
  background: #fee2e2;
  color: #991b1b;
}

.btn-danger:hover {
  background: #fca5a5;
}

.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.85rem 1.25rem;
  border-top: 1px solid #f1f5f9;
  font-size: 0.85rem;
}

.pagination-info {
  color: #64748b;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.btn-page {
  padding: 0.35rem 0.75rem;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  color: #334155;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-current {
  font-weight: 600;
  color: #0f172a;
}

.loading-state, .empty-state {
  padding: 4rem 1rem;
  text-align: center;
  color: #64748b;
}

.empty-icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 0.5rem;
}
</style>
