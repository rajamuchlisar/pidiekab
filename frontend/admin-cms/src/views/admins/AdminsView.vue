<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useAuthStore } from '../../stores/auth'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const authStore = useAuthStore()
const notification = useNotificationStore()

const admins = ref<any[]>([])
const isLoading = ref(true)

// Create Admin modal
const showCreateModal = ref(false)
const isSubmitting = ref(false)
const createForm = reactive({
  email: '',
  displayName: '',
  password: '',
})

// Change Password modal
const showPasswordModal = ref(false)
const isChangingPassword = ref(false)
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})

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
      password: createForm.password,
    })
    notification.success('Admin Dibuat', `Akun ${createForm.email} berhasil ditambahkan.`)
    showCreateModal.value = false
    createForm.email = ''
    createForm.displayName = ''
    createForm.password = ''
    loadAdmins()
  } catch (err: any) {
    notification.error('Gagal Membuat Admin', err.detail || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSubmitting.value = false
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
  loadAdmins()
})
</script>

<template>
  <AppLayout>
    <div class="admins-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Manajemen Administrator</h2>
          <p class="page-desc">Kelola akun administrator dengan akses penuh ke sistem PidieKab.</p>
        </div>
        <div class="header-actions">
          <button class="btn btn-secondary" @click="showPasswordModal = true">🔑 Ganti Kata Sandi Saya</button>
          <button class="btn btn-primary" @click="showCreateModal = true">+ Tambah Administrator</button>
        </div>
      </div>

      <div class="table-card">
        <div v-if="isLoading" class="loading-state">Memuat data administrator...</div>

        <table v-else class="data-table">
          <thead>
            <tr>
              <th>Nama Tampilan</th>
              <th>Alamat Email</th>
              <th>Status</th>
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
                <span
                  class="badge"
                  :class="adm.isActive !== false && adm.is_active !== false ? 'badge-published' : 'badge-draft'"
                >
                  {{ adm.isActive !== false && adm.is_active !== false ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="text-muted">
                {{ adm.createdAt || adm.created_at ? new Date(adm.createdAt || adm.created_at).toLocaleDateString('id-ID') : '-' }}
              </td>
              <td class="text-right">
                <button
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

      <!-- Create Admin Modal -->
      <div v-if="showCreateModal" class="modal-backdrop" @click.self="showCreateModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Tambah Administrator Baru</h3>
            <button class="modal-close" @click="showCreateModal = false">&times;</button>
          </div>
          <form @submit.prevent="handleCreateAdmin" class="modal-form">
            <div class="form-group">
              <label>Nama Tampilan *</label>
              <input v-model="createForm.displayName" type="text" class="input-control" placeholder="John Doe" required />
            </div>

            <div class="form-group">
              <label>Alamat Email *</label>
              <input v-model="createForm.email" type="email" class="input-control" placeholder="user@pidiekab.id" required />
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

      <!-- Change Password Modal -->
      <div v-if="showPasswordModal" class="modal-backdrop" @click.self="showPasswordModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>Ganti Kata Sandi Saya</h3>
            <button class="modal-close" @click="showPasswordModal = false">&times;</button>
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
  align-items: center;
}

.page-title { margin: 0 0 0.25rem 0; font-size: 1.35rem; font-weight: 700; color: #0f172a; }
.page-desc { margin: 0; font-size: 0.85rem; color: #64748b; }

.header-actions {
  display: flex;
  gap: 0.75rem;
}

.table-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
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

.font-medium { font-weight: 600; }
.text-muted { color: #64748b; font-size: 0.8rem; }
.text-right { text-align: right; }

.badge-self {
  font-size: 0.75rem;
  color: #059669;
  font-weight: 600;
  margin-left: 0.35rem;
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.badge-published { background: #d1fae5; color: #065f46; }
.badge-draft { background: #f1f5f9; color: #64748b; }

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
.btn-secondary { background: #f1f5f9; color: #334155; border: 1px solid #cbd5e1; }
.btn-secondary:hover { background: #e2e8f0; }

.btn-sm {
  padding: 0.35rem 0.65rem;
  font-size: 0.775rem;
  font-weight: 600;
  border-radius: 4px;
  cursor: pointer;
}

.btn-outline { border: 1px solid #cbd5e1; background: #ffffff; color: #334155; }
.btn-outline:hover:not(:disabled) { background: #f8fafc; }
.btn-danger { border: 1px solid #fecaca; background: #fee2e2; color: #991b1b; }
.btn-danger:hover:not(:disabled) { background: #fca5a5; }
.btn-sm:disabled { opacity: 0.4; cursor: not-allowed; }

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
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
  max-width: 480px;
  overflow: hidden;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.2);
}

.modal-header {
  padding: 1.1rem 1.25rem;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 { margin: 0; font-size: 1.15rem; }
.modal-close { background: none; border: none; font-size: 1.4rem; cursor: pointer; color: #94a3b8; }

.modal-form {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

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

.input-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid #f1f5f9;
}

.loading-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
