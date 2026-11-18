<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '../../components/layout/AppLayout.vue'
import { api } from '../../api/client'

interface AuditLogItem {
  id: number
  admin_id: number | null
  admin_email: string | null
  admin_display_name: string | null
  action: string
  entity_type: string
  entity_id: number | null
  metadata: Record<string, any> | null
  ip_address: string | null
  created_at: string
}

interface AuditResponse {
  logs: AuditLogItem[]
  total: number
  page: number
  limit: number
}

const logs = ref<AuditLogItem[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const loading = ref(false)
const error = ref('')

// Filter states
const filterAction = ref('')
const filterEntityType = ref('')

// Modal state
const selectedLog = ref<AuditLogItem | null>(null)
const showDetailModal = ref(false)

async function loadLogs() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams()
    params.set('page', page.value.toString())
    params.set('limit', limit.value.toString())
    if (filterAction.value.trim()) {
      params.set('action', filterAction.value.trim())
    }
    if (filterEntityType.value.trim()) {
      params.set('entity_type', filterEntityType.value.trim())
    }

    const res = await api.get<AuditResponse>(`/api/v1/admin/audit-logs?${params.toString()}`)
    logs.value = res.logs || []
    total.value = res.total || 0
  } catch (err: any) {
    error.value = err.message || 'Gagal memuat jejak audit log'
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  page.value = 1
  loadLogs()
}

function handleReset() {
  filterAction.value = ''
  filterEntityType.value = ''
  page.value = 1
  loadLogs()
}

function changePage(newPage: number) {
  if (newPage < 1 || (newPage - 1) * limit.value >= total.value) return
  page.value = newPage
  loadLogs()
}

function openDetail(item: AuditLogItem) {
  selectedLog.value = item
  showDetailModal.value = true
}

function formatDate(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  } catch {
    return iso
  }
}

function getActionBadgeClass(action: string): string {
  if (action.includes('delete') || action.includes('revoke')) return 'badge-danger'
  if (action.includes('create') || action.includes('publish')) return 'badge-success'
  if (action.includes('update') || action.includes('save')) return 'badge-info'
  if (action.includes('login')) return 'badge-primary'
  return 'badge-neutral'
}

onMounted(() => {
  loadLogs()
})
</script>

<template>
  <AppLayout>
    <div class="view-header">
      <div>
        <h1>Audit Trail & Aktivitas Sistem</h1>
        <p class="subtitle">Catatan historis yang tidak dapat diubah (append-only) untuk seluruh aktivitas keamanan dan perubahan data.</p>
      </div>
      <button class="btn btn-secondary" @click="loadLogs" :disabled="loading">
        🔄 Muat Ulang
      </button>
    </div>

    <!-- Filter Bar -->
    <div class="card filter-card">
      <form @submit.prevent="handleFilter" class="filter-form">
        <div class="filter-group">
          <label>Aksi / Aktivitas</label>
          <input
            v-model="filterAction"
            type="text"
            class="input-control"
            placeholder="Contoh: auth.login, content.update"
          />
        </div>
        <div class="filter-group">
          <label>Tipe Entitas</label>
          <input
            v-model="filterEntityType"
            type="text"
            class="input-control"
            placeholder="Contoh: content, admin, setting"
          />
        </div>
        <div class="filter-actions">
          <button type="submit" class="btn btn-primary" :disabled="loading">Terapkan Filter</button>
          <button type="button" class="btn btn-outline" @click="handleReset" :disabled="loading">Reset</button>
        </div>
      </form>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div class="card">
      <div v-if="loading" class="loading-state">Memuat data audit log...</div>
      <div v-else-if="logs.length === 0" class="empty-state">
        <p>Tidak ada catatan audit yang cocok dengan kriteria filter.</p>
      </div>
      <div v-else class="table-responsive">
        <table class="table">
          <thead>
            <tr>
              <th>Waktu (WIB)</th>
              <th>Operator / Admin</th>
              <th>Aksi</th>
              <th>Entitas</th>
              <th>Alamat IP</th>
              <th>Detail</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td class="text-nowrap font-mono text-sm">{{ formatDate(log.created_at) }}</td>
              <td>
                <div v-if="log.admin_email">
                  <strong>{{ log.admin_display_name || 'Admin' }}</strong>
                  <div class="text-xs text-muted">{{ log.admin_email }}</div>
                </div>
                <span v-else class="text-muted italic">Sistem / Background</span>
              </td>
              <td>
                <span :class="['badge', getActionBadgeClass(log.action)]">
                  {{ log.action }}
                </span>
              </td>
              <td>
                <span class="entity-pill">{{ log.entity_type }}</span>
                <span v-if="log.entity_id" class="text-xs text-muted font-mono ml-1">#{{ log.entity_id }}</span>
              </td>
              <td class="font-mono text-sm">{{ log.ip_address || '-' }}</td>
              <td>
                <button class="btn btn-sm btn-ghost" @click="openDetail(log)">Inspeksi</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="total > 0" class="pagination-bar">
        <div class="pagination-info">
          Menampilkan baris {{ (page - 1) * limit + 1 }} - {{ Math.min(page * limit, total) }} dari total {{ total }}
        </div>
        <div class="pagination-controls">
          <button
            class="btn btn-sm btn-outline"
            :disabled="page <= 1 || loading"
            @click="changePage(page - 1)"
          >
            &laquo; Sebelumnya
          </button>
          <span class="page-indicator">Halaman {{ page }}</span>
          <button
            class="btn btn-sm btn-outline"
            :disabled="page * limit >= total || loading"
            @click="changePage(page + 1)"
          >
            Berikutnya &raquo;
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Detail Metadata -->
    <div v-if="showDetailModal && selectedLog" class="modal-backdrop" @click.self="showDetailModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>Detail Rekaman Audit #{{ selectedLog.id }}</h3>
          <button class="close-btn" @click="showDetailModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-row">
            <strong>Waktu:</strong>
            <span>{{ formatDate(selectedLog.created_at) }}</span>
          </div>
          <div class="detail-row">
            <strong>Aksi:</strong>
            <span :class="['badge', getActionBadgeClass(selectedLog.action)]">{{ selectedLog.action }}</span>
          </div>
          <div class="detail-row">
            <strong>Operator:</strong>
            <span>{{ selectedLog.admin_display_name }} ({{ selectedLog.admin_email || 'System' }})</span>
          </div>
          <div class="detail-row">
            <strong>Target:</strong>
            <span>{{ selectedLog.entity_type }} #{{ selectedLog.entity_id ?? 'N/A' }}</span>
          </div>
          <div class="detail-row">
            <strong>IP Address:</strong>
            <span class="font-mono">{{ selectedLog.ip_address || 'N/A' }}</span>
          </div>

          <h4 class="mt-4 mb-2">Payload Metadata (JSON)</h4>
          <pre class="json-viewer">{{ JSON.stringify(selectedLog.metadata, null, 2) }}</pre>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDetailModal = false">Tutup</button>
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
  color: var(--color-text-muted);
  font-size: 0.95rem;
  margin-top: 0.25rem;
}
.filter-card {
  margin-bottom: 1.5rem;
  padding: 1rem 1.25rem;
}
.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: flex-end;
}
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 200px;
  flex: 1;
}
.filter-group label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-text-muted);
}
.filter-actions {
  display: flex;
  gap: 0.5rem;
}
.table-responsive {
  overflow-x: auto;
}
.table {
  width: 100%;
  border-collapse: collapse;
}
.table th,
.table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}
.table th {
  background: var(--color-bg-subtle);
  font-size: 0.85rem;
  text-transform: uppercase;
  color: var(--color-text-muted);
}
.badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  font-family: monospace;
}
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-success { background: #dcfce7; color: #166534; }
.badge-info { background: #e0f2fe; color: #075985; }
.badge-primary { background: #e0e7ff; color: #3730a3; }
.badge-neutral { background: #f1f5f9; color: #475569; }
.entity-pill {
  background: var(--color-bg-subtle);
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  font-size: 0.8rem;
  font-weight: 500;
}
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-top: 1px solid var(--color-border);
}
.pagination-info {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}
.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.page-indicator {
  font-size: 0.85rem;
  font-weight: 500;
}
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal {
  background: #ffffff;
  border-radius: 8px;
  width: 100%;
  max-width: 600px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border);
}
.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--color-text-muted);
}
.modal-body {
  padding: 1.25rem;
}
.modal-footer {
  padding: 0.75rem 1.25rem;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--color-border);
}
.detail-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}
.json-viewer {
  background: #0f172a;
  color: #f8fafc;
  padding: 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  overflow-x: auto;
  max-height: 250px;
}
.font-mono { font-family: monospace; }
.text-sm { font-size: 0.85rem; }
.text-xs { font-size: 0.75rem; }
.text-muted { color: var(--color-text-muted); }
.text-nowrap { white-space: nowrap; }
.mt-4 { margin-top: 1rem; }
.mb-2 { margin-bottom: 0.5rem; }
.ml-1 { margin-left: 0.25rem; }
</style>
