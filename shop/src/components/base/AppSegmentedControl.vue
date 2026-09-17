<script setup>
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] }, // [{ label, value }]
})
const emit = defineEmits(['update:modelValue'])

const root = ref(null)
const canScroll = ref(false)
const dragging = ref(false)
const DRAG_THRESHOLD = 5

let pointer = null // { id, startX, left, captured }
let suppressClick = false
let resetTimer = 0
let resizeObserver = null

function scrollable() {
  const el = root.value
  return !!(el && el.scrollWidth > el.clientWidth + 1)
}
function refreshScrollable() {
  canScroll.value = scrollable()
}

// 单击才选中；拖动结束时抑制紧随的 click，避免误触发选中/外层 @click 刷新
function onSelect(value, e) {
  if (suppressClick) {
    e.stopPropagation() // 同时掐断冒泡到根的调用方 @click
    return
  }
  emit('update:modelValue', value)
}

// 鼠标拖动滚动（touch/pen 一律放行，交给原生横滑）
function onPointerDown(e) {
  const el = root.value
  if (e.pointerType !== 'mouse' || !scrollable() || !el)
    return
  pointer = { id: e.pointerId, startX: e.clientX, left: el.scrollLeft }
}
function onPointerMove(e) {
  const el = root.value
  if (!pointer || !el || e.pointerId !== pointer.id)
    return
  const dx = e.clientX - pointer.startX
  if (!pointer.captured) {
    if (Math.abs(dx) <= DRAG_THRESHOLD)
      return
    pointer.captured = true
    dragging.value = true
    try {
      el.setPointerCapture(e.pointerId)
    }
    catch {
      // 捕获失败时忽略：未 capture 仅影响拖出条外的跟随
    }
  }
  el.scrollLeft = pointer.left - dx
}
function endDrag(e) {
  if (!pointer || e.pointerId !== pointer.id)
    return
  const moved = Math.abs(e.clientX - pointer.startX) > DRAG_THRESHOLD
  pointer = null
  dragging.value = false
  if (moved) {
    suppressClick = true
    clearTimeout(resetTimer)
    resetTimer = window.setTimeout(() => {
      suppressClick = false
    }, 0)
  }
}
const onPointerUp = endDrag
const onPointerCancel = endDrag // cancel 后无 click，置位再复位无副作用

// 选中项溢出时滚入可视区：只动容器 scrollLeft，绝不滚动页面
function scrollActiveIntoView() {
  const el = root.value
  if (!el || !scrollable())
    return
  const i = props.options.findIndex(o => o.value === props.modelValue)
  if (i < 0 || i >= el.children.length)
    return
  const btn = el.children[i]
  const cr = el.getBoundingClientRect()
  const br = btn.getBoundingClientRect()
  if (br.left < cr.left)
    el.scrollLeft -= cr.left - br.left
  else if (br.right > cr.right)
    el.scrollLeft += br.right - cr.right
}
watch(() => props.modelValue, () => nextTick(scrollActiveIntoView))

// 纵向滚轮 → 横向（仅可滚动且有位移余量时接管）
function onWheel(e) {
  const el = root.value
  if (!el || !scrollable())
    return
  const { deltaX, deltaY } = e
  if (Math.abs(deltaX) > Math.abs(deltaY))
    return
  if ((deltaY < 0 && el.scrollLeft <= 0)
    || (deltaY > 0 && el.scrollLeft >= el.scrollWidth - el.clientWidth - 1)) {
    return
  }
  e.preventDefault()
  el.scrollLeft += deltaY
}

onMounted(() => {
  nextTick(() => {
    refreshScrollable()
    scrollActiveIntoView()
  })
  resizeObserver = new ResizeObserver(refreshScrollable)
  resizeObserver.observe(root.value)
})
onUnmounted(() => {
  resizeObserver?.disconnect()
})
</script>

<template>
  <div
    ref="root"
    class="inline-flex max-w-full no-scrollbar select-none overflow-x-auto rounded-full bg-bg-soft p-1"
    :class="canScroll ? (dragging ? 'cursor-grabbing' : 'cursor-grab') : ''"
    @pointerdown="onPointerDown"
    @pointermove="onPointerMove"
    @pointerup="onPointerUp"
    @pointercancel="onPointerCancel"
    @wheel="onWheel"
  >
    <button
      v-for="o in options"
      :key="o.value"
      type="button"
      class="flex-1 min-w-max whitespace-nowrap rounded-full px-4 py-1.5 text-sm font-medium transition"
      :class="modelValue === o.value ? 'bg-surface text-primary shadow-card' : 'text-fg-2'"
      @click="onSelect(o.value, $event)"
    >
      {{ o.label }}
    </button>
  </div>
</template>
