<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'
import BladeIcon from '../../components/common/BladeIcon.vue'

const notification = useNotificationStore()

interface BannerItem {
  id: number
  title: string
  placement: string
  banner_type: 'image' | 'adsense'
  image_url: string
  target_url: string
  html_code: string
  open_in_new_tab: boolean
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

const banners = ref<BannerItem[]>([])
const isLoading = ref(true)
const isSubmitting = ref(false)
const selectedPlacement = ref<string>('all')
const searchQuery = ref<string>('')

// Placements definition
const placements = [
  { key: 'all', label: 'Semua Penempatan', icon: 'squares-2x2', desc: 'Seluruh banner dan slot iklan aktif' },
  { key: 'header_top', label: 'Header Atas', icon: 'arrow-up-circle', desc: 'Di bagian paling atas halaman portal (bar pengumuman/leaderboard)' },
  { key: 'below_hero', label: 'Di Bawah Hero', icon: 'view-columns', desc: 'Tepat di bawah kotak pengantar Hero di beranda utama' },
  { key: 'content_sidebar', label: 'Sidebar Konten', icon: 'bars-3-bottom-left', desc: 'Di kolom samping kurasi berita & artikel' },
  { key: 'footer_top', label: 'Di Atas Footer', icon: 'arrow-down-circle', desc: 'Di atas footer di seluruh halaman portal' },
  { key: 'popup_promo', label: 'Pop-up Promo', icon: 'sparkles', desc: 'Slot promosi pop-up interaktif' },
]

function getPlacementLabel(key: string): string {
  const found = placements.find((p) => p.key === key)
  return found ? found.label : key
}

function getPlacementBadgeClass(key: string): string {
  switch (key) {
    case 'header_top':
      return 'badge-blue'
    case 'below_hero':
      return 'badge-emerald'
    case 'content_sidebar':
      return 'badge-purple'
    case 'footer_top':
      return 'badge-amber'
    case 'popup_promo':
      return 'badge-rose'
    default:
      return 'badge-slate'
  }
}

// Modal state
const showModal = ref(false)
const isEditing = ref(false)
const currentId = ref<number | null>(null)

// Delete confirm modal
const showDeleteModal = ref(false)
const itemToDelete = ref<BannerItem | null>(null)

const form = reactive({
  title: '',
  placement: 'below_hero',
  banner_type: 'image' as 'image' | 'adsense',
  image_url: '',
  target_url: '',
  html_code: '',
  open_in_new_tab: true,
  is_active: true,
  sort_order: 1,
})

const filteredBanners = computed(() => {
  return banners.value.filter((b) => {
    const matchPlacement =
      selectedPlacement.value === 'all' || b.placement === selectedPlacement.value
    const q = searchQuery.value.toLowerCase().trim()
    const matchQuery =
      !q ||
      b.title.toLowerCase().includes(q) ||
      b.target_url.toLowerCase().includes(q) ||
      b.placement.toLowerCase().includes(q)
    return matchPlacement && matchQuery
  })
})

async function loadBanners() {
  isLoading.value = true
  try {
    const res = await api.get<{ data: BannerItem[] }>('/api/v1/admin/banners')
    if (res?.data) {
      banners.value = res.data
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Banner', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function openCreateModal(defaultPlacement?: string) {
  isEditing.value = false
  currentId.value = null
  form.title = ''
  form.placement = defaultPlacement && defaultPlacement !== 'all' ? defaultPlacement : 'below_hero'
  form.banner_type = 'image'
  form.image_url = ''
  form.target_url = ''
  form.html_code = ''
  form.open_in_new_tab = true
  form.is_active = true
  form.sort_order = banners.value.length + 1
  showModal.value = true
}

function openEditModal(item: BannerItem) {
  isEditing.value = true
  currentId.value = item.id
  form.title = item.title
  form.placement = item.placement
  form.banner_type = item.banner_type
  form.image_url = item.image_url
  form.target_url = item.target_url
  form.html_code = item.html_code
  form.open_in_new_tab = item.open_in_new_tab
  form.is_active = item.is_active
  form.sort_order = item.sort_order
  showModal.value = true
}

async function handleSubmit() {
  if (!form.title.trim()) {
    alert('Judul atau label banner wajib diisi.')
    return
  }

  if (form.banner_type === 'image' && !form.image_url.trim()) {
    alert('URL Gambar wajib diisi untuk banner gambar.')
    return
  }

  if (form.banner_type === 'adsense' && !form.html_code.trim()) {
    alert('Kode AdSense atau script HTML wajib diisi.')
    return
  }

  isSubmitting.value = true
  try {
    const payload = {
      title: form.title.trim(),
      placement: form.placement,
      banner_type: form.banner_type,
      image_url: form.image_url.trim(),
      target_url: form.target_url.trim(),
      html_code: form.html_code.trim(),
      open_in_new_tab: form.open_in_new_tab,
      is_active: form.is_active,
      sort_order: Number(form.sort_order) || 0,
    }

    if (isEditing.value && currentId.value) {
      await api.put(`/api/v1/admin/banners/${currentId.value}`, payload)
      notification.success('Berhasil Diperbarui', 'Banner dan tata letak berhasil diperbarui.')
    } else {
      await api.post('/api/v1/admin/banners', payload)
      notification.success('Berhasil Ditambahkan', 'Banner baru berhasil ditambahkan dan disinkronkan.')
    }

    showModal.value = false
    await loadBanners()
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan validasi.', err.requestId)
  } finally {
    isSubmitting.value = false
  }
}

async function handleToggleActive(item: BannerItem) {
  const newStatus = !item.is_active
  const prev = item.is_active
  item.is_active = newStatus // optimistic update

  try {
    await api.patch(`/api/v1/admin/banners/${item.id}/toggle`, { is_active: newStatus })
    notification.success(
      'Status Diperbarui',
      `Banner "${item.title}" sekarang ${newStatus ? 'Aktif' : 'Nonaktif'}.`
    )
  } catch (err: any) {
    item.is_active = prev // rollback
    notification.error('Gagal Mengubah Status', err.detail || 'Terjadi kesalahan.', err.requestId)
  }
}

function confirmDelete(item: BannerItem) {
  itemToDelete.value = item
  showDeleteModal.value = true
}

async function handleDelete() {
  if (!itemToDelete.value) return
  try {
    await api.delete(`/api/v1/admin/banners/${itemToDelete.value.id}`)
    notification.success('Banner Dihapus', `Banner "${itemToDelete.value.title}" telah dihapus.`)
    showDeleteModal.value = false
    itemToDelete.value = null
    await loadBanners()
  } catch (err: any) {
    notification.error('Gagal Menghapus', err.detail || 'Terjadi kesalahan.', err.requestId)
  }
}

onMounted(() => {
  loadBanners()
})
</script>

<template>
  <AppLayout>
    <div class="layout-page">
      <!-- Page Header -->
      <div class="page-header">
        <div>
          <h2 class="page-title">Tata Letak & Banner Promosi</h2>
          <p class="page-subtitle">
            Atur banner promosi, sponsor wisata daerah, pengumuman, dan kode Google AdSense pada tata letak portal publik.
          </p>
        </div>
        <div class="header-actions">
          <button class="btn btn-primary" @click="openCreateModal(selectedPlacement)">
            <BladeIcon name="plus" size="18" />
            <span>Tambah Banner / AdSense</span>
          </button>
        </div>
      </div>

      <!-- Placement Cards Overview -->
      <div class="placement-cards-grid">
        <div
          v-for="p in placements"
          :key="p.key"
          class="placement-card"
          :class="{ active: selectedPlacement === p.key }"
          @click="selectedPlacement = p.key"
        >
          <div class="placement-icon-box" :class="getPlacementBadgeClass(p.key)">
            <BladeIcon :name="p.icon" size="20" />
          </div>
          <div class="placement-info">
            <div class="placement-title-row">
              <span class="placement-title">{{ p.label }}</span>
              <span class="placement-count">
                {{ p.key === 'all' ? banners.length : banners.filter((b) => b.placement === p.key).length }}
              </span>
            </div>
            <p class="placement-desc">{{ p.desc }}</p>
          </div>
        </div>
      </div>

      <!-- Content Section -->
      <div class="content-card">
        <!-- Search & Filter Bar -->
        <div class="toolbar">
          <div class="search-input-wrapper">
            <BladeIcon name="magnifying-glass" size="18" class="search-icon" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari banner berdasarkan judul atau URL tautan..."
              class="search-input"
            />
          </div>
          <div class="toolbar-stats">
            Menampilkan <strong>{{ filteredBanners.length }}</strong> item
          </div>
        </div>

        <!-- Table Loading -->
        <div v-if="isLoading" class="loading-state">
          <div class="spinner"></div>
          <p>Memuat konfigurasi tata letak dan banner...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredBanners.length === 0" class="empty-state">
          <BladeIcon name="rectangle-group" size="48" class="empty-icon" />
          <h3>Belum Ada Banner di Penempatan Ini</h3>
          <p>Tambahkan banner promosi atau slot kode AdSense untuk mengisi ruang promosi portal Pidie.</p>
          <button class="btn btn-outline" @click="openCreateModal(selectedPlacement)">
            <BladeIcon name="plus" size="16" />
            <span>Tambah Banner Sekarang</span>
          </button>
        </div>

        <!-- Data Table -->
        <div v-else class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 70px;">Pratinjau</th>
                <th>Judul & Tipe</th>
                <th>Penempatan</th>
                <th>Tautan / Kode</th>
                <th style="width: 80px; text-align: center;">Urutan</th>
                <th style="width: 110px; text-align: center;">Status</th>
                <th style="width: 120px; text-align: right;">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in filteredBanners" :key="b.id">
                <!-- Preview Thumbnail / Type Icon -->
                <td>
                  <div class="preview-thumbnail">
                    <img
                      v-if="b.banner_type === 'image' && b.image_url"
                      :src="b.image_url"
                      :alt="b.title"
                      class="thumb-img"
                      @error="(e: any) => (e.target.style.display = 'none')"
                    />
                    <div v-else class="code-thumb">
                      <BladeIcon name="code-bracket" size="20" />
                    </div>
                  </div>
                </td>

                <!-- Title & Type -->
                <td>
                  <div class="title-cell">
                    <span class="banner-title">{{ b.title }}</span>
                    <span class="type-pill" :class="b.banner_type === 'image' ? 'type-image' : 'type-adsense'">
                      <BladeIcon :name="b.banner_type === 'image' ? 'photo' : 'code-bracket'" size="13" />
                      {{ b.banner_type === 'image' ? 'Banner Gambar' : 'Google AdSense / Script' }}
                    </span>
                  </div>
                </td>

                <!-- Placement Badge -->
                <td>
                  <span class="placement-badge" :class="getPlacementBadgeClass(b.placement)">
                    {{ getPlacementLabel(b.placement) }}
                  </span>
                </td>

                <!-- Target URL / Code Snippet -->
                <td>
                  <div v-if="b.banner_type === 'image'" class="url-cell">
                    <a
                      v-if="b.target_url"
                      :href="b.target_url"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="target-link"
                      title="Buka Tautan"
                    >
                      <BladeIcon name="arrow-top-right-on-square" size="14" />
                      <span class="url-text">{{ b.target_url }}</span>
                    </a>
                    <span v-else class="text-muted">Tanpa Tautan (Hanya Tampilan)</span>
                  </div>
                  <div v-else class="code-cell">
                    <code class="code-snippet">{{ b.html_code.slice(0, 50) }}...</code>
                  </div>
                </td>

                <!-- Sort Order -->
                <td style="text-align: center;">
                  <span class="order-badge">#{{ b.sort_order }}</span>
                </td>

                <!-- Status Toggle -->
                <td style="text-align: center;">
                  <button
                    type="button"
                    class="status-toggle-btn"
                    :class="{ active: b.is_active }"
                    @click="handleToggleActive(b)"
                    :title="b.is_active ? 'Klik untuk nonaktifkan' : 'Klik untuk aktifkan'"
                  >
                    <span class="toggle-dot"></span>
                    <span class="toggle-text">{{ b.is_active ? 'Aktif' : 'Nonaktif' }}</span>
                  </button>
                </td>

                <!-- Actions -->
                <td style="text-align: right;">
                  <div class="actions-group">
                    <button class="btn-icon" title="Edit Banner" @click="openEditModal(b)">
                      <BladeIcon name="pencil" size="16" />
                    </button>
                    <button class="btn-icon btn-icon-danger" title="Hapus Banner" @click="confirmDelete(b)">
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

    <!-- Create / Edit Modal -->
    <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
      <div class="modal-dialog">
        <div class="modal-header">
          <h3 class="modal-title">{{ isEditing ? 'Edit Banner / Slot Iklan' : 'Tambah Banner / Slot Iklan Baru' }}</h3>
          <button class="close-btn" @click="showModal = false">&times;</button>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="modal-body">
            <!-- Judul Banner -->
            <div class="form-group">
              <label class="form-label required">Judul / Label Banner</label>
              <input
                v-model="form.title"
                type="text"
                class="form-control"
                placeholder="Contoh: Banner Festival Kuliner Melinjo 2026 / AdSense Header"
                required
              />
              <span class="form-hint">Label internal pengingat tujuan promosi atau ID slot sponsor.</span>
            </div>

            <!-- Penempatan Tata Letak -->
            <div class="form-group">
              <label class="form-label required">Penempatan Tata Letak</label>
              <select v-model="form.placement" class="form-control" required>
                <option value="header_top">Header Atas (Top Announcement / Leaderboard)</option>
                <option value="below_hero">Di Bawah Hero Beranda (Posisi Utama Paling Strategis)</option>
                <option value="content_sidebar">Sidebar Konten (Di samping artikel & kurasi)</option>
                <option value="footer_top">Di Atas Footer (Full-width sebelum footer portal)</option>
                <option value="popup_promo">Pop-up Promo (Promosi interaktif)</option>
              </select>
            </div>

            <!-- Tipe Konten: Banner Gambar vs Kode AdSense -->
            <div class="form-group">
              <label class="form-label required">Tipe Konten</label>
              <div class="type-toggle-group">
                <button
                  type="button"
                  class="type-toggle-option"
                  :class="{ active: form.banner_type === 'image' }"
                  @click="form.banner_type = 'image'"
                >
                  <BladeIcon name="photo" size="18" />
                  <div>
                    <div class="toggle-opt-title">Banner Gambar</div>
                    <div class="toggle-opt-desc">Gambar promosi, sponsor, atau poster acara</div>
                  </div>
                </button>
                <button
                  type="button"
                  class="type-toggle-option"
                  :class="{ active: form.banner_type === 'adsense' }"
                  @click="form.banner_type = 'adsense'"
                >
                  <BladeIcon name="code-bracket" size="18" />
                  <div>
                    <div class="toggle-opt-title">Google AdSense / Script HTML</div>
                    <div class="toggle-opt-desc">Kode iklan otomatis atau tag script kustom</div>
                  </div>
                </button>
              </div>
            </div>

            <!-- Fields for Image Banner -->
            <template v-if="form.banner_type === 'image'">
              <div class="form-group">
                <label class="form-label required">URL Berkas Gambar</label>
                <input
                  v-model="form.image_url"
                  type="text"
                  class="form-control"
                  placeholder="https://pidiekab.web.id/media/2026/11/... atau URL eksternal"
                  required
                />
                <span class="form-hint">Bisa menyalin tautan gambar dari menu Media Library. Rekomendasi lebar: 1200x200px (Hero) atau 300x250px (Sidebar).</span>
              </div>

              <!-- Live Preview if URL entered -->
              <div v-if="form.image_url" class="preview-box">
                <div class="preview-label">Pratinjau Gambar:</div>
                <img :src="form.image_url" alt="Pratinjau Banner" class="live-preview-img" />
              </div>

              <div class="form-group">
                <label class="form-label">Tautan Tujuan (Klik)</label>
                <input
                  v-model="form.target_url"
                  type="text"
                  class="form-control"
                  placeholder="https://... atau /id/kuliner/kerupuk-mulieng/ (kosongkan jika tanpa tautan)"
                />
              </div>

              <div class="form-checkbox-row">
                <label class="checkbox-label">
                  <input v-model="form.open_in_new_tab" type="checkbox" />
                  <span>Buka tautan di tab baru (target="_blank")</span>
                </label>
              </div>
            </template>

            <!-- Fields for AdSense / HTML Script -->
            <template v-else>
              <div class="form-group">
                <label class="form-label required">Kode Script Google AdSense / Tag HTML</label>
                <textarea
                  v-model="form.html_code"
                  rows="6"
                  class="form-control code-textarea"
                  placeholder='<script async src="https://pagead2.googlesyndication.com/..."></script>&#10;<ins class="adsbygoogle" ...></ins>&#10;<script>(adsbygoogle = window.adsbygoogle || []).push({});</script>'
                  required
                ></textarea>
                <span class="form-hint">Kode iklan akan dirender langsung di posisi penempatan yang dipilih.</span>
              </div>
            </template>

            <!-- Sort Order & Active Toggle -->
            <div class="form-row">
              <div class="form-group col-half">
                <label class="form-label">Urutan Tampilan (Sort Order)</label>
                <input v-model.number="form.sort_order" type="number" min="0" class="form-control" />
                <span class="form-hint">Angka lebih kecil tampil lebih awal.</span>
              </div>
              <div class="form-group col-half">
                <label class="form-label">Status Tayang</label>
                <div class="checkbox-box">
                  <label class="checkbox-label">
                    <input v-model="form.is_active" type="checkbox" />
                    <span>Aktifkan dan tampilkan sekarang</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-outline" @click="showModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
              <span v-if="isSubmitting" class="spinner-sm"></span>
              <span>{{ isEditing ? 'Simpan Perubahan' : 'Tambahkan Banner' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-backdrop" @click.self="showDeleteModal = false">
      <div class="modal-dialog modal-sm">
        <div class="modal-header">
          <h3 class="modal-title">Konfirmasi Hapus Banner</h3>
          <button class="close-btn" @click="showDeleteModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <p>
            Apakah Anda yakin ingin menghapus banner <strong>"{{ itemToDelete?.title }}"</strong>?
          </p>
          <p class="text-muted" style="font-size: 0.875rem;">
            Banner ini akan dihapus dari tata letak dan situs publik akan diperbarui secara otomatis.
          </p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-outline" @click="showDeleteModal = false">Batal</button>
          <button type="button" class="btn btn-danger" @click="handleDelete">Hapus Sekarang</button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.layout-page {
  padding: 1.5rem;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  gap: 1rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

.page-subtitle {
  color: #64748b;
  font-size: 0.9375rem;
  margin: 0;
}

/* Placement Overview Cards */
.placement-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.placement-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem;
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.placement-card:hover {
  border-color: #cbd5e1;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.placement-card.active {
  border-color: #059669;
  background: #f0fdf4;
  box-shadow: 0 2px 8px rgba(5, 150, 105, 0.12);
}

.placement-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.placement-info {
  flex: 1;
  min-width: 0;
}

.placement-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.2rem;
}

.placement-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #1e293b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.placement-count {
  font-size: 0.75rem;
  font-weight: 700;
  background: #e2e8f0;
  color: #475569;
  padding: 0.15rem 0.45rem;
  border-radius: 9999px;
}

.placement-card.active .placement-count {
  background: #059669;
  color: #ffffff;
}

.placement-desc {
  font-size: 0.75rem;
  color: #64748b;
  margin: 0;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Badges */
.badge-blue { background: #eff6ff; color: #2563eb; }
.badge-emerald { background: #ecfdf5; color: #059669; }
.badge-purple { background: #faf5ff; color: #9333ea; }
.badge-amber { background: #fffbeb; color: #d97706; }
.badge-rose { background: #fff1f2; color: #e11d48; }
.badge-slate { background: #f8fafc; color: #475569; }

/* Content Card */
.content-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.toolbar {
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e2e8f0;
  gap: 1rem;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  max-width: 450px;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
}

.search-input {
  width: 100%;
  padding: 0.5rem 0.75rem 0.5rem 2.25rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.15s;
}

.search-input:focus {
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}

.toolbar-stats {
  font-size: 0.875rem;
  color: #64748b;
}

/* Table */
.table-responsive {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  text-align: left;
}

.data-table th {
  background: #f8fafc;
  color: #475569;
  font-weight: 600;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}

.data-table tr:hover td {
  background: #f8fafc;
}

/* Cell Elements */
.preview-thumbnail {
  width: 60px;
  height: 40px;
  border-radius: 6px;
  overflow: hidden;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
}

.thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.code-thumb {
  color: #64748b;
}

.title-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.banner-title {
  font-weight: 600;
  color: #0f172a;
}

.type-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-weight: 600;
  width: fit-content;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}

.type-image {
  background: #ecfdf5;
  color: #059669;
}

.type-adsense {
  background: #eff6ff;
  color: #2563eb;
}

.placement-badge {
  display: inline-block;
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
}

.url-cell {
  max-width: 250px;
}

.target-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: #059669;
  text-decoration: none;
  font-size: 0.8125rem;
}

.target-link:hover {
  text-decoration: underline;
}

.url-text {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.code-snippet {
  background: #f1f5f9;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  font-size: 0.75rem;
  color: #334155;
  font-family: monospace;
}

.order-badge {
  font-weight: 700;
  color: #64748b;
  font-size: 0.8125rem;
}

/* Status Switch Button */
.status-toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 9999px;
  padding: 0.25rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s ease;
}

.status-toggle-btn .toggle-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #94a3b8;
}

.status-toggle-btn.active {
  background: #ecfdf5;
  border-color: #a7f3d0;
  color: #047857;
}

.status-toggle-btn.active .toggle-dot {
  background: #10b981;
}

/* Action Buttons */
.actions-group {
  display: flex;
  justify-content: flex-end;
  gap: 0.4rem;
}

.btn-icon {
  background: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.4rem;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.btn-icon:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.btn-icon-danger:hover {
  background: #fef2f2;
  border-color: #fecaca;
  color: #dc2626;
}

/* Empty & Loading States */
.loading-state,
.empty-state {
  padding: 4rem 2rem;
  text-align: center;
  color: #64748b;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e2e8f0;
  border-top-color: #059669;
  border-radius: 50%;
  margin: 0 auto 1rem;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #ffffff;
  border-radius: 50%;
  display: inline-block;
  vertical-align: middle;
  animation: spin 0.8s linear infinite;
  margin-right: 0.4rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-icon {
  color: #cbd5e1;
  margin-bottom: 1rem;
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 0.5rem;
}

.empty-state p {
  max-width: 420px;
  margin: 0 auto 1.5rem;
  font-size: 0.875rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.55rem 1rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  border: 1px solid transparent;
}

.btn-primary {
  background: #059669;
  color: #ffffff;
}

.btn-primary:hover {
  background: #047857;
}

.btn-outline {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #334155;
}

.btn-outline:hover {
  background: #f8fafc;
  border-color: #94a3b8;
}

.btn-danger {
  background: #dc2626;
  color: #ffffff;
}

.btn-danger:hover {
  background: #b91c1c;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1.5rem;
}

.modal-dialog {
  background: #ffffff;
  border-radius: 14px;
  width: 100%;
  max-width: 650px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.2);
}

.modal-sm {
  max-width: 440px;
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  font-size: 1.5rem;
  color: #94a3b8;
  cursor: pointer;
  line-height: 1;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  background: #f8fafc;
  border-radius: 0 0 14px 14px;
}

/* Forms */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.form-label.required::after {
  content: ' *';
  color: #dc2626;
}

.form-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.15s;
  font-family: inherit;
}

.form-control:focus {
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}

.code-textarea {
  font-family: monospace;
  font-size: 0.8125rem;
  line-height: 1.5;
  background: #f8fafc;
}

.form-hint {
  font-size: 0.75rem;
  color: #64748b;
}

.form-row {
  display: flex;
  gap: 1rem;
}

.col-half {
  flex: 1;
}

/* Type Toggle Buttons */
.type-toggle-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.type-toggle-option {
  border: 1px solid #cbd5e1;
  background: #ffffff;
  border-radius: 10px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.type-toggle-option:hover {
  border-color: #94a3b8;
  background: #f8fafc;
}

.type-toggle-option.active {
  border-color: #059669;
  background: #f0fdf4;
  box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.2);
}

.toggle-opt-title {
  font-size: 0.875rem;
  font-weight: 700;
  color: #1e293b;
}

.toggle-opt-desc {
  font-size: 0.75rem;
  color: #64748b;
}

/* Image Preview in Modal */
.preview-box {
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  padding: 0.75rem;
  text-align: center;
}

.preview-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 0.5rem;
  text-align: left;
}

.live-preview-img {
  max-width: 100%;
  max-height: 160px;
  border-radius: 6px;
  display: inline-block;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

.checkbox-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: #334155;
  cursor: pointer;
}

.checkbox-box {
  padding-top: 0.5rem;
}
</style>
