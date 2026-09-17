<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: { type: String, default: 'primary' }, // primary | soft | ghost | outline | danger
  size: { type: String, default: 'md' }, // sm | md | lg
  block: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  type: { type: String, default: 'button' },
})

const base = 'inline-flex items-center justify-center gap-2 rounded-full font-semibold transition press select-none disabled:pointer-events-none disabled:opacity-50'

const sizes = {
  sm: 'h-9 px-4 text-sm',
  md: 'h-12 px-6 text-[15px]',
  lg: 'h-[52px] px-8 text-base',
}

const variants = {
  primary: 'bg-primary text-on-primary shadow-[var(--shadow-btn)]',
  soft: 'bg-primary-soft text-primary',
  ghost: 'bg-transparent text-fg-2',
  outline: 'border border-line-strong bg-surface text-fg',
  danger: 'bg-danger text-white',
}

const cls = computed(() => [
  base,
  sizes[props.size],
  variants[props.variant],
  props.block ? 'w-full' : '',
])
</script>

<template>
  <button :type="type" :class="cls" :disabled="disabled || loading">
    <AppIcon v-if="loading" name="refresh" :size="18" class="animate-spin" />
    <slot />
  </button>
</template>
