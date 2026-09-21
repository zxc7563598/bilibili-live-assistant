<template>
  <CommonPage>
    <MeCrud
      ref="$table"
      v-model:query-items="query"
      :columns="columns"
      :get-data="api.getList"
      :scroll-x="1250"
      export-module="order"
    >
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
import { NButton, NImage, NTag } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import { formatProductSpecs, getOptionsLabel, getOptionsType } from '@/utils'
import api from './api'

const router = useRouter()

const $table = ref(null)

const orderStatusOptions = ref([
  { label: '待付款', value: 0, type: 'info' },
  { label: '待发货', value: 1, type: 'warning' },
  { label: '待收货', value: 2, type: 'warning' },
  { label: '已完成', value: 3, type: 'success' },
  { label: '已取消', value: 4, type: 'warning' },
  { label: '售后中', value: 5, type: 'error' },
])
const payStatusOptions = ref([
  { label: '未支付', value: 0, type: 'warning' },
  { label: '已支付', value: 1, type: 'success' },
  { label: '已退款', value: 2, type: 'info' },
])
const shipStatusOptions = ref([
  { label: '未发货', value: 0, type: 'warning' },
  { label: '已发货', value: 1, type: 'info' },
  { label: '已签收', value: 2, type: 'success' },
])

function renderCover(row) {
  if (row.product_cover) {
    return h(
      'div',
      { class: 'h-48 w-48 shrink-0 overflow-hidden border border-light_border rounded-6 dark:border-dark_border' },
      h(NImage, { src: row.product_cover, width: 48, height: 48, objectFit: 'cover', class: 'block' }),
    )
  }
  return h(
    'div',
    { class: 'h-48 w-48 shrink-0 flex items-center justify-center border border-light_border rounded-6 text-16 text-gray-300 dark:border-dark_border dark:text-gray-600' },
    h('i', { class: 'i-fe:image' }),
  )
}

// 封面、名称、数量、规格摊在同一格，省得为了核对商品再进一次详情
function renderProduct(row) {
  const specs = formatProductSpecs(row.product_spec_properties)
  return h('div', { class: 'flex items-center gap-10' }, [
    renderCover(row),
    h('div', { class: 'min-w-0 flex flex-col gap-4' }, [
      h('div', { class: 'flex items-start gap-6' }, [
        h('div', { class: 'min-w-0 break-all text-13 font-medium leading-18' }, row.product_name),
        h('span', { class: 'shrink-0 text-12 text-primary leading-18' }, `×${row.quantity}`),
      ]),
      h('div', { class: 'break-all text-12 text-gray-400 leading-16 dark:text-gray-500' }, `规格：${specs || '无'}`),
    ]),
  ])
}

const columns = [
  { title: '订单编号', key: 'order_sn', width: 200, sorter: true },
  { title: '用户昵称', key: 'uname', minWidth: 80, sorter: true },
  {
    title: '商品信息',
    key: 'product_name',
    // 该单元格摊了商品名、数量、规格三项，导出取后端拼好的同名文案
    exportKey: 'product_info',
    minWidth: 260,
    sorter: true,
    render: renderProduct,
  },
  { title: '支付状态', key: 'pay_status', width: 100, sorter: true, render(row) {
    return h(
      NTag,
      { size: 'small', bordered: false, type: getOptionsType(payStatusOptions, row.pay_status) },
      { default: () => getOptionsLabel(payStatusOptions, row.pay_status) },
    )
  } },
  { title: '发货状态', key: 'ship_status', width: 100, sorter: true, render(row) {
    return h(
      NTag,
      { size: 'small', bordered: false, type: getOptionsType(shipStatusOptions, row.ship_status) },
      { default: () => getOptionsLabel(shipStatusOptions, row.ship_status) },
    )
  } },
  { title: '订单状态', key: 'order_status', width: 100, sorter: true, render(row) {
    return h(
      NTag,
      { size: 'small', bordered: false, type: getOptionsType(orderStatusOptions, row.order_status) },
      { default: () => getOptionsLabel(orderStatusOptions, row.order_status) },
    )
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
