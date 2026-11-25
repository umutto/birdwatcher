<script setup lang="ts">
import { computed } from 'vue'

import { isCrawlActive, type CrawlState } from '@/api/types'
import { getCrawlStateStyle } from '@/utils/crawl'

const { state } = defineProps<{
  state: CrawlState
}>()

const isActive = computed(() => isCrawlActive(state))
const stateStyle = computed(() => getCrawlStateStyle(state))
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded px-2 py-0.5 text-xs font-semibold lowercase"
    :class="stateStyle.badge"
  >
    <span
      v-if="isActive"
      aria-hidden="true"
      class="size-1.5 rounded-full bg-current motion-safe:animate-crawl-pulse"
    />
    {{ state }}
  </span>
</template>
