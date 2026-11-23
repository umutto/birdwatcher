<script setup lang="ts">
import { ref } from 'vue'
import { Plus } from 'lucide-vue-next'

import CrawlCard from '@/components/CrawlCard.vue'
import NewExpeditionModal from '@/components/NewExpeditionModal.vue'
import { CRAWLS_INITIAL_LIMIT, useCrawlsFeed } from '@/queries/crawls'

const {
  crawls,
  isPending,
  isError,
  error,
  hasNextPage,
  isFetchingNextPage,
  fetchNextPage,
  refetch,
} = useCrawlsFeed()

const isModalOpen = ref(false)
</script>

<template>
  <main>
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <button
        type="button"
        class="flex min-h-48 cursor-pointer flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-forest/45 bg-paper/65 p-6 text-center shadow-card transition-colors hover:border-solid hover:border-forest/75 hover:bg-paper"
        @click="isModalOpen = true"
      >
        <span
          aria-hidden="true"
          class="flex size-11 items-center justify-center rounded-full bg-forest text-paper shadow-xs"
        >
          <Plus class="size-5" :stroke-width="2.2" />
        </span>

        <div>
          <h2 class="font-serif text-xl font-bold text-stone-900">New expedition</h2>
          <p class="mt-0.5 text-sm text-stone-600">Survey a website for bird species.</p>
        </div>
      </button>

      <template v-if="isPending">
        <div
          v-for="slot in CRAWLS_INITIAL_LIMIT"
          :key="slot"
          aria-hidden="true"
          class="flex min-h-48 animate-pulse flex-col justify-between rounded-xl border border-filigree/20 bg-paper/75 px-5 py-4.5"
        >
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <div class="h-5 w-20 rounded bg-filigree/15" />
              <div class="h-3.5 w-24 rounded bg-filigree/10" />
            </div>
            <div class="space-y-2 border-l-3 border-l-filigree/20 pl-3">
              <div class="h-5 w-4/5 rounded bg-filigree/15" />
              <div class="h-3.5 w-3/5 rounded bg-filigree/10" />
            </div>
          </div>

          <div class="space-y-2 border-t border-filigree/15 pt-2.5">
            <div class="flex justify-between">
              <div class="h-3.5 w-16 rounded bg-filigree/10" />
              <div class="h-3.5 w-14 rounded bg-filigree/15" />
            </div>
            <div class="h-1.5 w-full rounded-full bg-filigree/12" />
          </div>
        </div>
      </template>

      <div
        v-else-if="isError"
        role="alert"
        class="flex min-h-48 flex-col items-start justify-between rounded-xl border border-red-700/25 bg-paper px-5 py-4.5 sm:col-span-1 lg:col-span-2"
      >
        <div>
          <h2 class="font-serif text-lg font-bold text-stone-900">Unable to load expeditions</h2>
          <p class="mt-1 text-sm text-stone-600">
            {{ error?.message ?? 'Could not reach the catalog service.' }}
          </p>
        </div>

        <button
          type="button"
          class="cursor-pointer rounded-lg border border-filigree/30 bg-paper px-4 py-1.5 text-sm font-semibold text-stone-800 shadow-card transition-colors hover:border-filigree/55"
          @click="refetch()"
        >
          Try again
        </button>
      </div>

      <template v-else>
        <CrawlCard v-for="crawl in crawls" :key="crawl.id" :crawl="crawl" />
      </template>
    </div>

    <div v-if="hasNextPage" class="mt-6 flex justify-center">
      <button
        type="button"
        :disabled="isFetchingNextPage"
        class="cursor-pointer rounded-lg border border-filigree/30 bg-paper/85 px-5 py-2 text-sm font-semibold text-stone-800 shadow-card transition-colors hover:border-filigree/55 hover:bg-paper disabled:cursor-not-allowed disabled:opacity-60"
        @click="fetchNextPage()"
      >
        {{ isFetchingNextPage ? 'Loading...' : 'Load more' }}
      </button>
    </div>
  </main>

  <NewExpeditionModal :open="isModalOpen" @close="isModalOpen = false" />
</template>
