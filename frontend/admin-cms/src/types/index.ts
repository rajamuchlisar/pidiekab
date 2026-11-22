// Core TypeScript definitions for PidieKab Admin CMS

export type AdminRole = 'super_admin' | 'admin' | 'editor' | 'author'

export interface RoleDefinition {
  key: AdminRole
  name: string
  description: string
  badge_color: string
  permissions: string[]
}

export interface AdminUser {
  id: number
  email: string
  displayName: string
  role?: AdminRole
  permissions?: string[]
  isActive?: boolean
  createdAt?: string
  lastLoginAt?: string
}

export interface Category {
  id: number
  code: string
  isActive: boolean
  displayOrder: number
  names: Record<string, string> // e.g. { id: "Wisata & Destinasi", en: "Tourism & Destinations" }
  slugs: Record<string, string> // e.g. { id: "wisata-destinasi", en: "tourism-destinations" }
  contentCount?: number
}

export interface LocationItem {
  id: number
  kemendagriCode: string
  name: string
  type: string // "kabupaten", "kecamatan"
  parentCode?: string
}

export interface ContentSource {
  id?: number
  contentId?: number
  sourceUrl: string
  sourceName: string
  sourceType: string // "official_gov", "news_media", "academic", "field_survey"
  notes?: string
}

export interface ContentTranslation {
  locale: string
  title: string
  slug: string
  summary: string
  bodyHtml: string
  bodyJson?: any
  status: 'draft' | 'published' | 'archived'
  seoTitle?: string
  seoDescription?: string
  publishedAt?: string | null
  updatedAt?: string
}

export interface ContentItem {
  id: number
  categoryId: number
  categoryName?: string
  locationKecamatanId?: number | null
  kecamatanName?: string
  address?: string
  latitude?: number | null
  longitude?: number | null
  mainPhotoId?: number | null
  mainPhotoUrl?: string
  translations: Record<string, ContentTranslation>
  sources?: ContentSource[]
  createdAt: string
  updatedAt: string
}

export interface MediaItem {
  id: number
  storageKey: string
  url: string
  fileName: string
  fileSize: number
  mimeType: string
  title: string
  altTextId: string
  altTextEn: string
  attribution?: string
  sourceOrigin?: string
  createdAt: string
  inUse?: boolean
}

export interface WebsiteSettings {
  id: number
  siteName: string
  taglineId: string
  taglineEn: string
  logoUrl?: string
  faviconUrl?: string
  contactEmail: string
  contactPhone?: string
  address?: string
  socialInstagram?: string
  socialFacebook?: string
  socialYoutube?: string
  socialTwitter?: string
  defaultSeoTitleId?: string
  defaultSeoTitleEn?: string
  defaultSeoDescId?: string
  defaultSeoDescEn?: string
}

export interface StaticPageTranslation {
  locale: string
  title: string
  slug: string
  bodyHtml: string
  status: 'draft' | 'published' | 'archived'
  updatedAt?: string
}

export interface StaticPage {
  id: number
  pageKey: string
  translations: Record<string, StaticPageTranslation>
  createdAt: string
  updatedAt: string
}

export interface CurationStory {
  contentId: number
  title: string
  slug: string
  categoryName: string
  mainPhotoUrl?: string
  position: number
}

export interface CurationData {
  featured: Array<{ contentId: number; position: number; title: string; categoryName: string }>
  stories: {
    id: { lead?: CurationStory; supporting: CurationStory[] }
    en: { lead?: CurationStory; supporting: CurationStory[] }
  }
}
