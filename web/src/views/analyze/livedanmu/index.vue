<template>
  <CommonPage>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="api.getList" :scroll-x="800">
      <MeQueryItem label="房间ID">
        <n-select v-model:value="query.room_id" :options="roomOptions" placeholder="请指定房间" />
      </MeQueryItem>
      <MeQueryItem label="用户UID">
        <n-input-number v-model:value="query.uid" :show-button="false" :precision="0" placeholder="用户完整UID" />
      </MeQueryItem>
      <MeQueryItem label="用户昵称">
        <n-input v-model:value="query.uname" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="弹幕信息">
        <n-input v-model:value="query.msg" placeholder="支持模糊搜索" />
      </MeQueryItem>
      <MeQueryItem label="发送时间">
        <n-date-picker v-model:value="query.send_at" type="daterange" clearable />
      </MeQueryItem>
    </MeCrud>
  </CommonPage>
</template>

<script setup>
import { NTag } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import api from './api'

const $table = ref(null)
const levelColor = {
  0: '',
  1: '#FFD700',
  2: '#8A2BE2',
  3: '#4A90E2',
}
const roomOptions = ref(null)
const columns = [
  { title: '用户UID', key: 'uid', minWidth: 170 },
  {
    title: '牌子',
    key: 'badge_room_id',
    width: 140,
    render(row) {
      return row.badge_room_id
        ? h(
            NTag,
            {
              size: 'small',
              color: {
                borderColor: levelColor[String(row.badge_type)],
                textColor: levelColor[String(row.badge_type)],
              },
              onClick: () => goToLive(row.badge_room_id),
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
  { title: '弹幕内容', key: 'msg' },
  { title: '发送时间', key: 'send_at', width: 200 },
]
function goToLive(room_id) {
  window.open(`https://live.bilibili.com/${room_id}`, '_blank', 'noopener,noreferrer')
}

const query = ref({
  room_id: null,
  uid: null,
  uname: null,
  msg: null,
  send_at: null,
})

onMounted(() => {
  $table.value?.handleSearch()
  api.fetchRoomGroups().then((res) => {
    roomOptions.value = res.data.option
  })
})
</script>
