<script setup>
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])
function toggle() {
  if (!props.disabled)
    emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <button type="button" role="switch" :aria-checked="modelValue" :disabled="disabled" class="relative h-7 w-12 shrink-0 rounded-full transition-colors disabled:opacity-50" :class="modelValue ? 'bg-primary' : 'bg-line-strong'" @click="toggle">
    <span class="absolute top-0.5 h-6 w-6 rounded-full bg-white shadow transition-all" :class="modelValue ? 'left-5.5' : 'left-0.5'" />
  </button>
</template>
