<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { isCrawlActive } from '@/api/types'
import CrawlProgress from '@/components/CrawlProgress.vue'
import CrawlStateBadge from '@/components/CrawlStateBadge.vue'
import { useCrawl, useCrawlCatalog } from '@/queries/crawls'
import { getCrawlTitle } from '@/utils/crawl'

const route = useRoute()
const crawlId = computed(() => String(route.params.id ?? ''))

const { data: crawl, isPending } = useCrawl(crawlId)
const isActive = computed(() => isCrawlActive(crawl.value?.state))
const { data: catalog } = useCrawlCatalog(crawlId, isActive)

const sightedSpeciesCount = computed(
  () => catalog.value?.filter((species) => species.sighting_count > 0).length ?? 0,
)
const imageSightings = computed(
  () => catalog.value?.reduce((sum, species) => sum + (species.image_sighting_count ?? 0), 0) ?? 0,
)
const textSightings = computed(
  () => catalog.value?.reduce((sum, species) => sum + (species.text_sighting_count ?? 0), 0) ?? 0,
)
</script>

<template>
  <template v-if="crawl">
    <div class="min-w-0">
      <div class="flex flex-wrap items-center gap-2">
        <CrawlStateBadge :state="crawl.state" />
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

    <CrawlProgress
      :state="crawl.state"
      :unique-species-count="sightedSpeciesCount"
      :image-sighting-count="imageSightings"
      :text-sighting-count="textSightings"
      :pages-crawled-count="crawl.pages_crawled_count"
      variant="navbar"
      class="w-full shrink-0 md:w-64 md:text-right"
    />
  </template>

  <div v-else-if="isPending" class="flex w-full items-center justify-between gap-6">
    <div class="space-y-1.5">
      <div class="h-4 w-44 animate-pulse rounded bg-filigree/15" />
      <div class="h-6 w-64 animate-pulse rounded bg-filigree/15" />
    </div>
    <div class="hidden w-64 space-y-1.5 md:block">
      <div class="h-4 w-full animate-pulse rounded bg-filigree/15" />
      <div class="h-1.5 w-full animate-pulse rounded-full bg-filigree/12" />
    </div>
  </div>
</template>
