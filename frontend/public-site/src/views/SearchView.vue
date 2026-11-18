<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useHead } from '@unhead/vue'
import SiteHeader from '../components/layout/SiteHeader.vue'
import SiteFooter from '../components/layout/SiteFooter.vue'

const route = useRoute()
const router = useRouter()
const locale = computed<'id' | 'en'>(() => (route.path.startsWith('/en') ? 'en' : 'id'))
const isId = computed(() => locale.value === 'id')

const query = ref((route.query.q as string) || '')
const selectedCategory = ref((route.query.category as string) || '')
const results = ref<any[]>([])
const isLoading = ref(false)
const totalResults = ref(0)

const pageTitle = computed(() =>
  isId.value
    ? `Pencarian: ${query.value || 'Semua Destinasi'} — PidieKab`
    : `Search: ${query.value || 'All Destinations'} — PidieKab`
)

useHead({
  title: pageTitle.value,
  meta: [
    { name: 'robots', content: 'noindex, follow' },
  ],
})

async function performSearch() {
  if (!query.value.trim()) {
    results.value = []
    totalResults.value = 0
    return
  }

  isLoading.value = true
  try {
    const params = new URLSearchParams()
    params.set('q', query.value.trim())
    if (selectedCategory.value) {
      params.set('category', selectedCategory.value)
    }

    const res = await fetch(`/api/v1/public/${locale.value}/search?${params.toString()}`)
    if (res.ok) {
      const data = await res.json()
      results.value = data.items || []
      totalResults.value = data.total || results.value.length
    } else {
      results.value = []
      totalResults.value = 0
    }
  } catch (err) {
    console.error('Search request failed', err)
    results.value = []
  } finally {
    isLoading.value = false
  }
}

function submitSearch() {
  router.push({
    path: `/${locale.value}/search/`,
    query: {
      q: query.value.trim(),
      category: selectedCategory.value || undefined,
    },
  })
}

watch(
  () => [route.query.q, route.query.category],
  () => {
    query.value = (route.query.q as string) || ''
    selectedCategory.value = (route.query.category as string) || ''
    performSearch()
  }
)

onMounted(() => {
  if (query.value) {
    performSearch()
  }
})
</script>

<template>
  <div class="site-layout">
    <SiteHeader />

    <main class="page-container">
      <div class="search-header">
        <h1 class="search-title">
          {{ isId ? 'Pencarian Destinasi & Potensi Pidie' : 'Search Pidie Destinations & Culture' }}
        </h1>
        <p class="search-sub">
          {{ isId
            ? 'Temukan objek wisata, aneka kuliner khas, dan sejarah di Kabupaten Pidie.'
            : 'Find scenic spots, authentic culinary delights, and heritage in Pidie Regency.'
          }}
        </p>

        <!-- Search Input Form -->
        <form @submit.prevent="submitSearch" class="search-box">
          <input
            v-model="query"
            type="text"
            :placeholder="isId ? 'Ketik nama tempat, kuliner, atau kecamatan...' : 'Type destination name, food, or district...'"
            class="search-input"
          />
          <button type="submit" class="search-btn">
            {{ isId ? 'Cari' : 'Search' }}
          </button>
        </form>

        <!-- Category Filter Pills -->
        <div class="filter-pills">
          <button
            class="pill-btn"
            :class="{ active: selectedCategory === '' }"
            @click="selectedCategory = ''; submitSearch()"
          >
            {{ isId ? 'Semua Kategori' : 'All Categories' }}
          </button>
          <button
            class="pill-btn"
            :class="{ active: selectedCategory === 'wisata' }"
            @click="selectedCategory = 'wisata'; submitSearch()"
          >
            {{ isId ? 'Wisata & Destinasi' : 'Tourism & Nature' }}
          </button>
          <button
            class="pill-btn"
            :class="{ active: selectedCategory === 'kuliner' }"
            @click="selectedCategory = 'kuliner'; submitSearch()"
          >
            {{ isId ? 'Kuliner & UMKM' : 'Culinary & Crafts' }}
          </button>
          <button
            class="pill-btn"
            :class="{ active: selectedCategory === 'budaya' }"
            @click="selectedCategory = 'budaya'; submitSearch()"
          >
            {{ isId ? 'Seni Budaya & Sejarah' : 'Heritage & History' }}
          </button>
        </div>
      </div>

      <!-- Results Section -->
      <section class="results-section">
        <div v-if="isLoading" class="loading-state">
          {{ isId ? 'Mencari data...' : 'Searching records...' }}
        </div>

        <div v-else-if="results.length === 0 && query" class="empty-state">
          <span class="empty-icon">🔍</span>
          <h3>{{ isId ? 'Tidak Ada Hasil yang Ditemukan' : 'No Results Found' }}</h3>
          <p>
            {{ isId
              ? `Tidak ditemukan hasil yang cocok untuk kata kunci "${query}". Coba periksa ejaan atau gunakan kata kunci umum lainnya.`
              : `No matches found for "${query}". Try checking your spelling or use general keywords.`
            }}
          </p>
        </div>

        <div v-else-if="results.length > 0" class="results-list">
          <p class="results-count">
            {{ isId
              ? `Ditemukan ${totalResults} hasil untuk "${query}":`
              : `Found ${totalResults} results for "${query}":`
            }}
          </p>

          <article v-for="item in results" :key="item.id" class="result-card">
            <div class="result-body">
              <div class="result-meta">
                <span class="result-cat">{{ item.categoryName || 'Wisata' }}</span>
                <span v-if="item.locationName" class="result-loc">📍 {{ item.locationName }}</span>
              </div>
              <h2 class="result-title">
                <a :href="`/${locale}/${item.categorySlug || 'wisata-destinasi'}/${item.slug}/`">
                  {{ item.title }}
                </a>
              </h2>
              <p class="result-snippet">{{ item.snippet || item.summary }}</p>
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
  max-width: 900px;
  width: 100%;
  margin: 0 auto;
  padding: 3rem 1.5rem;
  flex: 1;
}

.search-header {
  margin-bottom: 2.5rem;
  text-align: center;
}

.search-title {
  font-size: 2rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 0.5rem 0;
  letter-spacing: -0.02em;
}

.search-sub {
  font-size: 1rem;
  color: #64748b;
  margin: 0 0 1.75rem 0;
}

.search-box {
  display: flex;
  background: #ffffff;
  border: 2px solid #059669;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 10px 20px -5px rgba(5, 150, 105, 0.15);
  margin-bottom: 1.25rem;
}

.search-input {
  flex: 1;
  padding: 0.85rem 1.25rem;
  border: none;
  font-size: 1rem;
  outline: none;
  color: #0f172a;
}

.search-btn {
  background: #059669;
  color: #ffffff;
  border: none;
  padding: 0 1.75rem;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.15s;
}

.search-btn:hover {
  background: #047857;
}

.filter-pills {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pill-btn {
  padding: 0.4rem 0.85rem;
  border-radius: 20px;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #475569;
  font-size: 0.825rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.pill-btn:hover {
  background: #f1f5f9;
}

.pill-btn.active {
  background: #064e3b;
  color: #ffffff;
  border-color: #064e3b;
}

.results-count {
  font-size: 0.9rem;
  color: #64748b;
  margin-bottom: 1rem;
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.result-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.25rem;
  transition: transform 0.15s, border-color 0.15s;
}

.result-card:hover {
  transform: translateY(-2px);
  border-color: #cbd5e1;
}

.result-meta {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.4rem;
  font-size: 0.75rem;
}

.result-cat {
  font-weight: 700;
  color: #059669;
  text-transform: uppercase;
}

.result-loc {
  color: #64748b;
}

.result-title {
  margin: 0 0 0.5rem 0;
  font-size: 1.15rem;
  font-weight: 700;
}

.result-title a {
  color: #0f172a;
  text-decoration: none;
}

.result-title a:hover {
  color: #059669;
  text-decoration: underline;
}

.result-snippet {
  margin: 0;
  font-size: 0.875rem;
  color: #475569;
  line-height: 1.5;
}

.empty-state {
  padding: 4rem 1rem;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px dashed #cbd5e1;
}

.empty-icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 0.75rem;
}

.empty-state h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.2rem;
  color: #0f172a;
}

.empty-state p {
  margin: 0;
  font-size: 0.9rem;
  color: #64748b;
}

.loading-state {
  padding: 3rem;
  text-align: center;
  color: #64748b;
}
</style>
