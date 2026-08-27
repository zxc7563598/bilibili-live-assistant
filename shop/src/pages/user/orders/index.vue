<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="我的订单" />
    <div class="sticky top-14 z-30 bg-bg px-4 py-3">
      <AppSegmentedControl v-model="status" :options="options" class="w-full" @click="updateStatus" />
    </div>
    <main class="mx-auto w-full max-w-5xl space-y-3 px-4">
      <div v-for="o in data.pageData" :key="o.id" class="card p-4">
        <div class="flex items-center justify-between">
          <span class="text-xs text-fg-3">订单号 {{ o.id }}</span>
          <Tag :color="statusColor[o.status]">
            {{ o.statusText }}
          </Tag>
        </div>
        <div class="mt-3 flex items-center gap-3">
          <div class="w-16 shrink-0">
            <AppImage :src="o.cover" :label="o.title" ratio="1 / 1" rounded="rounded-lg" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="line-clamp-2 text-sm font-medium leading-snug">
              {{ o.title }}
            </p>
            <p class="mt-1 text-xs text-fg-3">
              {{ o.sku.join('·') }} · x{{ o.count }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-1" :class="o.type === 0 ? 'text-starlight' : 'text-primary'">
            <AppIcon :name="o.type === 0 ? 'star' : 'points'" :size="15" />
            <span class="font-bold tabular-nums">{{ o.amount }}</span>
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between border-t border-line pt-3 text-xs text-fg-3">
          <span class="flex items-center gap-1"><AppIcon name="clock" :size="13" />{{ o.time }}</span>
        </div>
      </div>
      <div v-if="!data.pageData.length" class="py-10 text-center text-sm text-fg-3">
        <span v-if="loading" class="inline-flex items-center gap-1">
          <AppIcon name="refresh" :size="14" class="animate-spin" />
          加载中...
        </span>
        <span v-else>暂无相关订单</span>
      </div>
      <div ref="sentinelRef" class="py-6 text-center text-sm text-fg-3">
        <span v-if="loading && data.pageData.length" class="inline-flex items-center gap-1">
          <AppIcon name="refresh" :size="14" class="animate-spin" />
          加载中...
        </span>
        <span v-else-if="finished && data.pageData.length">—— 没有更多了 ——</span>
      </div>
    </main>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import toast from '@/utils/toast'
import api from './api'

const status = ref(0)
const page = ref(1)
const options = [
  { label: '全部', value: 0 },
  { label: '待发货', value: 1 },
  { label: '已发货', value: 2 },
  { label: '已完成', value: 3 },
]

const data = ref({ pageData: [], total: 0 })
const loading = ref(false)
const finished = ref(false)
const sentinelRef = ref(null)
let requestSeq = 0
let observer = null
let prevStatus = 0

const statusColor = {
  1: 'warning',
  2: 'info',
  3: 'success',
}

function loadList() {
  if (loading.value || finished.value)
    return
  loading.value = true
  const seq = ++requestSeq
  api.getOrderList(page.value, status.value).then((res) => {
    if (seq !== requestSeq)
      return
    if (res.code === 0) {
      data.value.total = res.data.total
      data.value.pageData = [...data.value.pageData, ...res.data.pageData]
      page.value += 1
      if (res.data.pageData.length === 0 || data.value.pageData.length >= data.value.total)
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

// 切换状态：立即作废在途请求、清空列表并回到顶部，再拉取新状态的数据
function updateStatus() {
  // 重复点击当前 tab 不触发刷新，避免无意义的清空 + loading 闪烁
  if (prevStatus === status.value)
    return
  prevStatus = status.value
  requestSeq++
  data.value = { pageData: [], total: 0 }
  page.value = 1
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
