<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import { CATEGORIES, CONTENTS } from '../data/fixtures'

const route = useRoute()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))
const isId = computed(() => locale.value === 'id')

const categorySlug = computed(() => route.params.categorySlug as string)
const contentSlug = computed(() => route.params.contentSlug as string)

// Find category
const category = computed(() =>
  CATEGORIES.find(
    (c) =>
      c.translations.id.slug === categorySlug.value ||
      c.translations.en.slug === categorySlug.value
  )
)

// Find content item
const content = computed(() =>
  CONTENTS.find(
    (item) =>
      item.translations.id.slug === contentSlug.value ||
      item.translations.en.slug === contentSlug.value
  )
)

const activeTranslation = computed(() => {
  if (!content.value) return null
  return isId.value ? content.value.translations.id : content.value.translations.en
})

const counterpartTranslation = computed(() => {
  if (!content.value) return null
  return isId.value ? content.value.translations.en : content.value.translations.id
})

// Hreflang is strictly allowed only if both ID and EN translations are published
const isBothPublished = computed(() => {
  return (
    content.value?.translations.id.status === 'published' &&
    content.value?.translations.en.status === 'published'
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

const pageTitle = computed(() =>
  activeTranslation.value
    ? `${activeTranslation.value.title} — PidieKab`
    : 'Halaman Tidak Ditemukan'
)

const pageDesc = computed(() =>
  activeTranslation.value ? activeTranslation.value.description : ''
)

const canonicalUrl = computed(() =>
  activeTranslation.value && catTranslation.value
    ? `https://pidiekab.id/${locale.value}/${catTranslation.value.slug}/${activeTranslation.value.slug}/`
    : ''
)

const alternateUrl = computed(() => {
  if (!isBothPublished.value || !counterpartTranslation.value || !otherCatTranslation.value) {
    return ''
  }
  const altLocale = isId.value ? 'en' : 'id'
  return `https://pidiekab.id/${altLocale}/${otherCatTranslation.value.slug}/${counterpartTranslation.value.slug}/`
})

useHead({
  title: pageTitle.value,
  meta: [
    { name: 'description', content: pageDesc.value },
    { property: 'og:title', content: pageTitle.value },
    { property: 'og:description', content: pageDesc.value },
    { property: 'og:url', content: canonicalUrl.value },
    { property: 'og:type', content: 'article' },
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
      <RouterLink :to="`/${locale}/${catTranslation?.slug}/`">{{ catTranslation?.name }}</RouterLink>
      <span> / </span>
      <span>{{ activeTranslation?.title }}</span>
    </nav>

    <div v-if="alternateUrl" class="lang-switch-box">
      <span>{{ isId ? 'Tersedia dalam Bahasa Inggris:' : 'Available in Indonesian:' }} </span>
      <a :href="alternateUrl">{{ isId ? 'Read in English' : 'Baca dalam Bahasa Indonesia' }}</a>
    </div>

    <article v-if="activeTranslation" class="content-detail">
      <header class="detail-header">
        <h1>{{ activeTranslation.title }}</h1>
        <p class="summary">{{ activeTranslation.description }}</p>
      </header>

      <div class="article-body">
        <p>{{ activeTranslation.body }}</p>
      </div>
    </article>

    <div v-else class="not-found">
      <h1>{{ isId ? 'Konten Tidak Ditemukan' : 'Content Not Found' }}</h1>
      <p>{{ isId ? 'Konten ini tidak tersedia atau belum dipublikasikan.' : 'This content is not available or unpublished.' }}</p>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  max-width: 860px;
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
.lang-switch-box {
  background-color: #ecfdf5;
  border-left: 4px solid #10b981;
  padding: 0.75rem 1rem;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}
.lang-switch-box a {
  color: #047857;
  font-weight: 600;
  text-decoration: underline;
}
.detail-header {
  margin-bottom: 2rem;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 1.5rem;
}
.detail-header h1 {
  font-size: 2.25rem;
  color: #111827;
  line-height: 1.25;
  margin-bottom: 0.75rem;
}
.summary {
  font-size: 1.2rem;
  color: #4b5563;
  line-height: 1.5;
}
.article-body {
  font-size: 1.1rem;
  line-height: 1.75;
  color: #374151;
}
.not-found {
  padding: 3rem 0;
  text-align: center;
}
</style>
