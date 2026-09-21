<template>
  <CommonPage>
    <MeCrud
      ref="$table"
      v-model:query-items="query"
      :columns="columns"
      :get-data="getData"
      :scroll-x="800"
      export-module="livepk"
    >
      <MeQueryItem label="房间ID">
        <n-select v-model:value="query.room_id" :options="roomOptions" placeholder="请指定房间" />
      </MeQueryItem>
      <MeQueryItem label="对手UID">
        <n-input-number v-model:value="query.rival_uid" :show-button="false" :precision="0" placeholder="用户完整UID" />
      </MeQueryItem>
      <MeQueryItem label="对手昵称">
        <n-input v-model:value="query.rival_uname" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="我方胜负">
        <n-select v-model:value="query.result" :options="resultOptions" placeholder="请选择" />
      </MeQueryItem>
      <MeQueryItem label="发送时间">
        <n-date-picker v-model:value="query.start_at" type="daterange" clearable />
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
import { getOptionsLabel, getOptionsType } from '@/utils'
import api from './api'

const roomOptions = ref(null)

const resultOptions = ref([
  {
    label: '落败',
    value: -1,
    type: 'error',
  },
  {
    label: '获胜',
    value: 2,
    type: 'success',
  },
])

const pkStatusOptions = ref([
  {
    label: '即将开始',
    value: 101,
  },
  {
    label: '正常结束',
    value: 401,
  },
  {
    label: '异常结束',
    value: 101,
  },
])

const battleTypeOptions = ref([
  {
    label: '经典PK',
    value: 2,
  },
  {
    label: '大乱斗',
    value: 6,
  },
])

// 列表信息
const $table = ref(null)
const columns = [
  { title: 'PK ID', key: 'pk_id', width: 120, sorter: true },
  { title: 'PK 状态', key: 'pk_status', width: 100, sorter: true, render(row) {
    return getOptionsLabel(pkStatusOptions, row.pk_status)
  } },
  { title: '对战类型', key: 'battle_type', width: 100, sorter: true, render(row) {
    return getOptionsLabel(battleTypeOptions, row.battle_type)
  } },
  { title: '对手 UID', key: 'rival_uid', minWidth: 180, sorter: true },
  { title: '对手昵称', key: 'rival_uname', minWidth: 120, sorter: true },
  { title: '我方PK值', key: 'self_votes', width: 120, sorter: true, render(row) {
    return h(
      NTag,
      { size: 'small', bordered: false, type: getOptionsType(resultOptions, row.self_result) },
      { default: () => row.self_votes },
    )
  } },
  { title: '对方PK值', key: 'rival_votes', width: 120, sorter: true, render(row) {
    return h(
      NTag,
      { size: 'small', bordered: false, type: getOptionsType(resultOptions, row.rival_result) },
      { default: () => row.rival_votes },
    )
  } },
  { title: '开始时间', key: 'start_at', width: 180, sorter: true },
  { title: '结束时间', key: 'settle_at', width: 180, sorter: true },
]

const query = ref({
  room_id: null,
  rival_uid: null,
  rival_uname: null,
  result: null,
  start_at: null,
})

const statsData = reactive({ originalPrice: 0, currentPrice: 0 })

const stats = computed(() => [
  { label: '总场数', value: statsData.total_num, tooltip: 'PK总场数' },
  { label: '胜利场数', value: statsData.win_num, tooltip: 'PK胜利的场数' },
  { label: '失败场数', value: statsData.lose_num, tooltip: 'PK失败的场数' },
])

async function getData(params) {
  const res = await api.getList(params)
  const s = res.data?.stats
  if (s) {
    statsData.total_num = s.total_num ?? 0
    statsData.win_num = s.win_num ?? 0
    statsData.lose_num = s.lose_num ?? 0
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
