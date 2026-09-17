<template>
  <div class="min-h-dvh pb-28">
    <header class="safe-top bg-bg">
      <div class="mx-auto w-full max-w-5xl px-4 pt-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-on-primary">
              <template v-if="!userLoading">
                <AppImage v-if="user.avatar" :src="user.avatar" rounded="rounded-full" />
                <AppIcon v-else name="store" :size="20" />
              </template>
            </div>
            <div>
              <p v-if="userLoading" class="text-xs text-fg-3">
                <AppSkeleton class="h-3 w-16" rounded="rounded-full" />
              </p>
              <p v-else class="text-xs text-fg-3">
                你好，{{ user.name }}
              </p>
            </div>
          </div>
          <div v-if="userLoading" class="flex items-center gap-1 rounded-full bg-primary-soft px-3 py-1.5">
            <AppSkeleton class="h-5 w-24" rounded="rounded-full" />
          </div>
          <div v-else class="flex items-center gap-1 rounded-full bg-primary-soft px-3 py-1.5 text-primary">
            <AppIcon name="points" :size="16" />
            <span class="text-sm font-bold tabular-nums">{{ user.points }}</span>
            ｜
            <AppIcon name="star" :size="16" />
            <span class="text-sm font-bold tabular-nums">{{ user.stars }}</span>
          </div>
        </div>
        <div class="mt-3 flex items-center gap-2 rounded-2xl border border-line bg-surface px-4 py-2.5" @click="focusSearchInput">
          <AppIcon :name="searchLoading ? 'refresh' : 'search'" :size="18" :class="searchLoading ? 'animate-spin' : ''" class="text-fg-3" />
          <input ref="searchRef" v-model="search" class="flex-1 bg-transparent text-sm outline-none placeholder:text-fg-3" placeholder="搜索商品" enterkeyhint="search" @keyup.enter="handleSearch" @input="handleInputChange">
        </div>
      </div>
    </header>
    <div v-if="categories.length" class="no-scrollbar mx-auto mt-3 flex w-full max-w-5xl gap-2 overflow-x-auto px-4">
      <button v-for="(c, i) in categories" :key="c" class="shrink-0 rounded-full px-4 py-1.5 text-sm font-medium transition" :class="i === 0 ? 'bg-primary text-on-primary' : 'bg-surface text-fg-2'">
        {{ c }}
      </button>
    </div>
    <main class="mx-auto w-full max-w-5xl px-4 pt-4">
      <!-- 首屏 / 搜索骨架屏 -->
      <div v-if="!products.length && loading" class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
        <div v-for="n in 6" :key="n" class="card overflow-hidden">
          <AppSkeleton class="aspect-square" rounded="rounded-none" />
          <div class="space-y-2 p-3">
            <AppSkeleton class="h-4 w-3/4" />
            <AppSkeleton class="h-4 w-1/2" />
          </div>
        </div>
      </div>
      <!-- 商品列表 -->
      <TransitionGroup v-else name="list" tag="div" class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
        <router-link v-for="(p, i) in products" :key="`${p.id}-${i}`" :to="`/details/${p.id}`" class="card press block overflow-hidden">
          <AppImage :src="p.cover" :label="p.name" ratio="1 / 1" rounded="rounded-none" />
          <div class="p-3">
            <p class="line-clamp-2 text-sm font-medium leading-snug">
              {{ p.name }}
            </p>
            <div class="mt-2 flex items-end justify-between">
              <span class="flex items-center gap-1" :class="p.credit_type === 0 ? 'text-starlight' : 'text-primary'">
                <AppIcon :name="p.credit_type === 0 ? 'star' : 'points'" :size="15" />
                <span class="text-base font-bold leading-none tabular-nums">{{ p.price }}</span>
              </span>
              <span class="text-xs text-fg-3 tabular-nums" />
            </div>
          </div>
        </router-link>
      </TransitionGroup>
      <!-- 空态 -->
      <AppEmpty v-if="!loading && finished && !products.length" icon="store" title="暂无商品" description="换个关键词试试" />
      <div ref="sentinelRef" class="py-6 text-center text-sm text-fg-3">
        <span v-if="loading && products.length" class="inline-flex items-center gap-1">
          <AppIcon name="refresh" :size="14" class="animate-spin" />
          加载中...
        </span>
        <span v-else-if="finished && products.length">—— 没有更多了 ——</span>
      </div>
    </main>
    <TabBar />
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import TabBar from '@/components/base/TabBar.vue'
import toast from '@/utils/toast'
import api from './api'

const user = ref({
  uid: '00000',
  name: '未知用户',
  avatar: '',
  points: 0, // 积分余额
  stars: 0, // 星光余额
})

const userLoading = ref(true)

const categories = ref([])
const products = ref([])

const timer = ref(null)
const searchRef = ref(null)
const search = ref('')
const searchLoading = ref(false)

const page = ref(1)
const total = ref(0)
const loading = ref(false)
const finished = ref(false)
const sentinelRef = ref(null)
let requestSeq = 0
let observer = null

function focusSearchInput() {
  searchRef.value?.focus()
}

function handleInputChange() {
  // 清除上一次的定时器
  if (timer.value)
    clearTimeout(timer.value)
  // 如果搜索词为空，可以立即清空结果，不用等500ms
  if (!search.value.trim()) {
    resetAndLoad()
    return
  }
  // 设置防抖，500ms 后执行搜索
  if (!searchLoading.value) {
    timer.value = setTimeout(resetAndLoad, 500)
  }
}

function handleSearch() {
  // 清除未执行的防抖定时器，避免重复请求
  if (timer.value) {
    clearTimeout(timer.value)
    timer.value = null
  }
  const keyword = search.value.trim()
  if (!keyword) {
    console.warn('请输入搜索关键词')
    return
  }
  // 执行搜索逻辑
  if (!searchLoading.value)
    resetAndLoad()
}

function loadList() {
  if (loading.value || finished.value)
    return
  loading.value = true
  if (search.value) {
    searchLoading.value = true
  }
  const seq = ++requestSeq
  api.getShopList(page.value, search.value).then((res) => {
    if (seq !== requestSeq)
      return
    if (res.code === 0) {
      total.value = res.data.total
      products.value = [...products.value, ...res.data.pageData]
      page.value += 1
      if (res.data.pageData.length === 0 || products.value.length >= total.value)
        finished.value = true
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    if (seq === requestSeq)
      toast.error('加载失败，请重试')
  }).finally(() => {
    if (seq === requestSeq) {
      loading.value = false
      searchLoading.value = false
    }
  })
}

function resetAndLoad() {
  requestSeq++
  products.value = []
  page.value = 1
  total.value = 0
  finished.value = false
  loading.value = false
  searchLoading.value = false
  window.scrollTo({ top: 0 })
  loadList()
}

onMounted(() => {
  api.getUserInfo().then((res) => {
    if (res.code === 0) {
      Object.assign(user.value, res.data)
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
  }).finally(() => {
    userLoading.value = false
  })
  resetAndLoad()
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting)
      loadList()
  }, { rootMargin: '0px 0px 200px 0px' })
  observer.observe(sentinelRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
  if (timer.value)
    clearTimeout(timer.value)
})
</script>
