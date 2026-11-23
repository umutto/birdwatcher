import { createRouter, createWebHistory } from 'vue-router'

import CrawlsView from '@/views/CrawlsView.vue'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    noindex?: boolean
  }
}

export const APP_TITLE = 'birdwatcher'

const SITE_URL = (import.meta.env.VITE_SITE_URL ?? '').trim().replace(/\/+$/, '')

export function formatPageTitle(subtitle?: string): string {
  const trimmed = subtitle?.trim()
  return trimmed ? `${trimmed} — ${APP_TITLE}` : APP_TITLE
}

function syncCanonicalLink(path: string, noindex: boolean): void {
  const existing = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]')

  if (!SITE_URL || noindex) {
    existing?.remove()
    return
  }

  const href = `${SITE_URL}${path}`
  if (existing) {
    existing.href = href
    return
  }

  const link = document.createElement('link')
  link.rel = 'canonical'
  link.href = href
  document.head.append(link)
}

function syncRobotsMeta(noindex: boolean): void {
  const existing = document.head.querySelector<HTMLMetaElement>('meta[name="robots"]')

  if (!noindex) {
    existing?.remove()
    return
  }

  if (existing) {
    existing.content = 'noindex'
    return
  }

  const meta = document.createElement('meta')
  meta.name = 'robots'
  meta.content = 'noindex'
  document.head.append(meta)
}

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
      meta: { title: 'Expedition' },
      components: {
        default: () => import('@/views/CrawlCatalogView.vue'),
        navbar: () => import('@/components/CrawlNavbarSummary.vue'),
      },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      meta: { title: 'Page not found', noindex: true },
      component: () => import('@/components/ErrorState.vue'),
      props: { code: 404, message: 'Page not found' },
    },
  ],
})

router.afterEach((to) => {
  const noindex = Boolean(to.meta.noindex)
  document.title = formatPageTitle(to.meta.title)
  syncCanonicalLink(to.path, noindex)
  syncRobotsMeta(noindex)
})

export default router
