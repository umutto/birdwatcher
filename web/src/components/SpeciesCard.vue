<script setup lang="ts">
import { computed } from 'vue'

import { resolveMediaThumbUrl, resolveMediaUrl } from '@/api/client'
import type { CatalogSpecies } from '@/api/types'

const { species } = defineProps<{
  species: CatalogSpecies
}>()

const emit = defineEmits<{
  select: [species: CatalogSpecies]
}>()

const isSighted = computed(() => species.sighting_count > 0)
const thumbSrc = computed(() => resolveMediaThumbUrl(species.image_url))
const fullSrc = computed(() => resolveMediaUrl(species.image_url))
</script>

<template>
  <button
    type="button"
    class="content-auto flex cursor-pointer flex-col overflow-hidden rounded-xl border border-filigree/25 bg-paper/85 text-left shadow-card transition-colors hover:border-filigree/55 hover:bg-paper"
    @click="emit('select', species)"
  >
    <div
      class="relative flex h-56 w-full items-center justify-center overflow-hidden border-b border-filigree/15 p-3.5"
      :class="isSighted ? 'bg-filigree/5' : 'bg-canvas'"
    >
      <span
        class="absolute top-2.5 left-2.5 rounded bg-paper/75 px-1.5 py-0.5 font-mono text-xs font-medium tabular-nums text-stone-600"
      >
        Plate {{ species.plate_number }}
      </span>

      <img
        :src="thumbSrc"
        :srcset="`${thumbSrc} 480w, ${fullSrc} 1600w`"
        sizes="(max-width: 640px) 46vw, 220px"
        :alt="species.common_name"
        loading="lazy"
        decoding="async"
        fetchpriority="low"
        class="size-full object-contain mix-blend-multiply"
        :class="{ 'opacity-30 grayscale': !isSighted }"
      />

      <span
        v-if="isSighted"
        class="absolute right-2.5 bottom-2.5 rounded border border-filigree/25 bg-paper/90 px-1.5 py-0.5 font-mono text-xs font-semibold tabular-nums text-stone-700"
      >
        &times;{{ species.sighting_count }}
      </span>
    </div>

    <div class="px-3.5 py-2.5">
      <h2 class="font-serif text-base leading-snug font-bold text-stone-900">
        {{ species.common_name }}
      </h2>
      <p class="mt-0.5 text-xs italic text-stone-500">
        {{ species.scientific_name }}
      </p>
    </div>
  </button>
</template>
