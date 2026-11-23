<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useNotificationStore } from '../../stores/notification'
import { api } from '../../api/client'
import type { ModuleInfo } from '../../types'
import AppLayout from '../../components/layout/AppLayout.vue'
import BladeIcon from '../../components/common/BladeIcon.vue'

const authStore = useAuthStore()
const notification = useNotificationStore()

// Modules definition matching domain.AvailableModules (exact 13 modules)
const standardModules: ModuleInfo[] = [
  {
    key: 'dashboard',
    name: 'Dashboard',
    path: '/dashboard',
    icon: 'chart-bar',
    description: 'Melihat statistik pengunjung, total artikel, dan ringkasan operasional',
  },
  {
    key: 'contents',
    name: 'Kelola Konten',
    path: '/contents',
    icon: 'document-text',
    description: 'Membuat, mengedit, dan mempublikasikan artikel berita, budaya, dan potensi daerah',
  },
  {
    key: 'categories',
    name: 'Kategori',
    path: '/categories',
    icon: 'tag',
    description: 'Mengelola taksonomi kategori konten (nama, slug bilingual, status aktif)',
  },
  {
    key: 'media',
    name: 'Media Library',
    path: '/media',
    icon: 'photo',
    description: 'Mengunggah dan mengelola berkas gambar, foto ilustrasi, dan dokumen media',
  },
  {
    key: 'curation',
    name: 'Kurasi Beranda',
    path: '/curation',
    icon: 'sparkles',
    description: 'Mengatur penempatan berita utama (Lead Story) dan konten unggulan di beranda publik',
  },
  {
    key: 'static-pages',
    name: 'Halaman Statis',
    path: '/static-pages',
    icon: 'document-duplicate',
    description: 'Mengelola konten halaman Profil Daerah, Kontak, Kebijakan Privasi, dan Syarat Ketentuan',
  },
  {
    key: 'menus',
    name: 'Menu Navigasi',
    path: '/menus',
    icon: 'bars-3',
    description: 'Mengatur susunan dan tautan navigasi header portal publik',
  },
  {
    key: 'unduhan',
    name: 'Unduhan Dokumen',
    path: '/unduhan',
    icon: 'arrow-down-tray',
    description: 'Mengunggah dan mengelola berkas panduan publik, peta wisata, dan dokumen unduhan',
  },
  {
    key: 'tata-letak',
    name: 'Tata Letak & Banner',
    path: '/tata-letak',
    icon: 'rectangle-group',
    description: 'Mengatur banner promosi, sponsor, dan slot iklan AdSense pada tata letak situs',
  },
  {
    key: 'settings',
    name: 'Pengaturan Web',
    path: '/settings',
    icon: 'cog-6-tooth',
    description: 'Mengatur identitas website, meta SEO, kontak resmi, dan konfigurasi portal',
  },
  {
    key: 'redirects',
    name: 'URL Redirects',
    path: '/redirects',
    icon: 'arrows-right-left',
    description: 'Mengelola pengalihan URL (301/302) untuk pencegahan tautan rusak',
  },
  {
    key: 'admins',
    name: 'Administrator',
    path: '/admins',
    icon: 'users',
    description: 'Mengelola akun pengguna administrator dan konfigurasi hak akses menu',
  },
  {
    key: 'audit-logs',
    name: 'Audit Trail',
    path: '/audit-logs',
    icon: 'shield-check',
    description: 'Melihat riwayat log aktivitas keamanan dan jejak perubahan sistem',
  },
]

const availableModules = ref<ModuleInfo[]>(standardModules)

// Profile Form
const profileForm = reactive({
  displayName: '',
})
const isUpdatingProfile = ref(false)

// Password Form
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})
const isUpdatingPassword = ref(false)

function formatDateTime(val?: string): string {
  if (!val) return 'Belum ada data'
  try {
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    return (
      d.toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }) + ' WIB'
    )
  } catch {
    return val
  }
}

function formatDate(val?: string): string {
  if (!val) return 'Belum ada data'
  try {
    const d = new Date(val)
    if (isNaN(d.getTime())) return val
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
  } catch {
    return val
  }
}

// Check if user has access to a specific module
function hasAccessTo(key: string): boolean {
  return authStore.canAccess(key)
}

const accessibleCount = computed(() => {
  return availableModules.value.filter((m) => hasAccessTo(m.key)).length
})

async function fetchModules() {
  try {
    const res = await api.get<{ modules: ModuleInfo[] }>('/api/v1/admin/modules')
    if (res?.modules && Array.isArray(res.modules)) {
      availableModules.value = res.modules
    }
  } catch {
    availableModules.value = standardModules
  }
}

async function handleUpdateProfile() {
  if (!profileForm.displayName.trim()) {
    alert('Nama tampilan tidak boleh kosong.')
    return
  }

  isUpdatingProfile.value = true
  try {
    await api.put('/api/v1/admin/auth/profile', {
      display_name: profileForm.displayName.trim(),
    })
    notification.success('Profil Diperbarui', 'Nama tampilan akun Anda berhasil disimpan.')
    await authStore.checkSession()
  } catch (err: any) {
    notification.error('Gagal Memperbarui Profil', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isUpdatingProfile.value = false
  }
}

async function handleUpdatePassword() {
  if (!passwordForm.currentPassword) {
    alert('Masukkan kata sandi saat ini.')
    return
  }

  if (passwordForm.newPassword.length < 8) {
    alert('Kata sandi baru minimal 8 karakter.')
    return
  }

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert('Konfirmasi kata sandi baru tidak cocok.')
    return
  }

  isUpdatingPassword.value = true
  try {
    await api.post('/api/v1/admin/auth/password/change', {
      current_password: passwordForm.currentPassword,
      new_password: passwordForm.newPassword,
    })
    notification.success('Kata Sandi Diperbarui', 'Kata sandi akun Anda berhasil diperbarui.')
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (err: any) {
    notification.error('Gagal Mengubah Kata Sandi', err.detail || 'Kata sandi saat ini tidak cocok atau tidak valid.', err.requestId)
  } finally {
    isUpdatingPassword.value = false
  }
}

onMounted(async () => {
  await authStore.checkSession()
  if (authStore.user?.displayName) {
    profileForm.displayName = authStore.user.displayName
  }
  fetchModules()
})
</script>

<template>
  <AppLayout>
    <div class="profile-page">
      <!-- Profile Hero Card -->
      <div class="profile-hero-card">
        <div class="hero-left">
          <div class="user-avatar-large">
            {{ authStore.user?.displayName?.charAt(0) || 'A' }}
          </div>
          <div class="user-info">
            <div class="name-row">
              <h2 class="user-name">{{ authStore.user?.displayName || 'Administrator' }}</h2>
              <span class="role-badge" :class="authStore.roleBadgeClass">
                {{ authStore.roleLabel }}
              </span>
              <span class="status-pill active">
                <span class="status-dot"></span>
                <span>Akun Aktif</span>
              </span>
            </div>
            <p class="user-email">
              <BladeIcon name="envelope" size="16" />
              <span>{{ authStore.user?.email || 'admin@pidiekab.id' }}</span>
            </p>
          </div>
        </div>

        <div class="hero-stats">
          <div class="stat-box">
            <span class="stat-label">Hak Akses Modul</span>
            <span class="stat-value">{{ accessibleCount }} / {{ availableModules.length }}</span>
          </div>
          <div class="stat-box">
            <span class="stat-label">Akun Terdaftar Sejak</span>
            <span class="stat-value-sub">{{ formatDate(authStore.user?.createdAt) }}</span>
          </div>
          <div class="stat-box">
            <span class="stat-label">Waktu Masuk Terakhir</span>
            <span class="stat-value-sub">{{ formatDateTime(authStore.user?.lastLoginAt) }}</span>
          </div>
        </div>
      </div>

      <!-- Main Columns -->
      <div class="profile-content-grid">
        <!-- Left: RBAC Permissions Card -->
        <div class="permissions-card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Hak Akses Modul Saya (RBAC)</h3>
              <p class="card-subtitle">
                Daftar menu dan hak otorisasi yang dapat diakses oleh akun Anda pada sistem PidieKab.
              </p>
            </div>
            <span class="counter-pill">
              {{ accessibleCount }} Menu Diizinkan
            </span>
          </div>

          <!-- Super Admin Notice -->
          <div v-if="authStore.isSuperAdmin" class="super-admin-banner">
            <BladeIcon name="shield-check" size="22" class="shield-icon" />
            <div>
              <strong class="banner-title">Hak Akses Penuh (Super Administrator)</strong>
              <p class="banner-desc">
                Sebagai pemilik hak akses tertinggi, akun Anda memiliki otorisasi penuh untuk mengakses seluruh menu, mengatur hak akses administrator lain, dan konfigurasi sistem.
              </p>
            </div>
          </div>

          <div v-else class="role-access-banner">
            <BladeIcon name="key" size="20" class="key-icon" />
            <div>
              <strong class="banner-title">Otorisasi Berdasarkan Peran & Checklist</strong>
              <p class="banner-desc">
                Hak akses Anda telah dikonfigurasi berdasarkan peran <strong>{{ authStore.roleLabel }}</strong> dan checklist otorisasi yang diberikan oleh Super Administrator.
              </p>
            </div>
          </div>

          <!-- Modules Checklist Grid -->
          <div class="modules-grid">
            <div
              v-for="m in availableModules"
              :key="m.key"
              class="module-item-card"
              :class="{ allowed: hasAccessTo(m.key), forbidden: !hasAccessTo(m.key) }"
            >
              <div class="module-item-top">
                <div class="module-icon-wrap" :class="{ 'icon-allowed': hasAccessTo(m.key) }">
                  <BladeIcon :name="m.icon" size="20" />
                </div>
                <span v-if="hasAccessTo(m.key)" class="access-pill allowed">
                  <BladeIcon name="check" size="13" />
                  <span>Diberikan Akses</span>
                </span>
                <span v-else class="access-pill forbidden">
                  <BladeIcon name="lock-closed" size="13" />
                  <span>Tidak Diberikan</span>
                </span>
              </div>

              <div class="module-item-body">
                <h4 class="module-name">{{ m.name }}</h4>
                <code class="module-path">{{ m.path }}</code>
                <p class="module-desc">{{ m.description }}</p>
              </div>

              <div v-if="hasAccessTo(m.key)" class="module-item-footer">
                <router-link :to="m.path" class="module-link">
                  <span>Buka Menu</span>
                  <BladeIcon name="arrow-right" size="13" />
                </router-link>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Forms Column (Account Info & Password Change) -->
        <div class="forms-column">
          <!-- Edit Profile Card -->
          <div class="settings-card">
            <div class="card-header">
              <h3 class="card-title">Informasi Akun</h3>
              <p class="card-subtitle">Ubah nama lengkap atau nama tampilan profil Anda.</p>
            </div>

            <form @submit.prevent="handleUpdateProfile" class="card-body">
              <div class="form-group">
                <label class="form-label">Alamat Email</label>
                <div class="input-with-icon">
                  <BladeIcon name="envelope" size="18" class="field-icon" />
                  <input
                    :value="authStore.user?.email"
                    type="email"
                    class="form-control"
                    disabled
                    title="Alamat email akun bersifat permanen dan dilindungi."
                  />
                </div>
                <span class="form-hint">Email bersifat unik dan terikat pada akun Anda.</span>
              </div>

              <div class="form-group">
                <label class="form-label required">Nama Tampilan / Nama Lengkap</label>
                <div class="input-with-icon">
                  <BladeIcon name="user" size="18" class="field-icon" />
                  <input
                    v-model="profileForm.displayName"
                    type="text"
                    class="form-control"
                    placeholder="Contoh: Raja Muchlisar"
                    required
                  />
                </div>
                <span class="form-hint">Nama ini ditampilkan di sudut kanan atas panel dan riwayat audit log.</span>
              </div>

              <div class="form-group">
                <label class="form-label">Peran Administrator (Role)</label>
                <input
                  :value="authStore.roleLabel"
                  type="text"
                  class="form-control"
                  disabled
                />
              </div>

              <div class="card-actions">
                <button type="submit" class="btn btn-primary" :disabled="isUpdatingProfile">
                  <span v-if="isUpdatingProfile" class="spinner-sm"></span>
                  <span>Simpan Perubahan Profil</span>
                </button>
              </div>
            </form>
          </div>

          <!-- Change Password Card -->
          <div class="settings-card">
            <div class="card-header">
              <h3 class="card-title">Keamanan & Kata Sandi</h3>
              <p class="card-subtitle">Perbarui kata sandi akun Anda secara berkala untuk menjaga keamanan sistem.</p>
            </div>

            <form @submit.prevent="handleUpdatePassword" class="card-body">
              <div class="form-group">
                <label class="form-label required">Kata Sandi Saat Ini</label>
                <div class="input-with-icon">
                  <BladeIcon name="key" size="18" class="field-icon" />
                  <input
                    v-model="passwordForm.currentPassword"
                    type="password"
                    class="form-control"
                    placeholder="Masukkan kata sandi lama Anda"
                    required
                  />
                </div>
              </div>

              <div class="form-group">
                <label class="form-label required">Kata Sandi Baru</label>
                <div class="input-with-icon">
                  <BladeIcon name="lock-closed" size="18" class="field-icon" />
                  <input
                    v-model="passwordForm.newPassword"
                    type="password"
                    class="form-control"
                    placeholder="Minimal 8 karakter aman"
                    required
                    minlength="8"
                  />
                </div>
              </div>

              <div class="form-group">
                <label class="form-label required">Konfirmasi Kata Sandi Baru</label>
                <div class="input-with-icon">
                  <BladeIcon name="shield-check" size="18" class="field-icon" />
                  <input
                    v-model="passwordForm.confirmPassword"
                    type="password"
                    class="form-control"
                    placeholder="Ulangi kata sandi baru"
                    required
                    minlength="8"
                  />
                </div>
              </div>

              <div class="card-actions">
                <button type="submit" class="btn btn-primary" :disabled="isUpdatingPassword">
                  <span v-if="isUpdatingPassword" class="spinner-sm"></span>
                  <span>Perbarui Kata Sandi</span>
                </button>
              </div>
            </form>
          </div>

          <!-- Security Badge Card -->
          <div class="security-info-box">
            <BladeIcon name="shield-check" size="24" class="sec-icon" />
            <div>
              <strong class="sec-title">Proteksi Sesi Aktif</strong>
              <p class="sec-desc">
                Sesi login Anda dilindungi oleh algoritma enkripsi password modern Argon2id dan proteksi anti-pemalsuan token CSRF.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.profile-page {
  padding: 1.5rem;
  max-width: 1400px;
  margin: 0 auto;
}

/* Profile Hero Card */
.profile-hero-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 1.75rem 2rem;
  margin-bottom: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1.5rem;
  flex-wrap: wrap;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.hero-left {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.user-avatar-large {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  color: #ffffff;
  font-size: 2rem;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(5, 150, 105, 0.25);
  flex-shrink: 0;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.user-name {
  font-size: 1.45rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0;
}

.role-badge {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.25rem 0.65rem;
  border-radius: 9999px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.badge-role-super {
  background: #fdf2f8;
  color: #be185d;
  border: 1px solid #fbcfe8;
}

.badge-role-admin {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.badge-role-editor {
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
}

.badge-role-author {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #047857;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  padding: 0.2rem 0.55rem;
  border-radius: 9999px;
}

.status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
}

.user-email {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  color: #64748b;
  font-size: 0.875rem;
  margin: 0;
}

.hero-stats {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.stat-box {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  background: #f8fafc;
  padding: 0.65rem 1rem;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.stat-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.stat-value {
  font-size: 1.15rem;
  font-weight: 800;
  color: #059669;
}

.stat-value-sub {
  font-size: 0.85rem;
  font-weight: 700;
  color: #334155;
}

/* Grid Layout */
.profile-content-grid {
  display: grid;
  grid-template-columns: 1.35fr 1fr;
  gap: 1.5rem;
  align-items: flex-start;
}

@media (max-width: 1024px) {
  .profile-content-grid {
    grid-template-columns: 1fr;
  }
}

/* Permissions Card */
.permissions-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 1.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.card-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

.card-subtitle {
  color: #64748b;
  font-size: 0.875rem;
  margin: 0;
}

.counter-pill {
  background: #ecfdf5;
  color: #047857;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.35rem 0.75rem;
  border-radius: 9999px;
  border: 1px solid #a7f3d0;
  white-space: nowrap;
}

/* Banners */
.super-admin-banner {
  background: linear-gradient(135deg, #f0fdf4 0%, #ecfdf5 100%);
  border: 1px solid #a7f3d0;
  border-left: 4px solid #059669;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  margin-bottom: 1.25rem;
}

.shield-icon {
  color: #059669;
  flex-shrink: 0;
  margin-top: 0.15rem;
}

.role-access-banner {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-left: 4px solid #3b82f6;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  margin-bottom: 1.25rem;
}

.key-icon {
  color: #3b82f6;
  flex-shrink: 0;
  margin-top: 0.15rem;
}

.banner-title {
  display: block;
  font-size: 0.9rem;
  color: #0f172a;
  margin-bottom: 0.2rem;
}

.banner-desc {
  font-size: 0.8125rem;
  color: #475569;
  line-height: 1.45;
  margin: 0;
}

/* Modules Grid */
.modules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 0.875rem;
}

.module-item-card {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: all 0.15s ease;
  background: #ffffff;
}

.module-item-card.allowed {
  border-color: #d1fae5;
  background: #ffffff;
}

.module-item-card.allowed:hover {
  border-color: #059669;
  box-shadow: 0 4px 10px rgba(5, 150, 105, 0.08);
  transform: translateY(-2px);
}

.module-item-card.forbidden {
  background: #f8fafc;
  opacity: 0.7;
}

.module-item-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.module-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: #f1f5f9;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
}

.module-icon-wrap.icon-allowed {
  background: #ecfdf5;
  color: #059669;
}

.access-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
}

.access-pill.allowed {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.access-pill.forbidden {
  background: #f1f5f9;
  color: #94a3b8;
  border: 1px solid #cbd5e1;
}

.module-item-body {
  flex: 1;
}

.module-name {
  font-size: 0.95rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.2rem 0;
}

.module-path {
  display: inline-block;
  font-size: 0.75rem;
  color: #059669;
  background: #f0fdf4;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.module-desc {
  font-size: 0.775rem;
  color: #64748b;
  line-height: 1.4;
  margin: 0 0 0.75rem 0;
}

.module-item-footer {
  padding-top: 0.65rem;
  border-top: 1px solid #f1f5f9;
}

.module-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: #059669;
  text-decoration: none;
}

.module-link:hover {
  text-decoration: underline;
}

/* Forms Column */
.forms-column {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.settings-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 1.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
}

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

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.field-icon {
  position: absolute;
  left: 0.75rem;
  color: #94a3b8;
  pointer-events: none;
}

.form-control {
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.input-with-icon .form-control {
  padding-left: 2.25rem;
}

.form-control:focus {
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}

.form-control:disabled {
  background: #f8fafc;
  color: #64748b;
  cursor: not-allowed;
}

.form-hint {
  font-size: 0.75rem;
  color: #64748b;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

/* Security Box */
.security-info-box {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem 1.25rem;
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
}

.sec-icon {
  color: #059669;
  flex-shrink: 0;
  margin-top: 0.15rem;
}

.sec-title {
  display: block;
  font-size: 0.875rem;
  color: #0f172a;
  margin-bottom: 0.15rem;
}

.sec-desc {
  font-size: 0.775rem;
  color: #64748b;
  margin: 0;
  line-height: 1.4;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1.25rem;
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
</style>
