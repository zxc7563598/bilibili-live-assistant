<script setup>
import { computed } from 'vue'

// 颜色通过内联样式 + CSS 变量实现，避免对自定义变量使用 /opacity 修饰符。
const props = defineProps({
  variant: { type: String, default: 'soft' }, // soft | solid | outline
  color: { type: String, default: 'primary' }, // primary | success | warning | danger | info | starlight | neutral
})

const SPECS = {
  primary: { var: '--primary', on: 'var(--on-primary)' },
  success: { var: '--success', on: '#ffffff' },
  warning: { var: '--warning', on: '#ffffff' },
  danger: { var: '--danger', on: '#ffffff' },
  info: { var: '--info', on: '#ffffff' },
  starlight: { var: '--starlight', on: 'var(--on-starlight)' },
}

const style = computed(() => {
  if (props.color === 'neutral') {
    return props.variant === 'solid'
      ? { background: 'var(--fg-2)', color: 'var(--surface)' }
      : { background: 'var(--surface-2)', color: 'var(--fg-2)', borderColor: 'var(--line-strong)' }
  }

  const spec = SPECS[props.color] || SPECS.primary
  if (props.variant === 'solid') {
    return { background: `var(${spec.var})`, color: spec.on }
  }
  if (props.variant === 'outline') {
    return { color: `var(${spec.var})`, borderColor: `color-mix(in srgb, var(${spec.var}) 40%, transparent)` }
  }
  // soft
  return {
    background: `color-mix(in srgb, var(${spec.var}) 12%, transparent)`,
    color: `var(${spec.var})`,
    borderColor: 'transparent',
  }
})
</script>

<template>
  <span class="inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium leading-5" :style="style">
    <slot />
  </span>
</template>
