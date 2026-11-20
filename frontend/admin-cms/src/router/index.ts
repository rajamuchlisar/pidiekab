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
import SettingsView from '../views/settings/SettingsView.vue'
import AdminsView from '../views/admins/AdminsView.vue'
import RedirectsView from '../views/redirects/RedirectsView.vue'
import AuditLogsView from '../views/audit/AuditLogsView.vue'

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
    meta: { requiresAuth: true, title: 'Dashboard Operasional' },
  },
  {
    path: '/contents',
    name: 'contents',
    component: ContentListView,
    meta: { requiresAuth: true, title: 'Daftar Konten' },
  },
  {
    path: '/contents/new',
    name: 'content-new',
    component: ContentEditorView,
    meta: { requiresAuth: true, title: 'Tambah Konten Baru' },
  },
  {
    path: '/contents/:id/edit',
    name: 'content-edit',
    component: ContentEditorView,
    meta: { requiresAuth: true, title: 'Edit Konten' },
  },
  {
    path: '/categories',
    name: 'categories',
    component: CategoryListView,
    meta: { requiresAuth: true, title: 'Kategori Konten' },
  },
  {
    path: '/media',
    name: 'media',
    component: MediaLibraryView,
    meta: { requiresAuth: true, title: 'Media Library' },
  },
  {
    path: '/curation',
    name: 'curation',
    component: CurationView,
    meta: { requiresAuth: true, title: 'Kurasi Beranda' },
  },
  {
    path: '/static-pages',
    name: 'static-pages',
    component: StaticPagesView,
    meta: { requiresAuth: true, title: 'Halaman Statis' },
  },
  {
    path: '/menus',
    name: 'menus',
    component: MenuView,
    meta: { requiresAuth: true, title: 'Navigasi Menu Header' },
  },
  {
    path: '/unduhan',
    name: 'unduhan',
    component: UnduhanView,
    meta: { requiresAuth: true, title: 'Unduhan Dokumen' },
  },
  {
    path: '/settings',
    name: 'settings',
    component: SettingsView,
    meta: { requiresAuth: true, title: 'Pengaturan Website' },
  },
  {
    path: '/admins',
    name: 'admins',
    component: AdminsView,
    meta: { requiresAuth: true, title: 'Administrator' },
  },
  {
    path: '/redirects',
    name: 'redirects',
    component: RedirectsView,
    meta: { requiresAuth: true, title: 'URL Redirects' },
  },
  {
    path: '/audit-logs',
    name: 'audit-logs',
    component: AuditLogsView,
    meta: { requiresAuth: true, title: 'Audit Trail' },
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
  } else {
    next()
  }
})
