<script setup>
defineProps({
  modelValue: { type: [String, Number], default: '' },
  label: { type: String, default: '' },
  type: { type: String, default: 'text' }, // text | password | textarea | number | tel | email
  placeholder: { type: String, default: '' },
  error: { type: String, default: '' },
  rows: { type: Number, default: 3 },
  autocomplete: { type: String, default: 'off' },
})

const emit = defineEmits(['update:modelValue'])
function onInput(e) {
  emit('update:modelValue', e.target.value)
}
</script>

<template>
  <label class="block">
    <span v-if="label" class="mb-1.5 block text-sm font-medium text-fg-2">{{ label }}</span>
    <div class="relative">
      <slot name="prefix" />
      <textarea v-if="type === 'textarea'" :value="modelValue" :rows="rows" :placeholder="placeholder" class="w-full resize-none rounded-2xl border bg-surface px-4 py-3 text-[15px] outline-none transition placeholder:text-fg-3" :class="error ? 'border-danger' : 'border-line focus:border-primary'" @input="onInput" />
      <input v-else :type="type" :value="modelValue" :placeholder="placeholder" :autocomplete="autocomplete" class="h-12 w-full rounded-2xl border bg-surface px-4 text-[15px] outline-none transition placeholder:text-fg-3" :class="error ? 'border-danger' : 'border-line focus:border-primary'" @input="onInput">
      <slot name="suffix" />
    </div>
    <span v-if="error" class="mt-1.5 block text-xs text-danger">{{ error }}</span>
  </label>
</template>
