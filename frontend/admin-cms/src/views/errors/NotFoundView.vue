<script setup lang="ts">
import AppLayout from '../../components/layout/AppLayout.vue'
</script>

<template>
  <AppLayout>
    <div class="cms-error-page">
      <div class="error-box">
        <div class="error-badge">404 NOT FOUND</div>
        <div class="error-icon">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#047857" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon>
          </svg>
        </div>
        <h1 class="error-heading">Halaman Admin Tidak Ditemukan</h1>
        <p class="error-desc">
          Menu atau tautan halaman CMS yang Anda tuju tidak tersedia atau salah. Silakan kembali ke dasbor operasional.
        </p>
        <div class="error-cta">
          <router-link to="/dashboard" class="btn btn-primary">
            Kembali ke Dasbor
          </router-link>
          <a href="/id/" target="_blank" class="btn btn-secondary">
            Buka Website Publik ↗
          </a>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.cms-error-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 120px);
  padding: 2rem;
}
.error-box {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 3rem 2rem;
  max-width: 520px;
  width: 100%;
  text-align: center;
  box-shadow: 0 4px 16px -2px rgba(15, 23, 42, 0.05);
}
.error-badge {
  display: inline-block;
  padding: 0.35rem 0.85rem;
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  margin-bottom: 1.25rem;
}
.error-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  background: #f0fdf4;
  border-radius: 20px;
  margin-bottom: 1.25rem;
}
.error-heading {
  font-size: 1.5rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 0.75rem;
}
.error-desc {
  font-size: 0.95rem;
  color: #64748b;
  line-height: 1.6;
  margin: 0 0 2rem;
}
.error-cta {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}
.btn {
  display: inline-flex;
  align-items: center;
  padding: 0.7rem 1.4rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
}
.btn-primary {
  background: #047857;
  color: #ffffff;
  border: 1px solid #047857;
}
.btn-primary:hover {
  background: #065f46;
}
.btn-secondary {
  background: #f8fafc;
  color: #334155;
  border: 1px solid #cbd5e1;
}
.btn-secondary:hover {
  background: #f1f5f9;
}
</style>
