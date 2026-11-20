<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))
const isId = computed(() => locale.value === 'id')

const mobileMenuOpen = ref(false)

const navLinks = computed(() => {
  if (isId.value) {
    return [
      { name: 'Beranda', path: '/id/' },
      { name: 'Wisata & Destinasi', path: '/id/wisata-destinasi/' },
      { name: 'Kuliner & UMKM', path: '/id/kuliner-umkm/' },
      { name: 'Budaya & Sejarah', path: '/id/budaya-sejarah/' },
    ]
  }
  return [
    { name: 'Home', path: '/en/' },
    { name: 'Tourism', path: '/en/tourism-destinations/' },
    { name: 'Culinary', path: '/en/culinary-local-business/' },
    { name: 'Culture & History', path: '/en/culture-history/' },
  ]
})

// Counterpart route resolution
const counterpartUrl = computed(() => {
  const p = route.path
  if (isId.value) {
    if (p.includes('/wisata-destinasi/contoh-konten/')) return '/en/tourism-destinations/sample-content/'
    if (p.includes('/wisata-destinasi/')) return '/en/tourism-destinations/'
    return '/en/'
  } else {
    if (p.includes('/tourism-destinations/sample-content/')) return '/id/wisata-destinasi/contoh-konten/'
    if (p.includes('/tourism-destinations/')) return '/id/wisata-destinasi/'
    return '/id/'
  }
})
</script>

<template>
  <header class="site-header">
    <div class="header-inner">
      <!-- Logo / Brand -->
      <RouterLink :to="`/${locale}/`" class="site-brand">
        <span class="brand-badge">PK</span>
        <div class="brand-text">
          <span class="brand-name">PidieKab</span>
        </div>
      </RouterLink>

      <!-- Desktop Nav -->
      <nav class="desktop-nav">
        <RouterLink
          v-for="item in navLinks"
          :key="item.path"
          :to="item.path"
          class="nav-link"
          :class="{ active: route.path === item.path }"
        >
          {{ item.name }}
        </RouterLink>
      </nav>

      <!-- Right Controls: Language Switcher & Mobile Toggle -->
      <div class="header-right">
        <div class="lang-switcher">
          <RouterLink
            :to="isId ? route.path : counterpartUrl"
            class="lang-btn"
            :class="{ active: isId }"
          >
            ID
          </RouterLink>
          <span class="lang-divider">/</span>
          <RouterLink
            :to="!isId ? route.path : counterpartUrl"
            class="lang-btn"
            :class="{ active: !isId }"
          >
            EN
          </RouterLink>
        </div>

        <button
          class="mobile-toggle"
          @click="mobileMenuOpen = !mobileMenuOpen"
          :aria-expanded="mobileMenuOpen"
          aria-label="Menu"
        >
          <span class="hamburger-bar"></span>
          <span class="hamburger-bar"></span>
          <span class="hamburger-bar"></span>
        </button>
      </div>
    </div>

    <!-- Mobile Drawer -->
    <div v-if="mobileMenuOpen" class="mobile-drawer">
      <nav class="mobile-nav">
        <RouterLink
          v-for="item in navLinks"
          :key="item.path"
          :to="item.path"
          class="mobile-nav-link"
          @click="mobileMenuOpen = false"
        >
          {{ item.name }}
        </RouterLink>
      </nav>
    </div>
  </header>
</template>

<style scoped>
.site-header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid #e2e8f0;
  height: 65px;
  box-sizing: border-box;
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1.5rem;
  height: 65px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
}

.site-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: #064e3b;
}

.brand-badge {
  width: 38px;
  height: 38px;
  background: #064e3b;
  color: #10b981;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-size: 1.1rem;
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: #064e3b;
  line-height: 1.1;
}

.brand-tagline {
  font-size: 0.725rem;
  color: #64748b;
  font-weight: 500;
}

.desktop-nav {
  display: flex;
  gap: 1.5rem;
}

.nav-link {
  color: #334155;
  text-decoration: none;
  font-size: 0.925rem;
  font-weight: 500;
  padding: 0.35rem 0;
  position: relative;
  transition: color 0.15s;
}

.nav-link:hover, .nav-link.active {
  color: #064e3b;
  font-weight: 600;
}

.nav-link.active::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 0;
  right: 0;
  height: 2px;
  background: #059669;
  border-radius: 2px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.lang-switcher {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  background: #f1f5f9;
  padding: 0.25rem 0.65rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.lang-btn {
  color: #64748b;
  text-decoration: none;
  transition: color 0.15s;
}

.lang-btn.active {
  color: #064e3b;
  font-weight: 700;
}

.lang-divider {
  color: #cbd5e1;
}

.mobile-toggle {
  display: none;
  background: none;
  border: none;
  flex-direction: column;
  gap: 4px;
  cursor: pointer;
  padding: 0.25rem;
}

.hamburger-bar {
  width: 22px;
  height: 2px;
  background: #0f172a;
  border-radius: 2px;
}

.mobile-drawer {
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  padding: 1rem 1.5rem;
}

.mobile-nav {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.mobile-nav-link {
  color: #1e293b;
  text-decoration: none;
  font-size: 1rem;
  font-weight: 500;
  padding: 0.5rem 0;
}

@media (max-width: 768px) {
  .desktop-nav {
    display: none;
  }
  .mobile-toggle {
    display: flex;
  }
}
</style>
