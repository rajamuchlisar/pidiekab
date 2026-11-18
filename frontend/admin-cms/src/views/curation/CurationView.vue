<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const activeTab = ref<'featured' | 'stories'>('featured')
const activeLocale = ref<'id' | 'en'>('id')
const isLoading = ref(true)
const isSaving = ref(false)

// Published contents available to pick
const availableContents = ref<any[]>([])

// Featured state
const featuredList = ref<Array<{ contentId: number; title: string; categoryName: string; position: number }>>([])

// Stories state
const leadContentId = ref<number | null>(null)
const supportingList = ref<Array<{ contentId: number; title: string; categoryName: string; position: number }>>([])

async function loadData() {
  isLoading.value = true
  try {
    const [contentsRes, featRes, storiesRes] = await Promise.all([
      api.get<{ items: any[] }>('/api/v1/admin/contents?per_page=100'),
      api.get<{ items: any[] }>('/api/v1/admin/curation/featured'),
      api.get<any>('/api/v1/admin/curation/stories'),
    ])

    if (contentsRes?.items) {
      availableContents.value = contentsRes.items
    }

    if (featRes?.items) {
      featuredList.value = featRes.items
    }

    if (storiesRes && storiesRes[activeLocale.value]) {
      const locStories = storiesRes[activeLocale.value]
      leadContentId.value = locStories.lead?.contentId || null
      supportingList.value = locStories.supporting || []
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Kurasi', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function addFeatured(content: any) {
  if (featuredList.value.some((f) => f.contentId === content.id)) {
    alert('Artikel ini sudah ada di daftar Unggulan.')
    return
  }
  featuredList.value.push({
    contentId: content.id,
    title: content.translations?.id?.title || content.translations?.en?.title || 'Tanpa Judul',
    categoryName: content.categoryName || 'Umum',
    position: featuredList.value.length + 1,
  })
}

function removeFeatured(index: number) {
  featuredList.value.splice(index, 1)
  featuredList.value.forEach((f, idx) => {
    f.position = idx + 1
  })
}

function addSupporting(content: any) {
  if (content.id === leadContentId.value) {
    alert('Artikel yang dipilih sudah menjadi Lead Story.')
    return
  }
  if (supportingList.value.some((s) => s.contentId === content.id)) {
    alert('Artikel ini sudah ada di daftar Supporting Stories.')
    return
  }
  supportingList.value.push({
    contentId: content.id,
    title: content.translations?.[activeLocale.value]?.title || 'Tanpa Judul',
    categoryName: content.categoryName || 'Umum',
    position: supportingList.value.length + 1,
  })
}

function removeSupporting(index: number) {
  supportingList.value.splice(index, 1)
  supportingList.value.forEach((s, idx) => {
    s.position = idx + 1
  })
}

async function saveFeatured() {
  isSaving.value = true
  try {
    const payload = {
      items: featuredList.value.map((f, idx) => ({
        content_id: f.contentId,
        position: idx + 1,
      })),
    }
    await api.put('/api/v1/admin/curation/featured', payload)
    notification.success('Kurasi Disimpan', 'Daftar Artikel Unggulan berhasil diperbarui.')
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

async function saveStories() {
  isSaving.value = true
  try {
    const payload = {
      locale: activeLocale.value,
      lead_content_id: leadContentId.value,
      supporting: supportingList.value.map((s, idx) => ({
        content_id: s.contentId,
        position: idx + 1,
      })),
    }
    await api.put('/api/v1/admin/curation/stories', payload)
    notification.success('Cerita Beranda Disimpan', `Kurasi Cerita Pidie (${activeLocale.value.toUpperCase()}) berhasil disimpan.`)
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan kurasi cerita.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <AppLayout>
    <div class="curation-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Kurasi Editorial Beranda</h2>
          <p class="page-desc">Tentukan artikel unggulan dan cerita utama yang tampil di beranda website publik.</p>
        </div>
      </div>

      <!-- Tab Switcher -->
      <div class="curation-tabs">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'featured' }"
          @click="activeTab = 'featured'"
        >
          ⭐ Artikel Unggulan (Featured)
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'stories' }"
          @click="activeTab = 'stories'"
        >
          📖 Cerita tentang Pidie (Stories)
        </button>
      </div>

      <div v-if="isLoading" class="loading-state">Memuat data kurasi...</div>

      <!-- FEATURED TAB -->
      <div v-else-if="activeTab === 'featured'" class="tab-content">
        <div class="curation-grid">
          <!-- Active Featured Items -->
          <div class="card cur-card">
            <div class="card-header">
              <div>
                <h3 class="card-title">Daftar Artikel Unggulan Terpilih</h3>
                <p class="card-sub">Urutan posisi 1, 2, 3 menentukan posisi tampil di kartu grid beranda.</p>
              </div>
              <button class="btn btn-primary" :disabled="isSaving" @click="saveFeatured">
                {{ isSaving ? 'Menyimpan...' : 'Simpan Urutan' }}
              </button>
            </div>

            <div v-if="featuredList.length === 0" class="empty-list">
              Belum ada artikel unggulan yang dipilih. Tambahkan artikel dari daftar di samping kanan.
            </div>

            <ul v-else class="curation-items">
              <li v-for="(item, idx) in featuredList" :key="item.contentId" class="cur-item">
                <span class="item-pos">#{{ idx + 1 }}</span>
                <div class="item-info">
                  <span class="item-title">{{ item.title }}</span>
                  <span class="item-cat">{{ item.categoryName }}</span>
                </div>
                <button class="btn-remove" @click="removeFeatured(idx)">Hapus</button>
              </li>
            </ul>
          </div>

          <!-- Content Picker Pool -->
          <div class="card pool-card">
            <h3 class="card-title">Pilih dari Artikel Tersedia</h3>
            <p class="card-sub">Hanya artikel yang sudah terbit (Published) yang dapat diunggulkan.</p>

            <ul class="pool-items">
              <li v-for="c in availableContents" :key="c.id" class="pool-item">
                <div class="pool-info">
                  <span class="pool-title">
                    {{ c.translations?.id?.title || c.translations?.en?.title || 'Tanpa Judul' }}
                  </span>
                  <span class="pool-cat">{{ c.categoryName }}</span>
                </div>
                <button class="btn-sm btn-outline" @click="addFeatured(c)">+ Pilih</button>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- STORIES TAB -->
      <div v-else-if="activeTab === 'stories'" class="tab-content">
        <!-- Locale Switcher for Stories -->
        <div class="sub-tabs">
          <button
            class="sub-tab"
            :class="{ active: activeLocale === 'id' }"
            @click="activeLocale = 'id'; loadData()"
          >
            🇮🇩 Cerita Pidie (Bahasa Indonesia)
          </button>
          <button
            class="sub-tab"
            :class="{ active: activeLocale === 'en' }"
            @click="activeLocale = 'en'; loadData()"
          >
            🇬🇧 Pidie Stories (English)
          </button>
        </div>

        <div class="curation-grid">
          <div class="card cur-card">
            <div class="card-header">
              <div>
                <h3 class="card-title">Cerita Utama & Pendukung ({{ activeLocale.toUpperCase() }})</h3>
                <p class="card-sub">Maksimal 1 Lead Story (Cerita Utama) dan kumpulan Supporting Stories.</p>
              </div>
              <button class="btn btn-primary" :disabled="isSaving" @click="saveStories">
                {{ isSaving ? 'Menyimpan...' : 'Simpan Cerita' }}
              </button>
            </div>

            <!-- Lead Story Box -->
            <div class="lead-box">
              <span class="box-badge">LEAD STORY (Cerita Utama)</span>
              <div v-if="leadContentId" class="lead-selected">
                <div class="item-info">
                  <span class="item-title">
                    {{ availableContents.find(c => c.id === leadContentId)?.translations?.[activeLocale]?.title || 'Artikel #' + leadContentId }}
                  </span>
                </div>
                <button class="btn-remove" @click="leadContentId = null">Ganti / Hapus</button>
              </div>
              <div v-else class="lead-empty">
                Belum ada Lead Story yang dipilih. Klik tombol "+ Jadikan Lead" pada daftar artikel di sebelah kanan.
              </div>
            </div>

            <!-- Supporting Stories List -->
            <h4 class="section-subtitle">Supporting Stories (Cerita Pendukung)</h4>
            <div v-if="supportingList.length === 0" class="empty-list">
              Belum ada Supporting Story yang dipilih.
            </div>
            <ul v-else class="curation-items">
              <li v-for="(item, idx) in supportingList" :key="item.contentId" class="cur-item">
                <span class="item-pos">#{{ idx + 1 }}</span>
                <div class="item-info">
                  <span class="item-title">{{ item.title }}</span>
                  <span class="item-cat">{{ item.categoryName }}</span>
                </div>
                <button class="btn-remove" @click="removeSupporting(idx)">Hapus</button>
              </li>
            </ul>
          </div>

          <!-- Pool for Stories -->
          <div class="card pool-card">
            <h3 class="card-title">Artikel Tersedia ({{ activeLocale.toUpperCase() }})</h3>
            <p class="card-sub">Pilih artikel untuk dijadikan Lead Story atau Supporting Story.</p>

            <ul class="pool-items">
              <li v-for="c in availableContents" :key="c.id" class="pool-item">
                <div class="pool-info">
                  <span class="pool-title">{{ c.translations?.[activeLocale]?.title || 'Tanpa Judul' }}</span>
                  <span class="pool-cat">{{ c.categoryName }}</span>
                </div>
                <div class="pool-actions">
                  <button class="btn-xs btn-outline" @click="leadContentId = c.id">Lead</button>
                  <button class="btn-xs btn-outline ml-1" @click="addSupporting(c)">+ Support</button>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.curation-page {
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

.curation-tabs, .sub-tabs {
  display: flex;
  gap: 0.5rem;
}

.tab-btn {
  padding: 0.65rem 1.25rem;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
  color: #475569;
}

.tab-btn.active {
  background: #059669;
  color: #ffffff;
  border-color: #059669;
}

.sub-tab {
  padding: 0.5rem 1rem;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  color: #475569;
}

.sub-tab.active {
  background: #064e3b;
  color: #ffffff;
  border-color: #064e3b;
}

.curation-grid {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 1.5rem;
  align-items: start;
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.25rem;
  gap: 1rem;
}

.card-title { margin: 0 0 0.25rem 0; font-size: 1.05rem; font-weight: 700; color: #0f172a; }
.card-sub { margin: 0; font-size: 0.8rem; color: #64748b; }

.curation-items, .pool-items {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.cur-item, .pool-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 0.65rem 0.85rem;
  border-radius: 6px;
  gap: 0.5rem;
}

.item-pos {
  font-weight: 800;
  color: #059669;
  font-size: 0.95rem;
}

.item-info, .pool-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.item-title, .pool-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: #0f172a;
}

.item-cat, .pool-cat {
  font-size: 0.725rem;
  color: #64748b;
}

.btn-remove {
  background: none;
  border: none;
  color: #ef4444;
  font-size: 0.775rem;
  font-weight: 600;
  cursor: pointer;
}

.lead-box {
  background: #f0fdf4;
  border: 2px dashed #86efac;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.box-badge {
  display: inline-block;
  font-size: 0.7rem;
  font-weight: 800;
  background: #059669;
  color: #ffffff;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.lead-selected {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.lead-empty {
  font-size: 0.85rem;
  color: #166534;
}

.section-subtitle {
  margin: 0 0 0.75rem 0;
  font-size: 0.9rem;
  font-weight: 700;
  color: #334155;
}

.pool-card {
  max-height: 580px;
  overflow-y: auto;
}

.pool-actions { display: flex; }
.ml-1 { margin-left: 0.35rem; }

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
.btn-outline { border: 1px solid #cbd5e1; background: #ffffff; color: #334155; cursor: pointer; }
.btn-outline:hover { background: #f8fafc; }

.btn-sm { padding: 0.35rem 0.65rem; font-size: 0.775rem; border-radius: 4px; }
.btn-xs { padding: 0.2rem 0.45rem; font-size: 0.7rem; border-radius: 4px; }

.loading-state, .empty-list {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
