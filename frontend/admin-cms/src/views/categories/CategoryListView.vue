<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const categories = ref<any[]>([])
const isLoading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const currentId = ref<number | null>(null)
const isSubmitting = ref(false)

const form = reactive({
  code: '',
  displayOrder: 1,
  nameId: '',
  nameEn: '',
  slugId: '',
  slugEn: '',
})

async function loadCategories() {
  isLoading.value = true
  try {
    const res = await api.get<any>('/api/v1/admin/categories')
    const list = res?.categories || res?.data || []
    if (Array.isArray(list)) {
      categories.value = list.map((cat: any) => {
        const idTrans = cat.translations?.id || {}
        const enTrans = cat.translations?.en || {}
        return {
          id: cat.id,
          code: cat.code || idTrans.slug || `cat-${cat.id}`,
          displayOrder: cat.displayOrder || cat.display_order || cat.id,
          isActive: cat.isActive !== undefined ? cat.isActive : (cat.is_active !== undefined ? cat.is_active : true),
          names: {
            id: cat.names?.id || idTrans.name || cat.name || '-',
            en: cat.names?.en || enTrans.name || '-',
          },
          slugs: {
            id: cat.slugs?.id || idTrans.slug || '-',
            en: cat.slugs?.en || enTrans.slug || '-',
          },
          raw: cat,
        }
      })
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Kategori', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

function openCreateModal() {
  isEditing.value = false
  currentId.value = null
  form.code = ''
  form.displayOrder = categories.value.length + 1
  form.nameId = ''
  form.nameEn = ''
  form.slugId = ''
  form.slugEn = ''
  showModal.value = true
}

function openEditModal(cat: any) {
  isEditing.value = true
  currentId.value = cat.id
  form.code = cat.code || ''
  form.displayOrder = cat.displayOrder || cat.display_order || 1
  form.nameId = cat.names?.id !== '-' ? cat.names?.id : ''
  form.nameEn = cat.names?.en !== '-' ? cat.names?.en : ''
  form.slugId = cat.slugs?.id !== '-' ? cat.slugs?.id : ''
  form.slugEn = cat.slugs?.en !== '-' ? cat.slugs?.en : ''
  showModal.value = true
}

function generateSlugs() {
  if (form.nameId && !form.slugId) {
    form.slugId = form.nameId.toLowerCase().trim().replace(/[^a-z0-9\s-]/g, '').replace(/\s+/g, '-')
  }
  if (form.nameEn && !form.slugEn) {
    form.slugEn = form.nameEn.toLowerCase().trim().replace(/[^a-z0-9\s-]/g, '').replace(/\s+/g, '-')
  }
}

async function handleSubmit() {
  if (!form.nameId) {
    alert('Nama Kategori (Bahasa Indonesia) wajib diisi.')
    return
  }

  isSubmitting.value = true
  try {
    const slugId = form.slugId || form.nameId.toLowerCase().trim().replace(/[^a-z0-9\s-]/g, '').replace(/\s+/g, '-')
    const slugEn = form.slugEn || (form.nameEn ? form.nameEn.toLowerCase().trim().replace(/[^a-z0-9\s-]/g, '').replace(/\s+/g, '-') : slugId)
    const code = form.code || slugId

    const payload = {
      is_active: true,
      code,
      names: { id: form.nameId, en: form.nameEn || form.nameId },
      slugs: { id: slugId, en: slugEn },
      translations: [
        { locale: 'id', name: form.nameId, slug: slugId },
        { locale: 'en', name: form.nameEn || form.nameId, slug: slugEn },
      ],
    }

    if (isEditing.value && currentId.value) {
      await api.put(`/api/v1/admin/categories/${currentId.value}`, payload)
      await api.put(`/api/v1/admin/categories/${currentId.value}/translations/id`, {
        name: form.nameId,
        slug: slugId,
      }).catch(() => null)
      if (form.nameEn) {
        await api.put(`/api/v1/admin/categories/${currentId.value}/translations/en`, {
          name: form.nameEn,
          slug: slugEn,
        }).catch(() => null)
      }
      notification.success('Kategori Diperbarui', `Kategori "${form.nameId}" berhasil disimpan.`)
    } else {
      await api.post('/api/v1/admin/categories', payload)
      notification.success('Kategori Dibuat', `Kategori "${form.nameId}" berhasil ditambahkan.`)
    }
    showModal.value = false
    loadCategories()
  } catch (err: any) {
    notification.error('Gagal Menyimpan Kategori', err.detail || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSubmitting.value = false
  }
}

async function deleteCategory(cat: any) {
  const confirmed = confirm(
    `Apakah Anda yakin ingin menghapus kategori "${cat.names?.id || cat.code}"?\n\nKategori yang masih digunakan oleh artikel tidak dapat dihapus.`
  )
  if (!confirmed) return

  try {
    await api.del(`/api/v1/admin/categories/${cat.id}`)
    notification.success('Kategori Dihapus', 'Kategori berhasil dihapus.')
    loadCategories()
  } catch (err: any) {
    notification.error('Gagal Menghapus', err.detail || 'Kategori sedang digunakan atau tidak dapat dihapus.', err.requestId)
  }
}

onMounted(() => {
  loadCategories()
})
</script>

<template>
  <AppLayout>
    <div class="categories-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Manajemen Kategori</h2>
          <p class="page-desc">Kelola kategori konten publikasi dwibahasa (ID & EN).</p>
        </div>
        <button class="btn btn-primary" @click="openCreateModal">
          <BladeIcon name="plus" size="18" />
          <span>Tambah Kategori</span>
        </button>
      </div>

      <div class="table-card">
        <div v-if="isLoading" class="loading-state">Memuat kategori...</div>

        <div v-else-if="categories.length === 0" class="empty-state">
          Belum ada kategori yang terdaftar.
        </div>

        <table v-else class="data-table">
          <thead>
            <tr>
              <th style="width: 80px;">Urutan</th>
              <th>Kode</th>
              <th>Nama (ID)</th>
              <th>Nama (EN)</th>
              <th>Slug (ID / EN)</th>
              <th>Status</th>
              <th class="text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cat in categories" :key="cat.id">
              <td class="font-bold">{{ cat.displayOrder || cat.display_order }}</td>
              <td><code>{{ cat.code }}</code></td>
              <td class="font-medium">{{ cat.names?.id || '-' }}</td>
              <td>{{ cat.names?.en || '-' }}</td>
              <td class="text-muted">
                /{{ cat.slugs?.id || '-' }}<br />
                /{{ cat.slugs?.en || '-' }}
              </td>
              <td>
                <span class="badge" :class="cat.isActive !== false ? 'badge-published' : 'badge-draft'">
                  {{ cat.isActive !== false ? 'Aktif' : 'Nonaktif' }}
                </span>
              </td>
              <td class="text-right">
                <button class="btn-sm btn-outline" @click="openEditModal(cat)">
                  <BladeIcon name="pencil-square" size="14" />
                  <span>Edit</span>
                </button>
                <button class="btn-sm btn-danger ml-1" @click="deleteCategory(cat)">
                  <BladeIcon name="trash" size="14" />
                  <span>Hapus</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Create / Edit Modal -->
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal-card">
          <div class="modal-header">
            <h3>{{ isEditing ? 'Edit Kategori' : 'Tambah Kategori Baru' }}</h3>
            <button class="modal-close" @click="showModal = false" title="Tutup">
              <BladeIcon name="x-mark" size="20" />
            </button>
          </div>
          <form @submit.prevent="handleSubmit" class="modal-form">
            <div class="form-row">
              <div class="form-group flex-1">
                <label>Kode Unik (Alphanumeric) *</label>
                <input v-model="form.code" type="text" class="input-control" placeholder="wisata" required />
              </div>
              <div class="form-group" style="width: 100px;">
                <label>Urutan *</label>
                <input v-model="form.displayOrder" type="number" class="input-control" min="1" required />
              </div>
            </div>

            <div class="form-group">
              <label>Nama Kategori (Bahasa Indonesia) *</label>
              <input
                v-model="form.nameId"
                type="text"
                class="input-control"
                placeholder="Wisata & Destinasi"
                required
                @blur="generateSlugs"
              />
            </div>

            <div class="form-group">
              <label>Slug URL (Bahasa Indonesia) *</label>
              <input v-model="form.slugId" type="text" class="input-control" placeholder="wisata-destinasi" required />
            </div>

            <div class="form-group">
              <label>Nama Kategori (English) *</label>
              <input
                v-model="form.nameEn"
                type="text"
                class="input-control"
                placeholder="Tourism & Destinations"
                required
                @blur="generateSlugs"
              />
            </div>

            <div class="form-group">
              <label>Slug URL (English) *</label>
              <input v-model="form.slugEn" type="text" class="input-control" placeholder="tourism-destinations" required />
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showModal = false">Batal</button>
              <button type="submit" class="btn btn-primary" :disabled="isSubmitting">
                {{ isSubmitting ? 'Menyimpan...' : 'Simpan Kategori' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.categories-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  margin: 0 0 0.25rem 0;
  font-size: 1.35rem;
  font-weight: 700;
  color: #0f172a;
}

.page-desc {
  margin: 0;
  font-size: 0.85rem;
  color: #64748b;
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

.font-bold { font-weight: 700; }
.font-medium { font-weight: 600; }
.text-muted { color: #64748b; font-size: 0.8rem; }
.text-right { text-align: right; }
.ml-1 { margin-left: 0.4rem; }

code {
  background: #f1f5f9;
  padding: 0.15rem 0.35rem;
  border-radius: 4px;
  font-size: 0.8rem;
  color: #0f172a;
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
.btn-primary:hover { background: #047857; }
.btn-secondary { background: #f1f5f9; color: #334155; border: 1px solid #cbd5e1; }
.btn-secondary:hover { background: #e2e8f0; }

.btn-sm {
  padding: 0.35rem 0.65rem;
  font-size: 0.775rem;
  font-weight: 600;
  border-radius: 4px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.btn-outline { border: 1px solid #cbd5e1; background: #ffffff; color: #334155; }
.btn-outline:hover { background: #f8fafc; }
.btn-danger { border: 1px solid #fecaca; background: #fee2e2; color: #991b1b; }
.btn-danger:hover { background: #fca5a5; }

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
  max-width: 520px;
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
.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  border-radius: 4px;
}
.modal-close:hover {
  color: #475569;
  background: #f1f5f9;
}

.modal-form {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row { display: flex; gap: 0.75rem; }
.flex-1 { flex: 1; }

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

.loading-state, .empty-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
