<script setup>
import { useRouter } from 'vue-router'

defineProps({
  title: { type: String, default: '' },
  transparent: { type: Boolean, default: false }, // 悬浮在图片上方时使用
  back: { type: Boolean, default: true },
})

const router = useRouter()
function onBack() {
  if (window.history.length > 1)
    router.back()
  else router.replace('/')
}
</script>

<template>
  <header class="safe-top sticky top-0 z-40 flex h-14 items-center gap-1 px-2" :class="transparent ? 'bg-transparent text-white' : 'glass-strong text-fg'">
    <button v-if="back" class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full press" aria-label="返回" @click="onBack">
      <AppIcon name="chevron-left" :size="24" />
    </button>
    <h1 class="flex-1 truncate pr-2 text-center text-base font-semibold">
      {{ title }}
    </h1>
    <div class="flex w-10 shrink-0 items-center justify-center">
      <slot name="right" />
    </div>
  </header>
</template>
