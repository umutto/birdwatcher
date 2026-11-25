<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpRight } from 'lucide-vue-next'

import type { CatalogSpecies } from '@/api/types'
import { useLatestSpeciesSighting } from '@/queries/crawls'

const { crawlId, species } = defineProps<{
  crawlId: string
  species: CatalogSpecies
}>()

const { data: latestSighting, isPending: isLoading } = useLatestSpeciesSighting(
  () => crawlId,
  () => (species.sighting_count > 0 ? species.plate_number : null),
)

interface SnippetSegment {
  text: string
  bold: boolean
}

function escapeRegex(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

const snippetSegments = computed<SnippetSegment[]>(() => {
  const snippet = latestSighting.value?.snippet
  if (!snippet) return []

  const matchedText = latestSighting.value?.metadata?.matched_text?.trim()
  const terms = matchedText
    ? [matchedText]
    : [species.common_name, species.scientific_name, ...(species.aliases ?? [])]
        .map((term) => term.trim())
        .filter((term) => term.length >= 3)
        .sort((a, b) => b.length - a.length)

  if (terms.length === 0) return [{ text: snippet, bold: false }]

  const pattern = new RegExp(
    `(?<![\\p{L}\\p{N}_])(${terms.map(escapeRegex).join('|')})(?![\\p{L}\\p{N}_])`,
    'giu',
  )
  return snippet
    .split(pattern)
    .map((part, index) => ({ text: part, bold: index % 2 === 1 }))
    .filter((part) => part.text.length > 0)
})
</script>

<template>
  <div class="border-t border-filigree/15 pt-4">
    <div class="mb-2 flex items-center justify-between">
      <span class="text-xs font-semibold text-stone-600">Latest sighting</span>
      <span
        v-if="species.sighting_count > 0"
        class="font-mono text-xs font-semibold text-stone-700 tabular-nums"
      >
        &times;{{ species.sighting_count }} total
      </span>
    </div>

    <p v-if="species.sighting_count === 0" class="text-xs text-stone-500 italic">
      Not spotted in this expedition.
    </p>

    <div
      v-else-if="isLoading"
      aria-hidden="true"
      class="h-16 animate-pulse rounded-lg border border-filigree/15 bg-filigree/5"
    />

    <div
      v-else-if="latestSighting"
      class="rounded-lg border border-l-3 border-filigree/20 border-l-forest bg-filigree/5 px-3 py-2.5 text-xs"
    >
      <p v-if="snippetSegments.length > 0" class="leading-relaxed text-stone-800">
        <template v-for="(part, idx) in snippetSegments" :key="idx">
          <strong v-if="part.bold" class="font-semibold text-stone-950">{{ part.text }}</strong>
          <template v-else>{{ part.text }}</template>
        </template>
      </p>
      <p v-else class="text-stone-600 italic">
        Spotted via {{ latestSighting.source_type }} recognition.
      </p>

      <div class="mt-2 flex items-center justify-between text-stone-500">
        <span class="font-mono tabular-nums">
          {{ latestSighting.source_type }} &middot;
          {{ Math.round(latestSighting.confidence * 100) }}%
        </span>
        <a
          :href="latestSighting.page_url"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-0.5 font-semibold text-forest hover:underline"
        >
          Source
          <ArrowUpRight class="size-3" />
        </a>
      </div>
    </div>
  </div>
</template>
