<script setup lang="ts">
import { ArrowUpRight } from 'lucide-vue-next'

import type { CatalogSpecies } from '@/api/types'
import FiligreeModal from '@/components/FiligreeModal.vue'
import SpeciesLatestSighting from '@/components/SpeciesLatestSighting.vue'
import SpeciesPlateStage from '@/components/SpeciesPlateStage.vue'

const { crawlId, species } = defineProps<{
  crawlId: string
  species: CatalogSpecies | null
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <FiligreeModal
    :open="species !== null"
    :label="species?.common_name ?? 'Species details'"
    size="lg"
    @close="emit('close')"
  >
    <div v-if="species" class="grid grid-cols-1 lg:grid-cols-12">
      <SpeciesPlateStage :species="species" />

      <div class="flex flex-col justify-between gap-5 p-6 lg:col-span-5 lg:p-7">
        <div>
          <h2 class="pr-8 font-serif text-3xl leading-tight font-bold text-stone-900">
            {{ species.common_name }}
          </h2>
          <p class="mt-1 text-sm text-stone-500 italic">
            {{ species.scientific_name }}
          </p>

          <div v-if="species.audubon_url" class="mt-4 mb-3">
            <a
              :href="species.audubon_url"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1 rounded-md border border-forest/25 bg-forest/10 px-2.5 py-1 text-xs font-semibold text-forest transition-colors hover:bg-forest/15"
            >
              Audubon Birds of America
              <ArrowUpRight class="size-3.5" />
            </a>
          </div>

          <p v-if="species.description" class="text-sm leading-relaxed text-stone-700">
            {{ species.description }}
          </p>
        </div>

        <SpeciesLatestSighting :crawl-id="crawlId" :species="species" />
      </div>
    </div>
  </FiligreeModal>
</template>
