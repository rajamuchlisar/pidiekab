import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const distDir = path.resolve(__dirname, '../dist')

const DOMAIN = 'https://pidiekab.id'

function findHtmlPages(dir, base = '') {
  let pages = []
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  for (const ent of entries) {
    const fullPath = path.join(dir, ent.name)
    const relPath = base ? `${base}/${ent.name}` : ent.name
    if (ent.isDirectory()) {
      if (ent.name !== 'assets' && ent.name !== '.vite') {
        pages = pages.concat(findHtmlPages(fullPath, relPath))
      }
    } else if (ent.isFile() && ent.name === 'index.html' && base !== '') {
      pages.push(`/${base}/`)
    }
  }
  return pages
}

function generateSitemap() {
  if (!fs.existsSync(distDir)) {
    console.warn(`[sitemap] dist directory ${distDir} does not exist yet.`)
    return
  }

  const paths = findHtmlPages(distDir).sort()

  const urls = paths.map((p) => {
    let priority = '0.7'
    let changefreq = 'monthly'

    if (p === '/id/' || p === '/en/') {
      priority = '1.0'
      changefreq = 'daily'
    } else if (p.split('/').filter(Boolean).length === 2) {
      // Category page (e.g. /id/wisata-alam/)
      priority = '0.8'
      changefreq = 'weekly'
    }

    return `  <url>
    <loc>${DOMAIN}${p}</loc>
    <changefreq>${changefreq}</changefreq>
    <priority>${priority}</priority>
  </url>`
  }).join('\n')

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
${urls}
</urlset>
`

  const sitemapPath = path.join(distDir, 'sitemap.xml')
  fs.writeFileSync(sitemapPath, xml, 'utf-8')
  console.log(`[sitemap] Generated sitemap.xml with ${paths.length} indexable published routes at ${sitemapPath}`)
}

generateSitemap()
