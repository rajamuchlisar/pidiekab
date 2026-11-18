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
  'id/sejarah-budaya/index.html',
  'en/history-heritage/index.html',
  'id/sejarah-budaya/eksplorasi-guha-tujoh-laweung-keajaiban-karst/index.html',
  'en/history-heritage/exploring-guha-tujoh-laweung-coastal-karst/index.html',
  'id/kuliner-khas/kelezatan-kerupuk-mulieng-beureunuen-melinjo/index.html',
  'en/traditional-culinary/savory-beureunuen-kerupuk-mulieng-melinjo/index.html',
]

for (const r of expectedRoutes) {
  const filePath = path.join(distDir, r)
  assert(fs.existsSync(filePath), `Static File Exists: ${r}`, `Checked path ${filePath}`)
}

// 3. Content Isolation
const idDetailHtml = fs.readFileSync(path.join(distDir, 'id/sejarah-budaya/eksplorasi-guha-tujoh-laweung-keajaiban-karst/index.html'), 'utf-8')
const enDetailHtml = fs.readFileSync(path.join(distDir, 'en/history-heritage/exploring-guha-tujoh-laweung-coastal-karst/index.html'), 'utf-8')

assert(
  idDetailHtml.includes('Eksplorasi Guha Tujoh Laweung') &&
  idDetailHtml.includes('Gua karst ini memiliki tujuh lorong') &&
  !idDetailHtml.includes('breathtaking natural landscapes') &&
  !idDetailHtml.includes('geological wonder features seven primary caverns'),
  'Content Isolation (ID)',
  'Indonesian HTML contains ID title and body, and does NOT leak English body'
)

assert(
  enDetailHtml.includes('Exploring Guha Tujoh Laweung') &&
  enDetailHtml.includes('geological wonder features seven primary caverns') &&
  !enDetailHtml.includes('Gua karst ini memiliki tujuh lorong'),
  'Content Isolation (EN)',
  'English HTML contains EN title and body, and does NOT leak Indonesian body'
)

// 4. SEO Title
assert(
  idDetailHtml.includes('<title>Eksplorasi Guha Tujoh Laweung: Keajaiban Karst Pesisir Selat Malaka — PidieKab</title>'),
  'SEO Title (ID)',
  'Contains exact localized title'
)
assert(
  enDetailHtml.includes('<title>Exploring Guha Tujoh Laweung: The Coastal Karst Wonder on Malacca Strait — PidieKab</title>'),
  'SEO Title (EN)',
  'Contains exact localized title'
)

// 5. Meta Description
assert(
  idDetailHtml.includes('name="description" content="Gua alam bersejarah dengan formasi stalaktit'),
  'Meta Description (ID)',
  'Contains localized meta description'
)
assert(
  enDetailHtml.includes('name="description" content="A historic natural limestone cave featuring magnificent'),
  'Meta Description (EN)',
  'Contains localized meta description'
)

// 6. Canonical URL
assert(
  idDetailHtml.includes('rel="canonical" href="https://pidiekab.id/id/sejarah-budaya/eksplorasi-guha-tujoh-laweung-keajaiban-karst/"'),
  'Canonical URL (ID)',
  'Self-referential canonical URL present'
)
assert(
  enDetailHtml.includes('rel="canonical" href="https://pidiekab.id/en/history-heritage/exploring-guha-tujoh-laweung-coastal-karst/"'),
  'Canonical URL (EN)',
  'Self-referential canonical URL present'
)

// 7. Hreflang
assert(
  idDetailHtml.includes('rel="alternate" hreflang="en" href="https://pidiekab.id/en/history-heritage/exploring-guha-tujoh-laweung-coastal-karst/"') &&
  idDetailHtml.includes('rel="alternate" hreflang="id" href="https://pidiekab.id/id/sejarah-budaya/eksplorasi-guha-tujoh-laweung-keajaiban-karst/"'),
  'Hreflang Bilingual Pair (ID)',
  'Bilingual alternate links exist for Published items'
)
assert(
  enDetailHtml.includes('rel="alternate" hreflang="id" href="https://pidiekab.id/id/sejarah-budaya/eksplorasi-guha-tujoh-laweung-keajaiban-karst/"') &&
  enDetailHtml.includes('rel="alternate" hreflang="en" href="https://pidiekab.id/en/history-heritage/exploring-guha-tujoh-laweung-coastal-karst/"'),
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
  sitemapContent.includes('eksplorasi-guha-tujoh-laweung-keajaiban-karst') &&
  sitemapContent.includes('exploring-guha-tujoh-laweung-coastal-karst'),
  'Sitemap Contains Published Routes',
  'Authentic published routes present in sitemap.xml'
)

// 9. Draft / Unpublished Content Exclusion
const draftIdPath = path.join(distDir, 'id/sejarah-budaya/draft-rahasia/index.html')
const draftEnPath = path.join(distDir, 'en/history-heritage/secret-draft/index.html')

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
