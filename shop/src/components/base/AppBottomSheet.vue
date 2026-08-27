<script setup>
defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])
function close() {
  emit('update:modelValue', false)
}
</script>

<template>
  <Teleport to="body">
    <Transition name="sheet">
      <div v-if="modelValue" class="fixed inset-0 z-60">
        <div class="absolute inset-0 bg-black/40" @click="close" />
        <div class="absolute inset-x-0 bottom-0 mx-auto max-w-lg rounded-t-3xl bg-surface shadow-pop">
          <div class="flex items-center justify-between px-5 pb-2 pt-4">
            <span class="text-base font-semibold">{{ title }}</span>
            <button class="flex h-8 w-8 items-center justify-center rounded-full text-fg-3 press" aria-label="关闭" @click="close">
              <AppIcon name="close" :size="20" />
            </button>
          </div>
          <div class="safe-bottom max-h-[70vh] overflow-y-auto px-5 pb-6">
            <slot />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.sheet-enter-active,
.sheet-leave-active {
  transition: opacity 0.2s ease;
}
.sheet-enter-active > div:last-child,
.sheet-leave-active > div:last-child {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.sheet-enter-from,
.sheet-leave-to {
  opacity: 0;
}
.sheet-enter-from > div:last-child,
.sheet-leave-to > div:last-child {
  transform: translateY(24px);
}
</style>
