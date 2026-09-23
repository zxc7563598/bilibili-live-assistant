<template>
  <span class="mc-guard-tag" :class="level ? 'is-guard' : 'is-plain'" :style="level ? { '--mc-guard-color': color } : null" :title="`大航海身份：${label}`" @click="emit('click')">
    {{ label }}
  </span>
</template>

<script setup>
const props = defineProps({
  // 后端返回的 vip_type
  type: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['click'])

const LEVEL_LABEL = { 0: '普通', 3: '舰长', 2: '提督', 1: '总督' }
const LEVEL_COLOR = { 1: '#FFD700', 2: '#8A2BE2', 3: '#4A90E2' }

// 只认识的档位按普通处理，避免后端加档位时页面渲染出空白标签
const level = computed(() => (LEVEL_LABEL[props.type] ? props.type : 0))
const label = computed(() => LEVEL_LABEL[level.value])
const color = computed(() => LEVEL_COLOR[level.value])
</script>

<style scoped>
.mc-guard-tag {
  display: inline-flex;
  align-items: center;
  padding: 0 7px;
  border: 1px solid transparent;
  border-radius: 999px;
  font-size: 11px;
  line-height: 18px;
  white-space: nowrap;
  vertical-align: middle;
  cursor: pointer;
  user-select: none;
  transition: opacity 0.2s;
}

.mc-guard-tag:hover {
  opacity: 0.8;
}

/* 普通用户没有档位色，用中性灰 */
.mc-guard-tag.is-plain {
  color: #8a919f;
  background: #eef0f3;
  border-color: #e2e5ea;
}

.dark .mc-guard-tag.is-plain {
  color: #9aa1ab;
  background: #34373c;
  border-color: #3d4148;
}

/* 深色文字由主色混黑得来：金色这类亮色直接当文字色在浅色底上看不清 */
.mc-guard-tag.is-guard {
  color: color-mix(in srgb, var(--mc-guard-color) 62%, black);
  background: color-mix(in srgb, var(--mc-guard-color) 16%, transparent);
  border-color: color-mix(in srgb, var(--mc-guard-color) 38%, transparent);
}

.dark .mc-guard-tag.is-guard {
  color: var(--mc-guard-color);
  background: color-mix(in srgb, var(--mc-guard-color) 22%, transparent);
  border-color: color-mix(in srgb, var(--mc-guard-color) 44%, transparent);
}
</style>
