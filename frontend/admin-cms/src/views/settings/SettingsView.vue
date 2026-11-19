<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { api } from '../../api/client'
import { useNotificationStore } from '../../stores/notification'
import AppLayout from '../../components/layout/AppLayout.vue'

const notification = useNotificationStore()

const isLoading = ref(true)
const isSaving = ref(false)

const form = reactive({
  siteName: 'PidieKab',
  taglineId: 'Eksplorasi Pesona Wisata, Kuliner & Budaya Kabupaten Pidie',
  taglineEn: 'Explore the Beauty, Culinary & Culture of Pidie Regency',
  logoUrl: '',
  faviconUrl: '',
  contactEmail: 'kontak@pidiekab.id',
  contactPhone: '',
  address: 'Sigli, Kabupaten Pidie, Aceh, Indonesia',
  socialInstagram: '',
  socialFacebook: '',
  socialYoutube: '',
  socialTwitter: '',
  defaultSeoTitleId: 'PidieKab — Direktori Wisata, Kuliner & Budaya Kabupaten Pidie',
  defaultSeoTitleEn: 'PidieKab — Tourism, Culinary & Heritage of Pidie Regency',
  defaultSeoDescId: 'Portal informasi independen potensi wisata, warisan budaya, dan kekayaan kuliner khas Kabupaten Pidie.',
  defaultSeoDescEn: 'Independent digital portal exploring tourism, culinary delights, and rich cultural heritage of Pidie, Aceh.',
})

async function loadSettings() {
  isLoading.value = true
  try {
    const res = await api.get<{ settings: any }>('/api/v1/admin/settings')
    if (res?.settings) {
      const s = res.settings
      form.siteName = s.siteName || s.site_name || form.siteName
      form.taglineId = s.taglineId || s.tagline_id || form.taglineId
      form.taglineEn = s.taglineEn || s.tagline_en || form.taglineEn
      form.logoUrl = s.logoUrl || s.logo_url || ''
      form.faviconUrl = s.faviconUrl || s.favicon_url || ''
      form.contactEmail = s.contactEmail || s.contact_email || form.contactEmail
      form.contactPhone = s.contactPhone || s.contact_phone || ''
      form.address = s.address || form.address
      form.socialInstagram = s.socialInstagram || s.social_instagram || ''
      form.socialFacebook = s.socialFacebook || s.social_facebook || ''
      form.socialYoutube = s.socialYoutube || s.social_youtube || ''
      form.socialTwitter = s.socialTwitter || s.social_twitter || ''
      form.defaultSeoTitleId = s.defaultSeoTitleId || s.default_seo_title_id || form.defaultSeoTitleId
      form.defaultSeoTitleEn = s.defaultSeoTitleEn || s.default_seo_title_en || form.defaultSeoTitleEn
      form.defaultSeoDescId = s.defaultSeoDescId || s.default_seo_desc_id || form.defaultSeoDescId
      form.defaultSeoDescEn = s.defaultSeoDescEn || s.default_seo_desc_en || form.defaultSeoDescEn
    }
  } catch (err: any) {
    notification.error('Gagal Memuat Pengaturan', err.detail || 'Terjadi kesalahan sistem.', err.requestId)
  } finally {
    isLoading.value = false
  }
}

async function saveSettings() {
  isSaving.value = true
  try {
    const payload = {
      site_name: form.siteName,
      tagline_id: form.taglineId,
      tagline_en: form.taglineEn,
      logo_url: form.logoUrl,
      favicon_url: form.faviconUrl,
      contact_email: form.contactEmail,
      contact_phone: form.contactPhone,
      address: form.address,
      social_instagram: form.socialInstagram,
      social_facebook: form.socialFacebook,
      social_youtube: form.socialYoutube,
      social_twitter: form.socialTwitter,
      default_seo_title_id: form.defaultSeoTitleId,
      default_seo_title_en: form.defaultSeoTitleEn,
      default_seo_desc_id: form.defaultSeoDescId,
      default_seo_desc_en: form.defaultSeoDescEn,
    }

    await api.put('/api/v1/admin/settings', payload)
    notification.success('Pengaturan Disimpan', 'Konfigurasi website berhasil diperbarui.')
  } catch (err: any) {
    notification.error('Gagal Menyimpan', err.detail || 'Terjadi kesalahan.', err.requestId)
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <AppLayout>
    <div class="settings-page">
      <div class="page-header">
        <div>
          <h2 class="page-title">Pengaturan Website</h2>
          <p class="page-desc">Konfigurasi identitas situs, informasi kontak, media sosial, dan SEO default.</p>
        </div>
        <button class="btn btn-primary" :disabled="isSaving" @click="saveSettings">
          <BladeIcon name="check" size="18" />
          <span>{{ isSaving ? 'Menyimpan...' : 'Simpan Perubahan' }}</span>
        </button>
      </div>

      <div v-if="isLoading" class="loading-state">Memuat pengaturan...</div>

      <div v-else class="settings-form">
        <!-- Site Identity Card -->
        <div class="card">
          <h3 class="card-title">Identitas Situs</h3>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>Nama Situs *</label>
              <input v-model="form.siteName" type="text" class="input-control" />
            </div>
            <div class="form-group flex-1">
              <label>Logo URL (dari Media Library)</label>
              <input v-model="form.logoUrl" type="text" class="input-control" placeholder="/media/..." />
            </div>
          </div>

          <div class="form-group">
            <label>Tagline (Bahasa Indonesia)</label>
            <input v-model="form.taglineId" type="text" class="input-control" />
          </div>

          <div class="form-group">
            <label>Tagline (English)</label>
            <input v-model="form.taglineEn" type="text" class="input-control" />
          </div>
        </div>

        <!-- Contact & Location Card -->
        <div class="card">
          <h3 class="card-title">Kontak & Lokasi Kantor</h3>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>Email Kontak Resmi *</label>
              <input v-model="form.contactEmail" type="email" class="input-control" />
            </div>
            <div class="form-group flex-1">
              <label>Nomor Telepon / WhatsApp</label>
              <input v-model="form.contactPhone" type="text" class="input-control" placeholder="+62 8..." />
            </div>
          </div>

          <div class="form-group">
            <label>Alamat Sekretariat</label>
            <textarea v-model="form.address" rows="2" class="textarea-control"></textarea>
          </div>
        </div>

        <!-- Social Media Links Card -->
        <div class="card">
          <h3 class="card-title">Akun Media Sosial Resmi</h3>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>Instagram URL</label>
              <input v-model="form.socialInstagram" type="url" class="input-control" placeholder="https://instagram.com/..." />
            </div>
            <div class="form-group flex-1">
              <label>Facebook URL</label>
              <input v-model="form.socialFacebook" type="url" class="input-control" placeholder="https://facebook.com/..." />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group flex-1">
              <label>YouTube URL</label>
              <input v-model="form.socialYoutube" type="url" class="input-control" placeholder="https://youtube.com/..." />
            </div>
            <div class="form-group flex-1">
              <label>X / Twitter URL</label>
              <input v-model="form.socialTwitter" type="url" class="input-control" placeholder="https://x.com/..." />
            </div>
          </div>
        </div>

        <!-- Default SEO Card -->
        <div class="card">
          <h3 class="card-title">Default SEO Metadata (Homepage & Fallback)</h3>
          <div class="form-group">
            <label>Default Title (ID)</label>
            <input v-model="form.defaultSeoTitleId" type="text" class="input-control" />
          </div>
          <div class="form-group">
            <label>Default Meta Description (ID)</label>
            <textarea v-model="form.defaultSeoDescId" rows="2" class="textarea-control"></textarea>
          </div>

          <div class="form-group">
            <label>Default Title (EN)</label>
            <input v-model="form.defaultSeoTitleEn" type="text" class="input-control" />
          </div>
          <div class="form-group">
            <label>Default Meta Description (EN)</label>
            <textarea v-model="form.defaultSeoDescEn" rows="2" class="textarea-control"></textarea>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.settings-page {
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

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 860px;
}

.card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.card-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: #0f172a;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 0.75rem;
}

.form-row { display: flex; gap: 1rem; }
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

.input-control, .textarea-control {
  padding: 0.55rem 0.75rem;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 0.9rem;
  font-family: inherit;
}

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

.loading-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
