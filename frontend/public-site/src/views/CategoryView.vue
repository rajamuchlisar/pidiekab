<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import { CATEGORIES, CONTENTS } from '../data/fixtures'

const route = useRoute()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))
const isId = computed(() => locale.value === 'id')

// Find category matching current slug
const category = computed(() => {
  const slug = route.params.categorySlug as string
  return CATEGORIES.find(
    (c) => c.translations.id.slug === slug || c.translations.en.slug === slug
  )
})

const catTranslation = computed(() => {
  if (!category.value) return null
  return isId.value ? category.value.translations.id : category.value.translations.en
})

const otherCatTranslation = computed(() => {
  if (!category.value) return null
  return isId.value ? category.value.translations.en : category.value.translations.id
})

const items = computed(() => {
  if (!category.value) return []
  return CONTENTS.filter(
    (item) =>
      item.categoryId === category.value?.id &&
      (isId.value
        ? item.translations.id.status === 'published'
        : item.translations.en.status === 'published')
  )
})

const pageTitle = computed(() =>
  catTranslation.value
    ? `${catTranslation.value.name} — PidieKab`
    : 'Kategori Tidak Ditemukan'
)

const pageDesc = computed(() =>
  catTranslation.value
    ? catTranslation.value.description
    : ''
)

const canonicalUrl = computed(() =>
  catTranslation.value
    ? `https://pidiekab.id/${locale.value}/${catTranslation.value.slug}/`
    : ''
)

const alternateUrl = computed(() =>
  otherCatTranslation.value
    ? `https://pidiekab.id/${isId.value ? 'en' : 'id'}/${otherCatTranslation.value.slug}/`
    : ''
)

useHead({
  title: pageTitle.value,
  meta: [
    { name: 'description', content: pageDesc.value },
    { property: 'og:title', content: pageTitle.value },
    { property: 'og:description', content: pageDesc.value },
    { property: 'og:url', content: canonicalUrl.value },
  ],
  link: [
    { rel: 'canonical', href: canonicalUrl.value },
    { rel: 'alternate', hreflang: locale.value, href: canonicalUrl.value },
    ...(alternateUrl.value
      ? [{ rel: 'alternate', hreflang: isId.value ? 'en' : 'id', href: alternateUrl.value }]
      : []),
  ],
  htmlAttrs: {
    lang: locale.value,
  },
})
</script>

<template>
  <div class="page-container">
    <nav class="breadcrumb">
      <RouterLink :to="`/${locale}/`">{{ isId ? 'Beranda' : 'Home' }}</RouterLink>
      <span> / </span>
      <span>{{ catTranslation?.name }}</span>
    </nav>

    <header class="cat-header" v-if="catTranslation">
      <h1>{{ catTranslation.name }}</h1>
      <p>{{ catTranslation.description }}</p>
    </header>

    <main class="item-list">
      <div v-if="items.length === 0" class="empty">
        {{ isId ? 'Belum ada konten published.' : 'No published content yet.' }}
      </div>
      <div v-else class="cards">
        <article v-for="item in items" :key="item.id" class="card">
          <h2>{{ isId ? item.translations.id.title : item.translations.en.title }}</h2>
          <p>{{ isId ? item.translations.id.description : item.translations.en.description }}</p>
          <RouterLink
            :to="`/${locale}/${catTranslation?.slug}/${isId ? item.translations.id.slug : item.translations.en.slug}/`"
            class="read-more"
          >
            {{ isId ? 'Baca Selengkapnya →' : 'Read Detail →' }}
          </RouterLink>
        </article>
      </div>
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
.breadcrumb {
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  color: #6b7280;
}
.breadcrumb a {
  color: #065f46;
  text-decoration: none;
}
.cat-header {
  margin-bottom: 2rem;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 1rem;
}
.cat-header h1 {
  font-size: 2rem;
  color: #111827;
  margin-bottom: 0.5rem;
}
.cat-header p {
  font-size: 1.1rem;
  color: #4b5563;
}
.cards {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.card {
  padding: 1.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #ffffff;
}
.card h2 {
  font-size: 1.35rem;
  margin-bottom: 0.5rem;
  color: #065f46;
}
.card p {
  color: #4b5563;
  margin-bottom: 1rem;
  line-height: 1.5;
}
.read-more {
  color: #047857;
  font-weight: 600;
  text-decoration: none;
}
</style>
