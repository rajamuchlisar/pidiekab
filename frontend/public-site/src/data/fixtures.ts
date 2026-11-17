export interface Translation {
  locale: 'id' | 'en'
  title: string
  slug: string
  description: string
  body: string
  status: 'published' | 'draft' | 'archived'
}

export interface CategoryTranslation {
  locale: 'id' | 'en'
  name: string
  slug: string
  description: string
}

export interface Category {
  id: string
  translations: Record<'id' | 'en', CategoryTranslation>
}

export interface ContentItem {
  id: string
  categoryId: string
  translations: Record<'id' | 'en', Translation>
}

export const CATEGORIES: Category[] = [
  {
    id: 'cat-wisata',
    translations: {
      id: {
        locale: 'id',
        name: 'Wisata & Destinasi',
        slug: 'wisata-destinasi',
        description: 'Jelajahi keindahan alam, pantai, pegunungan, dan destinasi menarik di Pidie.',
      },
      en: {
        locale: 'en',
        name: 'Tourism & Destinations',
        slug: 'tourism-destinations',
        description: 'Explore scenic landscapes, beaches, mountains, and tourist attractions in Pidie.',
      },
    },
  },
]

export const CONTENTS: ContentItem[] = [
  {
    id: 'content-sample-01',
    categoryId: 'cat-wisata',
    translations: {
      id: {
        locale: 'id',
        title: 'Pesona Destinasi Wisata Pidie',
        slug: 'contoh-konten',
        description: 'Panduan lengkap menjelajahi keindahan alam dan destinasi wisata unggulan di Kabupaten Pidie.',
        body: 'Kabupaten Pidie menyimpan keindahan alam dan warisan budaya yang memikat hati. Dari hamparan pantai selat Malaka hingga perbukitan Tangse yang sejuk, Pidie menawarkan pengalaman wisata yang autentik.',
        status: 'published',
      },
      en: {
        locale: 'en',
        title: 'Exploring Pidie Tourism Destinations',
        slug: 'sample-content',
        description: 'A comprehensive discovery guide to exploring the natural wonders and destinations in Pidie Regency.',
        body: 'Pidie Regency offers breathtaking natural landscapes and fascinating cultural heritage. From scenic coastlines to the lush mountain valleys of Tangse, Pidie delivers an authentic discovery adventure.',
        status: 'published',
      },
    },
  },
  {
    id: 'content-draft-02',
    categoryId: 'cat-wisata',
    translations: {
      id: {
        locale: 'id',
        title: 'Rencana Rahasia Wisata Belum Rilis',
        slug: 'draft-rahasia',
        description: 'Konten internal draft yang tidak boleh dipublikasikan atau di-generate ke publik.',
        body: 'Draft body - strictly confidential.',
        status: 'draft',
      },
      en: {
        locale: 'en',
        title: 'Secret Upcoming Destination Draft',
        slug: 'secret-draft',
        description: 'Internal draft content that must not be published or pre-rendered.',
        body: 'Draft body - confidential.',
        status: 'draft',
      },
    },
  },
]

/**
 * Helper to get all published routes that must be generated into static HTML.
 * Strictly excludes any draft or archived routes.
 */
export function getPublishedRoutes(): string[] {
  const routes: string[] = ['/id/', '/en/']

  for (const cat of CATEGORIES) {
    routes.push(`/id/${cat.translations.id.slug}/`)
    routes.push(`/en/${cat.translations.en.slug}/`)
  }

  for (const item of CONTENTS) {
    const cat = CATEGORIES.find((c) => c.id === item.categoryId)
    if (!cat) continue

    if (item.translations.id.status === 'published') {
      routes.push(`/id/${cat.translations.id.slug}/${item.translations.id.slug}/`)
    }
    if (item.translations.en.status === 'published') {
      routes.push(`/en/${cat.translations.en.slug}/${item.translations.en.slug}/`)
    }
  }

  return routes
}
