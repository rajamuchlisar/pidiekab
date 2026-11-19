<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '../../components/layout/AppLayout.vue'
import { api } from '../../api/client'

interface RedirectItem {
  id: number
  old_path: string
  new_path: string
  status_code: number
  created_at: string
}

const redirects = ref<RedirectItem[]>([])
const loading = ref(false)
const error = ref('')
const successMsg = ref('')

// Form state
const showModal = ref(false)
const oldPath = ref('')
const newPath = ref('')
const isSubmitting = ref(false)

async function loadRedirects() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/api/v1/admin/redirects')
    redirects.value = res.data || []
  } catch (err: any) {
    error.value = err.message || 'Gagal memuat daftar URL redirects'
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  if (!oldPath.value.trim() || !newPath.value.trim()) {
    error.value = 'Path asal dan path tujuan harus diisi'
    return
  }
  isSubmitting.value = true
  error.value = ''
  try {
    await api.post('/api/v1/admin/redirects', {
      old_path: oldPath.value.trim(),
      new_path: newPath.value.trim(),
    })
    successMsg.value = 'URL redirect berhasil ditambahkan.'
    showModal.value = false
    oldPath.value = ''
    newPath.value = ''
    await loadRedirects()
  } catch (err: any) {
    error.value = err.message || 'Gagal menyimpan URL redirect'
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(id: number) {
  if (!confirm('Apakah Anda yakin ingin menghapus redirect ini?')) return
  try {
    await api.del(`/api/v1/admin/redirects/${id}`)
    successMsg.value = 'Redirect berhasil dihapus.'
    await loadRedirects()
  } catch (err: any) {
    error.value = err.message || 'Gagal menghapus redirect'
  }
}

onMounted(() => {
  loadRedirects()
})
</script>

<template>
  <AppLayout>
    <div class="view-header">
      <div>
        <h1>URL Redirects (301)</h1>
        <p class="subtitle">Kelola pengalihan URL permanen untuk menjaga ranking SEO dan tautan masuk lama.</p>
      </div>
      <button class="btn btn-primary" @click="showModal = true">
        <BladeIcon name="plus" size="18" />
        <span>Tambah Redirect</span>
      </button>
    </div>

    <div v-if="successMsg" class="alert alert-success">{{ successMsg }}</div>
    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div class="card">
      <div v-if="loading" class="loading-state">Memuat data redirects...</div>
      <div v-else-if="redirects.length === 0" class="empty-state">
        <p>Belum ada aturan URL redirect yang terdaftar.</p>
      </div>
      <table v-else class="table">
        <thead>
          <tr>
            <th>Path Asal (Old Path)</th>
            <th>Target Pengalihan (New Path)</th>
            <th>Status</th>
            <th>Tanggal Dibuat</th>
            <th>Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in redirects" :key="item.id">
            <td><code>{{ item.old_path }}</code></td>
            <td><code>{{ item.new_path }}</code></td>
            <td><span class="badge badge-success">{{ item.status_code }} Permanent</span></td>
            <td>{{ new Date(item.created_at).toLocaleDateString('id-ID') }}</td>
            <td>
              <button class="btn-sm btn-danger" @click="handleDelete(item.id)">
                <BladeIcon name="trash" size="14" />
                <span>Hapus</span>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal Form Tambah Redirect -->
    <div v-if="showModal" class="modal-backdrop">
      <div class="modal">
        <div class="modal-header">
          <h3>Tambah URL Redirect Baru</h3>
          <button class="btn-close" @click="showModal = false" title="Tutup">
            <BladeIcon name="x-mark" size="20" />
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Path Asal (Lama):</label>
            <input
              type="text"
              v-model="oldPath"
              placeholder="/id/kategori-lama/artikel-lama/"
              class="form-control"
            />
            <small class="form-hint">Path lama yang ingin dialihkan (misal: link rusak atau artikel yang diubah slug-nya).</small>
          </div>
          <div class="form-group">
            <label>Path Tujuan (Baru):</label>
            <input
              type="text"
              v-model="newPath"
              placeholder="/id/kategori-baru/artikel-baru/"
              class="form-control"
            />
            <small class="form-hint">Target URL baru ke mana pengunjung akan diarahkan.</small>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showModal = false">Batal</button>
          <button class="btn btn-primary" :disabled="isSubmitting" @click="handleCreate">
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Redirect' }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}
.subtitle {
  color: #6b7280;
  margin: 0.25rem 0 0 0;
  font-size: 0.9rem;
}
.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  border: 1px solid #e5e7eb;
  overflow: hidden;
}
.table {
  width: 100%;
  border-collapse: collapse;
}
.table th, .table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #f3f4f6;
  font-size: 0.9rem;
}
.table th {
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
}
code {
  background: #f3f4f6;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  color: #0f766e;
}
.badge-success {
  background: #dcfce7;
  color: #15803d;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
}
.btn {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}
.btn-primary {
  background: #0f766e;
  color: white;
}
.btn-primary:hover {
  background: #0d9488;
}
.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}
.btn-sm {
  padding: 0.25rem 0.5rem;
  font-size: 0.8rem;
  border-radius: 4px;
  border: none;
  cursor: pointer;
}
.btn-danger {
  background: #fee2e2;
  color: #b91c1c;
}
.btn-danger:hover {
  background: #fecaca;
}
.alert {
  padding: 0.75rem 1rem;
  border-radius: 6px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}
.alert-success {
  background: #dcfce7;
  color: #166534;
}
.alert-danger {
  background: #fee2e2;
  color: #991b1b;
}
.loading-state, .empty-state {
  padding: 2.5rem;
  text-align: center;
  color: #6b7280;
}
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal {
  background: white;
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
  overflow: hidden;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #e5e7eb;
}
.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
}
.btn-close {
  background: transparent;
  border: none;
  font-size: 1.25rem;
  cursor: pointer;
}
.modal-body {
  padding: 1.25rem;
}
.modal-footer {
  padding: 0.75rem 1.25rem;
  background: #f9fafb;
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
.form-group {
  margin-bottom: 1rem;
}
.form-group label {
  display: block;
  font-size: 0.85rem;
  font-weight: 500;
  margin-bottom: 0.35rem;
}
.form-control {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
}
.form-hint {
  display: block;
  font-size: 0.75rem;
  color: #6b7280;
  margin-top: 0.25rem;
}
</style>
