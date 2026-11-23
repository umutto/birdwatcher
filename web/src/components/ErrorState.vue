<script setup lang="ts">
const { code, message } = defineProps<{
  code: number | string
  message: string
}>()
</script>

<template>
  <main class="flex flex-1 flex-col items-center justify-center px-4 md:flex-row">
    <h1
      class="border-filigree/25 pt-6 font-serif text-8xl leading-none font-semibold tracking-tight text-stone-700 md:mr-6 md:border-e md:pr-6 md:pb-6"
    >
      {{ code }}
    </h1>

    <div class="flex flex-col gap-2 text-center md:text-left">
      <p class="font-light text-stone-600">{{ message }}</p>
    </div>
  </main>
</template>
