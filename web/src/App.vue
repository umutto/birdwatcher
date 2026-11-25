<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import BackToTop from '@/components/BackToTop.vue'
import FiligreeNavbar from '@/components/FiligreeNavbar.vue'
import MeadowFooter from '@/components/MeadowFooter.vue'

const route = useRoute()
const hasNavbarContent = computed(() =>
  route.matched.some((record) => Boolean(record.components?.navbar)),
)
</script>

<template>
  <div class="flex min-h-dvh flex-col bg-field-paper">
    <div class="mx-auto flex w-full max-w-7xl flex-1 flex-col px-4 sm:px-7">
      <FiligreeNavbar :has-content="hasNavbarContent">
        <RouterView name="navbar" />
      </FiligreeNavbar>

      <RouterView />
    </div>

    <MeadowFooter />
    <BackToTop />
  </div>
</template>
