<script setup>
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'

// 依赖 free 的轮播：scroll-snap 横向滑动 + 指示点
const props = defineProps({
  items: { type: Array, default: () => [] }, // [{ src, label }]
  ratio: { type: String, default: '1 / 1' },
  interval: { type: [String, Number], default: 0 }, // 自动轮播间隔（ms），0 表示不自动
})

const intervalMs = computed(() => Number(props.interval))
const track = ref(null)
const active = ref(0)
let timer = null

// 首尾各补一张克隆图，滚到克隆项后无缝跳回真实项
const clones = computed(() => {
  const list = props.items
  if (list.length <= 1)
    return list
  return [list.at(-1), ...list, list[0]]
})

function slideWidth() {
  return track.value?.clientWidth || 0
}

function scrollToSlide(i, smooth) {
  if (!track.value)
    return
  track.value.scrollTo({ left: i * slideWidth(), behavior: smooth ? 'smooth' : 'instant' })
}

function onScroll() {
  const w = slideWidth()
  if (!track.value || !w || props.items.length <= 1)
    return
  // clones 中真实项位于 1..n，前后克隆分别映射到最后一张 / 第一张
  const i = Math.round(track.value.scrollLeft / w)
  active.value = ((i - 1) + props.items.length) % props.items.length
}

// 滚动停止后，若停在首尾克隆项上则无缝跳回对应的真实项
function onScrollEnd() {
  const w = slideWidth()
  if (!track.value || !w || props.items.length <= 1)
    return
  const i = Math.round(track.value.scrollLeft / w)
  const n = props.items.length
  if (i === 0) {
    scrollToSlide(n, false)
    active.value = n - 1
  }
  else if (i === n + 1) {
    scrollToSlide(1, false)
    active.value = 0
  }
}

// 按方向切一页；到达克隆项后由 onScrollEnd 无缝跳回真实项，形成无限循环
function go(delta) {
  const w = slideWidth()
  if (!track.value || !w || props.items.length <= 1)
    return
  const cur = Math.round(track.value.scrollLeft / w)
  const last = props.items.length + 1
  const target = Math.min(Math.max(cur + delta, 0), last)
  scrollToSlide(target, true)
  active.value = ((target - 1) + props.items.length) % props.items.length
  // 手动切换后重置自动轮播计时，避免刚切完又立刻跳页
  if (timer) {
    clearInterval(timer)
    timer = setInterval(go, intervalMs.value, 1)
  }
}

function stop() {
  if (timer)
    clearInterval(timer)
  timer = null
}

function start() {
  if (intervalMs.value > 0 && props.items.length > 1 && !timer)
    timer = setInterval(go, intervalMs.value, 1)
}

// items 到达后（含首次渲染即有内容的情况）定位到第一张真实图并启动自动轮播
watch(clones, async () => {
  await nextTick()
  if (!track.value)
    return
  if (props.items.length <= 1) {
    stop()
    return
  }
  track.value.scrollTo({ left: slideWidth(), behavior: 'instant' })
  active.value = 0
  start()
}, { immediate: true })

onUnmounted(stop)
</script>

<template>
  <div class="relative" @mouseenter="stop" @mouseleave="start">
    <div ref="track" class="no-scrollbar flex snap-x snap-mandatory overflow-x-auto" @scroll="onScroll" @scrollend="onScrollEnd">
      <div v-for="(it, i) in clones" :key="i" class="w-full shrink-0 snap-center">
        <AppImage :src="it.src" :label="it.label" :ratio="ratio" rounded="rounded-none" />
      </div>
    </div>
    <div v-if="props.items.length > 1" class="absolute bottom-3 left-1/2 flex -translate-x-1/2 gap-1.5">
      <span v-for="(it, i) in props.items" :key="i" class="h-1.5 rounded-full bg-white transition-all" :class="i === active ? 'w-5' : 'w-1.5 opacity-60'" />
    </div>
  </div>
</template>
