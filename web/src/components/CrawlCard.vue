<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import type { CrawlSummary } from '@/api/types'
import CrawlProgress from '@/components/CrawlProgress.vue'
import CrawlStateBadge from '@/components/CrawlStateBadge.vue'
import { formatCrawlDate, getCrawlStateStyle, getCrawlTitle } from '@/utils/crawl'

const { crawl } = defineProps<{
  crawl: CrawlSummary
}>()

const stateStyle = computed(() => getCrawlStateStyle(crawl.state))
</script>

<template>
  <RouterLink
    :to="{ name: 'crawl-catalog', params: { id: crawl.id } }"
    class="flex min-h-48 flex-col justify-between rounded-xl border border-filigree/25 bg-paper/85 px-5 py-4.5 shadow-card transition-colors hover:border-filigree/55 hover:bg-paper"
  >
    <div class="w-full">
      <div class="mb-2.5 flex items-center justify-between gap-2">
        <CrawlStateBadge :state="crawl.state" />
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

    <CrawlProgress
      :state="crawl.state"
      :unique-species-count="crawl.unique_species_count"
      :image-sighting-count="crawl.image_sighting_count"
      :text-sighting-count="crawl.text_sighting_count"
      :pages-crawled-count="crawl.pages_crawled_count"
      class="mt-3.5 w-full border-t border-filigree/15 pt-2.5"
    />
  </RouterLink>
</template>
