import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const distDir = path.resolve(__dirname, '../dist')

console.log('====================================================')
console.log('       PIDIEKAB SSG POC ACCEPTANCE TEST SUITE       ')
console.log('====================================================')

const results = []

function assert(condition, testName, details = '') {
  if (condition) {
    console.log(`[PASS] ${testName}`)
    results.push({ testName, status: 'PASS', details })
  } else {
    console.error(`[FAIL] ${testName}: ${details}`)
    results.push({ testName, status: 'FAIL', details })
  }
}

// 1. Static HTML Generation & 2. Dynamic Localized Routes
const expectedRoutes = [
  'id/index.html',
  'en/index.html',
  'id/wisata-destinasi/index.html',
  'en/tourism-destinations/index.html',
  'id/wisata-destinasi/contoh-konten/index.html',
  'en/tourism-destinations/sample-content/index.html',
]

for (const r of expectedRoutes) {
  const filePath = path.join(distDir, r)
  assert(fs.existsSync(filePath), `Static File Exists: ${r}`, `Checked path ${filePath}`)
}

// 3. Content Isolation
const idDetailHtml = fs.readFileSync(path.join(distDir, 'id/wisata-destinasi/contoh-konten/index.html'), 'utf-8')
const enDetailHtml = fs.readFileSync(path.join(distDir, 'en/tourism-destinations/sample-content/index.html'), 'utf-8')

assert(
  idDetailHtml.includes('Pesona Destinasi Wisata Pidie') &&
  idDetailHtml.includes('Kabupaten Pidie menyimpan keindahan alam') &&
  !idDetailHtml.includes('breathtaking natural landscapes'),
  'Content Isolation (ID)',
  'Indonesian HTML contains ID title and body, and does NOT leak English body'
)

assert(
  enDetailHtml.includes('Exploring Pidie Tourism Destinations') &&
  enDetailHtml.includes('breathtaking natural landscapes') &&
  !enDetailHtml.includes('menyimpan keindahan alam'),
  'Content Isolation (EN)',
  'English HTML contains EN title and body, and does NOT leak Indonesian body'
)

// 4. SEO Title
assert(
  idDetailHtml.includes('<title>Pesona Destinasi Wisata Pidie — PidieKab</title>'),
  'SEO Title (ID)',
  'Contains exact localized title'
)
assert(
  enDetailHtml.includes('<title>Exploring Pidie Tourism Destinations — PidieKab</title>'),
  'SEO Title (EN)',
  'Contains exact localized title'
)

// 5. Meta Description
assert(
  idDetailHtml.includes('name="description" content="Panduan lengkap menjelajahi keindahan alam'),
  'Meta Description (ID)',
  'Contains localized meta description'
)
assert(
  enDetailHtml.includes('name="description" content="A comprehensive discovery guide to exploring'),
  'Meta Description (EN)',
  'Contains localized meta description'
)

// 6. Canonical URL
assert(
  idDetailHtml.includes('rel="canonical" href="https://pidiekab.id/id/wisata-destinasi/contoh-konten/"'),
  'Canonical URL (ID)',
  'Self-referential canonical URL present'
)
assert(
  enDetailHtml.includes('rel="canonical" href="https://pidiekab.id/en/tourism-destinations/sample-content/"'),
  'Canonical URL (EN)',
  'Self-referential canonical URL present'
)

// 7. Hreflang
assert(
  idDetailHtml.includes('rel="alternate" hreflang="en" href="https://pidiekab.id/en/tourism-destinations/sample-content/"') &&
  idDetailHtml.includes('rel="alternate" hreflang="id" href="https://pidiekab.id/id/wisata-destinasi/contoh-konten/"'),
  'Hreflang Bilingual Pair (ID)',
  'Bilingual alternate links exist for Published items'
)
assert(
  enDetailHtml.includes('rel="alternate" hreflang="id" href="https://pidiekab.id/id/wisata-destinasi/contoh-konten/"') &&
  enDetailHtml.includes('rel="alternate" hreflang="en" href="https://pidiekab.id/en/tourism-destinations/sample-content/"'),
  'Hreflang Bilingual Pair (EN)',
  'Bilingual alternate links exist for Published items'
)

// 8. Sitemap validation
const sitemapPath = path.join(distDir, 'sitemap.xml')
assert(fs.existsSync(sitemapPath), 'Sitemap Exists', 'dist/sitemap.xml exists')
const sitemapContent = fs.readFileSync(sitemapPath, 'utf-8')
assert(
  sitemapContent.includes('<loc>https://pidiekab.id/id/</loc>') &&
  sitemapContent.includes('<loc>https://pidiekab.id/en/</loc>') &&
  sitemapContent.includes('<loc>https://pidiekab.id/id/wisata-destinasi/contoh-konten/</loc>') &&
  sitemapContent.includes('<loc>https://pidiekab.id/en/tourism-destinations/sample-content/</loc>'),
  'Sitemap Contains Published Routes',
  'All 6 published routes present in sitemap.xml'
)

// 9. Draft / Unpublished Content Exclusion
const draftIdPath = path.join(distDir, 'id/wisata-destinasi/draft-rahasia/index.html')
const draftEnPath = path.join(distDir, 'en/tourism-destinations/secret-draft/index.html')

assert(!fs.existsSync(draftIdPath), 'Draft Exclusion: ID HTML Not Created', 'draft-rahasia file does not exist')
assert(!fs.existsSync(draftEnPath), 'Draft Exclusion: EN HTML Not Created', 'secret-draft file does not exist')
assert(!sitemapContent.includes('draft-rahasia'), 'Draft Exclusion: Not in Sitemap (ID)', 'No draft in sitemap')
assert(!sitemapContent.includes('secret-draft'), 'Draft Exclusion: Not in Sitemap (EN)', 'No draft in sitemap')

// Summary
const failed = results.filter((r) => r.status === 'FAIL')
console.log('====================================================')
console.log(`TOTAL TESTS: ${results.length} | PASSED: ${results.length - failed.length} | FAILED: ${failed.length}`)
console.log('====================================================')

if (failed.length > 0) {
  process.exit(1)
}
