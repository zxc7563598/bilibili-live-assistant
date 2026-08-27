<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="账户记录" />
    <div class="sticky top-14 z-30 bg-bg px-4 py-3">
      <AppSegmentedControl v-model="type" :options="options" class="w-full" @click="updateType" />
    </div>
    <main class="mx-auto w-full max-w-5xl space-y-3 px-4">
      <div v-for="l in list" :key="l.id" class="card flex items-center gap-3 p-4">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full" :class="l.positive ? 'text-success' : 'text-danger'" :style="{ background: `color-mix(in srgb, ${l.positive ? 'var(--success)' : 'var(--danger)'} 10%, transparent)` }">
          <AppIcon :name="l.positive ? 'plus' : 'minus'" :size="20" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium">
            {{ l.title }}
          </p>
          <p class="mt-0.5 flex items-center gap-1 text-xs text-fg-3">
            <AppIcon :name="l.type === 0 ? 'star' : 'points'" :size="12" />{{ l.time }}
          </p>
        </div>
        <div class="shrink-0 text-right">
          <p class="font-bold tabular-nums" :class="l.positive ? 'text-success' : 'text-fg'">
            {{ l.value }}
          </p>
          <p class="mt-0.5 text-xs text-fg-3 tabular-nums">
            余额 {{ l.balance }}
          </p>
        </div>
      </div>
      <div v-if="!list.length" class="py-10 text-center text-sm text-fg-3">
        <span v-if="loading" class="inline-flex items-center gap-1">
          <AppIcon name="refresh" :size="14" class="animate-spin" />
          加载中...
        </span>
        <span v-else>暂无记录</span>
      </div>
      <div ref="sentinelRef" class="py-6 text-center text-sm text-fg-3">
        <span v-if="loading && list.length" class="inline-flex items-center gap-1">
          <AppIcon name="refresh" :size="14" class="animate-spin" />
          加载中...
        </span>
        <span v-else-if="finished && list.length">—— 没有更多了 ——</span>
      </div>
    </main>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import toast from '@/utils/toast'
import api from './api'

const type = ref(0)
const options = [
  { label: '积分', value: 0 },
  { label: '星光', value: 1 },
]

const page = ref(1)
const total = ref(0)
const list = ref([])
const loading = ref(false)
const finished = ref(false)
const sentinelRef = ref(null)
let requestSeq = 0
let observer = null
let prevType = 0

function loadList() {
  if (loading.value || finished.value)
    return
  loading.value = true
  const seq = ++requestSeq
  api.getAssetList(page.value, type.value).then((res) => {
    if (seq !== requestSeq)
      return
    if (res.code === 0) {
      total.value = res.data.total
      list.value = [...list.value, ...res.data.pageData]
      page.value += 1
      if (res.data.pageData.length === 0 || list.value.length >= total.value)
        finished.value = true
    }
    else {
      toast.error(res.msg)
    }
  }).finally(() => {
    if (seq === requestSeq)
      loading.value = false
  })
}

// 切换类型：立即作废在途请求、清空列表并回到顶部，再拉取新类型的数据
function updateType() {
  // 重复点击当前 tab 不触发刷新，避免无意义的清空 + loading 闪烁
  if (prevType === type.value)
    return
  prevType = type.value
  requestSeq++
  list.value = []
  page.value = 1
  total.value = 0
  finished.value = false
  loading.value = false
  window.scrollTo({ top: 0 })
  loadList()
}

onMounted(() => {
  loadList()
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting)
      loadList()
  }, { rootMargin: '0px 0px 200px 0px' })
  observer.observe(sentinelRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
})
</script>
