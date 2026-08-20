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
      <MeQueryItem label="礼物类型">
        <n-select v-model:value="query.gift_type" :options="giftTypeOptions" placeholder="请指定礼物类型" />
      </MeQueryItem>
      <MeQueryItem label="原始礼物">
        <n-select v-model:value="query.original" :options="originalOptions" placeholder="是否是原始礼物" />
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
import { NTag, NTooltip } from 'naive-ui'
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
  {
    title: '牌子',
    key: 'badge_name',
    width: 140,
    render(row) {
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
    },
  },
  { title: '用户昵称', key: 'uname', minWidth: 120 },
  { title: '礼物名称', key: 'gift_name', render(row) {
    return row.message
      ? h(
          NTooltip,
          { showArrow: false, trigger: 'hover' },
          {
            trigger: () => row.gift_name,
            default: () => row.message,
          },
        )
      : row.gift_name
  } },
  {
    title: '单价',
    key: 'price',
    width: 100,
    render(row) {
      const price = row.price / 100
      return `¥${price.toFixed(2)}`
    },
  },
  { title: '数量', key: 'num', width: 70 },
  {
    title: '总价',
    key: 'total',
    width: 100,
    render(row) {
      const total = (row.price / 100) * row.num
      return `¥${total.toFixed(2)}`
    },
  },
  { title: '赠送时间', key: 'send_at', width: 200 },
]

const query = ref({
  room_id: null,
  uid: null,
  uname: null,
  gift_name: null,
  gift_type: null,
  original: null,
  send_at: null,
})
const giftTypeOptions = ref([
  {
    label: '普通',
    value: 0,
  },
  {
    label: '航海',
    value: 1,
  },
  {
    label: '醒目留言',
    value: 2,
  },
])
const originalOptions = ref([
  {
    label: '否',
    value: 0,
  },
  {
    label: '是',
    value: 1,
  },
])

const statsData = reactive({ totalNum: 0, totalAmount: 0 })

const stats = computed(() => [
  { label: '礼物数量', value: statsData.totalNum, tooltip: '统计截至目前的礼物总件数' },
  { label: '礼物总价', value: (statsData.totalAmount / 100).toFixed(2), prefix: '¥', suffix: '元', tooltip: '所有礼物的价值总和（数量 x 单价）' },
  { label: '平均单价', value: statsData.totalNum ? (statsData.totalAmount / 100 / statsData.totalNum).toFixed(2) : 0, prefix: '¥', suffix: '元', tooltip: '单份礼物的平均价值，数值高低反映了送礼群体的消费倾向' },
])

async function getData(params) {
  const res = await api.getList(params)
  const s = res.data?.stats
  if (s) {
    statsData.totalNum = s.total_num ?? 0
    statsData.totalAmount = s.total_amount ?? 0
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
