<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { Music } from 'lucide-vue-next'

import { resolveMediaUrl } from '@/api/client'
import type { CatalogSpecies } from '@/api/types'

const { species } = defineProps<{
  species: CatalogSpecies
}>()

const isZoomed = ref(false)
const zoomX = ref(50)
const zoomY = ref(50)

const isPlaying = ref(false)
let audio: HTMLAudioElement | null = null

const isSighted = computed(() => species.sighting_count > 0)
const imageSrc = computed(() => resolveMediaUrl(species.image_url))
const callSrc = computed(() => resolveMediaUrl(species.call_url))

function stopAudio() {
  if (audio) {
    audio.pause()
    audio.currentTime = 0
    audio = null
  }
  isPlaying.value = false
}

function toggleAudio() {
  if (!callSrc.value) return
  if (isPlaying.value) {
    stopAudio()
    return
  }

  stopAudio()
  audio = new Audio(callSrc.value)
  audio.addEventListener('ended', () => {
    isPlaying.value = false
  })
  audio.addEventListener('error', () => {
    isPlaying.value = false
  })
  isPlaying.value = true
  void audio.play().catch(() => {
    isPlaying.value = false
  })
}

watch(
  () => species.plate_number,
  () => {
    isZoomed.value = false
    zoomX.value = 50
    zoomY.value = 50
    stopAudio()
  },
)

onBeforeUnmount(stopAudio)

function handleMouseMove(event: MouseEvent) {
  if (!isZoomed.value) return
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  zoomX.value = Math.max(0, Math.min(100, ((event.clientX - rect.left) / rect.width) * 100))
  zoomY.value = Math.max(0, Math.min(100, ((event.clientY - rect.top) / rect.height) * 100))
}
</script>

<template>
  <div
    class="border-b border-filigree/15 p-5 lg:col-span-7 lg:border-r lg:border-b-0"
    :class="isSighted ? 'bg-filigree/5' : 'bg-canvas'"
  >
    <div
      class="relative flex h-96 w-full items-center justify-center overflow-hidden rounded-lg lg:h-120"
      :class="isZoomed ? 'cursor-zoom-out' : 'cursor-zoom-in'"
      @click="isZoomed = !isZoomed"
      @mousemove="handleMouseMove"
    >
      <span
        class="pointer-events-none absolute top-1.5 left-1.5 z-10 rounded bg-paper/60 px-2 py-0.5 font-mono text-xs text-stone-600/80 tabular-nums"
      >
        Plate {{ species.plate_number }}
      </span>

      <img
        :src="imageSrc"
        :alt="species.common_name"
        class="size-full object-contain mix-blend-multiply transition-transform duration-200"
        :class="{ 'scale-300': isZoomed, 'opacity-30 grayscale': !isSighted }"
        :style="isZoomed ? { transformOrigin: `${zoomX}% ${zoomY}%` } : undefined"
      />

      <button
        v-if="callSrc"
        type="button"
        :aria-label="isPlaying ? 'Stop bird call' : 'Play bird call'"
        :title="isPlaying ? 'Stop bird call' : 'Play bird call'"
        class="absolute right-2 bottom-2 z-10 flex size-8.5 cursor-pointer items-center justify-center rounded-full border bg-paper/85 transition-colors"
        :class="
          isPlaying
            ? 'border-forest/45 text-forest'
            : 'border-filigree/25 text-stone-700 hover:border-filigree/50'
        "
        @click.stop="toggleAudio"
      >
        <span
          v-if="isPlaying"
          aria-hidden="true"
          class="pointer-events-none absolute -inset-px rounded-full border border-forest/40 motion-safe:animate-note-ripple"
        />
        <Music
          class="size-4"
          :class="{ 'motion-safe:animate-note-breath': isPlaying }"
          :stroke-width="2"
        />
      </button>
    </div>
  </div>
</template>
