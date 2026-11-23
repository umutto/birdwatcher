<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main class="flex flex-1 flex-col items-center justify-center px-4 md:flex-row">
    <h1
      class="border-filigree/25 pt-6 font-serif text-8xl leading-none font-semibold tracking-tight text-stone-700 md:mr-6 md:border-e md:pr-6 md:pb-6"
    >
      404
    </h1>

    <div class="flex flex-col gap-2 text-center md:text-left">
      <p class="font-light text-stone-600">Page not found</p>
      <RouterLink
        :to="{ name: 'crawls' }"
        class="mt-4 border-t border-filigree/25 pt-4 text-xl text-forest hover:text-stone-700 md:mt-0 md:border-t-0 md:pt-0"
      >
        Return to index
      </RouterLink>
    </div>
  </main>
</template>
