<template>
  <CommonPage>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="api.getList" :scroll-x="960">
      <MeQueryItem label="用户UID">
        <n-input-number v-model:value="query.uid" :show-button="false" :precision="0" placeholder="用户完整UID" clearable />
      </MeQueryItem>
      <MeQueryItem label="用户昵称">
        <n-input v-model:value="query.uname" placeholder="支持模糊搜索" clearable />
      </MeQueryItem>
    </MeCrud>
    <ComplaintDetailModal ref="detailModalRef" />
  </CommonPage>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import api from './api'
import ComplaintDetailModal from './components/ComplaintDetailModal.vue'

const $table = ref(null)
const detailModalRef = ref(null)

const columns = [
  { title: '用户UID', key: 'uid', minWidth: 190, sorter: true },
  { title: '用户昵称', key: 'uname', minWidth: 140, sorter: true },
  { title: '问题类型', key: 'type', minWidth: 140, sorter: true },
  { title: '联系方式', key: 'contact', minWidth: 160, sorter: true },
  { title: '投诉时间', key: 'created_at', width: 180, sorter: true },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    align: 'right',
    fixed: 'right',
    hideInExcel: true,
    render(row) {
      return [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            secondary: true,
            onClick: () => detailModalRef.value?.open(row),
          },
          {
            default: () => '查看',
            icon: () => h('i', { class: 'i-fe:file-text text-14' }),
          },
        ),
      ]
    },
  },
]

const query = ref({
  uid: null,
  uname: null,
})

onMounted(() => {
  $table.value?.handleSearch()
})
</script>
