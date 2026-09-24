<script setup>
import MarkdownIt from 'markdown-it'
import { computed } from 'vue'

const props = defineProps({
  content: { type: String, default: '' },
})

// html 关闭：正文里的 HTML 一律转义输出，v-html 不存在注入面
const md = new MarkdownIt({
  html: false,
  linkify: false,
  breaks: true,
})

const html = computed(() => md.render(props.content || ''))
</script>

<template>
  <div class="md" v-html="html" />
</template>
