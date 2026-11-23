<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="flex min-h-dvh flex-col bg-field-paper">
    <div class="mx-auto flex w-full max-w-7xl flex-1 flex-col px-4 sm:px-7">
      <RouterView />
    </div>

    <footer aria-hidden="true" class="pt-9">
      <div class="h-40 bg-meadow opacity-90 mix-blend-multiply" />
    </footer>
  </div>
</template>
