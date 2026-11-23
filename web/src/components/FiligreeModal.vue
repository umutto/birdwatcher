<script setup lang="ts">
import { useTemplateRef, watch } from 'vue'

import FiligreeCorner from '@/components/FiligreeCorner.vue'

const {
  open,
  label,
  size = 'lg',
} = defineProps<{
  open: boolean
  label: string
  size?: 'sm' | 'lg'
}>()

const emit = defineEmits<{
  close: []
}>()

const dialogRef = useTemplateRef<HTMLDialogElement>('dialog')

watch(
  () => open,
  (isOpen) => {
    const dialog = dialogRef.value
    if (!dialog) return

    if (isOpen && !dialog.open) {
      dialog.showModal()
    } else if (!isOpen && dialog.open) {
      dialog.close()
    }
  },
  { immediate: true, flush: 'post' },
)

function handleNativeClose() {
  if (open) {
    emit('close')
  }
}

function handleBackdropClick(event: MouseEvent) {
  const dialog = dialogRef.value
  if (!dialog || 'closedBy' in HTMLDialogElement.prototype) return
  if (event.target !== dialog) return

  const rect = dialog.getBoundingClientRect()
  const isInsideDialog =
    rect.top <= event.clientY &&
    event.clientY <= rect.top + rect.height &&
    rect.left <= event.clientX &&
    event.clientX <= rect.left + rect.width

  if (!isInsideDialog) {
    dialog.close()
  }
}
</script>

<template>
  <dialog
    ref="dialog"
    closedby="any"
    :aria-label="label"
    class="m-auto max-h-11/12 w-11/12 overflow-auto rounded-xl border border-filigree/25 bg-paper text-stone-900 shadow-modal backdrop:bg-stone-900/60 backdrop:backdrop-blur-xs"
    :class="size === 'sm' ? 'max-w-md' : 'max-w-5xl'"
    @close="handleNativeClose"
    @click="handleBackdropClick"
  >
    <div class="relative">
      <slot />

      <FiligreeCorner class="absolute top-1 right-1 size-14" />
      <FiligreeCorner class="absolute right-1 bottom-1 size-12 -scale-y-100 opacity-65" />
    </div>
  </dialog>
</template>
