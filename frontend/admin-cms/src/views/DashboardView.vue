<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api } from '../api/client'
import AppLayout from '../components/layout/AppLayout.vue'

const isLoading = ref(true)
const isRefreshing = ref(false)

const stats = ref<any>({
  visitor_kpis: {
    active_visitors_now: 1,
    total_visitors_today: 0,
    total_views_today: 0,
    total_views_all_time: 0,
    mobile_percentage: 65,
    desktop_percentage: 35,
    indonesian_traffic_pct: 88,
    english_traffic_pct: 12,
  },
  content_kpis: {
    total_contents: 0,
    published_id: 0,
    published_en: 0,
    draft_contents: 0,
    total_categories: 0,
    total_media: 0,
    total_downloads: 0,
    total_static_pages: 0,
  },
  traffic_trend_7d: [],
  top_articles: [],
  category_traffic: [],
  recent_articles: [],
})

let refreshTimer: any = null

async function loadDashboardData(silent = false) {
  if (!silent) {
    if (!stats.value.content_kpis.total_contents) {
      isLoading.value = true
    } else {
      isRefreshing.value = true
    }
  }

  try {
    const res = await api.get<any>('/api/v1/admin/dashboard/stats')
    if (res && res.data) {
      stats.value = res.data
    }
  } catch (err) {
    console.error('Failed to load dashboard analytics', err)
  } finally {
    isLoading.value = false
    isRefreshing.value = false
  }
}

const maxTrendViews = computed(() => {
  const list = stats.value.traffic_trend_7d || []
  if (list.length === 0) return 100
  let m = 1
  for (const day of list) {
    if (day.views > m) m = day.views
  }
  return m
})

onMounted(() => {
  loadDashboardData()
  // Auto-refresh active stats every 25 seconds
  refreshTimer = setInterval(() => {
    loadDashboardData(true)
  }, 25000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <AppLayout>
    <div class="dashboard-view">
      <!-- Quick Action Toolbar -->
      <div class="action-banner">
        <div>
          <div class="title-with-badge">
            <h2 class="banner-title">Ikhtisar &amp; Analitik Real-Time PidieKab</h2>
            <span class="live-pill">
              <span class="live-dot"></span>
              <span>LIVE SYSTEM</span>
            </span>
          </div>
          <p class="banner-sub">
            Pantau pengunjung saat ini, statistik pembaca artikel, performa publikasi dwibahasa, dan aset media secara langsung.
          </p>
        </div>
        <div class="banner-buttons">
          <button
            type="button"
            class="btn btn-refresh"
            :disabled="isRefreshing"
            @click="loadDashboardData(false)"
            title="Muat Ulang Statistik"
          >
            <BladeIcon name="arrow-path" size="16" :class="{ 'spin-icon': isRefreshing }" />
            <span>{{ isRefreshing ? 'Memperbarui...' : 'Segarkan Data' }}</span>
          </button>
          <router-link to="/contents/new" class="btn btn-primary">
            <BladeIcon name="plus" size="16" />
            <span>Tambah Konten</span>
          </router-link>
          <router-link to="/media" class="btn btn-secondary">
            <BladeIcon name="arrow-up-tray" size="16" />
            <span>Upload Media</span>
          </router-link>
          <router-link to="/curation" class="btn btn-secondary">
            <BladeIcon name="sparkles" size="16" />
            <span>Kurasi Beranda</span>
          </router-link>
        </div>
      </div>

      <!-- Real-Time Traffic & Visitor Cards (Primary Row) -->
      <div class="visitor-kpi-grid">
        <!-- Active Live Visitors Now -->
        <div class="stat-card card-highlight">
          <div class="stat-icon bg-emerald-glow">
            <BladeIcon name="bolt" size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-header-row">
              <span class="stat-value text-emerald">
                {{ stats.visitor_kpis.active_visitors_now }}
              </span>
              <span class="live-tag">Online</span>
            </div>
            <span class="stat-label font-semibold">Pengunjung Saat Ini</span>
            <span class="stat-subtext">Aktif dalam 5 menit terakhir</span>
          </div>
        </div>

        <!-- Total Visitors Today -->
        <div class="stat-card">
          <div class="stat-icon bg-blue">
            <BladeIcon name="users" size="24" />
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.visitor_kpis.total_visitors_today?.toLocaleString('id-ID') }}</span>
            <span class="stat-label">Pengunjung Hari Ini</span>
            <span class="stat-subtext">Pengunjung unik harian</span>
          </div>
        </div>

        <!-- Total Views Today -->
        <div class="stat-card">
          <div class="stat-icon bg-indigo">
            <BladeIcon name="eye" size="24" />
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.visitor_kpis.total_views_today?.toLocaleString('id-ID') }}</span>
            <span class="stat-label">Views Artikel Hari Ini</span>
            <span class="stat-subtext">Total impresi halaman</span>
          </div>
        </div>

        <!-- All-Time Total Views -->
        <div class="stat-card">
          <div class="stat-icon bg-purple">
            <BladeIcon name="chart-bar" size="24" />
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ stats.visitor_kpis.total_views_all_time?.toLocaleString('id-ID') }}</span>
            <span class="stat-label">Total Views Keseluruhan</span>
            <span class="stat-subtext">Akumulasi pembaca publik</span>
          </div>
        </div>
      </div>

      <!-- Operational Content Metrics Grid (Secondary Row) -->
      <div class="content-kpi-grid">
        <div class="mini-card">
          <span class="mini-label">Total Konten</span>
          <span class="mini-value">{{ stats.content_kpis.total_contents }}</span>
          <span class="mini-hint">Artikel tersimpan</span>
        </div>
        <div class="mini-card border-green">
          <span class="mini-label text-emerald">Published (ID)</span>
          <span class="mini-value text-emerald">{{ stats.content_kpis.published_id }}</span>
          <span class="mini-hint">Bahasa Indonesia aktif</span>
        </div>
        <div class="mini-card border-indigo">
          <span class="mini-label text-indigo">Published (EN)</span>
          <span class="mini-value text-indigo">{{ stats.content_kpis.published_en }}</span>
          <span class="mini-hint">English aktif</span>
        </div>
        <div class="mini-card border-amber">
          <span class="mini-label text-amber">Draft Konten</span>
          <span class="mini-value text-amber">{{ stats.content_kpis.draft_contents }}</span>
          <span class="mini-hint">Belum rilis publik</span>
        </div>
        <div class="mini-card">
          <span class="mini-label">Aset Media</span>
          <span class="mini-value">{{ stats.content_kpis.total_media }}</span>
          <span class="mini-hint">Foto &amp; berkas visual</span>
        </div>
        <div class="mini-card">
          <span class="mini-label">Kategori Aktif</span>
          <span class="mini-value">{{ stats.content_kpis.total_categories }}</span>
          <span class="mini-hint">Kanal informasi</span>
        </div>
        <div class="mini-card">
          <span class="mini-label">Dokumen Unduhan</span>
          <span class="mini-value">{{ stats.content_kpis.total_downloads }}</span>
          <span class="mini-hint">File publik</span>
        </div>
        <div class="mini-card">
          <span class="mini-label">Halaman Statis</span>
          <span class="mini-value">{{ stats.content_kpis.total_static_pages }}</span>
          <span class="mini-hint">Tentang, Kontak, dll</span>
        </div>
      </div>

      <!-- Analytics Visual Section (2 Columns) -->
      <div class="analytics-split-grid">
        <!-- Left: 7-Day Traffic Trend -->
        <div class="section-card">
          <div class="section-header">
            <div>
              <h3 class="section-title">Tren Kunjungan &amp; Views (7 Hari Terakhir)</h3>
              <p class="section-sub">Statistik pembaca artikel dan pengunjung unik harian.</p>
            </div>
            <div class="trend-legend">
              <span class="legend-item"><span class="legend-color bg-bar-views"></span> Views</span>
              <span class="legend-item"><span class="legend-color bg-bar-visitors"></span> Pengunjung</span>
            </div>
          </div>

          <div v-if="stats.traffic_trend_7d.length === 0" class="empty-state-sm">
            Belum ada catatan trafik untuk 7 hari terakhir.
          </div>

          <div v-else class="chart-container">
            <div class="bars-wrapper">
              <div
                v-for="day in stats.traffic_trend_7d"
                :key="day.date"
                class="bar-column"
              >
                <div class="bar-pair">
                  <!-- Views bar -->
                  <div
                    class="bar-fill bar-views"
                    :style="{ height: `${Math.max(12, Math.round((day.views / maxTrendViews) * 140))}px` }"
                    :title="`${day.views} Views pada ${day.date}`"
                  >
                    <span v-if="day.views > 0" class="bar-tooltip-val">{{ day.views }}</span>
                  </div>
                  <!-- Visitors bar -->
                  <div
                    class="bar-fill bar-visitors"
                    :style="{ height: `${Math.max(8, Math.round((day.visitors / maxTrendViews) * 140))}px` }"
                    :title="`${day.visitors} Pengunjung pada ${day.date}`"
                  ></div>
                </div>
                <div class="bar-label">
                  <span class="day-name">{{ day.day_name }}</span>
                  <span class="date-short">{{ day.date.slice(5) }}</span>
                </div>
              </div>
            </div>

            <!-- Demographics and Platform split -->
            <div class="demographics-row">
              <div class="demo-box">
                <div class="demo-header">
                  <span class="demo-title">Perangkat Pembaca</span>
                  <span class="demo-val">Mobile {{ Math.round(stats.visitor_kpis.mobile_percentage) }}% &bull; Desktop {{ Math.round(stats.visitor_kpis.desktop_percentage) }}%</span>
                </div>
                <div class="progress-bar-wrap">
                  <div class="prog-fill prog-mobile" :style="{ width: `${stats.visitor_kpis.mobile_percentage}%` }"></div>
                  <div class="prog-fill prog-desktop" :style="{ width: `${stats.visitor_kpis.desktop_percentage}%` }"></div>
                </div>
              </div>

              <div class="demo-box">
                <div class="demo-header">
                  <span class="demo-title">Bahasa Pembaca</span>
                  <span class="demo-val">ID {{ Math.round(stats.visitor_kpis.indonesian_traffic_pct) }}% &bull; EN {{ Math.round(stats.visitor_kpis.english_traffic_pct) }}%</span>
                </div>
                <div class="progress-bar-wrap">
                  <div class="prog-fill prog-id" :style="{ width: `${stats.visitor_kpis.indonesian_traffic_pct}%` }"></div>
                  <div class="prog-fill prog-en" :style="{ width: `${stats.visitor_kpis.english_traffic_pct}%` }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Top Viewed Articles -->
        <div class="section-card">
          <div class="section-header">
            <div>
              <h3 class="section-title">Artikel Paling Banyak Dibaca</h3>
              <p class="section-sub">Peringkat artikel dengan impresi dan views tertinggi.</p>
            </div>
            <BladeIcon name="fire" size="20" class="text-amber" />
          </div>

          <div v-if="stats.top_articles.length === 0" class="empty-state-sm">
            Belum ada data artikel terpopuler.
          </div>

          <div v-else class="top-articles-list">
            <div
              v-for="(art, idx) in stats.top_articles"
              :key="art.content_id"
              class="top-article-item"
            >
              <div class="rank-badge" :class="'rank-' + (idx + 1)">
                #{{ idx + 1 }}
              </div>
              <div class="article-info">
                <router-link :to="`/contents/${art.content_id}/edit`" class="article-title" :title="art.title">
                  {{ art.title }}
                </router-link>
                <div class="article-meta">
                  <span class="badge badge-category-sm">{{ art.category_name }}</span>
                  <span class="views-badge">
                    <BladeIcon name="eye" size="13" />
                    <strong>{{ art.views?.toLocaleString('id-ID') }}</strong> views
                  </span>
                </div>
              </div>
              <div class="article-action">
                <router-link
                  :to="`/contents/${art.content_id}/edit`"
                  class="btn-icon"
                  title="Edit Konten"
                >
                  <BladeIcon name="pencil" size="14" />
                </router-link>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Contents Table with Views -->
      <div class="section-card">
        <div class="section-header">
          <div>
            <h3 class="section-title">Daftar Konten Terkini</h3>
            <p class="section-sub">Artikel yang baru diperbarui beserta status dan data pembacanya.</p>
          </div>
          <router-link to="/contents" class="link-more">
            <span>Kelola Semua Konten</span>
            <BladeIcon name="arrow-right" size="14" />
          </router-link>
        </div>

        <div v-if="isLoading" class="loading-state">
          <BladeIcon name="arrow-path" size="22" class="spin-icon" />
          <span>Memuat data operasional dan statistik...</span>
        </div>

        <div v-else-if="stats.recent_articles.length === 0" class="empty-state">
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
                <th>Total Views</th>
                <th>Terakhir Diubah</th>
                <th class="text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in stats.recent_articles" :key="item.id">
                <td class="font-medium article-title-cell">
                  <router-link :to="`/contents/${item.id}/edit`" class="table-article-link">
                    {{ item.title_id || item.title_en || 'Tanpa Judul' }}
                  </router-link>
                </td>
                <td>
                  <span class="badge badge-category">{{ item.category_name || 'Umum' }}</span>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="item.status_id === 'published' ? 'badge-published' : 'badge-draft'"
                  >
                    {{ item.status_id }}
                  </span>
                </td>
                <td>
                  <span
                    class="badge"
                    :class="item.status_en === 'published' ? 'badge-published' : 'badge-draft'"
                  >
                    {{ item.status_en }}
                  </span>
                </td>
                <td>
                  <span class="views-counter">
                    <BladeIcon name="eye" size="14" class="text-muted" />
                    <strong>{{ item.views?.toLocaleString('id-ID') }}</strong>
                  </span>
                </td>
                <td class="text-muted">
                  {{ new Date(item.updated_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) }}
                </td>
                <td class="text-right">
                  <router-link :to="`/contents/${item.id}/edit`" class="btn-sm btn-outline">
                    <BladeIcon name="pencil-square" size="14" />
                    <span>Edit</span>
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
  gap: 1.5rem;
}

/* Action Banner */
.action-banner {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.4rem 1.75rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.banner-title {
  margin: 0;
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
}

.live-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: #ecfdf5;
  color: #059669;
  border: 1px solid #a7f3d0;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.live-dot {
  width: 7px;
  height: 7px;
  background: #10b981;
  border-radius: 50%;
  animation: pulse-live 1.5s infinite;
}

@keyframes pulse-live {
  0% { transform: scale(0.95); opacity: 1; box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1.2); opacity: 0.8; box-shadow: 0 0 0 5px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); opacity: 1; box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

.banner-sub {
  margin: 0.35rem 0 0 0;
  font-size: 0.85rem;
  color: #64748b;
  max-width: 650px;
}

.banner-buttons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 0.95rem;
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 6px;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.15s ease;
  border: none;
}

.btn-refresh {
  background: #ffffff;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.btn-refresh:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #059669;
  color: #059669;
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

/* Visitor KPI Grid */
.visitor-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.15rem;
}

@media (max-width: 1024px) {
  .visitor-kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .visitor-kpi-grid {
    grid-template-columns: 1fr;
  }
}

.stat-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.25rem 1.4rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.card-highlight {
  border: 1px solid #a7f3d0;
  background: linear-gradient(to bottom right, #ffffff, #f0fdf4);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.bg-emerald-glow {
  background: #d1fae5;
  color: #047857;
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.25);
}

.bg-blue { background: #dbeafe; color: #1e40af; }
.bg-indigo { background: #e0e7ff; color: #3730a3; }
.bg-purple { background: #f3e8ff; color: #6b21a8; }

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-header-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stat-value {
  font-size: 1.7rem;
  font-weight: 800;
  color: #0f172a;
  line-height: 1.1;
  letter-spacing: -0.02em;
}

.text-emerald { color: #059669; }
.text-indigo { color: #4338ca; }
.text-amber { color: #d97706; }

.live-tag {
  background: #10b981;
  color: #ffffff;
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.stat-label {
  font-size: 0.825rem;
  color: #334155;
  margin-top: 0.25rem;
}

.stat-subtext {
  font-size: 0.725rem;
  color: #94a3b8;
  margin-top: 0.15rem;
}

/* Content Mini KPIs */
.content-kpi-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 0.75rem;
}

@media (max-width: 1200px) {
  .content-kpi-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 640px) {
  .content-kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.mini-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.border-green { border-color: #a7f3d0; background: #fafdfb; }
.border-indigo { border-color: #c7d2fe; background: #fafaff; }
.border-amber { border-color: #fde68a; background: #fffdfa; }

.mini-label {
  font-size: 0.725rem;
  font-weight: 600;
  color: #64748b;
}

.mini-value {
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
}

.mini-hint {
  font-size: 0.675rem;
  color: #94a3b8;
}

/* Analytics Split Grid */
.analytics-split-grid {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 1.25rem;
}

@media (max-width: 1024px) {
  .analytics-split-grid {
    grid-template-columns: 1fr;
  }
}

/* Section Card */
.section-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.4rem 1.6rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.25rem;
  gap: 1rem;
}

.section-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
}

.section-sub {
  margin: 0.2rem 0 0 0;
  font-size: 0.785rem;
  color: #64748b;
}

.trend-legend {
  display: flex;
  gap: 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #475569;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.legend-color {
  width: 10px;
  height: 10px;
  border-radius: 2px;
}

.bg-bar-views { background: #059669; }
.bg-bar-visitors { background: #93c5fd; }

/* Chart Visualization */
.chart-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.bars-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  height: 180px;
  padding: 1rem 0.5rem 0 0.5rem;
  border-bottom: 1px solid #f1f5f9;
}

.bar-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  height: 100%;
  justify-content: flex-end;
}

.bar-pair {
  display: flex;
  align-items: flex-end;
  gap: 4px;
}

.bar-fill {
  width: 16px;
  border-radius: 4px 4px 0 0;
  transition: height 0.3s ease;
  position: relative;
}

.bar-views {
  background: #059669;
}

.bar-visitors {
  background: #93c5fd;
}

.bar-tooltip-val {
  position: absolute;
  top: -18px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 0.65rem;
  font-weight: 700;
  color: #059669;
}

.bar-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 0.7rem;
}

.day-name {
  font-weight: 700;
  color: #334155;
}

.date-short {
  color: #94a3b8;
  font-size: 0.65rem;
}

/* Demographics Row */
.demographics-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  padding-top: 0.5rem;
}

@media (max-width: 640px) {
  .demographics-row {
    grid-template-columns: 1fr;
  }
}

.demo-box {
  background: #f8fafc;
  padding: 0.75rem 0.95rem;
  border-radius: 8px;
  border: 1px solid #f1f5f9;
}

.demo-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  margin-bottom: 0.4rem;
}

.demo-title {
  font-weight: 600;
  color: #334155;
}

.demo-val {
  color: #64748b;
  font-weight: 500;
}

.progress-bar-wrap {
  height: 7px;
  border-radius: 999px;
  background: #e2e8f0;
  overflow: hidden;
  display: flex;
}

.prog-fill {
  height: 100%;
}

.prog-mobile { background: #059669; }
.prog-desktop { background: #3b82f6; }
.prog-id { background: #059669; }
.prog-en { background: #6366f1; }

/* Top Articles List */
.top-articles-list {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.top-article-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.75rem;
  border-radius: 8px;
  border: 1px solid #f1f5f9;
  background: #ffffff;
  transition: all 0.15s ease;
}

.top-article-item:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.rank-badge {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 800;
  flex-shrink: 0;
  background: #f1f5f9;
  color: #475569;
}

.rank-1 { background: #fef3c7; color: #b45309; }
.rank-2 { background: #e0e7ff; color: #3730a3; }
.rank-3 { background: #d1fae5; color: #065f46; }

.article-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.article-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: #1e293b;
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.article-title:hover {
  color: #059669;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.badge-category-sm {
  background: #f1f5f9;
  color: #475569;
  font-size: 0.675rem;
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.views-badge {
  font-size: 0.725rem;
  color: #64748b;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.views-badge strong {
  color: #059669;
}

.btn-icon {
  background: none;
  border: 1px solid transparent;
  color: #94a3b8;
  padding: 0.35rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  text-decoration: none;
}

.btn-icon:hover {
  background: #f1f5f9;
  color: #059669;
}

/* Recent Articles Table */
.table-article-link {
  color: #0f172a;
  text-decoration: none;
  font-weight: 600;
}

.table-article-link:hover {
  color: #059669;
}

.views-counter {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  color: #334155;
  font-size: 0.85rem;
}

.views-counter strong {
  color: #059669;
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
}

.article-title-cell {
  max-width: 320px;
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.725rem;
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

.link-more {
  font-size: 0.825rem;
  color: #059669;
  font-weight: 600;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.link-more:hover {
  text-decoration: underline;
}

.btn-sm {
  padding: 0.3rem 0.65rem;
  font-size: 0.775rem;
  font-weight: 600;
  border-radius: 4px;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.btn-outline {
  border: 1px solid #cbd5e1;
  color: #334155;
  background: #ffffff;
}

.btn-outline:hover {
  background: #f8fafc;
  border-color: #059669;
  color: #059669;
}

.text-right {
  text-align: right;
}

.text-muted {
  color: #94a3b8;
}

.loading-state, .empty-state {
  padding: 3rem 1rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.empty-state-sm {
  padding: 2rem 1rem;
  text-align: center;
  color: #94a3b8;
  font-size: 0.825rem;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
