<template>
  <CommonPage>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="getData" :scroll-x="800">
      <MeQueryItem label="房间ID">
        <n-select v-model:value="query.room_id" :options="roomOptions" placeholder="请指定房间" />
      </MeQueryItem>
      <MeQueryItem label="用户UID">
        <n-input-number v-model:value="query.uid" :show-button="false" :precision="0" placeholder="用户完整UID" />
      </MeQueryItem>
      <MeQueryItem label="用户昵称">
        <n-input v-model:value="query.uname" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="礼物名称">
        <n-input v-model:value="query.gift_name" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="盲盒名称">
        <n-input v-model:value="query.original_gift_name" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="发送时间">
        <n-date-picker v-model:value="query.send_at" type="daterange" clearable />
      </MeQueryItem>
      <template #statistic>
        <StatisticsCard :stats="stats" />
      </template>
    </MeCrud>
  </CommonPage>
</template>

<script setup>
import { NTag } from 'naive-ui'
import { MeCrud, MeQueryItem, StatisticsCard } from '@/components'
import api from './api'

// 房间信息
const roomOptions = ref(null)

// 列表信息
const $table = ref(null)
const levelColor = {
  0: '',
  1: '#FFD700',
  2: '#8A2BE2',
  3: '#4A90E2',
}
const columns = [
  { title: '用户UID', key: 'uid', minWidth: 170 },
  { title: '牌子', key: 'badge_name', width: 140, render(row) {
    return row.badge_name
      ? h(
          NTag,
          {
            size: 'small',
            color: {
              borderColor: levelColor[String(row.badge_type)],
              textColor: levelColor[String(row.badge_type)],
            },
          },
          {
            default: () => row.badge_name,
            icon: () => Number(row.badge_type) > 0 ? [h('i', { class: 'i-fe:anchor text-14' }), h('i', { class: `i-fe:tabler-number-${row.badge_level}-small text-14` })] : h('i', { class: `i-fe:tabler-number-${row.badge_level}-small text-14` }),
          },
        )
      : ''
  } },
  { title: '用户昵称', key: 'uname', minWidth: 120 },
  { title: '礼物名称', key: 'gift_name', minWidth: 120 },
  { title: '单价', key: 'price', width: 100, render(row) {
    const price = row.price / 100
    return `¥${price.toFixed(2)}`
  } },
  { title: '数量', key: 'num', width: 70 },
  { title: '总价', key: 'total', width: 100, render(row) {
    const total = (row.price / 100) * row.num
    return `¥${total.toFixed(2)}`
  } },
  { title: '对应盲盒', key: 'original_gift_name', minWidth: 120 },
  { title: '盲盒价格', key: 'original_gift_price', width: 100, render(row) {
    const original_gift_price = row.original_gift_price / 100
    return `¥${original_gift_price.toFixed(2)}`
  } },
  { title: '盈利金额', key: 'profit', width: 100, render(row) {
    const profit = (((row.price - row.original_gift_price) * row.num) / 100)
    return `¥${profit.toFixed(2)}`
  } },
  { title: '赠送时间', key: 'send_at', width: 200 },
]

const query = ref({
  room_id: null,
  uid: null,
  uname: null,
  gift_name: null,
  original_gift_name: null,
  send_at: null,
})

const statsData = reactive({ originalPrice: 0, currentPrice: 0 })

const stats = computed(() => [
  { label: '礼物金额', value: (statsData.currentPrice / 100).toFixed(2), prefix: '¥', suffix: '元', tooltip: '收到的所有礼物的价值总和' },
  { label: '盲盒金额', value: (statsData.originalPrice / 100).toFixed(2), prefix: '¥', suffix: '元', tooltip: '用户赠送的所有盲盒的价格总和' },
  { label: '盲盒盈利', value: ((statsData.currentPrice - statsData.originalPrice) / 100).toFixed(2), prefix: '¥', suffix: '元', tooltip: '所有盲盒的盈利情况' },
])

async function getData(params) {
  const res = await api.getList(params)
  const s = res.data?.stats
  if (s) {
    statsData.originalPrice = s.original_price ?? 0
    statsData.currentPrice = s.current_price ?? 0
  }
  return res
}

onMounted(() => {
  $table.value?.handleSearch()
  api.fetchRoomGroups().then((res) => {
    roomOptions.value = res.data.option
  })
})
</script>
