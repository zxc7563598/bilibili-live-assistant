<template>
  <n-dynamic-tags :value="tags" :max="30" :disabled="disabled" :placeholder="placeholder" @update:value="onChange" />
</template>

<script setup>
defineOptions({ name: 'TagEditor' })

defineProps({
  /** 是否禁用编辑 */
  disabled: { type: Boolean, default: false },
  /** 新增输入框的占位提示 */
  placeholder: { type: String, default: '输入后回车添加标签' },
})

const tags = defineModel('tags', { type: Array, default: () => [] })

/** 后端用逗号拼接存储标签，所以输入里的英文逗号一律剥掉 */
const COMMA_RE = /,/g

/**
 * 标签既不能为空、也不能含英文逗号（后端用逗号拼接存储），
 * 因此在这里统一清洗并去重后再写回，避免脏数据流到提交体里。
 */
function onChange(value) {
  const seen = new Set()
  const clean = []
  for (const item of value) {
    const tag = String(item).replace(COMMA_RE, '').trim()
    if (tag && !seen.has(tag)) {
      seen.add(tag)
      clean.push(tag)
    }
  }
  tags.value = clean
}
</script>
