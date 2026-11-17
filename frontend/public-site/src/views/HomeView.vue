<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import { CATEGORIES } from '../data/fixtures'

const route = useRoute()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))

const isId = computed(() => locale.value === 'id')

const title = computed(() =>
  isId.value
    ? 'PidieKab — Promosi Potensi & Wisata Kabupaten Pidie'
    : 'PidieKab — Discover Pidie Regency Culture & Tourism'
)

const description = computed(() =>
  isId.value
    ? 'Platform digital independen untuk menjelajahi potensi wisata, kuliner, dan budaya di Kabupaten Pidie, Aceh.'
    : 'An independent visitor discovery platform exploring tourism, culinary arts, and heritage of Pidie Regency, Aceh.'
)

const canonicalUrl = computed(() => `https://pidiekab.id/${locale.value}/`)

useHead({
  title: title.value,
  meta: [
    { name: 'description', content: description.value },
    { property: 'og:title', content: title.value },
    { property: 'og:description', content: description.value },
    { property: 'og:url', content: canonicalUrl.value },
    { property: 'og:type', content: 'website' },
  ],
  link: [
    { rel: 'canonical', href: canonicalUrl.value },
    { rel: 'alternate', hreflang: 'id', href: 'https://pidiekab.id/id/' },
    { rel: 'alternate', hreflang: 'en', href: 'https://pidiekab.id/en/' },
  ],
  htmlAttrs: {
    lang: locale.value,
  },
})
</script>

<template>
  <div class="page-container">
    <header class="navbar">
      <div class="brand">PidieKab</div>
      <nav class="lang-switch">
        <RouterLink :to="'/id/'" :class="{ active: isId }">Bahasa Indonesia</RouterLink>
        <span> | </span>
        <RouterLink :to="'/en/'" :class="{ active: !isId }">English</RouterLink>
      </nav>
    </header>

    <main class="hero">
      <h1 class="hero-title">{{ title }}</h1>
      <p class="hero-desc">{{ description }}</p>

      <section class="categories-section">
        <h2>{{ isId ? 'Kategori Eksplorasi' : 'Discovery Categories' }}</h2>
        <div class="category-cards">
          <article v-for="cat in CATEGORIES" :key="cat.id" class="card">
            <h3>{{ isId ? cat.translations.id.name : cat.translations.en.name }}</h3>
            <p>{{ isId ? cat.translations.id.description : cat.translations.en.description }}</p>
            <RouterLink
              :to="`/${locale}/${isId ? cat.translations.id.slug : cat.translations.en.slug}/`"
              class="link"
            >
              {{ isId ? 'Lihat Destinasi →' : 'Explore Destinations →' }}
            </RouterLink>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 960px;
  margin: 0 auto;
  padding: 1.5rem;
  font-family: system-ui, -apple-system, sans-serif;
  color: #1f2937;
}
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}
.brand {
  font-size: 1.5rem;
  font-weight: 700;
  color: #065f46;
}
.lang-switch a {
  color: #4b5563;
  text-decoration: none;
  font-size: 0.9rem;
}
.lang-switch a.active {
  font-weight: bold;
  color: #065f46;
}
.hero {
  margin-top: 2.5rem;
}
.hero-title {
  font-size: 2.25rem;
  color: #111827;
  margin-bottom: 1rem;
}
.hero-desc {
  font-size: 1.125rem;
  color: #4b5563;
  line-height: 1.6;
  margin-bottom: 2rem;
}
.categories-section h2 {
  font-size: 1.5rem;
  margin-bottom: 1rem;
}
.category-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
}
.card {
  padding: 1.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background-color: #fafafa;
}
.card h3 {
  margin-bottom: 0.5rem;
  color: #065f46;
}
.card p {
  color: #6b7280;
  margin-bottom: 1rem;
  font-size: 0.95rem;
}
.link {
  color: #047857;
  font-weight: 600;
  text-decoration: none;
}
</style>
