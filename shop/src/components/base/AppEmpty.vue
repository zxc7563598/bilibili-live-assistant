<script setup>
// 空态：图标 + 标题 + 描述 + 可选操作插槽，统一各列表页空状态视觉
defineProps({
  icon: { type: String, default: 'box' },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 text-center">
    <div class="flex h-14 w-14 items-center justify-center rounded-full bg-bg-soft text-fg-3">
      <AppIcon :name="icon" :size="26" />
    </div>
    <p v-if="title" class="mt-3 text-[15px] font-medium text-fg-2">
      {{ title }}
    </p>
    <p v-if="description" class="mt-1 text-xs text-fg-3">
      {{ description }}
    </p>
    <div v-if="$slots.default" class="mt-4">
      <slot />
    </div>
  </div>
</template>
