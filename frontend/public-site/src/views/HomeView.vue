<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import { CATEGORIES, CONTENTS } from '../data/fixtures'
import SiteHeader from '../components/layout/SiteHeader.vue'
import SiteFooter from '../components/layout/SiteFooter.vue'

const route = useRoute()
const router = useRouter()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))
const isId = computed(() => locale.value === 'id')

const searchQuery = ref('')

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

// Published contents only
const publishedContents = computed(() => {
  return CONTENTS.filter((c) => c.translations[locale.value]?.status === 'published')
})

function onSearch() {
  if (searchQuery.value.trim()) {
    router.push({
      path: `/${locale.value}/search/`,
      query: { q: searchQuery.value.trim() },
    })
  }
}
</script>

<template>
  <div class="site-layout">
    <SiteHeader />

    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content">
        <span class="hero-tag">
          {{ isId ? '📍 Eksplorasi Daerah Pidie' : '📍 Explore Pidie Region' }}
        </span>
        <h1 class="hero-title">{{ title }}</h1>
        <p class="hero-desc">{{ description }}</p>

        <!-- Global Search Bar in Hero -->
        <form @submit.prevent="onSearch" class="hero-search-form">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="isId ? 'Cari destinasi, kuliner khas, atau sejarah...' : 'Search destinations, cuisine, or heritage...'"
            class="hero-search-input"
          />
          <button type="submit" class="hero-search-btn">
            {{ isId ? 'Cari' : 'Search' }}
          </button>
        </form>
      </div>
    </section>

    <!-- Main Container -->
    <main class="page-container">
      <!-- Categories Section -->
      <section class="categories-section">
        <div class="section-heading">
          <h2>{{ isId ? 'Kategori Eksplorasi' : 'Discovery Categories' }}</h2>
          <p class="section-sub">
            {{ isId
              ? 'Tiga pilar keindahan dan potensi lokal Kabupaten Pidie'
              : 'Three core pillars of natural and cultural potential of Pidie'
            }}
          </p>
        </div>

        <div class="category-cards">
          <article v-for="cat in CATEGORIES" :key="cat.id" class="cat-card">
            <div class="cat-card-body">
              <span class="cat-icon">🌿</span>
              <h3>{{ isId ? cat.translations.id.name : cat.translations.en.name }}</h3>
              <p>{{ isId ? cat.translations.id.description : cat.translations.en.description }}</p>
              <RouterLink
                :to="`/${locale}/${isId ? cat.translations.id.slug : cat.translations.en.slug}/`"
                class="link"
              >
                {{ isId ? 'Lihat Destinasi →' : 'Explore Destinations →' }}
              </RouterLink>
            </div>
          </article>
        </div>
      </section>

      <!-- Featured / Discovery Grid -->
      <section class="discovery-section">
        <div class="section-heading">
          <h2>{{ isId ? 'Jelajahi Pidie' : 'Explore Pidie' }}</h2>
          <p class="section-sub">
            {{ isId
              ? 'Destinasi dan daya tarik terpilih yang wajib Anda kunjungi'
              : 'Selected destinations and cultural highlights you must experience'
            }}
          </p>
        </div>

        <div class="content-grid">
          <article v-for="item in publishedContents" :key="item.id" class="content-card">
            <div class="card-thumb">
              <div class="photo-placeholder">🏔️ Foto Terverifikasi</div>
            </div>
            <div class="card-content">
              <span class="card-cat">
                {{ isId ? 'Wisata & Destinasi' : 'Tourism & Destinations' }}
              </span>
              <h3 class="card-title">{{ item.translations[locale].title }}</h3>
              <p class="card-desc">{{ item.translations[locale].description }}</p>
              <RouterLink
                :to="`/${locale}/${isId ? 'wisata-destinasi' : 'tourism-destinations'}/${item.translations[locale].slug}/`"
                class="card-link"
              >
                {{ isId ? 'Baca Selengkapnya →' : 'Read Discovery →' }}
              </RouterLink>
            </div>
          </article>
        </div>
      </section>

      <!-- Story Section -->
      <section class="story-section">
        <div class="story-card">
          <div class="story-badge">
            {{ isId ? 'CERITA TENTANG PIDIE' : 'STORY OF PIDIE' }}
          </div>
          <h2 class="story-title">
            {{ isId
              ? 'Menyusuri Keindahan Alam & Warisan Sejarah Tanah Rencong'
              : 'Unveiling the Scenic Landscapes & Historic Legacy of Pidie'
            }}
          </h2>
          <p class="story-text">
            {{ isId
              ? 'Pidie memiliki bentang alam yang memadukan pesisir pantai Selat Malaka, lembah pegunungan Tangse yang subur, serta kekayaan kuliner khas seperti meuseukat dan emping melinjo.'
              : 'Pidie boasts diverse topography spanning from the Malacca Strait coastal beaches to the lush mountain valleys of Tangse, accompanied by authentic culinary traditions.'
            }}
          </p>
          <RouterLink
            :to="`/${locale}/${isId ? 'wisata-destinasi/contoh-konten/' : 'tourism-destinations/sample-content/'}`"
            class="story-cta"
          >
            {{ isId ? 'Mulai Eksplorasi Sekarang' : 'Start Discovery Now' }}
          </RouterLink>
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

/* Hero Section */
.hero-section {
  background: linear-gradient(135deg, #064e3b 0%, #022c22 100%);
  color: #ffffff;
  padding: 4.5rem 1.5rem 4rem 1.5rem;
  text-align: center;
}

.hero-content {
  max-width: 820px;
  margin: 0 auto;
}

.hero-tag {
  display: inline-block;
  background: rgba(16, 185, 129, 0.2);
  color: #a7f3d0;
  border: 1px solid rgba(16, 185, 129, 0.35);
  font-size: 0.8rem;
  font-weight: 600;
  padding: 0.3rem 0.75rem;
  border-radius: 20px;
  margin-bottom: 1.25rem;
}

.hero-title {
  font-size: 2.25rem;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: -0.02em;
  margin: 0 0 1rem 0;
}

@media (max-width: 640px) {
  .hero-title {
    font-size: 1.75rem;
  }
}

.hero-desc {
  font-size: 1.1rem;
  line-height: 1.6;
  color: #d1fae5;
  margin: 0 0 2rem 0;
  opacity: 0.95;
}

.hero-search-form {
  display: flex;
  max-width: 560px;
  margin: 0 auto;
  background: #ffffff;
  border-radius: 10px;
  padding: 0.35rem;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
}

.hero-search-input {
  flex: 1;
  border: none;
  padding: 0.75rem 1rem;
  font-size: 0.95rem;
  border-radius: 8px;
  outline: none;
  color: #0f172a;
}

.hero-search-btn {
  background: #059669;
  color: #ffffff;
  border: none;
  padding: 0.75rem 1.5rem;
  font-weight: 600;
  font-size: 0.95rem;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
}

.hero-search-btn:hover {
  background: #047857;
}

/* Page Container */
.page-container {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 3rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 3.5rem;
  flex: 1;
}

.section-heading {
  margin-bottom: 1.75rem;
}

.section-heading h2 {
  font-size: 1.65rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.35rem 0;
}

.section-sub {
  font-size: 0.95rem;
  color: #64748b;
  margin: 0;
}

/* Category Cards */
.category-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
}

.cat-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.75rem;
  transition: transform 0.2s, box-shadow 0.2s;
  display: flex;
  flex-direction: column;
}

.cat-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 20px -5px rgba(0, 0, 0, 0.08);
  border-color: #cbd5e1;
}

.cat-icon {
  font-size: 1.75rem;
  margin-bottom: 0.75rem;
  display: block;
}

.cat-card h3 {
  font-size: 1.25rem;
  font-weight: 700;
  color: #064e3b;
  margin: 0 0 0.5rem 0;
}

.cat-card p {
  font-size: 0.9rem;
  line-height: 1.5;
  color: #475569;
  margin: 0 0 1.25rem 0;
  flex: 1;
}

.link {
  color: #059669;
  font-weight: 600;
  text-decoration: none;
  font-size: 0.9rem;
  display: inline-flex;
  align-items: center;
}

.link:hover {
  text-decoration: underline;
}

/* Discovery Grid */
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
  background: linear-gradient(135deg, #10b981 0%, #064e3b 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-placeholder {
  color: #ffffff;
  font-size: 0.9rem;
  font-weight: 600;
  background: rgba(0, 0, 0, 0.2);
  padding: 0.4rem 0.85rem;
  border-radius: 20px;
}

.card-content {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.card-cat {
  font-size: 0.75rem;
  font-weight: 700;
  color: #059669;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.35rem;
}

.card-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.65rem 0;
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

/* Story Section */
.story-section {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 14px;
  padding: 2.5rem;
}

.story-card {
  max-width: 760px;
}

.story-badge {
  font-size: 0.75rem;
  font-weight: 800;
  color: #065f46;
  letter-spacing: 0.08em;
  margin-bottom: 0.75rem;
}

.story-title {
  font-size: 1.5rem;
  font-weight: 800;
  color: #064e3b;
  margin: 0 0 0.75rem 0;
  line-height: 1.3;
}

.story-text {
  font-size: 1rem;
  line-height: 1.6;
  color: #334155;
  margin: 0 0 1.5rem 0;
}

.story-cta {
  display: inline-block;
  background: #064e3b;
  color: #ffffff;
  padding: 0.7rem 1.35rem;
  font-weight: 600;
  font-size: 0.9rem;
  border-radius: 8px;
  text-decoration: none;
  transition: background 0.15s;
}

.story-cta:hover {
  background: #047857;
}
</style>
