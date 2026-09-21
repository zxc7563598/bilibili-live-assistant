<template>
  <CommonPage>
    <MeCrud
      ref="$table"
      v-model:query-items="query"
      :columns="columns"
      :get-data="api.getList"
      :scroll-x="1320"
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
    </MeCrud>
    <ShipStatusModal ref="shipModalRef" @success="onShipSuccess" />
  </CommonPage>
</template>

<script setup>
import { NButton, NImage, NTag } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import { formatProductSpecs, getOptionsLabel } from '@/utils'
import ShipStatusModal from '../list/components/ShipStatusModal.vue'
import api from './api'

const router = useRouter()

const $table = ref(null)
const shipModalRef = ref(null)

const productTypeOptions = ref([
  { label: '虚拟商品', value: 0 },
  { label: '实体商品', value: 1 },
])

// 后端 timeutil.Format 对空值返回空串，这里统一成占位符，避免出现只有标签没内容的空行
function dash(value) {
  return value === '' || value === null || value === undefined ? '—' : value
}

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

// 发货页看的是「发什么、发多少」，所以封面、规格、数量都摊在同一格，管理员不用再进详情核对
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

// 收货信息是发货动作的直接依据：实体订单要姓名、电话、完整地址，虚拟订单只要邮箱
function renderReceiver(row) {
  if (Number(row.receiver_type) === 1) {
    const address = [row.receiver_region, row.receiver_detail].filter(Boolean).join(' ')
    return h('div', { class: 'flex flex-col gap-4' }, [
      h('div', { class: 'flex items-center gap-8 leading-18' }, [
        h('span', { class: 'break-all text-13 font-medium' }, dash(row.receiver_name)),
        h('span', { class: 'shrink-0 text-12 text-gray-400 dark:text-gray-500' }, dash(row.receiver_phone)),
      ]),
      h('div', { class: 'break-all text-12 text-gray-400 leading-17 dark:text-gray-500' }, address || '—'),
    ])
  }
  return h('div', { class: 'break-all text-13 leading-18' }, dash(row.receiver_email))
}

const columns = [
  { title: '订单编号', key: 'order_sn', width: 200, sorter: true },
  { title: '用户昵称', key: 'uname', minWidth: 100, sorter: true },
  {
    title: '商品信息',
    key: 'product_name',
    // 该单元格摊了商品名、数量、规格三项，导出取后端拼好的同名文案
    exportKey: 'product_info',
    minWidth: 260,
    render: renderProduct,
  },
  { title: '商品类型', key: 'receiver_type', width: 100, render(row) {
    const isActual = Number(row.receiver_type) === 1
    return h(
      NTag,
      { size: 'small', bordered: false, type: isActual ? 'success' : 'warning' },
      { default: () => getOptionsLabel(productTypeOptions, row.receiver_type) },
    )
  } },
  {
    title: '收货信息',
    key: 'receiver_name',
    // 实体单的姓名/电话/地址与虚拟单的邮箱在后端拼成同一列文案
    exportKey: 'receiver_info',
    minWidth: 260,
    render: renderReceiver,
  },
  { title: '下单时间', key: 'created_at', width: 180, sorter: true },
  {
    title: '操作',
    key: 'actions',
    width: 270,
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
            icon: () => h('i', { class: 'i-fe:file-text text-14' }),
          },
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            class: 'ml-12px',
            secondary: true,
            onClick: () => shipModalRef.value?.open(row),
          },
          {
            default: () => '变更发货状态',
            icon: () => h('i', { class: 'i-fe:truck text-14' }),
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
  order_status: 1,
  ship_status: 0,
})

// 标记发货后该订单不再满足固定条件，会从列表中消失，留在当前页刷新即可
function onShipSuccess() {
  $table.value?.handleSearch(true)
}

onMounted(() => {
  $table.value?.handleSearch()
})
</script>
