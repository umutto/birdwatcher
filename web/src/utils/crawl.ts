import type { Crawl, CrawlState } from '@/api/types'
import { TOTAL_SPECIES } from '@/queries/crawls'

export interface CrawlStateStyle {
  badge: string
  border: string
  fill: string
}

const CRAWL_STATE_STYLES: Record<CrawlState, CrawlStateStyle> = {
  completed: {
    badge: 'bg-forest/12 text-forest',
    border: 'border-l-forest',
    fill: 'bg-forest',
  },
  running: {
    badge: 'bg-yellow-700/12 text-yellow-700',
    border: 'border-l-yellow-700',
    fill: 'bg-yellow-700',
  },
  pending: {
    badge: 'bg-yellow-700/12 text-yellow-700',
    border: 'border-l-yellow-700',
    fill: 'bg-yellow-700',
  },
  failed: {
    badge: 'bg-red-700/12 text-red-700',
    border: 'border-l-red-700',
    fill: 'bg-red-700',
  },
  cancelled: {
    badge: 'bg-stone-500/12 text-stone-600',
    border: 'border-l-stone-400',
    fill: 'bg-stone-500',
  },
}

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
})

export function getCrawlStateStyle(state: CrawlState): CrawlStateStyle {
  return CRAWL_STATE_STYLES[state]
}

export function getCrawlTitle(crawl: Pick<Crawl, 'title' | 'target_url'>): string {
  const trimmed = crawl.title?.trim()
  if (trimmed) return trimmed

  try {
    return new URL(crawl.target_url).hostname
  } catch {
    return crawl.target_url
  }
}

export function formatCrawlDate(isoDate: string): string {
  return dateFormatter.format(new Date(isoDate))
}

export function getSpeciesProgressPercent(uniqueCount: number): number {
  if (uniqueCount <= 0) return 0
  return Math.max(2, Math.min(100, (uniqueCount / TOTAL_SPECIES) * 100))
}
