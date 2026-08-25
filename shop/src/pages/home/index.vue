<template>
  <div class="min-h-dvh pb-28">
    <header class="safe-top bg-bg">
      <div class="mx-auto w-full max-w-5xl px-4 pt-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-on-primary">
              <AppIcon name="store" :size="20" />
            </div>
            <div>
              <p class="text-xs text-fg-3">
                你好，{{ user.name }}
              </p>
              <p class="text-sm font-semibold">
                积分商城
              </p>
            </div>
          </div>
          <div class="flex items-center gap-1 rounded-full bg-primary-soft px-3 py-1.5 text-primary">
            <AppIcon name="points" :size="16" />
            <span class="text-sm font-bold tabular-nums">{{ user.points }}</span>
          </div>
        </div>
        <div class="mt-3 flex items-center gap-2 rounded-2xl border border-line bg-surface px-4 py-2.5">
          <AppIcon name="search" :size="18" class="text-fg-3" />
          <input class="flex-1 bg-transparent text-sm outline-none placeholder:text-fg-3" placeholder="搜索商品">
        </div>
      </div>
    </header>
    <div class="no-scrollbar mx-auto mt-3 flex w-full max-w-5xl gap-2 overflow-x-auto px-4">
      <button v-for="(c, i) in categories" :key="c" class="shrink-0 rounded-full px-4 py-1.5 text-sm font-medium transition" :class="i === 0 ? 'bg-primary text-on-primary' : 'bg-surface text-fg-2'">
        {{ c }}
      </button>
    </div>
    <main class="mx-auto w-full max-w-5xl px-4 pt-4">
      <div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
        <router-link v-for="p in products" :key="p.id" :to="`/product/${p.id}`" class="card press block overflow-hidden">
          <AppImage :src="p.cover" :label="p.name" ratio="1 / 1" rounded="rounded-none" />
          <div class="p-3">
            <p class="line-clamp-2 text-sm font-medium leading-snug">
              {{ p.name }}
            </p>
            <div class="mt-2 flex items-end justify-between">
              <span class="flex items-center gap-1" :class="[p.color]">
                <AppIcon :name="p.icon" :size="15" />
                <span class="text-base font-bold leading-none tabular-nums">{{ p.amount }}</span>
              </span>
              <span class="text-xs text-fg-3 tabular-nums">¥{{ p.rmb }}</span>
            </div>
          </div>
        </router-link>
      </div>
    </main>
    <TabBar />
  </div>
</template>

<script setup>
import TabBar from '@/components/base/TabBar.vue'

const user = {
  uid: '1234567',
  name: '小明同学',
  avatar: '',
  points: 3280, // 积分余额
  starlight: 1260, // 星光余额
  level: 'Lv.5',
}

const products = [
  {
    id: 1,
    name: '无线蓝牙耳机 主动降噪',
    cover: '',
    amount: 12800,
    icon: 'points',
    color: 'text-primary',
    label: '积分',
    rmb: 129,
    sold: 2341,
    stock: 99,
    tags: ['热卖', '包邮'],
    skuGroups: [
      { name: '颜色', options: ['星空灰', '月光白', '曜石黑'] },
      { name: '版本', options: ['标准版', 'Pro 版'] },
    ],
    desc: '一款主打舒适的主动降噪耳机，支持多设备连接。',
  },
  {
    id: 2,
    name: '316 不锈钢保温杯 500ml',
    cover: '',
    amount: 6800,
    icon: 'points',
    color: 'text-primary',
    label: '积分',
    rmb: 79,
    sold: 5120,
    stock: 300,
    tags: ['热卖'],
    skuGroups: [{ name: '颜色', options: ['燕麦白', '雾霾蓝', '松石绿'] }],
    desc: '316 内胆，长效保温，便携杯盖。',
  },
  {
    id: 3,
    name: '定制帆布手提包',
    cover: '',
    amount: 4200,
    icon: 'points',
    color: 'text-primary',
    label: '积分',
    rmb: 49,
    sold: 889,
    stock: 0,
    tags: ['限量'],
    skuGroups: [{ name: '款式', options: ['竖版', '横版'] }],
    desc: '厚实帆布，支持商户定制印刷。',
  },
  {
    id: 4,
    name: '智能运动手环',
    cover: '',
    amount: 360,
    icon: 'star',
    color: 'text-starlight',
    label: '星光',
    rmb: 199,
    sold: 3456,
    stock: 120,
    tags: ['新品', '包邮'],
    skuGroups: [{ name: '颜色', options: ['午夜黑', '珊瑚粉'] }],
    desc: '全天候心率监测，14 天续航。',
  },
  {
    id: 5,
    name: '桌面绿植摆件（多肉）',
    cover: '',
    amount: 120,
    icon: 'star',
    color: 'text-starlight',
    label: '星光',
    rmb: 39,
    sold: 1200,
    stock: 50,
    tags: [],
    skuGroups: [],
    desc: '懒人好养的多肉小盆栽，办公桌治愈好物。',
  },
  {
    id: 6,
    name: '便携蓝牙音箱',
    cover: '',
    amount: 9800,
    icon: 'points',
    color: 'text-primary',
    label: '积分',
    rmb: 159,
    sold: 1903,
    stock: 66,
    tags: ['热卖'],
    skuGroups: [{ name: '颜色', options: ['黑色', '军绿'] }],
    desc: 'IPX5 防水，360° 环绕音效。',
  },
]

const categories = ['全部', '数码', '生活', '潮玩', '绿植']
</script>
