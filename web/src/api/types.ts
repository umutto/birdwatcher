export type CrawlState = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

export type SightingSourceType = 'text' | 'image'

export interface Crawl {
  id: string
  target_url: string
  title: string | null
  state: CrawlState
  pages_crawled_count: number
  images_scanned_count: number
  created_at: string
  updated_at: string
  started_at: string | null
  completed_at: string | null
}

export interface CrawlSummary extends Crawl {
  total_sightings: number
  unique_species_count: number
}

export interface CatalogSpecies {
  plate_number: number
  common_name: string
  aliases: string[] | null
  scientific_name: string
  description: string | null
  image_url: string
  call_url: string | null
  audubon_url: string | null
  metadata: Record<string, unknown> | null
  sighting_count: number
}

export interface CrawlSighting {
  id: string
  crawl_id: string
  species_plate_number: number
  source_type: SightingSourceType
  confidence: number
  page_url: string
  snippet: string | null
  image_url: string | null
  metadata: Record<string, unknown> | null
  created_at: string
  common_name: string
  scientific_name: string
}

export interface CreateCrawlRequest {
  target_url: string
}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export function isCrawlActive(state: CrawlState | undefined): boolean {
  return state === 'pending' || state === 'running'
}
