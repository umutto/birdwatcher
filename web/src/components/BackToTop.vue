<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { ChevronUp } from 'lucide-vue-next'

const isVisible = ref(false)

function updateVisibility() {
  const threshold = Math.max(500, window.innerHeight * 0.85)
  isVisible.value = window.scrollY > threshold
}

function scrollToTop() {
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  window.scrollTo({ top: 0, behavior: reducedMotion ? 'auto' : 'smooth' })
}

onMounted(() => {
  updateVisibility()
  window.addEventListener('scroll', updateVisibility, { passive: true })
  window.addEventListener('resize', updateVisibility, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updateVisibility)
  window.removeEventListener('resize', updateVisibility)
})
</script>

<template>
  <button
    type="button"
    aria-label="Back to top"
    title="Back to top"
    :inert="!isVisible"
    class="fixed right-0 bottom-8 z-30 flex cursor-pointer items-center justify-center rounded-l-lg border border-r-0 border-filigree/40 bg-paper/90 py-2.5 pr-3 pl-2.5 text-stone-600 shadow-tag backdrop-blur-xs transition-all duration-200 hover:border-filigree/65 hover:bg-paper hover:text-forest"
    :class="
      isVisible
        ? 'translate-x-0 opacity-85 hover:opacity-100'
        : 'pointer-events-none translate-x-full opacity-0'
    "
    @click="scrollToTop"
  >
    <ChevronUp aria-hidden="true" class="size-4" :stroke-width="2.2" />
  </button>
</template>
