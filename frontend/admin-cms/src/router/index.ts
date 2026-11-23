import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import ContentListView from '../views/contents/ContentListView.vue'
import ContentEditorView from '../views/contents/ContentEditorView.vue'
import CategoryListView from '../views/categories/CategoryListView.vue'
import MediaLibraryView from '../views/media/MediaLibraryView.vue'
import CurationView from '../views/curation/CurationView.vue'
import StaticPagesView from '../views/staticpages/StaticPagesView.vue'
import MenuView from '../views/menus/MenuView.vue'
import UnduhanView from '../views/unduhan/UnduhanView.vue'
import LayoutView from '../views/layout/LayoutView.vue'
import SettingsView from '../views/settings/SettingsView.vue'
import AdminsView from '../views/admins/AdminsView.vue'
import RedirectsView from '../views/redirects/RedirectsView.vue'
import AuditLogsView from '../views/audit/AuditLogsView.vue'
import NotFoundView from '../views/errors/NotFoundView.vue'
import ForbiddenView from '../views/errors/ForbiddenView.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: { guestOnly: true, title: 'Login Admin' },
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: DashboardView,
    meta: { requiresAuth: true, title: 'Dashboard Operasional', module: 'dashboard' },
  },
  {
    path: '/contents',
    name: 'contents',
    component: ContentListView,
    meta: { requiresAuth: true, title: 'Daftar Konten', module: 'contents' },
  },
  {
    path: '/contents/new',
    name: 'content-new',
    component: ContentEditorView,
    meta: { requiresAuth: true, title: 'Tambah Konten Baru', module: 'contents' },
  },
  {
    path: '/contents/:id/edit',
    name: 'content-edit',
    component: ContentEditorView,
    meta: { requiresAuth: true, title: 'Edit Konten', module: 'contents' },
  },
  {
    path: '/categories',
    name: 'categories',
    component: CategoryListView,
    meta: { requiresAuth: true, title: 'Kategori Konten', module: 'categories' },
  },
  {
    path: '/media',
    name: 'media',
    component: MediaLibraryView,
    meta: { requiresAuth: true, title: 'Media Library', module: 'media' },
  },
  {
    path: '/curation',
    name: 'curation',
    component: CurationView,
    meta: { requiresAuth: true, title: 'Kurasi Beranda', module: 'curation' },
  },
  {
    path: '/static-pages',
    name: 'static-pages',
    component: StaticPagesView,
    meta: { requiresAuth: true, title: 'Halaman Statis', module: 'static-pages' },
  },
  {
    path: '/menus',
    name: 'menus',
    component: MenuView,
    meta: { requiresAuth: true, title: 'Navigasi Menu Header', module: 'menus' },
  },
  {
    path: '/unduhan',
    name: 'unduhan',
    component: UnduhanView,
    meta: { requiresAuth: true, title: 'Unduhan Dokumen', module: 'unduhan' },
  },
  {
    path: '/tata-letak',
    name: 'layout',
    component: LayoutView,
    meta: { requiresAuth: true, title: 'Tata Letak & Banner Promosi', module: 'tata-letak' },
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsView,
    meta: { requiresAuth: true, title: 'Pengaturan Website', module: 'settings' },
  },
  {
    path: '/admins',
    name: 'admins',
    component: AdminsView,
    meta: { requiresAuth: true, title: 'Administrator', module: 'admins' },
  },
  {
    path: '/redirects',
    name: 'redirects',
    component: RedirectsView,
    meta: { requiresAuth: true, title: 'URL Redirects', module: 'redirects' },
  },
  {
    path: '/audit-logs',
    name: 'audit-logs',
    component: AuditLogsView,
    meta: { requiresAuth: true, title: 'Audit Trail', module: 'audit-logs' },
  },
  {
    path: '/403',
    name: 'forbidden',
    component: ForbiddenView,
    meta: { requiresAuth: true, title: 'Akses Ditolak (403)' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: NotFoundView,
    meta: { requiresAuth: true, title: 'Halaman Tidak Ditemukan (404)' },
  },
]

export const router = createRouter({
  history: createWebHistory('/cms/'),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // If user state is not yet checked, verify session with backend
  if (authStore.user === null && authStore.isChecking) {
    await authStore.checkSession()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login' })
  } else if (to.meta.guestOnly && authStore.isAuthenticated) {
    next({ name: 'dashboard' })
  } else if (to.meta.module && !authStore.canAccess(to.meta.module as string)) {
    next({ name: 'forbidden' })
  } else {
    next()
  }
})
