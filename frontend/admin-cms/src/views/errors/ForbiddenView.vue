<script setup lang="ts">
import AppLayout from '../../components/layout/AppLayout.vue'
</script>

<template>
  <AppLayout>
    <div class="cms-error-page">
      <div class="error-box">
        <div class="error-badge">403 FORBIDDEN</div>
        <div class="error-icon">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#e11d48" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
            <line x1="9.5" y1="9.5" x2="14.5" y2="14.5"></line>
            <line x1="14.5" y1="9.5" x2="9.5" y2="14.5"></line>
          </svg>
        </div>
        <h1 class="error-heading">Akses Ditolak (403)</h1>
        <p class="error-desc">
          Akun Anda tidak memiliki izin atau peran (role) yang memadai untuk membuka menu ini. Hubungi Superadmin jika Anda memerlukan akses tambahan.
        </p>
        <div class="error-cta">
          <router-link to="/dashboard" class="btn btn-primary">
            Kembali ke Dasbor
          </router-link>
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
  background: #ffe4e6;
  color: #be123c;
  border: 1px solid #fecdd3;
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
  background: #fff1f2;
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
</style>
