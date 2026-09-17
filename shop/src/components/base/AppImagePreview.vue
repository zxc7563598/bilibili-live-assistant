<script setup>
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

// 全屏大图预览（Lightbox）：轮播图点击放大后用。
// - 整屏（图片 + 黑色背景）点按即关闭；横向滑动在 items 之间切换（scroll-snap，同 AppCarousel）
// - 复用 AppDialog/AppBottomSheet 的 Teleport + Transition + fixed 浮层约定
const props = defineProps({
  open: { type: Boolean, default: false },
  items: { type: Array, default: () => [] }, // [{ src, label }]
  initial: { type: Number, default: 0 }, // 打开时定位到第几张
})
const emit = defineEmits(['update:open'])

const track = ref(null)
const current = ref(0)

// 区分「点按关闭」与「滑动切图」：记录按下起点，位移或时长超阈值视为滑动而非点按
const tapStart = { x: 0, y: 0, t: 0 }
const TAP_MAX_DISTANCE = 10
const TAP_MAX_DURATION = 500

function onPointerDown(e) {
  tapStart.x = e.clientX
  tapStart.y = e.clientY
  tapStart.t = Date.now()
}

function onRootClick(e) {
  const dx = e.clientX - tapStart.x
  const dy = e.clientY - tapStart.y
  if (Math.hypot(dx, dy) < TAP_MAX_DISTANCE && Date.now() - tapStart.t < TAP_MAX_DURATION)
    emit('update:open', false)
}

function scrollToIndex(i, smooth) {
  const el = track.value
  if (!el)
    return
  el.scrollTo({ left: i * el.clientWidth, behavior: smooth ? 'smooth' : 'auto' })
}

function go(delta) {
  const n = props.items.length
  const i = Math.min(Math.max(current.value + delta, 0), n - 1)
  if (i !== current.value) {
    current.value = i
    scrollToIndex(i, true)
  }
}

// 滑动结束后同步当前下标（用于底部计数）
function onScroll() {
  const el = track.value
  if (!el || !el.clientWidth)
    return
  current.value = Math.round(el.scrollLeft / el.clientWidth)
}

// 桌面便捷操作：Esc 关闭、←/→ 切图
function onKeydown(e) {
  if (e.key === 'Escape') {
    emit('update:open', false)
  }
  else if (e.key === 'ArrowLeft') {
    e.preventDefault()
    go(-1)
  }
  else if (e.key === 'ArrowRight') {
    e.preventDefault()
    go(1)
  }
}

watch(() => props.open, async (v) => {
  // 锁定背景滚动：全屏预览需要整屏滑动手势，避免带动下层页面（Dialog/BottomSheet 未锁，此处为必要的例外）
  document.body.style.overflow = v ? 'hidden' : ''
  window.removeEventListener('keydown', onKeydown)
  if (!v)
    return
  window.addEventListener('keydown', onKeydown)
  await nextTick()
  current.value = props.initial
  scrollToIndex(props.initial, false)
})

onBeforeUnmount(() => {
  document.body.style.overflow = ''
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="preview">
      <div v-if="open" class="fixed inset-0 z-80 select-none bg-black/95" role="dialog" aria-modal="true" @pointerdown="onPointerDown" @click="onRootClick">
        <div ref="track" class="no-scrollbar flex h-full snap-x snap-mandatory overflow-x-auto" @scroll="onScroll">
          <div v-for="(it, i) in items" :key="i" class="flex h-full w-full shrink-0 snap-center items-center justify-center">
            <img :src="it.src" :alt="it.label" draggable="false" class="max-h-full max-w-full select-none object-contain">
          </div>
        </div>
        <div v-if="items.length > 1" class="pointer-events-none absolute inset-x-0 bottom-6 flex justify-center">
          <span class="rounded-full bg-black/60 px-3 py-1 text-xs tabular-nums text-white">{{ current + 1 }} / {{ items.length }}</span>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* 纯淡入淡出（同 AppDialog 外层过渡，无 @keyframes） */
.preview-enter-active,
.preview-leave-active {
  transition: opacity 0.2s ease;
}
.preview-enter-from,
.preview-leave-to {
  opacity: 0;
}
</style>
