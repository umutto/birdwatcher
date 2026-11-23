<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import type { CrawlSummary } from '@/api/types'
import { TOTAL_SPECIES } from '@/queries/crawls'
import {
  formatCrawlDate,
  getCrawlStateStyle,
  getCrawlTitle,
  getSpeciesProgressPercent,
} from '@/utils/crawl'

const { crawl } = defineProps<{
  crawl: CrawlSummary
}>()

const stateStyle = computed(() => getCrawlStateStyle(crawl.state))
const progressPercent = computed(() => getSpeciesProgressPercent(crawl.unique_species_count))
</script>

<template>
  <RouterLink
    :to="{ name: 'crawl-catalog', params: { id: crawl.id } }"
    class="flex min-h-48 flex-col justify-between rounded-xl border border-filigree/25 bg-paper/85 px-5 py-4.5 shadow-card transition-colors hover:border-filigree/55 hover:bg-paper"
  >
    <div class="w-full">
      <div class="mb-2.5 flex items-center justify-between gap-2">
        <span
          class="inline-flex items-center rounded px-2 py-0.5 text-xs font-semibold lowercase"
          :class="stateStyle.badge"
        >
          {{ crawl.state }}
        </span>
        <time :datetime="crawl.created_at" class="font-mono text-xs text-stone-500 tabular-nums">
          {{ formatCrawlDate(crawl.created_at) }}
        </time>
      </div>

      <div class="border-l-3 pl-3" :class="stateStyle.border">
        <h2 class="line-clamp-2 font-serif text-lg leading-snug font-bold text-stone-900">
          {{ getCrawlTitle(crawl) }}
        </h2>
        <p class="mt-0.5 truncate font-mono text-xs text-stone-500">
          {{ crawl.target_url }}
        </p>
      </div>
    </div>

    <div class="mt-3.5 w-full border-t border-filigree/15 pt-2.5">
      <div class="mb-1.5 flex items-baseline justify-between text-xs">
        <span class="text-stone-600">Identified</span>
        <strong class="font-mono font-semibold text-stone-900 tabular-nums">
          {{ crawl.unique_species_count }} / {{ TOTAL_SPECIES }}
        </strong>
      </div>

      <div class="h-1.5 overflow-hidden rounded-full bg-stone-900/10">
        <div
          class="h-full rounded-full transition-[width] duration-200"
          :class="stateStyle.fill"
          :style="{ width: `${progressPercent}%` }"
        />
      </div>

      <div class="mt-2 flex justify-between text-xs text-stone-500 tabular-nums">
        <span>{{ crawl.total_sightings }} sightings</span>
        <span>{{ crawl.pages_crawled_count }} pages</span>
      </div>
    </div>
  </RouterLink>
</template>
