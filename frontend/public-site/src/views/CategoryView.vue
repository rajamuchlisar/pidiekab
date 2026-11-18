<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import { CATEGORIES, CONTENTS } from '../data/fixtures'
import SiteHeader from '../components/layout/SiteHeader.vue'
import SiteFooter from '../components/layout/SiteFooter.vue'

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
  <div class="site-layout">
    <SiteHeader />

    <main class="page-container">
      <nav class="breadcrumb">
        <RouterLink :to="`/${locale}/`">{{ isId ? 'Beranda' : 'Home' }}</RouterLink>
        <span> / </span>
        <span>{{ catTranslation?.name }}</span>
      </nav>

      <header class="cat-header" v-if="catTranslation">
        <span class="cat-tag">{{ isId ? 'Kategori Eksplorasi' : 'Discovery Category' }}</span>
        <h1>{{ catTranslation.name }}</h1>
        <p>{{ catTranslation.description }}</p>
      </header>

      <section class="item-list">
        <div v-if="items.length === 0" class="empty">
          <p>{{ isId ? 'Belum ada konten published pada kategori ini.' : 'No published content yet in this category.' }}</p>
        </div>

        <div v-else class="content-grid">
          <article v-for="item in items" :key="item.id" class="content-card">
            <div class="card-thumb">
              <div class="photo-placeholder">🏔️ Foto Terverifikasi</div>
            </div>
            <div class="card-content">
              <h3 class="card-title">{{ item.translations[locale].title }}</h3>
              <p class="card-desc">{{ item.translations[locale].description }}</p>
              <RouterLink
                :to="`/${locale}/${catTranslation?.slug}/${item.translations[locale].slug}/`"
                class="card-link"
              >
                {{ isId ? 'Baca Selengkapnya →' : 'Read Discovery →' }}
              </RouterLink>
            </div>
          </article>
        </div>
      </section>
    </main>

    <SiteFooter />
  </div>
</template>

<style scoped>
.site-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #fdfdfd;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  color: #1e293b;
}

.page-container {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 2.5rem 1.5rem;
  flex: 1;
}

.breadcrumb {
  margin-bottom: 1.5rem;
  font-size: 0.85rem;
  color: #64748b;
}

.breadcrumb a {
  color: #064e3b;
  text-decoration: none;
  font-weight: 500;
}

.breadcrumb a:hover {
  text-decoration: underline;
}

.cat-header {
  margin-bottom: 2.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.cat-tag {
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 700;
  color: #059669;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.35rem;
}

.cat-header h1 {
  font-size: 2.25rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 0.5rem 0;
  letter-spacing: -0.02em;
}

.cat-header p {
  font-size: 1.05rem;
  color: #64748b;
  margin: 0;
  line-height: 1.6;
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.content-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.card-thumb {
  height: 180px;
  background: linear-gradient(135deg, #064e3b 0%, #10b981 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-placeholder {
  color: #ffffff;
  font-size: 0.85rem;
  font-weight: 600;
  background: rgba(0, 0, 0, 0.25);
  padding: 0.35rem 0.75rem;
  border-radius: 20px;
}

.card-content {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.card-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.5rem 0;
  line-height: 1.35;
}

.card-desc {
  font-size: 0.875rem;
  line-height: 1.5;
  color: #64748b;
  margin: 0 0 1.25rem 0;
  flex: 1;
}

.card-link {
  color: #059669;
  font-weight: 600;
  text-decoration: none;
  font-size: 0.875rem;
}

.card-link:hover {
  text-decoration: underline;
}

.empty {
  padding: 4rem 1rem;
  text-align: center;
  color: #64748b;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px dashed #cbd5e1;
}
</style>
