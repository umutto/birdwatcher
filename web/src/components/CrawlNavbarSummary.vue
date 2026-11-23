<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { isCrawlActive } from '@/api/types'
import { TOTAL_SPECIES, useCrawl, useCrawlCatalog } from '@/queries/crawls'
import { getCrawlStateStyle, getCrawlTitle, getSpeciesProgressPercent } from '@/utils/crawl'

const route = useRoute()
const crawlId = computed(() => String(route.params.id ?? ''))

const { data: crawl, isPending } = useCrawl(crawlId)
const { data: catalog } = useCrawlCatalog(crawlId, () => isCrawlActive(crawl.value?.state))

const sightedSpeciesCount = computed(
  () => catalog.value?.filter((species) => species.sighting_count > 0).length ?? 0,
)
const totalSightings = computed(
  () => catalog.value?.reduce((sum, species) => sum + species.sighting_count, 0) ?? 0,
)
const stateStyle = computed(() => (crawl.value ? getCrawlStateStyle(crawl.value.state) : null))
const progressPercent = computed(() => getSpeciesProgressPercent(sightedSpeciesCount.value))
</script>

<template>
  <template v-if="crawl && stateStyle">
    <div class="min-w-0">
      <div class="flex flex-wrap items-center gap-2">
        <span
          class="inline-flex items-center rounded px-2 py-0.5 text-xs font-semibold lowercase"
          :class="stateStyle.badge"
        >
          {{ crawl.state }}
        </span>
        <span class="truncate font-mono text-xs text-stone-500">
          {{ crawl.target_url }}
        </span>
      </div>

      <h1
        class="mt-0.5 font-serif text-xl leading-tight font-bold text-stone-900 md:truncate md:text-2xl"
      >
        {{ getCrawlTitle(crawl) }}
      </h1>
    </div>

    <div class="w-full shrink-0 md:w-52 md:text-right">
      <div class="mb-1 flex items-baseline justify-between text-xs">
        <span class="font-serif text-stone-600 italic">Identified</span>
        <strong class="font-mono text-sm font-semibold text-stone-900 tabular-nums">
          {{ sightedSpeciesCount }} / {{ TOTAL_SPECIES }}
        </strong>
      </div>

      <div class="h-1.5 overflow-hidden rounded-full bg-stone-900/10">
        <div
          class="h-full rounded-full transition-[width] duration-200"
          :class="stateStyle.fill"
          :style="{ width: `${progressPercent}%` }"
        />
      </div>

      <p class="mt-1 text-xs text-stone-500 tabular-nums">
        {{ totalSightings }} sightings &middot; {{ crawl.pages_crawled_count }} pages
      </p>
    </div>
  </template>

  <div v-else-if="isPending" class="flex w-full items-center justify-between gap-6">
    <div class="space-y-1.5">
      <div class="h-4 w-44 animate-pulse rounded bg-filigree/15" />
      <div class="h-6 w-64 animate-pulse rounded bg-filigree/15" />
    </div>
    <div class="hidden w-52 space-y-1.5 md:block">
      <div class="h-4 w-full animate-pulse rounded bg-filigree/15" />
      <div class="h-1.5 w-full animate-pulse rounded-full bg-filigree/12" />
    </div>
  </div>
</template>
