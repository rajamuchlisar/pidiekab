<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const activeTab = ref<'featured' | 'stories'>('featured')
const activeLocale = ref<'id' | 'en'>('id')
const isLoading = ref(true)
const isSaving = ref(false)
const searchQuery = ref('')

// Published contents available to pick
const availableContents = ref<any[]>([])
const categoryMap = ref<Record<string, string>>({})

// Featured state
const featuredList = ref<Array<{ contentId: number; title: string; categoryName: string; position: number }>>([])

// Stories state
const leadContentId = ref<number | null>(null)
const supportingList = ref<Array<{ contentId: number; title: string; categoryName: string; position: number }>>([])
const allStoriesData = ref<{ id?: any; en?: any }>({})

function getCategoryName(catId: any): string {
  if (!catId) return 'Umum'
  return categoryMap.value[String(catId)] || 'Umum'
}

function isContentPublished(content: any, locale?: 'id' | 'en'): boolean {
  if (!content || !content.translations) return false
  if (locale) {
    return content.translations[locale]?.status === 'published'
  }
  return content.translations.id?.status === 'published' || content.translations.en?.status === 'published'
}

const publishedContents = computed(() => {
  const reqLocale = activeTab.value === 'stories' ? activeLocale.value : undefined
  return availableContents.value.filter((c: any) => isContentPublished(c, reqLocale))
})

const filteredContents = computed(() => {
  const pool = publishedContents.value
  if (!searchQuery.value.trim()) return pool
  const q = searchQuery.value.toLowerCase()
  return pool.filter((c: any) => {
    const tActive = c.translations?.[activeLocale.value]?.title?.toLowerCase() || ''
    const tID = c.translations?.id?.title?.toLowerCase() || ''
    const tEN = c.translations?.en?.title?.toLowerCase() || ''
    const cat = (c.categoryName || '').toLowerCase()
    return tActive.includes(q) || tID.includes(q) || tEN.includes(q) || cat.includes(q)
  })
})

async function loadData() {
  isLoading.value = true
  try {
    const [contentsRes, featRes, storiesRes, catRes] = await Promise.all([
      api.get<{ items: any[] }>('/api/v1/admin/contents?per_page=100&status=published'),
      api.get<{ items: any[] }>('/api/v1/admin/curation/featured'),
      api.get<any>('/api/v1/admin/curation/stories'),
      api.get<{ items: any[] }>('/api/v1/admin/categories').catch(() => ({ items: [] })),
    ])

    if (catRes?.items) {
      const map: Record<string, string> = {}
      catRes.items.forEach((cat: any) => {
        map[String(cat.id)] = cat.translations?.id?.name || cat.translations?.en?.name || cat.slug || 'Umum'
      })
      categoryMap.value = map
    }

    if (contentsRes?.items) {
      availableContents.value = contentsRes.items.map((c: any) => ({
        ...c,
        categoryName: getCategoryName(c.category_id),
      }))
    }

    if (featRes?.items) {
      featuredList.value = featRes.items.map((f: any, idx: number) => ({
        contentId: f.contentId,
        title: f.title,
        categoryName: f.categoryName || getCategoryName(availableContents.value.find((c) => c.id === f.contentId)?.category_id),
        position: f.position || idx + 1,
      }))
    }

    if (storiesRes) {
      allStoriesData.value = storiesRes
      syncStoriesForActiveLocale()
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Kurasi', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function syncStoriesForActiveLocale() {
  const locStories = allStoriesData.value[activeLocale.value]
  if (locStories) {
    leadContentId.value = locStories.lead?.contentId || null
    supportingList.value = (locStories.supporting || []).map((s: any, idx: number) => ({
      contentId: s.contentId,
      title: s.title,
      categoryName: s.categoryName || getCategoryName(availableContents.value.find((c) => c.id === s.contentId)?.category_id),
      position: idx + 1,
    }))
  } else {
    leadContentId.value = null
    supportingList.value = []
  }
}

function switchLocale(loc: 'id' | 'en') {
  activeLocale.value = loc
  syncStoriesForActiveLocale()
}

function addFeatured(content: any) {
  if (!isContentPublished(content)) {
    notification.show('error', 'Tidak Dapat Dipilih', 'Artikel belum berstatus Terbit (Published).')
    return
  }
  if (featuredList.value.some((f) => f.contentId === content.id)) {
    notification.show('warning', 'Sudah Ada', 'Artikel ini sudah ada di daftar Unggulan.')
    return
  }
  featuredList.value.push({
    contentId: content.id,
    title: content.translations?.id?.title || content.translations?.en?.title || 'Tanpa Judul',
    categoryName: content.categoryName || getCategoryName(content.category_id),
    position: featuredList.value.length + 1,
  })
}

function removeFeatured(index: number) {
  featuredList.value.splice(index, 1)
  featuredList.value.forEach((f, idx) => {
    f.position = idx + 1
  })
}

function moveFeatured(index: number, direction: 'up' | 'down') {
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= featuredList.value.length) return
  const [item] = featuredList.value.splice(index, 1)
  if (item) {
    featuredList.value.splice(targetIndex, 0, item)
    featuredList.value.forEach((f, idx) => {
      f.position = idx + 1
    })
  }
}

function setLeadStory(content: any) {
  if (!isContentPublished(content, activeLocale.value)) {
    notification.show('error', 'Tidak Dapat Dipilih', `Artikel belum berstatus Terbit dalam bahasa ${activeLocale.value === 'id' ? 'Indonesia' : 'Inggris'}.`)
    return
  }
  leadContentId.value = content.id
  // If this item was in supporting, remove it from supporting
  const existingIdx = supportingList.value.findIndex((s) => s.contentId === content.id)
  if (existingIdx !== -1) {
    removeSupporting(existingIdx)
  }
}

function addSupporting(content: any) {
  if (!isContentPublished(content, activeLocale.value)) {
    notification.show('error', 'Tidak Dapat Dipilih', `Artikel belum berstatus Terbit dalam bahasa ${activeLocale.value === 'id' ? 'Indonesia' : 'Inggris'}.`)
    return
  }
  if (content.id === leadContentId.value) {
    notification.show('warning', 'Perhatian', 'Artikel yang dipilih saat ini adalah Cerita Utama (Lead Story).')
    return
  }
  if (supportingList.value.some((s) => s.contentId === content.id)) {
    notification.show('warning', 'Sudah Ada', 'Artikel ini sudah ada di daftar Supporting Stories.')
    return
  }
  supportingList.value.push({
    contentId: content.id,
    title: content.translations?.[activeLocale.value]?.title || content.translations?.id?.title || content.translations?.en?.title || 'Tanpa Judul',
    categoryName: content.categoryName || getCategoryName(content.category_id),
    position: supportingList.value.length + 1,
  })
}

function removeSupporting(index: number) {
  supportingList.value.splice(index, 1)
  supportingList.value.forEach((s, idx) => {
    s.position = idx + 1
  })
}

function moveSupporting(index: number, direction: 'up' | 'down') {
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= supportingList.value.length) return
  const [item] = supportingList.value.splice(index, 1)
  if (item) {
    supportingList.value.splice(targetIndex, 0, item)
    supportingList.value.forEach((s, idx) => {
      s.position = idx + 1
    })
  }
}

async function saveFeatured() {
  isSaving.value = true
  try {
    const payload = {
      content_ids: featuredList.value.map((f) => f.contentId),
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
      supporting_ids: supportingList.value.map((s) => s.contentId),
      supporting: supportingList.value.map((s, idx) => ({
        content_id: s.contentId,
        position: idx + 1,
      })),
    }
    await api.put('/api/v1/admin/curation/stories', payload)
    // Update local store
    allStoriesData.value[activeLocale.value] = {
      lead: leadContentId.value ? { contentId: leadContentId.value } : null,
      supporting: [...supportingList.value],
    }
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
          <p class="page-desc">Tentukan artikel unggulan dan cerita utama yang tampil di beranda website publik (SSG).</p>
        </div>
      </div>

      <!-- Tab Switcher -->
      <div class="curation-tabs">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'featured' }"
          @click="activeTab = 'featured'"
        >
          <BladeIcon name="star" size="18" />
          <span>Artikel Unggulan (Pilihan Redaksi)</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'stories' }"
          @click="activeTab = 'stories'"
        >
          <BladeIcon name="book-open" size="18" />
          <span>Cerita tentang Pidie (Lead & Supporting)</span>
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
                <h3 class="card-title">Daftar Artikel Unggulan (Pilihan Redaksi)</h3>
                <p class="card-sub">Urutan posisi 1, 2, 3 menentukan posisi tampil di kartu grid beranda. Gunakan tombol ▲ ▼ untuk mengatur urutan.</p>
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
                <div class="item-actions">
                  <button class="btn-arrow" :disabled="idx === 0" @click="moveFeatured(idx, 'up')" title="Geser ke Atas">▲</button>
                  <button class="btn-arrow" :disabled="idx === featuredList.length - 1" @click="moveFeatured(idx, 'down')" title="Geser ke Bawah">▼</button>
                  <button class="btn-remove" @click="removeFeatured(idx)">Hapus</button>
                </div>
              </li>
            </ul>
          </div>

          <!-- Content Picker Pool -->
          <div class="card pool-card">
            <div class="pool-header">
              <h3 class="card-title">Pilih dari Artikel Terbit</h3>
              <span class="badge-filter-status">Hanya Artikel Terbit</span>
            </div>
            <p class="card-sub">Menampilkan artikel yang sudah berstatus <strong>Terbit (Published)</strong> agar valid disimpan ke beranda.</p>

            <div class="search-box">
              <input
                v-model="searchQuery"
                type="text"
                class="form-input"
                placeholder="Cari artikel atau kategori..."
              />
            </div>

            <div v-if="filteredContents.length === 0" class="empty-pool">
              <BladeIcon name="information-circle" size="22" class="empty-icon" />
              <span>Tidak ada artikel berstatus Terbit yang cocok dengan pencarian.</span>
            </div>

            <ul v-else class="pool-items">
              <li v-for="c in filteredContents" :key="c.id" class="pool-item">
                <div class="pool-info">
                  <div class="pool-title-row">
                    <span class="pool-title">
                      {{ c.translations?.id?.title || c.translations?.en?.title || 'Tanpa Judul' }}
                    </span>
                    <span class="pill-terbit">Terbit</span>
                  </div>
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
            @click="switchLocale('id')"
          >
            <BladeIcon name="language" size="16" />
            <span>Cerita Pidie (Bahasa Indonesia)</span>
          </button>
          <button
            class="sub-tab"
            :class="{ active: activeLocale === 'en' }"
            @click="switchLocale('en')"
          >
            <BladeIcon name="globe-alt" size="16" />
            <span>Pidie Stories (English)</span>
          </button>
        </div>

        <div class="curation-grid">
          <div class="card cur-card">
            <div class="card-header">
              <div>
                <h3 class="card-title">Cerita Utama & Pendukung ({{ activeLocale.toUpperCase() }})</h3>
                <p class="card-sub">Tentukan 1 Cerita Utama (Lead Story) berukuran besar dan beberapa cerita pendukung di sampingnya.</p>
              </div>
              <button class="btn btn-primary" :disabled="isSaving" @click="saveStories">
                {{ isSaving ? 'Menyimpan...' : 'Simpan Cerita' }}
              </button>
            </div>

            <!-- Lead Story Box -->
            <div class="lead-box">
              <span class="box-badge">CERITA UTAMA (LEAD STORY)</span>
              <div v-if="leadContentId" class="lead-selected">
                <div class="item-info">
                  <span class="item-title">
                    {{ availableContents.find((c) => c.id === leadContentId)?.translations?.[activeLocale]?.title || availableContents.find((c) => c.id === leadContentId)?.translations?.id?.title || 'Artikel #' + leadContentId }}
                  </span>
                  <span class="item-cat">
                    {{ availableContents.find((c) => c.id === leadContentId)?.categoryName || 'Umum' }}
                  </span>
                </div>
                <button class="btn-remove" @click="leadContentId = null">Hapus Lead</button>
              </div>
              <div v-else class="lead-empty">
                Belum ada Cerita Utama yang dipilih. Klik tombol "+ Jadikan Lead" pada artikel di sebelah kanan.
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
                <div class="item-actions">
                  <button class="btn-arrow" :disabled="idx === 0" @click="moveSupporting(idx, 'up')" title="Geser ke Atas">▲</button>
                  <button class="btn-arrow" :disabled="idx === supportingList.length - 1" @click="moveSupporting(idx, 'down')" title="Geser ke Bawah">▼</button>
                  <button class="btn-remove" @click="removeSupporting(idx)">Hapus</button>
                </div>
              </li>
            </ul>
          </div>

          <!-- Pool for Stories -->
          <div class="card pool-card">
            <div class="pool-header">
              <h3 class="card-title">Artikel Terbit ({{ activeLocale.toUpperCase() }})</h3>
              <span class="badge-filter-status">Terbit: {{ activeLocale === 'id' ? 'Indonesia' : 'English' }}</span>
            </div>
            <p class="card-sub">Hanya artikel yang berstatus <strong>Terbit (Published)</strong> dalam bahasa {{ activeLocale === 'id' ? 'Indonesia' : 'Inggris' }} yang dapat dijadikan cerita beranda.</p>

            <div class="search-box">
              <input
                v-model="searchQuery"
                type="text"
                class="form-input"
                placeholder="Cari artikel atau kategori..."
              />
            </div>

            <div v-if="filteredContents.length === 0" class="empty-pool">
              <BladeIcon name="information-circle" size="22" class="empty-icon" />
              <span>Tidak ada artikel terbit dalam bahasa {{ activeLocale === 'id' ? 'Indonesia' : 'Inggris' }}. Terbitkan artikel terlebih dahulu di menu Kelola Konten.</span>
            </div>

            <ul v-else class="pool-items">
              <li v-for="c in filteredContents" :key="c.id" class="pool-item">
                <div class="pool-info">
                  <div class="pool-title-row">
                    <span class="pool-title">{{ c.translations?.[activeLocale]?.title || c.translations?.id?.title || 'Tanpa Judul' }}</span>
                    <span class="pill-terbit">Terbit ({{ activeLocale.toUpperCase() }})</span>
                  </div>
                  <span class="pool-cat">{{ c.categoryName }}</span>
                </div>
                <div class="pool-actions">
                  <button class="btn-xs btn-primary-outline" @click="setLeadStory(c)">+ Lead</button>
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
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1.25rem;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9rem;
  color: #475569;
  transition: all 0.15s ease;
}

.tab-btn.active {
  background: #059669;
  color: #ffffff;
  border-color: #059669;
}

.sub-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  color: #475569;
  transition: all 0.15s ease;
}

.sub-tab.active {
  background: #064e3b;
  color: #ffffff;
  border-color: #064e3b;
}

.curation-grid {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 1.5rem;
  align-items: start;
}

@media (max-width: 900px) {
  .curation-grid {
    grid-template-columns: 1fr;
  }
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

.pool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.25rem;
}

.badge-filter-status {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.2rem 0.55rem;
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
  border-radius: 9999px;
  white-space: nowrap;
}

.pool-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pill-terbit {
  font-size: 0.62rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  background: #dcfce7;
  color: #15803d;
  border: 1px solid #bbf7d0;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.025em;
  flex-shrink: 0;
  line-height: 1.2;
}

.empty-pool {
  padding: 1.75rem 1rem;
  text-align: center;
  color: #64748b;
  font-size: 0.85rem;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px dashed #cbd5e1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  line-height: 1.4;
}

.empty-icon {
  color: #94a3b8;
}

.card-title { margin: 0 0 0.25rem 0; font-size: 1.05rem; font-weight: 700; color: #0f172a; }
.card-sub { margin: 0; font-size: 0.8rem; color: #64748b; line-height: 1.4; }

.search-box {
  margin: 0.75rem 0 1rem;
}

.form-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #059669;
}

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
  gap: 0.6rem;
}

.item-pos {
  font-weight: 800;
  color: #059669;
  font-size: 0.95rem;
  min-width: 24px;
}

.item-info, .pool-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.item-title, .pool-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-cat, .pool-cat {
  font-size: 0.75rem;
  color: #047857;
  font-weight: 600;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.btn-arrow {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  padding: 0.2rem 0.45rem;
  font-size: 0.75rem;
  color: #334155;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-arrow:hover:not(:disabled) {
  background: #059669;
  color: #ffffff;
  border-color: #059669;
}

.btn-arrow:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-remove {
  background: none;
  border: none;
  color: #ef4444;
  font-size: 0.775rem;
  font-weight: 600;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
}

.btn-remove:hover {
  text-decoration: underline;
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
  letter-spacing: 0.04em;
}

.lead-selected {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
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
  max-height: 640px;
  overflow-y: auto;
}

.pool-actions { display: flex; align-items: center; }
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
  transition: background 0.15s;
}

.btn-primary { background: #059669; color: #ffffff; }
.btn-primary:hover:not(:disabled) { background: #047857; }
.btn-outline { border: 1px solid #cbd5e1; background: #ffffff; color: #334155; cursor: pointer; }
.btn-outline:hover { background: #f8fafc; }
.btn-primary-outline { border: 1px solid #059669; background: #ecfdf5; color: #059669; cursor: pointer; font-weight: 600; }
.btn-primary-outline:hover { background: #059669; color: #ffffff; }

.btn-sm { padding: 0.35rem 0.65rem; font-size: 0.775rem; border-radius: 4px; }
.btn-xs { padding: 0.25rem 0.5rem; font-size: 0.725rem; border-radius: 4px; }

.loading-state, .empty-list {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
