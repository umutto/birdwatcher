import {
  ApiError,
  type CatalogSpecies,
  type Crawl,
  type CrawlSighting,
  type CrawlSummary,
  type CreateCrawlRequest,
} from '@/api/types'

interface ErrorPayload {
  error?: string
}

async function requestJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  })

  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const payload = (await response.json()) as ErrorPayload
      if (payload.error) {
        message = payload.error
      }
    } catch {
      // Non-JSON error body; keep status fallback message.
    }
    throw new ApiError(response.status, message)
  }

  return (await response.json()) as T
}

export function listCrawls(options: {
  limit: number
  offset: number
  signal?: AbortSignal
}): Promise<CrawlSummary[]> {
  const params = new URLSearchParams({
    limit: String(options.limit),
    offset: String(options.offset),
  })
  return requestJSON<CrawlSummary[]>(`/api/v1/crawls?${params.toString()}`, {
    signal: options.signal,
  })
}

export function createCrawl(payload: CreateCrawlRequest): Promise<Crawl> {
  return requestJSON<Crawl>('/api/v1/crawls', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function getCrawl(id: string, signal?: AbortSignal): Promise<Crawl> {
  return requestJSON<Crawl>(`/api/v1/crawls/${encodeURIComponent(id)}`, { signal })
}

export function getCrawlCatalog(id: string, signal?: AbortSignal): Promise<CatalogSpecies[]> {
  return requestJSON<CatalogSpecies[]>(`/api/v1/crawls/${encodeURIComponent(id)}/catalog`, {
    signal,
  })
}

export function listCrawlSightings(
  id: string,
  options?: {
    plate?: number
    limit?: number
    offset?: number
    signal?: AbortSignal
  },
): Promise<CrawlSighting[]> {
  const params = new URLSearchParams({
    limit: String(options?.limit ?? 1),
    offset: String(options?.offset ?? 0),
  })
  if (options?.plate !== undefined) {
    params.set('plate', String(options.plate))
  }

  return requestJSON<CrawlSighting[]>(
    `/api/v1/crawls/${encodeURIComponent(id)}/sightings?${params.toString()}`,
    { signal: options?.signal },
  )
}

export function resolveMediaUrl(path: string | null | undefined): string {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path

  const base = (import.meta.env.VITE_MEDIA_BASE_URL ?? '').replace(/\/+$/, '')
  const cleanPath = path.replace(/^\/+/, '')
  return base ? `${base}/${cleanPath}` : `/${cleanPath}`
}

export function resolveMediaThumbUrl(path: string | null | undefined): string {
  if (!path) return ''
  return resolveMediaUrl(path.replace(/\/([^/]+)$/, '/thumbs/$1'))
}
