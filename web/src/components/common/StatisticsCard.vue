<template>
  <div class="relative w-full flex items-center justify-around py-16">
    <div v-for="(item, index) in stats" :key="item.label" class="relative flex-1 text-center">
      <div class="mb-3 flex items-center justify-center gap-1 text-12 text-gray-400">
        <span>{{ item.label }}</span>
        <n-tooltip v-if="item.tooltip" placement="top" :show-arrow="false">
          <template #trigger>
            <i class="i-fe:alert-circle ml-2 cursor-help text-12 text-gray-300 transition-colors dark:text-gray-500 hover:text-gray-500 dark:hover:text-gray-300" />
          </template>
          {{ item.tooltip }}
        </n-tooltip>
      </div>
      <div class="flex items-baseline justify-center">
        <span v-if="item.prefix" class="mr-4 translate-y--1 text-14 text-gray-400 font-medium">
          {{ item.prefix }}
        </span>
        <span class="text-20 font-500">
          {{ item.value }}
        </span>
        <span v-if="item.suffix" class="ml-4 text-12 text-gray-400 font-normal">
          {{ item.suffix }}
        </span>
      </div>
      <div v-if="index < stats.length - 1" class="absolute right-0 top-1/2 h-32 w-px bg-gray-200 -translate-y-1/2 dark:bg-gray-700" />
    </div>
  </div>
</template>

<script setup>
defineProps({
  stats: {
    type: Array,
    required: true,
    validator: (value) => {
      return value.every(item =>
        item.label !== undefined
        && item.value !== undefined,
      )
    },
  },
})
</script>
