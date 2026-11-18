<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useNotificationStore } from '../../stores/notification'

const route = useRoute()
const authStore = useAuthStore()
const notificationStore = useNotificationStore()

const currentRouteName = computed(() => route.name)

async function handleLogout() {
  if (confirm('Apakah Anda yakin ingin keluar dari Admin CMS?')) {
    await authStore.logout()
  }
}
</script>

<template>
  <div class="cms-shell">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-logo">PK</div>
        <div class="brand-info">
          <span class="brand-title">PidieKab</span>
          <span class="brand-sub">Admin CMS v79</span>
        </div>
      </div>

      <nav class="nav-menu">
        <router-link to="/dashboard" class="nav-item" :class="{ active: currentRouteName === 'dashboard' }">
          <span class="nav-icon">📊</span>
          <span class="nav-label">Dashboard</span>
        </router-link>

        <router-link to="/contents" class="nav-item" :class="{ active: String(currentRouteName).startsWith('content') }">
          <span class="nav-icon">📝</span>
          <span class="nav-label">Kelola Konten</span>
        </router-link>

        <router-link to="/categories" class="nav-item" :class="{ active: currentRouteName === 'categories' }">
          <span class="nav-icon">🏷️</span>
          <span class="nav-label">Kategori</span>
        </router-link>

        <router-link to="/media" class="nav-item" :class="{ active: currentRouteName === 'media' }">
          <span class="nav-icon">🖼️</span>
          <span class="nav-label">Media Library</span>
        </router-link>

        <router-link to="/curation" class="nav-item" :class="{ active: currentRouteName === 'curation' }">
          <span class="nav-icon">⭐</span>
          <span class="nav-label">Kurasi Beranda</span>
        </router-link>

        <router-link to="/static-pages" class="nav-item" :class="{ active: currentRouteName === 'static-pages' }">
          <span class="nav-icon">📄</span>
          <span class="nav-label">Halaman Statis</span>
        </router-link>

        <div class="nav-divider"></div>

        <router-link to="/settings" class="nav-item" :class="{ active: currentRouteName === 'settings' }">
          <span class="nav-icon">⚙️</span>
          <span class="nav-label">Pengaturan Web</span>
        </router-link>

        <router-link to="/admins" class="nav-item" :class="{ active: currentRouteName === 'admins' }">
          <span class="nav-icon">👥</span>
          <span class="nav-label">Administrator</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <a href="/" target="_blank" class="public-link">
          <span>🌐</span> Buka Situs Publik
        </a>
      </div>
    </aside>

    <!-- Main Container -->
    <div class="main-wrapper">
      <!-- Header -->
      <header class="topbar">
        <div class="topbar-left">
          <h1 class="page-title">{{ route.meta?.title || 'Admin Panel' }}</h1>
        </div>
        <div class="topbar-right">
          <div class="admin-badge">
            <div class="admin-avatar">{{ authStore.user?.displayName?.charAt(0) || 'A' }}</div>
            <div class="admin-details">
              <span class="admin-name">{{ authStore.user?.displayName || 'Administrator' }}</span>
              <span class="admin-email">{{ authStore.user?.email || 'admin@pidiekab.id' }}</span>
            </div>
          </div>
          <button class="btn-logout" @click="handleLogout" title="Keluar">
            Keluar
          </button>
        </div>
      </header>

      <!-- Page Body -->
      <main class="page-content">
        <slot />
      </main>
    </div>

    <!-- Toast Notifications -->
    <div class="toast-container" v-if="notificationStore.toasts.length > 0">
      <div
        v-for="toast in notificationStore.toasts"
        :key="toast.id"
        class="toast-item"
        :class="toast.type"
      >
        <div class="toast-header">
          <strong class="toast-title">{{ toast.title }}</strong>
          <button class="toast-close" @click="notificationStore.remove(toast.id)">&times;</button>
        </div>
        <p class="toast-msg">{{ toast.message }}</p>
        <span v-if="toast.requestId" class="toast-req-id">Ref ID: {{ toast.requestId }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cms-shell {
  display: flex;
  min-height: 100vh;
  background-color: #f8fafc;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  color: #1e293b;
}

/* Sidebar */
.sidebar {
  width: 250px;
  background-color: #064e3b; /* Deep Pidie Evergreen */
  color: #ecfdf5;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
}

.brand {
  padding: 1.5rem 1.25rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.brand-logo {
  width: 38px;
  height: 38px;
  background: #10b981;
  color: #064e3b;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-size: 1.1rem;
}

.brand-info {
  display: flex;
  flex-direction: column;
}

.brand-title {
  font-weight: 700;
  font-size: 1.15rem;
  letter-spacing: -0.02em;
}

.brand-sub {
  font-size: 0.75rem;
  opacity: 0.7;
}

.nav-menu {
  flex: 1;
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  color: #d1fae5;
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.nav-item:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.nav-item.active {
  background-color: #059669;
  color: #ffffff;
  font-weight: 600;
}

.nav-icon {
  font-size: 1.1rem;
}

.nav-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.1);
  margin: 0.75rem 0.5rem;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.public-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #a7f3d0;
  text-decoration: none;
  font-size: 0.85rem;
  padding: 0.5rem;
  border-radius: 6px;
  transition: background 0.15s;
}

.public-link:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

/* Main Wrapper */
.main-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  background-color: #ffffff;
  height: 64px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
}

.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.admin-badge {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.admin-avatar {
  width: 34px;
  height: 34px;
  background: #064e3b;
  color: #ffffff;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 0.9rem;
}

.admin-details {
  display: flex;
  flex-direction: column;
}

.admin-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: #1e293b;
}

.admin-email {
  font-size: 0.75rem;
  color: #64748b;
}

.btn-logout {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fecaca;
  padding: 0.4rem 0.85rem;
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-logout:hover {
  background: #fca5a5;
  color: #7f1d1d;
}

.page-content {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
}

/* Toast Container */
.toast-container {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  z-index: 9999;
  max-width: 360px;
}

.toast-item {
  padding: 0.85rem 1rem;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  border-left: 4px solid #64748b;
}

.toast-item.success {
  border-left-color: #10b981;
}

.toast-item.error {
  border-left-color: #ef4444;
}

.toast-item.warning {
  border-left-color: #f59e0b;
}

.toast-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toast-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: #0f172a;
}

.toast-close {
  background: transparent;
  border: none;
  font-size: 1.1rem;
  cursor: pointer;
  color: #94a3b8;
}

.toast-msg {
  font-size: 0.825rem;
  color: #334155;
  margin: 0.25rem 0 0 0;
}

.toast-req-id {
  display: block;
  font-size: 0.7rem;
  color: #94a3b8;
  margin-top: 0.35rem;
  font-family: monospace;
}
</style>
