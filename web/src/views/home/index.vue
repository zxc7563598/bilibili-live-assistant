<template>
  <AppPage show-footer full>
    <section class="relative overflow-hidden rounded-12 px-24 py-22">
      <div class="absolute inset-0" style="background: linear-gradient(118deg, rgba(var(--primary-color), 0.98) 0%, rgba(var(--primary-color), 0.85) 45%, rgba(var(--primary-color), 0.62) 100%)" />
      <div class="absolute h-160 w-160 rounded-full bg-white/10 blur-50 -right-30 -top-50" />
      <div class="absolute right-160 h-140 w-140 rounded-full bg-white/8 blur-45 -bottom-60" />
      <div class="relative flex items-center justify-between gap-20">
        <div class="min-w-0">
          <h1 class="text-22 text-white font-medium">
            {{ greeting }}，{{ displayName }}
          </h1>
          <p class="mt-8 text-13 text-white/75 leading-relaxed">
            这里是你的直播间助手控制台。第一次使用的话，先把「房间配置」里的机器人和直播间设好，后面几步按下面走一遍就差不多了。
          </p>
          <div class="mt-16 inline-flex cursor-pointer items-center gap-6 rounded-8 bg-white px-14 py-7 text-13 text-primary transition-all-300 hover:bg-white/88" @click="router.push('/room')">
            <i class="i-fe:tv text-14" />
            <span>去配置直播间</span>
            <i class="i-fe:arrow-right text-13" />
          </div>
        </div>
      </div>
    </section>
    <section class="mt-12 flex gap-10 border border-amber-200 rounded-12 bg-amber-50 px-16 py-13 dark:border-amber-500/25 dark:bg-amber-500/10">
      <i class="i-fe:alert-triangle mt-1 shrink-0 text-15 text-amber-500" />
      <div class="min-w-0">
        <div class="text-13 text-amber-700 font-medium dark:text-amber-300">
          机器人建议只在本项目登录
        </div>
        <div class="mt-4 text-12 text-amber-700/80 leading-relaxed dark:text-amber-200/70">
          同一账号在多处同时登录并且都比较活跃时，登录状态可能会失效。如果发现机器人不再记录数据，重新登录一下基本就能解决。
        </div>
      </div>
    </section>
    <section class="mt-18">
      <div class="mb-12 flex items-center gap-8">
        <div class="h-14 w-3 rounded-l-2 bg-primary" />
        <h2 class="text-15 font-medium">
          先让机器人跑起来
        </h2>
        <span class="text-12 text-gray-400">按这个顺序配置就好</span>
      </div>

      <div class="grid grid-cols-1 gap-12 md:grid-cols-2 xl:grid-cols-3">
        <div v-for="item in startGuides" :key="item.title" class="group flex flex-col cursor-pointer card-border rounded-12 auto-bg p-16 transition-all-300 hover:card-shadow hover:-translate-y-4" @click="router.push(item.path)">
          <div class="flex items-start justify-between">
            <div class="h-36 w-36 f-c-c rounded-10 text-17 transition-all-300 group-hover:scale-108" :class="item.tone">
              <i :class="item.icon" />
            </div>
            <i class="i-fe:arrow-right mt-10 text-15 text-gray-300 transition-all-300 group-hover:translate-x-3 dark:text-gray-600 group-hover:text-primary" />
          </div>
          <div class="mt-12 text-14 font-medium">
            {{ item.title }}
          </div>
          <div class="mt-5 flex-1 text-12 text-gray-500 leading-relaxed dark:text-gray-400">
            {{ item.desc }}
          </div>
        </div>
      </div>
    </section>
    <section class="mt-18">
      <div class="mb-12 flex items-center gap-8">
        <div class="h-14 w-3 rounded-l-2 bg-primary" />
        <h2 class="text-15 font-medium">
          平时会用到的地方
        </h2>
      </div>
      <div class="grid grid-cols-1 gap-12 md:grid-cols-2">
        <div v-for="item in dailyGuides" :key="item.title" class="group flex flex-col cursor-pointer card-border rounded-12 auto-bg p-16 transition-all-300 hover:card-shadow hover:-translate-y-4" @click="router.push(item.path)">
          <div class="flex items-start justify-between">
            <div class="h-36 w-36 f-c-c rounded-10 text-17 transition-all-300 group-hover:scale-108" :class="item.tone">
              <i :class="item.icon" />
            </div>
            <i class="i-fe:arrow-right mt-10 text-15 text-gray-300 transition-all-300 group-hover:translate-x-3 dark:text-gray-600 group-hover:text-primary" />
          </div>
          <div class="mt-12 text-14 font-medium">
            {{ item.title }}
          </div>
          <div class="mt-5 text-12 text-gray-500 leading-relaxed dark:text-gray-400">
            {{ item.desc }}
          </div>
          <div v-if="item.links" class="mt-10 flex flex-wrap gap-6">
            <span v-for="link in item.links" :key="link.path" class="rounded-6 bg-gray-100 px-8 py-3 text-12 text-gray-600 transition-all-300 dark:bg-gray-800 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-700" @click.stop="router.push(link.path)">
              {{ link.label }}
            </span>
          </div>
        </div>
      </div>
    </section>
    <section class="mt-18">
      <div class="mb-12 flex items-center gap-8">
        <div class="h-14 w-3 rounded-l-2 bg-primary" />
        <h2 class="text-15 font-medium">
          遇到问题
        </h2>
      </div>

      <div class="grid grid-cols-1 gap-12 md:grid-cols-2">
        <a v-for="item in helpLinks" :key="item.title" :href="item.url" target="_blank" rel="noreferrer" class="group flex items-center gap-12 card-border rounded-12 auto-bg p-16 transition-all-300 hover:card-shadow hover:-translate-y-4">
          <div class="h-34 w-34 f-c-c shrink-0 rounded-10 text-16" style="background: rgba(var(--primary-color), 0.1); color: rgba(var(--primary-color))">
            <i :class="item.icon" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-13 font-medium">
              {{ item.title }}
            </div>
            <div class="mt-3 text-12 text-gray-500 dark:text-gray-400">
              {{ item.desc }}
            </div>
          </div>
          <i class="i-fe:arrow-up-right text-14 text-gray-300 transition-all-300 dark:text-gray-600 group-hover:text-primary" />
        </a>
      </div>
    </section>
  </AppPage>
</template>

<script setup>
import { useUserStore } from '@/store'

const router = useRouter()
const userStore = useUserStore()

const displayName = computed(() => userStore.nickName || userStore.username || '主播')

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 5)
    return '夜深了'
  if (hour < 11)
    return '早上好'
  if (hour < 14)
    return '中午好'
  if (hour < 18)
    return '下午好'
  return '晚上好'
})

const startGuides = [
  {
    title: '房间配置',
    desc: '在这里登录机器人账号，并配置要监听的直播间。想让机器人开始工作，这一步是必须的。',
    icon: 'i-fe:tv',
    tone: 'bg-blue-50 text-blue-500 dark:bg-blue-500/10 dark:text-blue-400',
    path: '/room',
  },
  {
    title: '机器人配置',
    desc: '机器人的各种行为都在这里管理：进场欢迎、礼物感谢、弹幕回复，以及用户积分怎么获取。',
    icon: 'i-fe:sliders',
    tone: 'bg-violet-50 text-violet-500 dark:bg-violet-500/10 dark:text-violet-400',
    path: '/robot',
  },
  {
    title: 'App 配置',
    desc: '打算用积分商城的话，先在这里完成基础配置，比如商品图片存到哪里。不用商城可以先跳过。',
    icon: 'i-fe:grid',
    tone: 'bg-amber-50 text-amber-500 dark:bg-amber-500/10 dark:text-amber-400',
    path: '/app',
  },
]

const dailyGuides = [
  {
    title: '数据分析',
    desc: '弹幕、礼物、盲盒、PK 和用户数据都能在这里查到，直播结束后可以回头看看都发生了什么。',
    icon: 'i-fe:bar-chart-2',
    tone: 'bg-emerald-50 text-emerald-500 dark:bg-emerald-500/10 dark:text-emerald-400',
    path: '/livedanmu/list',
  },
  {
    title: '商城管理',
    desc: '商品管理用来上架和设置商品，用户管理可以查看、调整用户余额。用户余额是怎么产生的，在「机器人配置」里能看到。',
    icon: 'i-fe:shopping-cart',
    tone: 'bg-rose-50 text-rose-500 dark:bg-rose-500/10 dark:text-rose-400',
    path: '/shop/product',
    links: [
      { label: '商品管理', path: '/shop/product' },
      { label: '用户管理', path: '/shop/user' },
    ],
  },
]

const helpLinks = [
  {
    title: '去作者网站留言',
    desc: '使用中遇到问题，或者有想要的功能，都可以写在留言板',
    icon: 'i-fe:message-circle',
    url: 'https://hejunjie.life/danmusuite/local',
  },
  {
    title: '直接联系作者',
    desc: '需要单独沟通的话，可以通过这里找到我',
    icon: 'i-fe:mail',
    url: 'https://hejunjie.life/projects#direct-contact',
  },
]
</script>
