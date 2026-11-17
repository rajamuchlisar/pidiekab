<script setup lang="ts">
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
</script>

<template>
  <div class="cms-dashboard">
    <header class="header">
      <h2>PidieKab CMS Dashboard</h2>
      <div>
        <span>{{ authStore.user?.displayName || 'Admin' }}</span>
      </div>
    </header>
    <main class="content">
      <p>Selamat datang di panel administrasi PidieKab (Phase 0 Foundation).</p>
      <ul>
        <li>Konten</li>
        <li>Kategori</li>
        <li>Media Library</li>
        <li>Pengaturan</li>
      </ul>
    </main>
  </div>
</template>

<style scoped>
.cms-dashboard {
  font-family: system-ui, -apple-system, sans-serif;
}
.header {
  display: flex;
  justify-content: space-between;
  padding: 1rem 2rem;
  background: #064e3b;
  color: #fff;
}
.content {
  padding: 2rem;
}
</style>
