import type { RouteRecordRaw } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import CategoryView from '../views/CategoryView.vue'
import DetailView from '../views/DetailView.vue'

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/id/',
  },
  {
    path: '/id/',
    name: 'home-id',
    component: HomeView,
  },
  {
    path: '/en/',
    name: 'home-en',
    component: HomeView,
  },
  {
    path: '/id/:categorySlug/',
    name: 'category-id',
    component: CategoryView,
  },
  {
    path: '/en/:categorySlug/',
    name: 'category-en',
    component: CategoryView,
  },
  {
    path: '/id/:categorySlug/:contentSlug/',
    name: 'detail-id',
    component: DetailView,
  },
  {
    path: '/en/:categorySlug/:contentSlug/',
    name: 'detail-en',
    component: DetailView,
  },
]
