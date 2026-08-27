<script setup>
// 骨架屏微光块：尺寸/圆角由调用方 class 控制（继承到根节点），默认 rounded-md
defineProps({
  rounded: { type: String, default: 'rounded-md' },
})
</script>

<template>
  <div class="skeleton" :class="rounded" />
</template>
