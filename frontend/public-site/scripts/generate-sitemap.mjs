import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const distDir = path.resolve(__dirname, '../dist')

const DOMAIN = 'https://pidiekab.id'

// Published routes definition (mirrored from fixtures snapshot)
const PUBLISHED_PAGES = [
  {
    path: '/id/',
    locales: { id: '/id/', en: '/en/' },
    changefreq: 'daily',
    priority: '1.0',
  },
  {
    path: '/en/',
    locales: { id: '/id/', en: '/en/' },
    changefreq: 'daily',
    priority: '1.0',
  },
  {
    path: '/id/wisata-destinasi/',
    locales: { id: '/id/wisata-destinasi/', en: '/en/tourism-destinations/' },
    changefreq: 'weekly',
    priority: '0.8',
  },
  {
    path: '/en/tourism-destinations/',
    locales: { id: '/id/wisata-destinasi/', en: '/en/tourism-destinations/' },
    changefreq: 'weekly',
    priority: '0.8',
  },
  {
    path: '/id/wisata-destinasi/contoh-konten/',
    locales: { id: '/id/wisata-destinasi/contoh-konten/', en: '/en/tourism-destinations/sample-content/' },
    changefreq: 'monthly',
    priority: '0.7',
  },
  {
    path: '/en/tourism-destinations/sample-content/',
    locales: { id: '/id/wisata-destinasi/contoh-konten/', en: '/en/tourism-destinations/sample-content/' },
    changefreq: 'monthly',
    priority: '0.7',
  },
]

function generateSitemap() {
  if (!fs.existsSync(distDir)) {
    console.warn(`[sitemap] dist directory ${distDir} does not exist yet.`)
    return
  }

  const urls = PUBLISHED_PAGES.map((page) => {
    let altLinks = ''
    if (page.locales) {
      for (const [lang, lPath] of Object.entries(page.locales)) {
        altLinks += `\n    <xhtml:link rel="alternate" hreflang="${lang}" href="${DOMAIN}${lPath}"/>`
      }
    }
    return `  <url>
    <loc>${DOMAIN}${page.path}</loc>${altLinks}
    <changefreq>${page.changefreq}</changefreq>
    <priority>${page.priority}</priority>
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
  console.log(`[sitemap] Generated sitemap.xml with ${PUBLISHED_PAGES.length} indexable published routes at ${sitemapPath}`)
}

generateSitemap()
