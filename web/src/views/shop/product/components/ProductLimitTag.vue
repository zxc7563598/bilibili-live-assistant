<template>
  <span class="mc-limit-tag" :class="level ? 'is-limit' : 'is-plain'" :style="level ? { '--mc-limit-color': color } : null" :title="`限购：${label}`" @click="emit('click')">
    {{ label }}
  </span>
</template>

<script setup>
import { MIN_MEMBER_LEVEL_COLOR, MIN_MEMBER_LEVEL_LABEL } from '../limit-utils'

const props = defineProps({
  // 后端返回的 min_member_level
  type: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['click'])

// 只认识的档位按不限制处理，避免后端加档位时页面渲染出空白标签
const level = computed(() => (MIN_MEMBER_LEVEL_LABEL[props.type] ? props.type : 0))
const label = computed(() => MIN_MEMBER_LEVEL_LABEL[level.value])
const color = computed(() => MIN_MEMBER_LEVEL_COLOR[level.value])
</script>

<style scoped>
.mc-limit-tag {
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

.mc-limit-tag:hover {
  opacity: 0.8;
}

/* 不限制没有档位色，用中性灰 */
.mc-limit-tag.is-plain {
  color: #8a919f;
  background: #eef0f3;
  border-color: #e2e5ea;
}

.dark .mc-limit-tag.is-plain {
  color: #9aa1ab;
  background: #34373c;
  border-color: #3d4148;
}

/* 深色文字由主色混黑得来：金色这类亮色直接当文字色在浅色底上看不清 */
.mc-limit-tag.is-limit {
  color: color-mix(in srgb, var(--mc-limit-color) 62%, black);
  background: color-mix(in srgb, var(--mc-limit-color) 16%, transparent);
  border-color: color-mix(in srgb, var(--mc-limit-color) 38%, transparent);
}

.dark .mc-limit-tag.is-limit {
  color: var(--mc-limit-color);
  background: color-mix(in srgb, var(--mc-limit-color) 22%, transparent);
  border-color: color-mix(in srgb, var(--mc-limit-color) 44%, transparent);
}
</style>
