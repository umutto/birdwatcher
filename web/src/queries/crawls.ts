import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import {
  infiniteQueryOptions,
  queryOptions,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/vue-query'

import {
  createCrawl,
  getCrawl,
  getCrawlCatalog,
  listCrawls,
  listCrawlSightings,
} from '@/api/client'
import { isCrawlActive, type CrawlSummary, type CreateCrawlRequest } from '@/api/types'

export const POLL_INTERVAL_MS = 3_000
export const TOTAL_SPECIES = 435
export const CRAWLS_INITIAL_LIMIT = 8
export const CRAWLS_NEXT_LIMIT = 6

export interface CrawlsPage {
  items: CrawlSummary[]
  hasMore: boolean
}

interface CrawlsPageParam {
  offset: number
  limit: number
}

export const crawlKeys = {
  all: ['crawls'] as const,
  lists: () => [...crawlKeys.all, 'list'] as const,
  detail: (crawlId: MaybeRefOrGetter<string>) =>
    computed(() => [...crawlKeys.all, toValue(crawlId)] as const),
  catalog: (crawlId: MaybeRefOrGetter<string>) =>
    computed(() => [...crawlKeys.all, toValue(crawlId), 'catalog'] as const),
  latestSighting: (crawlId: MaybeRefOrGetter<string>, plate: MaybeRefOrGetter<number | null>) =>
    computed(() => [...crawlKeys.all, toValue(crawlId), 'sightings', toValue(plate)] as const),
}

export function crawlsInfiniteOptions() {
  return infiniteQueryOptions({
    queryKey: crawlKeys.lists(),
    queryFn: async ({ pageParam, signal }): Promise<CrawlsPage> => {
      // Request one extra row to determine accurately if another page exists.
      const rows = await listCrawls({
        limit: pageParam.limit + 1,
        offset: pageParam.offset,
        signal,
      })
      return {
        items: rows.slice(0, pageParam.limit),
        hasMore: rows.length > pageParam.limit,
      }
    },
    initialPageParam: { offset: 0, limit: CRAWLS_INITIAL_LIMIT } satisfies CrawlsPageParam,
    getNextPageParam: (lastPage, _allPages, lastPageParam): CrawlsPageParam | undefined => {
      if (!lastPage.hasMore) return undefined
      return {
        offset: lastPageParam.offset + lastPageParam.limit,
        limit: CRAWLS_NEXT_LIMIT,
      }
    },
    refetchInterval: (query) =>
      query.state.data?.pages.some((page) => page.items.some((crawl) => isCrawlActive(crawl.state)))
        ? POLL_INTERVAL_MS
        : false,
  })
}

export function crawlDetailOptions(crawlId: MaybeRefOrGetter<string>) {
  return queryOptions({
    queryKey: crawlKeys.detail(crawlId),
    queryFn: ({ signal }) => getCrawl(toValue(crawlId), signal),
    enabled: () => Boolean(toValue(crawlId)),
    refetchInterval: (query) => (isCrawlActive(query.state.data?.state) ? POLL_INTERVAL_MS : false),
  })
}

export function crawlCatalogOptions(
  crawlId: MaybeRefOrGetter<string>,
  active: MaybeRefOrGetter<boolean>,
) {
  return queryOptions({
    queryKey: crawlKeys.catalog(crawlId),
    queryFn: ({ signal }) => getCrawlCatalog(toValue(crawlId), signal),
    enabled: () => Boolean(toValue(crawlId)),
    refetchInterval: computed(() => (toValue(active) ? POLL_INTERVAL_MS : false)),
  })
}

export function latestSpeciesSightingOptions(
  crawlId: MaybeRefOrGetter<string>,
  plate: MaybeRefOrGetter<number | null>,
) {
  return queryOptions({
    queryKey: crawlKeys.latestSighting(crawlId, plate),
    queryFn: ({ signal }) => {
      const plateNumber = toValue(plate)
      if (plateNumber === null) return Promise.resolve([])
      return listCrawlSightings(toValue(crawlId), {
        plate: plateNumber,
        limit: 1,
        offset: 0,
        signal,
      })
    },
    select: (rows) => rows[0] ?? null,
    enabled: () => Boolean(toValue(crawlId)) && toValue(plate) !== null,
  })
}

export function useCrawlsFeed() {
  const query = useInfiniteQuery(crawlsInfiniteOptions())
  const crawls = computed(() => query.data.value?.pages.flatMap((page) => page.items) ?? [])

  return { ...query, crawls }
}

export function useCreateCrawl() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreateCrawlRequest) => createCrawl(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: crawlKeys.lists() })
    },
  })
}

export function useCrawl(crawlId: MaybeRefOrGetter<string>) {
  return useQuery(crawlDetailOptions(crawlId))
}

export function useCrawlCatalog(
  crawlId: MaybeRefOrGetter<string>,
  active: MaybeRefOrGetter<boolean>,
) {
  return useQuery(crawlCatalogOptions(crawlId, active))
}

export function useLatestSpeciesSighting(
  crawlId: MaybeRefOrGetter<string>,
  plate: MaybeRefOrGetter<number | null>,
) {
  return useQuery(latestSpeciesSightingOptions(crawlId, plate))
}
