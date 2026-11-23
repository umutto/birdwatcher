<script setup lang="ts">
import { computed, useSlots } from 'vue'
import { RouterLink } from 'vue-router'

import FiligreeCorner from '@/components/FiligreeCorner.vue'

const slots = useSlots()
const hasSlotContent = computed(() => Boolean(slots.default))
</script>

<template>
  <header
    class="relative mb-4 grid items-center gap-2.5 border-b border-filigree/30 px-3.5 pt-4 pb-3.5 md:h-23 md:grid-cols-[auto_1fr] md:gap-6 md:px-4 md:py-0"
  >
    <FiligreeCorner class="absolute top-1.5 -left-2 size-12 -scale-x-100" />
    <FiligreeCorner class="absolute top-1.5 -right-2 size-12" />

    <div
      class="flex items-center px-2.5 md:h-14 md:pr-6 md:pl-5"
      :class="{ 'md:border-r md:border-filigree/25': hasSlotContent }"
    >
      <RouterLink
        :to="{ name: 'crawls' }"
        title="Return to index"
        class="py-1 font-serif text-xl font-bold tracking-tight transition-opacity hover:opacity-80"
      >
        birdwatcher
      </RouterLink>
    </div>

    <div
      v-if="hasSlotContent"
      class="flex min-w-0 flex-col justify-between gap-2.5 px-2.5 md:flex-row md:items-center md:gap-6 md:pr-3 md:pl-0"
    >
      <slot />
    </div>
  </header>
</template>
