import { createRouter, createWebHistory } from 'vue-router'

import CrawlsView from '@/views/CrawlsView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'crawls',
      component: CrawlsView,
    },
    {
      path: '/crawls/:id',
      name: 'crawl-catalog',
      component: () => import('@/views/CrawlCatalogView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
    },
  ],
})

export default router
