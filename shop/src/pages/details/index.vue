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
        <AppCarousel :items="product.slides" ratio="1 / 1" interval="5000" />
        <div class="safe-top absolute inset-x-0 top-0 z-10 flex items-center justify-between px-3 pt-2">
          <button class="flex h-9 w-9 items-center justify-center rounded-full bg-black/25 text-white backdrop-blur press" aria-label="返回" @click="onBack">
            <AppIcon name="chevron-left" :size="22" />
          </button>
          <!-- <div class="flex items-center gap-2">
            <button class="flex h-9 w-9 items-center justify-center rounded-full bg-black/25 text-white backdrop-blur press" :class="liked ? 'text-danger' : ''" aria-label="收藏" @click="liked = !liked">
              <AppIcon name="heart" :size="20" :class="liked ? 'fill-current' : ''" />
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
          <div class="flex shrink-0 items-center gap-1" :class="[product.color]">
            <AppIcon :name="product.type === 0 ? 'star' : 'points'" :size="18" />
            <span class="text-xl font-bold tabular-nums">{{ product.amount }}</span>
          </div>
        </div>
        <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-fg-3">
          <span>已兑换 {{ product.sold }}</span>
          <span :class="product.stock > 0 ? 'text-success' : 'text-danger'">
            {{ product.stock > 0 ? `库存 ${product.stock}` : '已售罄' }}
          </span>
        </div>
        <div v-if="product.tags.length" class="mt-3 flex flex-wrap gap-1.5">
          <Tag v-for="t in product.tags" :key="t" color="primary">
            {{ t }}
          </Tag>
        </div>
      </section>
      <section v-if="product.skuGroups.length" class="card mx-auto mt-3 w-[calc(100%-2rem)] max-w-5xl p-4">
        <p class="text-sm font-semibold">
          选择规格
        </p>
        <div v-for="g in product.skuGroups" :key="g.name" class="mt-3">
          <p class="text-xs text-fg-3">
            {{ g.name }}
          </p>
          <div class="mt-2 flex flex-wrap gap-2">
            <button
              v-for="opt in g.options" :key="opt"
              class="rounded-full border px-3.5 py-1.5 text-sm transition"
              :class="selected[g.name] === opt ? 'border-primary bg-primary-soft text-primary' : 'border-line text-fg-2'"
              @click="select(g.name, opt)"
            >
              {{ opt }}
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
            <button class="flex h-8 w-8 items-center justify-center rounded-full border border-line press" @click="count++">
              <AppIcon name="plus" :size="16" />
            </button>
          </div>
        </div>
      </section>
      <section class="mx-auto mt-3 w-full max-w-5xl">
        <p class="px-4 py-2 text-sm font-semibold">
          商品详情
        </p>
        <p class="px-4 pb-2 text-sm text-fg-3">
          {{ product.desc }}
        </p>
        <AppImage v-for="(item, key) in product.detail_images" :key="key" :src="item.src" :label="item.label" ratio="none" rounded="rounded-none" />
      </section>
      <div class="glass-strong safe-bottom fixed inset-x-0 bottom-0 z-40">
        <div class="mx-auto flex w-full max-w-5xl items-center gap-4 px-4 py-2.5">
          <div class="shrink-0">
            <p class="text-xs text-fg-3">
              {{ product.type === 0 ? '星光' : '积分' }}合计
            </p>
            <p class="mt-0.5 flex items-center gap-1" :class="product.type === 0 ? 'text-starlight' : 'text-primary'">
              <AppIcon :name="product.type === 0 ? 'star' : 'points'" :size="18" />
              <span class="text-xl font-bold tabular-nums">{{ product.amount * count }}</span>
            </p>
          </div>
          <AppButton size="md" class="flex-1" @click="router.push('/confirm')">
            立即兑换
          </AppButton>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()
const route = useRoute()
// 死数据：直接取第一件商品。接入接口时替换为「当前点击的那件商品」即可。
const product = ref({
  id: 0,
  name: '',
  slides: [],
  amount: 0,
  type: 0,
  sold: 0,
  stock: 0,
  tags: [],
  skuGroups: [],
  desc: '',
  detail_images: [],
})

const loading = ref(true)
const selected = ref({})
const count = ref(1)

function select(group, opt) {
  selected.value = { ...selected.value, [group]: opt }
}
function onBack() {
  router.back()
}

onMounted(() => {
  api.getShopDetails(route.params.id).then((res) => {
    if (res.code === 0) {
      Object.assign(product.value, res.data)
    }
    else {
      toast.error(res.msg)
      router.replace('/')
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
    router.replace('/')
  }).finally(() => {
    loading.value = false
  })
})
</script>
