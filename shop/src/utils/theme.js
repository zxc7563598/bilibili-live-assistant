import { ref } from 'vue'

const STORAGE_KEY = 'shop-theme'

export const isDark = ref(false)

function resolveInitialTheme() {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'dark' || saved === 'light')
    return saved === 'dark'
  // 未手动设置时跟随系统
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

export function applyTheme(dark) {
  document.documentElement.dataset.theme = dark ? 'dark' : 'light'
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta)
    meta.setAttribute('content', dark ? '#0b0d12' : '#f5f6f8')
}

export function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem(STORAGE_KEY, isDark.value ? 'dark' : 'light')
  applyTheme(isDark.value)
}

export function initTheme() {
  isDark.value = resolveInitialTheme()
  applyTheme(isDark.value)
}
