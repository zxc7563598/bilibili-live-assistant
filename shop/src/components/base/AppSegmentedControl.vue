<script setup>
defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] }, // [{ label, value }]
})
const emit = defineEmits(['update:modelValue'])
</script>

<template>
  <div class="inline-flex rounded-full bg-bg-soft p-1">
    <button v-for="o in options" :key="o.value" type="button" class="flex-1 whitespace-nowrap rounded-full px-4 py-1.5 text-sm font-medium transition" :class="modelValue === o.value ? 'bg-surface text-primary shadow-card' : 'text-fg-2'" @click="emit('update:modelValue', o.value)">
      {{ o.label }}
    </button>
  </div>
</template>
