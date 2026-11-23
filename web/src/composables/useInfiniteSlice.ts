import { computed, ref, toValue, watch, type MaybeRefOrGetter, type ShallowRef } from 'vue'

export function useInfiniteSlice<T>(
  items: MaybeRefOrGetter<readonly T[]>,
  sentinel: Readonly<ShallowRef<HTMLElement | null>>,
  resetKey: MaybeRefOrGetter<unknown>,
  batchSize = 36,
) {
  const visibleCount = ref(batchSize)

  const visibleItems = computed(() => toValue(items).slice(0, visibleCount.value))
  const hasMore = computed(() => visibleCount.value < toValue(items).length)

  watch(
    () => toValue(resetKey),
    () => {
      visibleCount.value = batchSize
    },
  )

  watch(
    sentinel,
    (el, _prev, onCleanup) => {
      if (!el) return

      const observer = new IntersectionObserver(
        (entries) => {
          if (entries.some((entry) => entry.isIntersecting) && hasMore.value) {
            visibleCount.value += batchSize
          }
        },
        { rootMargin: '400px' },
      )

      observer.observe(el)
      onCleanup(() => observer.disconnect())
    },
    { flush: 'post' },
  )

  return { visibleItems, hasMore }
}
