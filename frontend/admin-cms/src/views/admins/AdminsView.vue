<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { api } from '../../api/client'
import { useAuthStore } from '../../stores/auth'
import { useNotificationStore } from '../../stores/notification'
import type { AdminRole, RoleDefinition } from '../../types'
import AppLayout from '../../components/layout/AppLayout.vue'

const authStore = useAuthStore()
const notification = useNotificationStore()

const activeTab = ref<'list' | 'matrix'>('list')
const admins = ref<any[]>([])
const availableRoles = ref<RoleDefinition[]>([])
const isLoading = ref(true)

// Create Admin modal
const showCreateModal = ref(false)
const isSubmitting = ref(false)
const createForm = reactive({
  email: '',
  displayName: '',
  role: 'editor' as AdminRole,
  password: '',
})

// Edit Role modal
const showRoleModal = ref(false)
const isUpdatingRole = ref(false)
const roleEditTarget = ref<any>(null)
const selectedRole = ref<AdminRole>('editor')

// Change Password modal
const showPasswordModal = ref(false)
const isChangingPassword = ref(false)
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// Fallback roles in case API fails
const defaultRoles: RoleDefinition[] = [
  {
    key: 'super_admin',
    name: 'Super Administrator',
    description: 'Akses penuh tanpa batasan ke seluruh modul sistem, termasuk kelola admin, role, dan konfigurasi inti.',
    badge_color: 'purple',
    permissions: [
      'admins:manage', 'settings:manage', 'audit:view',
      'content:read', 'content:write', 'content:publish', 'content:delete',
      'category:manage', 'media:manage', 'curation:manage',
      'pages:manage', 'menus:manage', 'downloads:manage',
      'redirects:manage', 'analytics:view',
    ],
  },
  {
    key: 'admin',
    name: 'Administrator',
    description: 'Mengelola operasional konten, kategori, media library, kurasi beranda, halaman statis, menu, dan unduhan dokumen.',
    badge_color: 'green',
    permissions: [
      'settings:manage', 'audit:view',
      'content:read', 'content:write', 'content:publish', 'content:delete',
      'category:manage', 'media:manage', 'curation:manage',
      'pages:manage', 'menus:manage', 'downloads:manage',
      'redirects:manage', 'analytics:view',
    ],
  },
  {
    key: 'editor',
    name: 'Editor Konten',
    description: 'Meninjau, mengedit, mempublikasikan artikel berita/budaya/wisata, kurasi beranda, serta mengelola media.',
    badge_color: 'blue',
    permissions: [
      'content:read', 'content:write', 'content:publish',
      'category:manage', 'media:manage', 'curation:manage',
      'downloads:manage', 'analytics:view',
    ],
  },
  {
    key: 'author',
    name: 'Penulis Konten',
    description: 'Menulis dan mengedit draf artikel sendiri dan mengunggah berkas foto/media pendukung liputan.',
    badge_color: 'amber',
    permissions: [
      'content:read', 'content:write',
      'media:manage', 'analytics:view',
    ],
  },
]

const matrixModules = [
  { name: 'Dashboard & Statistik Pengunjung', key: 'analytics:view', roles: ['super_admin', 'admin', 'editor', 'author'] },
  { name: 'Tulis & Edit Draf Konten', key: 'content:write', roles: ['super_admin', 'admin', 'editor', 'author'] },
  { name: 'Publikasikan Konten ke Publik', key: 'content:publish', roles: ['super_admin', 'admin', 'editor'] },
  { name: 'Hapus Konten', key: 'content:delete', roles: ['super_admin', 'admin'] },
  { name: 'Kelola Kategori Konten', key: 'category:manage', roles: ['super_admin', 'admin', 'editor'] },
  { name: 'Media Library (Unggah & Kelola)', key: 'media:manage', roles: ['super_admin', 'admin', 'editor', 'author'] },
  { name: 'Kurasi & Urutan Beranda (Lead Story)', key: 'curation:manage', roles: ['super_admin', 'admin', 'editor'] },
  { name: 'Unduhan Dokumen Publik', key: 'downloads:manage', roles: ['super_admin', 'admin', 'editor'] },
  { name: 'Halaman Statis & Legalitas', key: 'pages:manage', roles: ['super_admin', 'admin'] },
  { name: 'Menu Navigasi Header', key: 'menus:manage', roles: ['super_admin', 'admin'] },
  { name: 'URL Redirects & Aturan SEO', key: 'redirects:manage', roles: ['super_admin', 'admin'] },
  { name: 'Audit Trail & Log Keamanan', key: 'audit:view', roles: ['super_admin', 'admin'] },
  { name: 'Pengaturan Website & Konfigurasi', key: 'settings:manage', roles: ['super_admin', 'admin'] },
  { name: 'Kelola Akun Admin & Penugasan Role (RBAC)', key: 'admins:manage', roles: ['super_admin'] },
]

const activeCreateRoleMeta = computed<RoleDefinition>(() => {
  const list = availableRoles.value.length > 0 ? availableRoles.value : defaultRoles
  const r = list.find(x => x.key === createForm.role)
  return (r || defaultRoles[2]) as RoleDefinition
})

const activeEditRoleMeta = computed<RoleDefinition>(() => {
  const list = availableRoles.value.length > 0 ? availableRoles.value : defaultRoles
  const r = list.find(x => x.key === selectedRole.value)
  return (r || defaultRoles[2]) as RoleDefinition
})

function getRoleName(roleKey?: string): string {
  const match = (availableRoles.value.length ? availableRoles.value : defaultRoles).find(x => x.key === roleKey)
  return match ? match.name : (roleKey || 'Administrator')
}

function getRoleBadgeClass(roleKey?: string): string {
  switch (roleKey) {
    case 'super_admin':
      return 'badge-role-super'
    case 'admin':
      return 'badge-role-admin'
    case 'editor':
      return 'badge-role-editor'
    case 'author':
      return 'badge-role-author'
    default:
      return 'badge-role-admin'
  }
}

async function loadRoles() {
  try {
    const res = await api.get<{ roles: RoleDefinition[] }>('/api/v1/admin/roles')
    if (res?.roles && res.roles.length > 0) {
      availableRoles.value = res.roles
    } else {
      availableRoles.value = defaultRoles
    }
  } catch {
    availableRoles.value = defaultRoles
  }
}

async function loadAdmins() {
  isLoading.value = true
  try {
    const res = await api.get<{ admins: any[] }>('/api/v1/admin/admins')
    if (res?.admins) {
      admins.value = res.admins
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Administrator', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

async function handleCreateAdmin() {
  if (!createForm.email || !createForm.displayName || !createForm.password) {
    alert('Seluruh field wajib diisi.')
    return
  }
  if (createForm.password.length < 8) {
    alert('Kata sandi minimal 8 karakter.')
    return
  }

  isSubmitting.value = true
  try {
    await api.post('/api/v1/admin/admins', {
      email: createForm.email,
      display_name: createForm.displayName,
      role: createForm.role,
      password: createForm.password,
    })
    notification.success('Admin Dibuat', `Akun ${createForm.email} berhasil ditambahkan dengan peran "${getRoleName(createForm.role)}".`)
    showCreateModal.value = false
    createForm.email = ''
    createForm.displayName = ''
    createForm.role = 'editor'
    createForm.password = ''
    loadAdmins()
  } catch (err: any) {
    notification.error('Gagal Membuat Admin', err.detail || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSubmitting.value = false
  }
}

function openEditRoleModal(admin: any) {
  roleEditTarget.value = admin
  selectedRole.value = (admin.role as AdminRole) || 'editor'
  showRoleModal.value = true
}

async function handleSaveRole() {
  if (!roleEditTarget.value) return
  isUpdatingRole.value = true
  try {
    await api.put(`/api/v1/admin/admins/${roleEditTarget.value.id}/role`, {
      role: selectedRole.value,
    })
    notification.success('Peran Diperbarui', `Peran administrator "${roleEditTarget.value.email}" berhasil diubah menjadi "${getRoleName(selectedRole.value)}".`)
    showRoleModal.value = false
    loadAdmins()
  } catch (err: any) {
    notification.error('Gagal Mengubah Peran', err.detail || 'Tidak dapat mengubah peran admin.', err.requestId)
  } finally {
    isUpdatingRole.value = false
  }
}

async function toggleAdminStatus(admin: any) {
  const newStatus = !admin.is_active && admin.isActive !== true
  const actionText = newStatus ? 'mengaktifkan' : 'menonaktifkan'
  const confirmed = confirm(`Apakah Anda yakin ingin ${actionText} administrator "${admin.email}"?`)
  if (!confirmed) return

  try {
    await api.put(`/api/v1/admin/admins/${admin.id}/status`, {
      is_active: newStatus,
    })
    notification.success('Status Diperbarui', `Status administrator berhasil diubah.`)
    loadAdmins()
  } catch (err: any) {
    notification.error('Gagal Mengubah Status', err.detail || 'Tidak dapat menonaktifkan admin.', err.requestId)
  }
}

async function handleChangePassword() {
  if (!passwordForm.currentPassword || !passwordForm.newPassword) {
    alert('Kata sandi saat ini dan kata sandi baru wajib diisi.')
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    alert('Konfirmasi kata sandi tidak cocok.')
    return
  }
  if (passwordForm.newPassword.length < 8) {
    alert('Kata sandi baru minimal 8 karakter.')
    return
  }

  isChangingPassword.value = true
  try {
    await api.post('/api/v1/admin/auth/password/change', {
      current_password: passwordForm.currentPassword,
      new_password: passwordForm.newPassword,
    })
    notification.success('Kata Sandi Diubah', 'Kata sandi Anda berhasil diperbarui.')
    showPasswordModal.value = false
    passwordForm.currentPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (err: any) {
    notification.error('Gagal Mengubah Kata Sandi', err.detail || 'Kata sandi saat ini salah.', err.requestId)
  } finally {
    isChangingPassword.value = false
  }
}

onMounted(() => {
  loadRoles()
  loadAdmins()
})
</script>

<template>
  <AppLayout>
    <div class="admins-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Manajemen Administrator & Hak Akses (RBAC)</h2>
          <p class="page-desc">Kelola akun administrator dan atur batasan hak akses berdasarkan peran peran pengguna.</p>
        </div>
        <div class="header-actions">
          <button class="btn btn-secondary" @click="showPasswordModal = true">
            <BladeIcon name="key" size="16" />
            <span>Ganti Kata Sandi Saya</span>
          </button>
          <button v-if="authStore.canManageAdmins" class="btn btn-primary" @click="showCreateModal = true">
            <BladeIcon name="plus" size="16" />
            <span>Tambah Administrator</span>
          </button>
        </div>
      </div>

      <!-- Tab Bar -->
      <div class="tab-bar">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'list' }"
          @click="activeTab = 'list'"
        >
          <BladeIcon name="users" size="16" />
          <span>Daftar Administrator</span>
          <span class="tab-count">{{ admins.length }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'matrix' }"
          @click="activeTab = 'matrix'"
        >
          <BladeIcon name="shield-check" size="16" />
          <span>Matriks Hak Akses (RBAC)</span>
        </button>
      </div>

      <!-- Tab 1: Admins List -->
      <div v-if="activeTab === 'list'" class="table-card">
        <div v-if="isLoading" class="loading-state">Memuat data administrator...</div>

        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Nama Tampilan</th>
              <th>Alamat Email</th>
              <th>Peran (Role)</th>
              <th>Status</th>
              <th>Login Terakhir</th>
              <th>Tanggal Dibuat</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="adm in admins" :key="adm.id">
              <td class="font-medium">
                {{ adm.displayName || adm.display_name }}
                <span v-if="adm.id === authStore.user?.id" class="badge-self">(Anda)</span>
              </td>
              <td>{{ adm.email }}</td>
              <td>
                <span class="role-pill" :class="getRoleBadgeClass(adm.role)">
                  {{ getRoleName(adm.role) }}
                </span>
              </td>
              <td>
                <span
                  class="badge"
                  :class="adm.isActive !== false && adm.is_active !== false ? 'badge-published' : 'badge-draft'"
                >
                  {{ adm.isActive !== false && adm.is_active !== false ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="text-muted text-sm">
                {{ adm.lastLoginAt || adm.last_login_at ? new Date(adm.lastLoginAt || adm.last_login_at).toLocaleString('id-ID') : 'Belum pernah' }}
              </td>
              <td class="text-muted text-sm">
                {{ adm.createdAt || adm.created_at ? new Date(adm.createdAt || adm.created_at).toLocaleDateString('id-ID') : '-' }}
              </td>
              <td class="text-right actions-cell">
                <button
                  v-if="authStore.canManageAdmins"
                  class="btn-sm btn-outline"
                  @click="openEditRoleModal(adm)"
                  title="Ubah peran administrator"
                >
                  Ubah Peran
                </button>
                <button
                  v-if="authStore.canManageAdmins"
                  class="btn-sm"
                  :class="adm.isActive !== false && adm.is_active !== false ? 'btn-danger' : 'btn-outline'"
                  :disabled="adm.id === authStore.user?.id"
                  @click="toggleAdminStatus(adm)"
                  :title="adm.id === authStore.user?.id ? 'Tidak dapat menonaktifkan akun sendiri' : ''"
                >
                  {{ adm.isActive !== false && adm.is_active !== false ? 'Nonaktifkan' : 'Aktifkan' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Tab 2: RBAC Matrix View -->
      <div v-else-if="activeTab === 'matrix'" class="matrix-wrapper">
        <!-- Roles Overview Cards -->
        <div class="roles-grid">
          <div
            v-for="role in (availableRoles.length ? availableRoles : defaultRoles)"
            :key="role.key"
            class="role-card"
            :class="'role-card-' + role.key"
          >
            <div class="role-card-header">
              <span class="role-pill" :class="getRoleBadgeClass(role.key)">{{ role.name }}</span>
            </div>
            <p class="role-card-desc">{{ role.description }}</p>
            <div class="role-card-footer">
              <span class="perm-count">{{ role.permissions.length }} Hak Akses Terdaftar</span>
            </div>
          </div>
        </div>

        <!-- Permission Comparison Matrix Table -->
        <div class="table-card matrix-card">
          <div class="matrix-header">
            <h3>Matriks Perbandingan Perizinan Modul</h3>
            <p>Rincian batasan operasional antarmuka dan API backend untuk setiap level peran.</p>
          </div>
          <table class="data-table matrix-table">
            <thead>
              <tr>
                <th style="width: 38%;">Modul & Fitur Sistem</th>
                <th class="text-center" style="width: 15.5%;">Super Admin</th>
                <th class="text-center" style="width: 15.5%;">Administrator</th>
                <th class="text-center" style="width: 15.5%;">Editor Konten</th>
                <th class="text-center" style="width: 15.5%;">Penulis Konten</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="mod in matrixModules" :key="mod.key">
                <td class="font-medium text-slate-800">
                  <div class="mod-row">
                    <span>{{ mod.name }}</span>
                    <code class="mod-key">{{ mod.key }}</code>
                  </div>
                </td>
                <td class="text-center">
                  <span v-if="mod.roles.includes('super_admin')" class="check-allowed" title="Diizinkan">✓ Diizinkan</span>
                  <span v-else class="check-denied">—</span>
                </td>
                <td class="text-center">
                  <span v-if="mod.roles.includes('admin')" class="check-allowed" title="Diizinkan">✓ Diizinkan</span>
                  <span v-else class="check-denied">—</span>
                </td>
                <td class="text-center">
                  <span v-if="mod.roles.includes('editor')" class="check-allowed" title="Diizinkan">✓ Diizinkan</span>
                  <span v-else class="check-denied">—</span>
                </td>
                <td class="text-center">
                  <span v-if="mod.roles.includes('author')" class="check-allowed" title="Diizinkan">✓ Diizinkan</span>
                  <span v-else class="check-denied">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Create Admin Modal -->
      <div v-if="showCreateModal" class="modal-backdrop" @click.self="showCreateModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Tambah Administrator Baru</h3>
            <button class="modal-close" @click="showCreateModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <form @submit.prevent="handleCreateAdmin" class="modal-form">
            <div class="form-group">
              <label>Nama Tampilan *</label>
              <input v-model="createForm.displayName" type="text" class="input-control" placeholder="Contoh: Raja Muchlisar" required />
            </div>

            <div class="form-group">
              <label>Alamat Email *</label>
              <input v-model="createForm.email" type="email" class="input-control" placeholder="user@pidiekab.web.id" required />
            </div>

            <div class="form-group">
              <label>Peran / Role Pengguna *</label>
              <select v-model="createForm.role" class="input-control" required>
                <option value="editor">Editor Konten (Tulis, Edit & Publikasikan)</option>
                <option value="author">Penulis Konten (Draf & Media)</option>
                <option value="admin">Administrator (Kelola Semua Modul Operasional)</option>
                <option value="super_admin">Super Administrator (Akses Penuh Tanpa Batas)</option>
              </select>
            </div>

            <!-- Role Explainer Box -->
            <div class="role-explainer-box" :class="'explainer-' + createForm.role">
              <div class="explainer-header">
                <span class="role-pill" :class="getRoleBadgeClass(createForm.role)">{{ activeCreateRoleMeta.name }}</span>
                <span class="explainer-perm-count">{{ activeCreateRoleMeta.permissions.length }} izin aktif</span>
              </div>
              <p class="explainer-desc">{{ activeCreateRoleMeta.description }}</p>
            </div>

            <div class="form-group">
              <label>Kata Sandi (Minimal 8 karakter) *</label>
              <input v-model="createForm.password" type="password" class="input-control" placeholder="••••••••" required />
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showCreateModal = false">Batal</button>
              <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
                {{ isSubmitting ? 'Memproses...' : 'Buat Administrator' }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Edit Role Modal -->
      <div v-if="showRoleModal" class="modal-backdrop" @click.self="showRoleModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Ubah Peran Administrator</h3>
            <button class="modal-close" @click="showRoleModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <form @submit.prevent="handleSaveRole" class="modal-form">
            <div class="user-summary-card">
              <div class="summary-avatar">{{ roleEditTarget?.displayName?.charAt(0) || 'A' }}</div>
              <div>
                <h4 class="summary-name">{{ roleEditTarget?.displayName || roleEditTarget?.display_name }}</h4>
                <p class="summary-email">{{ roleEditTarget?.email }}</p>
              </div>
            </div>

            <div class="form-group">
              <label>Pilih Peran Baru *</label>
              <select v-model="selectedRole" class="input-control" required>
                <option value="author">Penulis Konten (Draf & Media)</option>
                <option value="editor">Editor Konten (Tulis, Edit & Publikasikan)</option>
                <option value="admin">Administrator (Kelola Semua Modul Operasional)</option>
                <option value="super_admin">Super Administrator (Akses Penuh Tanpa Batas)</option>
              </select>
            </div>

            <!-- Role Explainer Box -->
            <div class="role-explainer-box" :class="'explainer-' + selectedRole">
              <div class="explainer-header">
                <span class="role-pill" :class="getRoleBadgeClass(selectedRole)">{{ activeEditRoleMeta.name }}</span>
                <span class="explainer-perm-count">{{ activeEditRoleMeta.permissions.length }} izin aktif</span>
              </div>
              <p class="explainer-desc">{{ activeEditRoleMeta.description }}</p>
            </div>

            <div v-if="roleEditTarget?.role === 'super_admin' && selectedRole !== 'super_admin'" class="alert-box alert-warning">
              <strong>Peringatan:</strong> Anda akan menurunkan hak akses akun Super Administrator ini. Pastikan setidaknya masih terdapat satu Super Administrator lain yang aktif.
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showRoleModal = false">Batal</button>
              <button type="submit" class="btn btn-primary" :disabled="isUpdatingRole">
                {{ isUpdatingRole ? 'Menyimpan...' : 'Simpan Perubahan Peran' }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Change Password Modal -->
      <div v-if="showPasswordModal" class="modal-backdrop" @click.self="showPasswordModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Ganti Kata Sandi Saya</h3>
            <button class="modal-close" @click="showPasswordModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <form @submit.prevent="handleChangePassword" class="modal-form">
            <div class="form-group">
              <label>Kata Sandi Saat Ini *</label>
              <input v-model="passwordForm.currentPassword" type="password" class="input-control" required />
            </div>

            <div class="form-group">
              <label>Kata Sandi Baru (Minimal 8 karakter) *</label>
              <input v-model="passwordForm.newPassword" type="password" class="input-control" required />
            </div>

            <div class="form-group">
              <label>Konfirmasi Kata Sandi Baru *</label>
              <input v-model="passwordForm.confirmPassword" type="password" class="input-control" required />
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showPasswordModal = false">Batal</button>
              <button type="submit" class="btn btn-primary" :disabled="isChangingPassword">
                {{ isChangingPassword ? 'Menyimpan...' : 'Perbarui Kata Sandi' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.admins-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
}

.page-desc {
  font-size: 0.875rem;
  color: #64748b;
  margin-top: 0.25rem;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

/* Tab Bar */
.tab-bar {
  display: flex;
  gap: 0.5rem;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 0.25rem;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1.15rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #64748b;
  background: transparent;
  border: none;
  border-radius: 8px 8px 0 0;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  bottom: -2px;
}

.tab-btn:hover {
  color: #0f172a;
  background: #f1f5f9;
}

.tab-btn.active {
  color: #064e3b;
  border-bottom: 2px solid #064e3b;
  background: #ffffff;
}

.tab-count {
  font-size: 0.75rem;
  background: #e2e8f0;
  color: #475569;
  padding: 0.15rem 0.45rem;
  border-radius: 9999px;
}

.tab-btn.active .tab-count {
  background: #d1fae5;
  color: #064e3b;
}

/* Table card */
.table-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.data-table th {
  background: #f8fafc;
  color: #475569;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.85rem 1.25rem;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 1rem 1.25rem;
  font-size: 0.875rem;
  color: #1e293b;
  border-bottom: 1px solid #f1f5f9;
}

.actions-cell {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

/* Role Pills */
.role-pill {
  display: inline-block;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.2rem 0.6rem;
  border-radius: 9999px;
  letter-spacing: 0.025em;
  text-transform: uppercase;
  white-space: nowrap;
}

.badge-role-super {
  background: #f3e8ff;
  color: #7e22ce;
  border: 1px solid #d8b4fe;
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

.badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.badge-published {
  background: #dcfce7;
  color: #15803d;
}

.badge-draft {
  background: #f1f5f9;
  color: #64748b;
}

.badge-self {
  font-size: 0.7rem;
  color: #0284c7;
  font-weight: 600;
  margin-left: 0.35rem;
}

.text-muted {
  color: #64748b;
}

.text-sm {
  font-size: 0.8rem;
}

.text-right {
  text-align: right;
}

.font-medium {
  font-weight: 500;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.55rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid transparent;
}

.btn-primary {
  background: #064e3b;
  color: #ffffff;
}

.btn-primary:hover:not(:disabled) {
  background: #043d2e;
}

.btn-secondary {
  background: #ffffff;
  color: #334155;
  border-color: #cbd5e1;
}

.btn-secondary:hover {
  background: #f8fafc;
}

.btn-sm {
  padding: 0.3rem 0.65rem;
  font-size: 0.775rem;
  font-weight: 600;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.btn-danger {
  background: #fee2e2;
  color: #b91c1c;
  border-color: #fecaca;
}

.btn-danger:hover:not(:disabled) {
  background: #fca5a5;
  color: #991b1b;
}

.btn-outline {
  background: #ffffff;
  color: #334155;
  border-color: #cbd5e1;
}

.btn-outline:hover:not(:disabled) {
  background: #f1f5f9;
  color: #0f172a;
}

.btn-sm:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* RBAC Matrix Tab */
.matrix-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1rem;
}

.role-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.role-card-super_admin {
  border-top: 4px solid #7e22ce;
}

.role-card-admin {
  border-top: 4px solid #047857;
}

.role-card-editor {
  border-top: 4px solid #1d4ed8;
}

.role-card-author {
  border-top: 4px solid #b45309;
}

.role-card-desc {
  font-size: 0.825rem;
  color: #475569;
  line-height: 1.45;
  margin: 0;
  flex: 1;
}

.role-card-footer {
  border-top: 1px solid #f1f5f9;
  padding-top: 0.5rem;
}

.perm-count {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
}

.matrix-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.matrix-header h3 {
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.matrix-header p {
  font-size: 0.825rem;
  color: #64748b;
  margin: 0.25rem 0 0 0;
}

.matrix-table th {
  padding: 0.85rem 1rem;
}

.matrix-table td {
  padding: 0.85rem 1rem;
}

.mod-row {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.mod-key {
  font-size: 0.72rem;
  font-family: monospace;
  color: #64748b;
  background: #f1f5f9;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  width: fit-content;
}

.check-allowed {
  font-size: 0.775rem;
  font-weight: 700;
  color: #059669;
  background: #ecfdf5;
  padding: 0.2rem 0.5rem;
  border-radius: 9999px;
  border: 1px solid #a7f3d0;
}

.check-denied {
  color: #cbd5e1;
  font-size: 1.1rem;
}

/* Modals */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-card {
  background: #ffffff;
  border-radius: 12px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.15);
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
  font-size: 1.125rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.modal-close {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  border-radius: 4px;
}

.modal-close:hover {
  color: #0f172a;
  background: #f1f5f9;
}

.modal-form {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #334155;
}

.input-control {
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  padding: 0.6rem 0.85rem;
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.15s ease;
  background: #ffffff;
}

.input-control:focus {
  border-color: #064e3b;
  box-shadow: 0 0 0 2px rgba(6, 78, 59, 0.15);
}

/* Role Explainer Box */
.role-explainer-box {
  padding: 0.85rem 1rem;
  border-radius: 8px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.explainer-super_admin {
  background: #faf5ff;
  border-color: #e9d5ff;
}

.explainer-admin {
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.explainer-editor {
  background: #f0f9ff;
  border-color: #bae6fd;
}

.explainer-author {
  background: #fffbeb;
  border-color: #fde68a;
}

.explainer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.explainer-perm-count {
  font-size: 0.72rem;
  color: #64748b;
  font-weight: 600;
}

.explainer-desc {
  font-size: 0.8rem;
  color: #334155;
  line-height: 1.4;
  margin: 0;
}

/* User summary card in Edit Role */
.user-summary-card {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.75rem 1rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.summary-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: #064e3b;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.summary-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.summary-email {
  font-size: 0.8rem;
  color: #64748b;
  margin: 0;
}

.alert-box {
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 0.825rem;
  line-height: 1.4;
}

.alert-warning {
  background: #fffbeb;
  color: #92400e;
  border: 1px solid #fde68a;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.loading-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
  font-size: 0.9rem;
}
</style>
