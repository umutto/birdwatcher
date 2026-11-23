<script setup lang="ts">
import { computed, ref, useTemplateRef, watchEffect } from 'vue'
import { useRoute } from 'vue-router'

import { ApiError, isCrawlActive, type CatalogSpecies } from '@/api/types'
import CatalogToolbar from '@/components/CatalogToolbar.vue'
import ErrorState from '@/components/ErrorState.vue'
import SpeciesCard from '@/components/SpeciesCard.vue'
import SpeciesDetailModal from '@/components/SpeciesDetailModal.vue'
import { useInfiniteSlice } from '@/composables/useInfiniteSlice'
import { useStoredBoolean } from '@/composables/useStoredBoolean'
import { useCrawl, useCrawlCatalog } from '@/queries/crawls'
import { formatPageTitle } from '@/router'
import { getCrawlTitle } from '@/utils/crawl'

const route = useRoute()
const crawlId = computed(() => String(route.params.id ?? ''))

const {
  data: crawl,
  isError: isCrawlError,
  error: crawlError,
  refetch: refetchCrawl,
} = useCrawl(crawlId)

const {
  data: catalog,
  isPending: isCatalogPending,
  isError: isCatalogError,
  error: catalogError,
  refetch: refetchCatalog,
} = useCrawlCatalog(crawlId, () => isCrawlActive(crawl.value?.state))

const hideMissing = useStoredBoolean('birdwatcher:hide-missing')
const searchQuery = ref('')
const selectedSpecies = ref<CatalogSpecies | null>(null)
const sentinelRef = useTemplateRef<HTMLElement>('sentinel')

const activeError = computed(() => crawlError.value ?? catalogError.value)
const errorCode = computed(() =>
  activeError.value instanceof ApiError ? activeError.value.status : 'Error',
)
const errorMessage = computed(() => {
  const err = activeError.value
  if (err instanceof ApiError) {
    if (err.status === 404) return 'Expedition not found'
    if (err.status === 400) return 'Invalid expedition ID'
  }
  return err?.message ?? 'Unable to load expedition'
})
const canRetry = computed(
  () => !(activeError.value instanceof ApiError && activeError.value.status < 500),
)

watchEffect(() => {
  if (route.name !== 'crawl-catalog') return

  if (crawl.value) {
    document.title = formatPageTitle(getCrawlTitle(crawl.value))
  } else if (isCrawlError.value) {
    document.title = formatPageTitle(errorMessage.value)
  } else {
    document.title = formatPageTitle('Expedition')
  }
})

const sightedSpeciesCount = computed(
  () => catalog.value?.filter((species) => species.sighting_count > 0).length ?? 0,
)

const filteredSpecies = computed(() => {
  const list = catalog.value ?? []
  const query = searchQuery.value.trim().toLowerCase()

  return list.filter((species) => {
    if (hideMissing.value && species.sighting_count === 0) return false
    if (!query) return true

    return (
      species.common_name.toLowerCase().includes(query) ||
      species.scientific_name.toLowerCase().includes(query) ||
      String(species.plate_number).includes(query) ||
      Boolean(species.aliases?.some((alias) => alias.toLowerCase().includes(query)))
    )
  })
})

const { visibleItems: visibleSpecies, hasMore } = useInfiniteSlice(
  filteredSpecies,
  sentinelRef,
  () => `${crawlId.value}:${hideMissing.value}:${searchQuery.value.trim()}`,
)

function retryLoad() {
  void refetchCrawl()
  void refetchCatalog()
}
</script>

<template>
  <ErrorState v-if="isCrawlError || isCatalogError" :code="errorCode" :message="errorMessage">
    <template v-if="canRetry" #actions>
      <button
        type="button"
        class="cursor-pointer text-xl text-stone-600 hover:text-stone-900"
        @click="retryLoad"
      >
        Try again
      </button>
    </template>
  </ErrorState>

  <main v-else>
    <CatalogToolbar v-model:hide-missing="hideMissing" v-model:search="searchQuery" />

    <div
      v-if="isCatalogPending"
      class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
    >
      <div
        v-for="slot in 15"
        :key="slot"
        aria-hidden="true"
        class="h-72 animate-pulse rounded-xl border border-filigree/20 bg-paper/75"
      />
    </div>

    <div
      v-else-if="filteredSpecies.length === 0"
      class="rounded-xl border border-filigree/20 bg-paper/75 px-6 py-12 text-center"
    >
      <p class="font-serif text-lg font-bold text-stone-900">No matching species</p>
      <p class="mt-1 text-sm text-stone-600">
        {{
          hideMissing && sightedSpeciesCount === 0
            ? 'No birds have been spotted in this expedition yet.'
            : 'Try clearing your search or showing missing plates.'
        }}
      </p>
    </div>

    <template v-else>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        <SpeciesCard
          v-for="species in visibleSpecies"
          :key="species.plate_number"
          :species="species"
          @select="selectedSpecies = $event"
        />
      </div>

      <div v-if="hasMore" ref="sentinel" aria-hidden="true" class="h-4" />
    </template>
  </main>

  <SpeciesDetailModal
    :crawl-id="crawlId"
    :species="selectedSpecies"
    @close="selectedSpecies = null"
  />
</template>
