<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Binoculars } from 'lucide-vue-next'

import FiligreeModal from '@/components/FiligreeModal.vue'
import { useCreateCrawl } from '@/queries/crawls'

const SAMPLE_URLS = [
  { label: 'audubon.org', url: 'https://www.audubon.org/birds-of-america' },
  { label: 'hacker news', url: 'https://news.ycombinator.com' },
  { label: 'fictional birds', url: 'https://en.wikipedia.org/wiki/List_of_fictional_birds' },
] as const

const { open } = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()
const targetUrl = ref('')

const { mutateAsync: submitCrawl, isPending: isCreating, isError, error, reset } = useCreateCrawl()

watch(
  () => open,
  (isOpen) => {
    if (isOpen) reset()
  },
)

async function handleSubmit() {
  try {
    const crawl = await submitCrawl({ target_url: targetUrl.value })
    targetUrl.value = ''
    emit('close')
    await router.push({ name: 'crawl-catalog', params: { id: crawl.id } })
  } catch {
    // API error is displayed inline via error.
  }
}
</script>

<template>
  <FiligreeModal :open="open" label="New expedition" size="sm" @close="emit('close')">
    <div class="p-6">
      <div class="flex items-center gap-2.5">
        <span
          aria-hidden="true"
          class="flex size-8 shrink-0 items-center justify-center rounded-full border border-forest/25 bg-forest/10 text-forest"
        >
          <Binoculars class="size-4.5" :stroke-width="1.8" />
        </span>
        <h2 class="font-serif text-2xl leading-none font-bold text-stone-900">New expedition</h2>
      </div>

      <p class="mt-2.5 text-sm text-stone-600">
        Pick a website and see which Audubon birds turn up.
      </p>

      <form class="mt-4" @submit.prevent="handleSubmit">
        <label for="target-url" class="mb-1.5 block text-xs font-semibold text-stone-600">
          Starting URL
        </label>
        <input
          id="target-url"
          v-model.trim="targetUrl"
          type="url"
          required
          autofocus
          pattern="https?://.+"
          placeholder="https://www.audubon.org/birds-of-america"
          class="w-full rounded-lg border border-filigree/30 bg-white/85 px-3 py-2.5 font-mono text-sm text-stone-900 placeholder:text-stone-400 focus:border-filigree focus:bg-white focus:outline-none"
        />

        <div class="mt-2.5 flex flex-wrap items-center gap-1.5">
          <span class="text-xs text-stone-500">Try:</span>
          <button
            v-for="sample in SAMPLE_URLS"
            :key="sample.label"
            type="button"
            class="cursor-pointer rounded-md border border-filigree/20 bg-filigree/5 px-2 py-0.5 font-mono text-xs text-stone-700 transition-colors hover:border-forest/40 hover:text-forest"
            @click="targetUrl = sample.url"
          >
            {{ sample.label }}
          </button>
        </div>

        <p v-if="isError" role="alert" class="mt-2 text-xs text-red-700">
          {{ error?.message }}
        </p>

        <div class="mt-5 flex justify-end gap-2 border-t border-filigree/15 pt-4">
          <button
            type="button"
            class="cursor-pointer rounded-lg border border-filigree/30 px-3.5 py-1.5 text-sm font-semibold text-stone-700 transition-colors hover:border-filigree/55 hover:bg-filigree/5"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="isCreating"
            class="cursor-pointer rounded-lg border border-forest bg-forest px-4 py-1.5 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{ isCreating ? 'Starting...' : 'Begin' }}
          </button>
        </div>
      </form>
    </div>
  </FiligreeModal>
</template>
