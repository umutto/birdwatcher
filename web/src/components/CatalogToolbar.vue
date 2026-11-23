<script setup lang="ts">
import { Search } from 'lucide-vue-next'

const hideMissing = defineModel<boolean>('hideMissing', { required: true })
const search = defineModel<string>('search', { required: true })
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-3 pt-2 pb-4.5">
    <button
      type="button"
      role="switch"
      :aria-checked="hideMissing"
      class="inline-flex cursor-pointer items-center gap-2.5 text-sm font-semibold text-stone-800 select-none"
      @click="hideMissing = !hideMissing"
    >
      <span
        class="relative h-5 w-9 rounded-full border transition-colors"
        :class="hideMissing ? 'border-forest bg-forest' : 'border-filigree/35 bg-filigree/20'"
      >
        <span
          class="absolute top-0.5 left-0.5 size-3.5 rounded-full bg-white shadow-xs transition-transform"
          :class="{ 'translate-x-4': hideMissing }"
        />
      </span>
      Hide missing
    </button>

    <div class="relative w-full sm:w-64">
      <Search
        aria-hidden="true"
        class="pointer-events-none absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-stone-400"
      />
      <input
        v-model="search"
        type="search"
        aria-label="Search species or plate"
        placeholder="Search species or plate..."
        class="w-full rounded-lg border border-filigree/30 bg-white/80 py-1.5 pr-3 pl-8 text-sm text-stone-900 placeholder:text-stone-400 focus:border-filigree focus:bg-white focus:outline-none"
      />
    </div>
  </div>
</template>
