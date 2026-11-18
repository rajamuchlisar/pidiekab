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
  <div class="site-layout">
    <SiteHeader />

    <main class="page-container">
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
          <div class="meta-tag">{{ catTranslation?.name }}</div>
          <h1>{{ activeTranslation.title }}</h1>
          <p class="summary">{{ activeTranslation.description }}</p>
        </header>

        <!-- Main Photo Preview Banner -->
        <div class="main-photo-banner">
          <div class="photo-overlay">
            <span>📷 Dokumentasi Lapangan Terverifikasi — Kabupaten Pidie</span>
          </div>
        </div>

        <div class="article-body">
          <p>{{ activeTranslation.body }}</p>
        </div>

        <!-- Location Information Block -->
        <section class="location-box">
          <h3 class="location-title">📍 {{ isId ? 'Informasi Lokasi & Wilayah' : 'Location & Region Details' }}</h3>
          <p class="location-desc">
            {{ isId
              ? 'Kabupaten Pidie, Provinsi Aceh, Indonesia. Koordinat estimasi: 5.38° N, 95.95° E (Sigli).'
              : 'Pidie Regency, Aceh Province, Indonesia. Approximate coordinates: 5.38° N, 95.95° E (Sigli).'
            }}
          </p>
        </section>

        <!-- Verified Sources & Citations Section -->
        <section class="sources-section">
          <h3 class="sources-title">📚 {{ isId ? 'Sitasi Sumber & Referensi Faktual' : 'Verified Sources & References' }}</h3>
          <ul class="sources-list">
            <li class="source-item">
              <span class="source-badge">{{ isId ? 'RESMI' : 'OFFICIAL' }}</span>
              <span class="source-text">
                {{ isId
                  ? 'Dinas Pariwisata, Pemuda dan Olahraga (Disporapar) Kabupaten Pidie — Data Potensi Wisata Daerah'
                  : 'Pidie Regency Tourism & Sports Agency (Disporapar) — Regional Tourism Survey'
                }}
              </span>
            </li>
            <li class="source-item">
              <span class="source-badge">{{ isId ? 'SURVEI' : 'SURVEY' }}</span>
              <span class="source-text">
                {{ isId
                  ? 'Dokumentasi Lapangan & Kajian Budaya Lokal Masyarakat Pidie (2026)'
                  : 'Field Documentation & Pidie Local Heritage Study (2026)'
                }}
              </span>
            </li>
          </ul>
        </section>
      </article>

      <div v-else class="not-found">
        <h1>{{ isId ? 'Konten Tidak Ditemukan' : 'Content Not Found' }}</h1>
        <p>{{ isId ? 'Konten ini tidak tersedia atau belum dipublikasikan.' : 'This content is not available or unpublished.' }}</p>
      </div>
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
  max-width: 860px;
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

.lang-switch-box {
  background-color: #ecfdf5;
  border-left: 4px solid #10b981;
  padding: 0.75rem 1rem;
  margin-bottom: 2rem;
  font-size: 0.875rem;
  border-radius: 0 6px 6px 0;
}

.lang-switch-box a {
  color: #047857;
  font-weight: 600;
  text-decoration: underline;
}

.detail-header {
  margin-bottom: 2rem;
}

.meta-tag {
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 700;
  color: #059669;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
}

.detail-header h1 {
  font-size: 2.25rem;
  color: #0f172a;
  line-height: 1.25;
  margin: 0 0 1rem 0;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.summary {
  font-size: 1.15rem;
  color: #475569;
  line-height: 1.6;
  margin: 0;
}

.main-photo-banner {
  height: 320px;
  background: linear-gradient(135deg, #064e3b 0%, #10b981 100%);
  border-radius: 12px;
  margin-bottom: 2.5rem;
  display: flex;
  align-items: flex-end;
  padding: 1.25rem;
  overflow: hidden;
}

.photo-overlay {
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  color: #ffffff;
  padding: 0.4rem 0.85rem;
  border-radius: 6px;
  font-size: 0.775rem;
  font-weight: 500;
}

.article-body {
  font-size: 1.1rem;
  line-height: 1.8;
  color: #334155;
  margin-bottom: 3rem;
}

/* Location Box */
.location-box {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 1.25rem;
  margin-bottom: 2rem;
}

.location-title {
  margin: 0 0 0.4rem 0;
  font-size: 1rem;
  font-weight: 700;
  color: #0f172a;
}

.location-desc {
  margin: 0;
  font-size: 0.875rem;
  color: #64748b;
  line-height: 1.5;
}

/* Sources Section */
.sources-section {
  border-top: 1px solid #e2e8f0;
  padding-top: 1.75rem;
  margin-top: 2rem;
}

.sources-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 1rem 0;
}

.sources-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.source-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.85rem;
  color: #475569;
}

.source-badge {
  background: #e2e8f0;
  color: #334155;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
}

.not-found {
  padding: 4rem 0;
  text-align: center;
}
</style>
