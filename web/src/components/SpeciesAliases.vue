<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { CatalogSpecies } from '@/api/types'

const MAX_COLLAPSED_ALIASES = 3

const { species } = defineProps<{
  species: CatalogSpecies
}>()

const expanded = ref(false)

watch(
  () => species.plate_number,
  () => {
    expanded.value = false
  },
)

const aliases = computed(() => species.aliases ?? [])
const visibleAliases = computed(() =>
  expanded.value ? aliases.value : aliases.value.slice(0, MAX_COLLAPSED_ALIASES),
)
const hiddenCount = computed(() => aliases.value.length - visibleAliases.value.length)
</script>

<template>
  <div
    v-if="aliases.length > 0"
    class="flex items-baseline justify-between gap-3 border-t border-filigree/15 pt-3 text-xs"
  >
    <span class="shrink-0 font-semibold text-stone-600">Other names</span>
    <p class="text-right leading-relaxed text-stone-500">
      {{ visibleAliases.join(', ') }}
      <button
        v-if="hiddenCount > 0"
        type="button"
        class="ml-1 cursor-pointer font-semibold text-forest hover:underline"
        :aria-expanded="expanded"
        @click="expanded = true"
      >
        +{{ hiddenCount }}
      </button>
    </p>
  </div>
</template>
