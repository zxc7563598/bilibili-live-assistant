<template>
  <div class="min-h-dvh bg-bg pb-28">
    <header class="safe-top bg-linear-to-b from-primary-soft to-transparent mb-4">
      <div class="mx-auto w-full max-w-5xl px-4 pt-6">
        <div class="flex items-center gap-4">
          <div class="flex h-16 w-16 items-center justify-center rounded-full bg-primary-soft text-primary">
            <template v-if="!userLoading">
              <AppImage v-if="user.avatar" :src="user.avatar" rounded="rounded-full" />
              <AppIcon v-else name="user" :size="32" />
            </template>
          </div>
          <div class="min-w-0 flex-1">
            <template v-if="userLoading">
              <AppSkeleton class="h-6 w-28" />
              <AppSkeleton class="mt-2 h-4 w-20 rounded-full" />
            </template>
            <template v-else>
              <p class="text-lg font-bold">
                {{ user.name }}
              </p>
              <button class="mt-1 flex items-center gap-1.5 rounded-full bg-surface px-2.5 py-1 text-xs text-fg-2 press" @click="copyUid">
                UID：{{ user.uid }}
                <AppIcon :name="copied ? 'check' : 'copy'" :size="13" :class="copied ? 'text-success' : ''" />
              </button>
            </template>
          </div>
          <Tag color="primary">
            普通用户
          </Tag>
        </div>
      </div>
    </header>
    <main class="mx-auto w-full max-w-5xl px-4">
      <div class="card relative overflow-hidden p-5">
        <div class="pointer-events-none absolute -right-10 -top-10 h-40 w-40 rounded-full opacity-20 blur-2xl" style="background: var(--primary)" />
        <p class="text-sm text-fg-3">
          我的资产
        </p>
        <div class="mt-3 grid grid-cols-2 gap-3">
          <div class="rounded-2xl bg-primary-soft p-3.5">
            <p class="flex items-center gap-1 text-xs text-fg-2">
              <AppIcon name="points" :size="14" class="text-primary" />积分
            </p>
            <p v-if="userLoading" class="mt-1.5">
              <AppSkeleton class="h-7 w-16" />
            </p>
            <p v-else class="mt-1.5 text-2xl font-extrabold text-primary tabular-nums">
              {{ user.points }}
            </p>
          </div>
          <div class="rounded-2xl bg-starlight-soft p-3.5">
            <p class="flex items-center gap-1 text-xs text-fg-2">
              <AppIcon name="star" :size="14" class="text-starlight" />星光
            </p>
            <p v-if="userLoading" class="mt-1.5">
              <AppSkeleton class="h-7 w-16" />
            </p>
            <p v-else class="mt-1.5 text-2xl font-extrabold text-starlight tabular-nums">
              {{ user.stars }}
            </p>
          </div>
        </div>
        <div v-if="roomLoading" class="mt-4">
          <AppSkeleton class="h-9 w-full rounded-full" />
        </div>
        <div v-else-if="roomID > 0" class="mt-4 flex flex-wrap gap-2">
          <AppButton size="sm" block @click="enterLive">
            <AppIcon name="tv" :size="16" />火速进入直播间爆米
          </AppButton>
        </div>
      </div>
      <section class="card mt-4 overflow-hidden p-1">
        <router-link v-for="m in menus" :key="m.to" :to="m.to" class="flex items-center gap-3 rounded-2xl px-4 py-3.5 press">
          <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary-soft text-primary">
            <AppIcon :name="m.icon" :size="18" />
          </div>
          <span class="flex-1 text-[15px] font-medium">{{ m.label }}</span>
          <AppIcon name="chevron-right" :size="18" class="text-fg-3" />
        </router-link>
      </section>
      <section class="card mt-4 overflow-hidden p-1">
        <div class="flex items-center gap-3 rounded-2xl px-4 py-3.5">
          <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary-soft text-primary">
            <AppIcon :name="isDark ? 'moon' : 'sun'" :size="18" />
          </div>
          <span class="flex-1 text-[15px] font-medium">暗夜模式</span>
          <AppSwitch :model-value="isDark" @update:model-value="toggleTheme" />
        </div>
        <button class="flex w-full items-center gap-3 rounded-2xl px-4 py-3.5 press disabled:opacity-60" :disabled="logoutLoading" @click="logout">
          <div class="flex h-9 w-9 items-center justify-center rounded-full text-danger" style="background: color-mix(in srgb, var(--danger) 10%, transparent)">
            <AppIcon :name="logoutLoading ? 'refresh' : 'logout'" :size="18" :class="logoutLoading ? 'animate-spin' : ''" />
          </div>
          <span class="flex-1 text-left text-[15px] font-medium text-danger">{{ logoutLoading ? '退出中...' : '退出登录' }}</span>
        </button>
      </section>
    </main>
    <TabBar />
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import TabBar from '@/components/base/TabBar.vue'
import { isDark, toggleTheme } from '@/utils/theme'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()
const copied = ref(false)

const user = ref({
  uid: '00000',
  name: '未知用户',
  avatar: '',
  points: 0, // 积分余额
  stars: 0, // 星光余额
})

const roomID = ref(0)
const userLoading = ref(true)
const roomLoading = ref(true)
const logoutLoading = ref(false)

async function copyUid() {
  try {
    await navigator.clipboard.writeText(user.value.uid)
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  }
  catch (e) {
    console.warn('复制失败', e)
  }
}

function enterLive() {
  window.open(`https://live.bilibili.com/${roomID.value}`, '_blank')
}

function logout() {
  if (logoutLoading.value)
    return
  logoutLoading.value = true
  api.logout().then((res) => {
    if (res.code === 0) {
      toast.success('退出成功')
      router.replace('/login')
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
  }).finally(() => {
    logoutLoading.value = false
  })
}

const menus = [
  { icon: 'box', label: '我的订单', to: '/user/orders' },
  { icon: 'clock', label: '账户记录', to: '/user/assets' },
  { icon: 'map-pin', label: '收货地址', to: '/user/address' },
  { icon: 'message', label: '投诉建议', to: '/user/feedback' },
  { icon: 'lock', label: '修改密码', to: '/user/password' },
]

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
  api.getRoomID().then((res) => {
    if (res.code === 0) {
      roomID.value = res.data.roomID
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
  }).finally(() => {
    roomLoading.value = false
  })
})
</script>
