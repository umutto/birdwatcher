import { ref, watch } from 'vue'

export function useStoredBoolean(key: string, defaultValue = false) {
  let initial = defaultValue
  try {
    const saved = window.localStorage.getItem(key)
    if (saved === 'true' || saved === 'false') {
      initial = saved === 'true'
    }
  } catch {
    // Ignore storage access errors (e.g. private browsing).
  }

  const state = ref(initial)

  watch(state, (value) => {
    try {
      window.localStorage.setItem(key, String(value))
    } catch {
      // Ignore storage write errors.
    }
  })

  return state
}
