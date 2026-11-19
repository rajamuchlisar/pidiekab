<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api/client'
import AppLayout from '../components/layout/AppLayout.vue'

const isLoading = ref(true)
const stats = ref({
  totalContents: 0,
  publishedId: 0,
  publishedEn: 0,
  draftContents: 0,
  totalCategories: 0,
  totalMedia: 0,
})

const recentContents = ref<any[]>([])

async function loadDashboardData() {
  isLoading.value = true
  try {
    // 1. Fetch contents
    const contentsRes = await api.get<any>('/api/v1/admin/contents?per_page=5')
    const cItems = contentsRes?.items || contentsRes?.data || []
    if (Array.isArray(cItems)) {
      recentContents.value = cItems
      stats.value.totalContents = contentsRes.total || contentsRes.meta?.total_items || cItems.length
      
      let pubId = 0
      let pubEn = 0
      let drafts = 0
      cItems.forEach((c: any) => {
        if (c.translations?.id?.status === 'published') pubId++
        if (c.translations?.en?.status === 'published') pubEn++
        if (c.translations?.id?.status === 'draft' || c.translations?.en?.status === 'draft') drafts++
      })
      stats.value.publishedId = pubId
      stats.value.publishedEn = pubEn
      stats.value.draftContents = drafts
    }

    // 2. Fetch categories
    const categoriesRes = await api.get<{ categories: any[] }>('/api/v1/admin/categories')
    if (categoriesRes && categoriesRes.categories) {
      stats.value.totalCategories = categoriesRes.categories.length
    }

    // 3. Fetch media
    const mediaRes = await api.get<{ items: any[]; total: number }>('/api/v1/admin/media?per_page=1')
    if (mediaRes) {
      stats.value.totalMedia = mediaRes.total || 0
    }
  } catch (err) {
    console.error('Failed to load dashboard data', err)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadDashboardData()
})
</script>

<template>
  <AppLayout>
    <div class="dashboard-view">
      <!-- Quick Action Toolbar -->
      <div class="action-banner">
        <div>
          <h2 class="banner-title">Ikhtisar Operasional PidieKab</h2>
          <p class="banner-sub">Kelola publikasi dwibahasa, kurasi beranda, dan aset media daerah.</p>
        </div>
        <div class="banner-buttons">
          <router-link to="/contents/new" class="btn btn-primary">+ Tambah Konten</router-link>
          <router-link to="/media" class="btn btn-secondary">Upload Media</router-link>
          <router-link to="/curation" class="btn btn-secondary">Kurasi Beranda</router-link>
        </div>
      </div>

      <!-- KPI Metric Cards -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon bg-emerald">📝</div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.totalContents }}</span>
            <span class="stat-label">Total Konten</span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-blue">🇮🇩</div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.publishedId }}</span>
            <span class="stat-label">Published (ID)</span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-indigo">🇬🇧</div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.publishedEn }}</span>
            <span class="stat-label">Published (EN)</span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-amber">⏳</div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.draftContents }}</span>
            <span class="stat-label">Draft / Belum Rilis</span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-purple">🖼️</div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.totalMedia }}</span>
            <span class="stat-label">Aset Media</span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-teal">🏷️</div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.totalCategories }}</span>
            <span class="stat-label">Kategori Aktif</span>
          </div>
        </div>
      </div>

      <!-- Recent Contents Table -->
      <div class="section-card">
        <div class="section-header">
          <h3 class="section-title">Konten Terbaru</h3>
          <router-link to="/contents" class="link-more">Lihat Semua Konten &rarr;</router-link>
        </div>

        <div v-if="isLoading" class="loading-state">
          Memuat data operasional...
        </div>

        <div v-else-if="recentContents.length === 0" class="empty-state">
          Belum ada konten yang dibuat. Mulai dengan membuat artikel pertama.
        </div>

        <div v-else class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th>Judul Artikel</th>
                <th>Kategori</th>
                <th>Status (ID)</th>
                <th>Status (EN)</th>
                <th>Terakhir Diubah</th>
                <th class="text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in recentContents" :key="item.id">
                <td class="font-medium">
                  {{ item.translations?.id?.title || item.translations?.en?.title || 'Tanpa Judul' }}
                </td>
                <td>
                  <span class="badge badge-category">{{ item.categoryName || 'Umum' }}</span>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="item.translations?.id?.status === 'published' ? 'badge-published' : 'badge-draft'"
                  >
                    {{ item.translations?.id?.status || 'none' }}
                  </span>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="item.translations?.en?.status === 'published' ? 'badge-published' : 'badge-draft'"
                  >
                    {{ item.translations?.en?.status || 'none' }}
                  </span>
                </td>
                <td class="text-muted">
                  {{ new Date(item.updatedAt || item.createdAt).toLocaleDateString('id-ID') }}
                </td>
                <td class="text-right">
                  <router-link :to="`/contents/${item.id}/edit`" class="btn-sm btn-outline">
                    Edit
                  </router-link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.dashboard-view {
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

/* Action Banner */
.action-banner {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.5rem 1.75rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.banner-title {
  margin: 0 0 0.35rem 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
}

.banner-sub {
  margin: 0;
  font-size: 0.875rem;
  color: #64748b;
}

.banner-buttons {
  display: flex;
  gap: 0.75rem;
  flex-shrink: 0;
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
  transition: all 0.15s ease;
  border: none;
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

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.bg-emerald { background: #d1fae5; color: #065f46; }
.bg-blue { background: #dbeafe; color: #1e40af; }
.bg-indigo { background: #e0e7ff; color: #3730a3; }
.bg-amber { background: #fef3c7; color: #92400e; }
.bg-purple { background: #f3e8ff; color: #6b21a8; }
.bg-teal { background: #ccfbf1; color: #115e59; }

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.1;
}

.stat-label {
  font-size: 0.775rem;
  color: #64748b;
  font-weight: 500;
  margin-top: 0.2rem;
}

/* Section Card */
.section-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.section-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
}

.link-more {
  font-size: 0.85rem;
  color: #059669;
  font-weight: 600;
  text-decoration: none;
}

.link-more:hover {
  text-decoration: underline;
}

/* Table */
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
}

.font-medium {
  font-weight: 600;
}

.text-muted {
  color: #64748b;
}

.text-right {
  text-align: right;
}

.badge {
  display: inline-block;
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

.badge-draft {
  background: #f1f5f9;
  color: #64748b;
}

.badge-category {
  background: #e2e8f0;
  color: #334155;
}

.btn-sm {
  padding: 0.3rem 0.65rem;
  font-size: 0.775rem;
  font-weight: 600;
  border-radius: 4px;
  text-decoration: none;
  display: inline-block;
}

.btn-outline {
  border: 1px solid #cbd5e1;
  color: #334155;
  background: #ffffff;
}

.btn-outline:hover {
  background: #f8fafc;
  border-color: #94a3b8;
}

.loading-state, .empty-state {
  padding: 3rem 1rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9rem;
}
</style>
