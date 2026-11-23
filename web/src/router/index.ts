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
      components: {
        default: () => import('@/views/CrawlCatalogView.vue'),
        navbar: () => import('@/components/CrawlNavbarSummary.vue'),
      },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/components/ErrorState.vue'),
      props: { code: 404, message: 'Page not found' },
    },
  ],
})

export default router
