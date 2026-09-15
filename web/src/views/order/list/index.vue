<template>
  <CommonPage>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="api.getList" :scroll-x="1250">
      <MeQueryItem label="用户UID">
        <n-input-number v-model:value="query.uid" :show-button="false" :precision="0" placeholder="用户完整UID" clearable />
      </MeQueryItem>
      <MeQueryItem label="用户昵称">
        <n-input v-model:value="query.uname" placeholder="支持模糊搜索" clearable />
      </MeQueryItem>
      <MeQueryItem label="订单号">
        <n-input v-model:value="query.order_sn" placeholder="支持模糊搜索" clearable />
      </MeQueryItem>
      <MeQueryItem label="订单状态">
        <n-select v-model:value="query.order_status" :options="orderStatusOptions" placeholder="请指定订单状态" clearable />
      </MeQueryItem>
      <MeQueryItem label="支付状态">
        <n-select v-model:value="query.pay_status" :options="payStatusOptions" placeholder="请指定支付状态" clearable />
      </MeQueryItem>
      <MeQueryItem label="发货状态">
        <n-select v-model:value="query.ship_status" :options="shipStatusOptions" placeholder="请指定发货状态" clearable />
      </MeQueryItem>
    </MeCrud>
  </CommonPage>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import { getOptionsLabel } from '@/utils'
import api from './api'

const router = useRouter()

const $table = ref(null)

const orderStatusOptions = ref([
  { label: '待付款', value: 0 },
  { label: '待发货', value: 1 },
  { label: '待收货', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已取消', value: 4 },
  { label: '售后中', value: 5 },
])
const payStatusOptions = ref([
  { label: '未支付', value: 0 },
  { label: '已支付', value: 1 },
  { label: '已退款', value: 2 },
])
const shipStatusOptions = ref([
  { label: '未发货', value: 0 },
  { label: '已发货', value: 1 },
  { label: '已签收', value: 2 },
])

const columns = [
  { title: '订单编号', key: 'order_sn', width: 200, sorter: true },
  { title: '用户昵称', key: 'uname', minWidth: 80, sorter: true },
  { title: '商品名称', key: 'product_name', minWidth: 80, sorter: true },
  { title: '支付状态', key: 'pay_status', width: 100, sorter: true, render(row) {
    return getOptionsLabel(payStatusOptions, row.pay_status)
  } },
  { title: '发货状态', key: 'ship_status', width: 100, sorter: true, render(row) {
    return getOptionsLabel(shipStatusOptions, row.ship_status)
  } },
  { title: '订单状态', key: 'order_status', width: 100, sorter: true, render(row) {
    return getOptionsLabel(orderStatusOptions, row.order_status)
  } },
  { title: '下单时间', key: 'created_at', width: 180, sorter: true },
  {
    title: '操作',
    key: 'actions',
    width: 120,
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
            onClick: () => router.push({ path: '/order/list/details', query: { id: row.id } }),
          },
          {
            default: () => '查看详情',
            icon: () => h('i', { class: 'i-fe:calendar text-14' }),
          },
        ),
      ]
    },
  },
]

const query = ref({
  uid: null,
  uname: null,
  order_sn: null,
  order_status: null,
  pay_status: null,
  ship_status: null,
})

onMounted(() => {
  $table.value?.handleSearch()
})
</script>
