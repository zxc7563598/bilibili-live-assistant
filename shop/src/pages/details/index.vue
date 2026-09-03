<template>
  <div class="min-h-dvh pb-24">
    <!-- 骨架屏 -->
    <template v-if="loading">
      <div class="relative">
        <AppSkeleton class="aspect-square w-full" rounded="rounded-none" />
        <div class="safe-top absolute inset-x-0 top-0 z-10 flex items-center justify-between px-3 pt-2">
          <div class="flex h-9 w-9 items-center justify-center rounded-full bg-black/25 text-white backdrop-blur">
            <AppIcon name="chevron-left" :size="22" />
          </div>
        </div>
      </div>
      <section class="card mx-auto mt-4 w-[calc(100%-2rem)] max-w-5xl space-y-3 p-4">
        <AppSkeleton class="h-6 w-3/4" />
        <div class="flex items-center justify-between">
          <AppSkeleton class="h-4 w-1/2" />
          <AppSkeleton class="h-6 w-16" />
        </div>
        <AppSkeleton class="h-4 w-2/3" />
      </section>
      <section class="card mx-auto mt-3 w-[calc(100%-2rem)] max-w-5xl space-y-3 p-4">
        <AppSkeleton class="h-4 w-20" />
        <div class="flex gap-2">
          <AppSkeleton class="h-8 w-16" rounded="rounded-full" />
          <AppSkeleton class="h-8 w-16" rounded="rounded-full" />
          <AppSkeleton class="h-8 w-16" rounded="rounded-full" />
        </div>
      </section>
    </template>
    <template v-else-if="product.id">
      <div class="relative">
        <AppCarousel :items="carouselImages" ratio="4 / 3" interval="5000" />
        <div class="safe-top absolute inset-x-0 top-0 z-10 flex items-center justify-between px-3 pt-2">
          <button class="flex h-9 w-9 items-center justify-center rounded-full bg-black/25 text-white backdrop-blur press" aria-label="返回" @click="onBack">
            <AppIcon name="chevron-left" :size="22" />
          </button>
          <!-- <div class="flex items-center gap-2">
            <button class="flex h-9 w-9 items-center justify-center rounded-full bg-black/25 text-white backdrop-blur press" aria-label="收藏">
              <AppIcon name="heart" :size="20" />
            </button>
            <button class="flex h-9 w-9 items-center justify-center rounded-full bg-black/25 text-white backdrop-blur press" aria-label="分享">
              <AppIcon name="share" :size="20" />
            </button>
          </div> -->
        </div>
      </div>
      <section class="card mx-auto mt-4 w-[calc(100%-2rem)] max-w-5xl p-4">
        <div class="flex items-start justify-between gap-3">
          <h1 class="text-lg font-bold leading-snug">
            {{ product.name }}
          </h1>
          <div class="flex shrink-0 items-center gap-1" :class="priceColor">
            <AppIcon :name="isStars ? 'star' : 'points'" :size="18" />
            <span class="text-xl font-bold tabular-nums">{{ displayPrice }}</span>
          </div>
        </div>
        <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-fg-3">
          <span>已兑换 {{ product.sold }}</span>
          <span v-if="specs.length && !activeSku" class="text-fg-3">请选择规格</span>
          <span v-else :class="soldOut ? 'text-danger' : 'text-success'">
            {{ soldOut ? '已售罄' : `库存 ${displayStock}` }}
          </span>
        </div>
        <div v-if="tagList.length" class="mt-3 flex flex-wrap gap-1.5">
          <Tag v-for="t in tagList" :key="t" color="primary">
            {{ t }}
          </Tag>
        </div>
      </section>
      <section v-if="specs.length" class="card mx-auto mt-3 w-[calc(100%-2rem)] max-w-5xl p-4">
        <p class="text-sm font-semibold">
          选择规格
        </p>
        <div v-for="sp in specs" :key="sp.id" class="mt-3">
          <p class="text-xs text-fg-3">
            {{ sp.key_name }}
          </p>
          <div class="mt-2 flex flex-wrap gap-2">
            <button v-for="opt in sp.values" :key="opt.id" class="rounded-full border px-3.5 py-1.5 text-sm transition" :class="optionClass(sp.key_name, opt.value_name)" :disabled="!optionAvailable(sp.key_name, opt.value_name)" @click="select(sp.key_name, opt.value_name)">
              {{ opt.value_name }}
            </button>
          </div>
        </div>
        <div class="mt-4 flex items-center justify-between">
          <p class="text-sm font-semibold">
            数量
          </p>
          <div class="flex items-center gap-3">
            <button class="flex h-8 w-8 items-center justify-center rounded-full border border-line press" :disabled="count <= 1" @click="count--">
              <AppIcon name="minus" :size="16" />
            </button>
            <span class="w-6 text-center font-semibold tabular-nums">{{ count }}</span>
            <button class="flex h-8 w-8 items-center justify-center rounded-full border border-line press" :disabled="soldOut || count >= maxCount" @click="count++">
              <AppIcon name="plus" :size="16" />
            </button>
          </div>
        </div>
      </section>
      <section class="mx-auto mt-3 w-full max-w-5xl">
        <p class="px-4 py-2 text-sm font-semibold">
          商品详情
        </p>
        <p v-if="product.describe" class="px-4 pb-2 text-sm text-fg-3">
          {{ product.describe }}
        </p>
        <AppImage v-for="(item, key) in detailImages" :key="key" :src="item.src" :alt="product.name" ratio="none" rounded="rounded-none" />
      </section>
      <div class="glass-strong safe-bottom fixed inset-x-0 bottom-0 z-40">
        <div class="mx-auto flex w-full max-w-5xl items-center gap-4 px-4 py-2.5">
          <div class="shrink-0">
            <p class="text-xs text-fg-3">
              {{ isStars ? '星光' : '积分' }}合计
            </p>
            <p class="mt-0.5 flex items-center gap-1" :class="priceColor">
              <AppIcon :name="isStars ? 'star' : 'points'" :size="18" />
              <span class="text-xl font-bold tabular-nums">{{ displayPrice * count }}</span>
            </p>
          </div>
          <AppButton size="md" class="flex-1" :disabled="!canBuy" :loading="placeOrderLoading" @click="placeOrder">
            立即兑换
          </AppButton>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()
const route = useRoute()
// 商品详情原始返回（/api/shop/product/detail），字段对齐后端契约
const product = ref({})
const loading = ref(true)
// 已选规格组合：{ 规格key名: 规格值名 }，允许部分选择（行缺席 = 该行未选）。
// 交互：点可选值 = 选中/替换；再点已选值 = 取消选中、释放该行约束。
const selected = ref({})
const count = ref(1)

// —— 派生展示数据 ——

const specs = computed(() => product.value.specs ?? [])
const specKeys = computed(() => specs.value.map(s => s.key_name))

// 解析每个 SKU 的 spec_properties（JSON 字符串，元素 {规格key名: 规格值名}），得到可匹配的 combo
const rawSkus = computed(() => {
  const list = []
  for (const sku of product.value.skus ?? []) {
    try {
      const props = JSON.parse(sku.spec_properties)
      if (!Array.isArray(props))
        continue
      const combo = {}
      for (const item of props) {
        const key = Object.keys(item)[0]
        if (key)
          combo[key] = item[key]
      }
      if (Object.keys(combo).length)
        list.push({ id: sku.id, price: sku.price, stock: sku.stock, combo })
    }
    catch {
      // 单条规格快照异常时跳过该 SKU
      continue
    }
  }
  return list
})

// 轮播图：images 中 type=0；缺失时兜底封面图
const carouselImages = computed(() => {
  const banners = (product.value.images ?? []).filter(i => i.type === 0)
  const list = banners.map(i => ({ src: i.image_path, label: product.value.name ?? '' }))
  if (list.length)
    return list
  return product.value.cover ? [{ src: product.value.cover, label: product.value.name ?? '' }] : []
})

// 底部详情图：images 中 type=1
const detailImages = computed(() =>
  (product.value.images ?? []).filter(i => i.type === 1).map(i => ({ src: i.image_path, label: product.value.name ?? '' })),
)

const tagList = computed(() =>
  (product.value.tags ?? '').split(',').map(t => t.trim()).filter(Boolean),
)

// 价格货币类型：0 星光 / 1 积分
const isStars = computed(() => (product.value.credit_type ?? 0) === 0)
const priceColor = computed(() => (isStars.value ? 'text-starlight' : 'text-primary'))

// —— SKU 匹配 ——

// 按 key_name 精确比对，找到与 combo 完全一致的 SKU
function resolveSku(combo) {
  if (specKeys.value.length === 0)
    return null
  return rawSkus.value.find(s => specKeys.value.every(k => s.combo[k] === combo[k])) ?? null
}

// 是否已点齐全部规格维度（无规格商品恒为 true）
const selectionComplete = computed(() =>
  specs.value.length === 0 || specKeys.value.every(k => selected.value[k] !== undefined),
)

// 只有「点齐全部维度」才可能命中真实 SKU
const activeSku = computed(() =>
  selectionComplete.value ? resolveSku(selected.value) : null,
)

// 展示价/库存：选中 SKU 优先，否则回退商品顶层字段
const displayPrice = computed(() => activeSku.value?.price ?? product.value.price ?? 0)
const displayStock = computed(() => activeSku.value?.stock ?? product.value.stock ?? 0)
const soldOut = computed(() => displayStock.value <= 0)
// 多规格商品未选齐时不允许先调数量（上限钳到 1，对应「+」禁用）
const maxCount = computed(() => {
  if (specs.value.length && !activeSku.value)
    return 1
  return Math.max(activeSku.value ? activeSku.value.stock : (product.value.stock ?? 0), 0)
})
const canBuy = computed(() => {
  if (specs.value.length && !activeSku.value)
    return false
  return displayStock.value > 0
})
const placeOrderLoading = ref(false)

// 值是否可点：
// - 当前已选中的值 → 恒可点（用于再次点击取消）
// - 其它值 → 须存在一条「该维度=val、且与当前已选其它行一致、且库存>0」的 SKU
function optionAvailable(specKey, val) {
  if (selected.value[specKey] === val)
    return true
  return rawSkus.value.some((s) => {
    if (s.combo[specKey] !== val || s.stock <= 0)
      return false
    return Object.entries(selected.value).every(([k, v]) => k === specKey || s.combo[k] === v)
  })
}

function optionClass(specKey, val) {
  if (selected.value[specKey] === val)
    return 'border-primary bg-primary-soft text-primary'
  if (!optionAvailable(specKey, val))
    return 'border-line text-fg-3 opacity-50'
  return 'border-line text-fg-2'
}

// 点击规格值：点其它可用值 = 选中/替换；再点已选值 = 取消选中、解除该行约束。
// 不做自动协调——可用性已由 optionAvailable 保证，selected 始终是某条真实 SKU 投影的子集。
function select(specKey, val) {
  const next = { ...selected.value }
  if (next[specKey] === val) {
    delete next[specKey]
    count.value = 1
  }
  else {
    next[specKey] = val
    const sku = resolveSku(next)
    if (sku && count.value > sku.stock)
      count.value = Math.max(1, sku.stock)
  }
  selected.value = next
}

// 数据就绪后初始化选择状态：多规格商品默认不预选；
// 仅当商品只有一个 SKU 组合（如 Q版板绘）时兜底选中，用户仍可再点取消
function initSelection() {
  const only = rawSkus.value.length === 1 && specs.value.length ? { ...rawSkus.value[0].combo } : {}
  selected.value = only
  count.value = 1
}

function onBack() {
  router.back()
}

function fetchDetail(id) {
  const num = Number(id)
  loading.value = true
  api.getShopDetails(num).then((res) => {
    // 详情页间切换时，旧请求的结果可能后到，按当前路由 id 丢弃过期响应
    if (Number(route.params.id) !== num)
      return
    if (res.code === 0) {
      product.value = res.data ?? {}
      initSelection()
    }
    else {
      toast.error(res.msg)
      router.replace('/')
    }
  }).catch(() => {
    if (Number(route.params.id) !== num)
      return
    toast.error('加载失败，请重试')
    router.replace('/')
  }).finally(() => {
    if (Number(route.params.id) === num)
      loading.value = false
  })
}

function placeOrder() {
  placeOrderLoading.value = true
  api.orderPlace(activeSku.value.id, count.value).then((res) => {
    // 详情页间切换时，旧请求的结果可能后到，按当前路由 id 丢弃过期响应
    if (res.code === 0) {
      toast.success('下单成功')
      router.push('/confirm')
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
    router.replace('/')
  }).finally(() => {
    placeOrderLoading.value = false
  })
}

onMounted(() => {
  fetchDetail(route.params.id)
})

// 详情页若发生直接跳转（如后续的推荐位），组件实例会被复用、onMounted 不会重新触发，
// 需监听路由参数变化重新拉取数据
watch(() => route.params.id, id => fetchDetail(id))
</script>
