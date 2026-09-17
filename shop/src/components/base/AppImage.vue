<script setup>
import { ref, watch } from 'vue'

// 图片占位：无 src 或加载失败（@error）时显示渐变底 + 图标 + 说明文字，避免骨架阶段引入真实图片
const props = defineProps({
  src: { type: String, default: '' },
  alt: { type: String, default: '' },
  ratio: { type: String, default: '1 / 1' }, // 宽高比，如 '4 / 3'；传 'none' 则由父容器决定高度
  label: { type: String, default: '' },
  icon: { type: String, default: 'image' },
  rounded: { type: String, default: 'rounded-xl' },
})

const error = ref(false)

// src 变化时重置错误态，避免同一位置复用后一直显示占位
watch(() => props.src, () => {
  error.value = false
})
</script>

<template>
  <div class="relative w-full overflow-hidden bg-bg-soft" :class="rounded" :style="ratio === 'none' ? {} : { aspectRatio: ratio }">
    <img v-if="src && !error" :src="src" :alt="alt" loading="lazy" :class="ratio === 'none' ? 'block w-full' : 'absolute inset-0 h-full w-full object-cover'" @error="error = true">
    <div v-else class="flex flex-col items-center justify-center gap-1.5 text-fg-3" :class="ratio === 'none' ? 'min-h-28' : 'absolute inset-0'">
      <AppIcon :name="icon" :size="30" class="opacity-40" />
      <span v-if="label" class="text-xs">{{ label }}</span>
    </div>
  </div>
</template>
