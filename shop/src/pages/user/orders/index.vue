<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="我的订单" />
    <div class="sticky top-14 z-30 bg-bg px-4 py-3">
      <AppSegmentedControl v-model="status" :options="options" class="w-full" />
    </div>
    <main class="mx-auto w-full max-w-5xl px-4">
      <TransitionGroup name="list" tag="div" class="space-y-3">
        <div v-for="o in data.pageData" :key="o.id" class="card p-4">
          <div class="flex items-center justify-between">
            <span class="text-xs text-fg-3">订单号 {{ o.order_sn }}</span>
            <Tag :color="statusColor[o.order_status]">
              {{ getLabel(o.order_status) }}
            </Tag>
          </div>
          <div class="mt-3 flex items-center gap-3">
            <div class="w-16 shrink-0">
              <AppImage :src="o.product_cover" :label="o.product_name" ratio="1 / 1" rounded="rounded-lg" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="line-clamp-2 text-sm font-medium leading-snug">
                {{ o.product_name }}
              </p>
              <p class="mt-1 text-xs text-fg-3">
                {{ getSku(o.product_spec_properties) }} · x{{ o.quantity }}
              </p>
            </div>
            <div class="flex shrink-0 items-center gap-1" :class="o.credit_type === 0 ? 'text-starlight' : 'text-primary'">
              <AppIcon :name="o.credit_type === 0 ? 'star' : 'points'" :size="15" />
              <span class="font-bold tabular-nums">{{ o.price * o.quantity }}</span>
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between border-t border-line pt-3 text-xs text-fg-3">
            <span class="flex items-center gap-1"><AppIcon name="clock" :size="13" />{{ o.pay_at || o.created_at }}</span>
          </div>
        </div>
      </TransitionGroup>
      <div v-if="loading && !data.pageData.length" class="py-10 text-center text-sm text-fg-3">
        <span class="inline-flex items-center gap-1">
          <AppIcon name="refresh" :size="14" class="animate-spin" />
          加载中...
        </span>
      </div>
      <AppEmpty v-else-if="!data.pageData.length" icon="box" title="暂无相关订单" />
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
import { onMounted, onUnmounted, ref, watch } from 'vue'
import toast from '@/utils/toast'
import api from './api'

const status = ref(null)
const page = ref(1)
const options = [
  { label: '全部', value: null },
  { label: '待付款', value: 0 },
  { label: '待发货', value: 1 },
  { label: '待收货', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已取消', value: 4 },
  { label: '售后中', value: 5 },
]
const statusColor = {
  0: 'warning',
  1: 'info',
  2: 'info',
  3: 'success',
  4: 'warning',
  5: 'danger',
}
function getLabel(value) {
  const option = options.find(item => item.value === value)
  return option ? option.label : ''
}
function getSku(sku) {
  try {
    const list = JSON.parse(sku)
    return Array.isArray(list) ? list.map(item => Object.values(item)[0]).join('、') : ''
  }
  catch {
    return ''
  }
}

const data = ref({ pageData: [], total: 0 })
const loading = ref(false)
const finished = ref(false)
const sentinelRef = ref(null)
let requestSeq = 0
let observer = null

function loadList() {
  if (loading.value || finished.value)
    return
  loading.value = true
  const seq = ++requestSeq
  api.getOrderList(page.value, 20, status.value).then((res) => {
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
  }).catch(() => {
    if (seq === requestSeq)
      toast.error('加载失败，请重试')
  }).finally(() => {
    if (seq === requestSeq)
      loading.value = false
  })
}

// 切换状态：立即作废在途请求、清空列表并回到顶部，再拉取新状态的数据
// 只监听 status 值真实变化（重复点击当前 tab、拖动分段条不触发），彻底摆脱 click 事件目标不确定带来的误刷新
watch(status, () => {
  requestSeq++
  data.value = { pageData: [], total: 0 }
  page.value = 1
  finished.value = false
  loading.value = false
  window.scrollTo({ top: 0 })
  loadList()
})

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
