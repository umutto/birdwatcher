<script setup lang="ts">
import { computed } from 'vue'
import { FileText, Image } from 'lucide-vue-next'

import { isCrawlActive, type CrawlState } from '@/api/types'
import { TOTAL_SPECIES } from '@/queries/crawls'
import { getCrawlStateStyle, getSpeciesProgressPercent } from '@/utils/crawl'

const {
  state,
  uniqueSpeciesCount,
  imageSightingCount,
  textSightingCount,
  pagesCrawledCount,
  variant = 'card',
} = defineProps<{
  state: CrawlState
  uniqueSpeciesCount: number
  imageSightingCount: number
  textSightingCount: number
  pagesCrawledCount: number
  variant?: 'card' | 'navbar'
}>()

const isActive = computed(() => isCrawlActive(state))
const stateStyle = computed(() => getCrawlStateStyle(state))
const progressPercent = computed(() => getSpeciesProgressPercent(uniqueSpeciesCount))
</script>

<template>
  <div>
    <div
      class="flex items-baseline justify-between text-xs"
      :class="variant === 'navbar' ? 'mb-1' : 'mb-1.5'"
    >
      <span class="text-stone-600" :class="{ 'font-serif italic': variant === 'navbar' }">
        Identified
      </span>
      <strong
        class="font-mono font-semibold text-stone-900 tabular-nums"
        :class="{ 'text-sm': variant === 'navbar' }"
      >
        {{ uniqueSpeciesCount }} / {{ TOTAL_SPECIES }}
      </strong>
    </div>

    <div class="relative h-1.5 overflow-hidden rounded-full bg-stone-900/10">
      <div
        class="h-full rounded-full transition-[width] duration-200"
        :class="stateStyle.fill"
        :style="{ width: `${progressPercent}%` }"
      />
      <div
        v-if="isActive"
        aria-hidden="true"
        class="pointer-events-none absolute inset-y-0 left-0 w-1/3 rounded-full bg-linear-to-r from-transparent via-yellow-700/45 to-transparent motion-safe:animate-crawl-shimmer"
      />
    </div>

    <div
      class="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-xs text-stone-500 tabular-nums sm:flex-nowrap"
      :class="variant === 'navbar' ? 'mt-1 md:justify-end' : 'mt-2 justify-between'"
    >
      <span class="inline-flex items-center gap-1.5 whitespace-nowrap">
        <span>Sightings:</span>
        <span class="inline-flex items-center gap-0.5" title="Image sightings">
          <Image aria-hidden="true" class="size-3 text-stone-400" :stroke-width="2" />
          <span class="sr-only">image</span>
          {{ imageSightingCount }}
        </span>
        <span class="inline-flex items-center gap-0.5" title="Text sightings">
          <FileText aria-hidden="true" class="size-3 text-stone-400" :stroke-width="2" />
          <span class="sr-only">text</span>
          {{ textSightingCount }}
        </span>
      </span>
      <span v-if="variant === 'navbar'" aria-hidden="true">&middot;</span>
      <span class="whitespace-nowrap">Pages: {{ pagesCrawledCount }}</span>
    </div>
  </div>
</template>
