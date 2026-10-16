<template>
  <CommonPage>
    <template #action>
      <NButton v-permission="'AddGuard'" type="primary" @click="handleBackfill()">
        <i class="i-material-symbols:add mr-4 text-18" />
        补录开通
      </NButton>
    </template>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="getData" :scroll-x="1400" export-module="liveguard">
      <MeQueryItem label="月份" :label-width="90">
        <n-date-picker v-model:formatted-value="query.month" type="month" value-format="yyyy-MM" :clearable="false" class="w-full" />
      </MeQueryItem>
      <MeQueryItem label="大航海等级" :label-width="90">
        <n-select v-model:value="query.guard_level" :options="guardLevelOptions" :clearable="false" />
      </MeQueryItem>
    </MeCrud>
    <BackfillModal ref="backfillModalRef" @success="onBackfillSuccess" />
    <OrderModal ref="orderModalRef" @success="onOrderSuccess" />
  </CommonPage>
</template>

<script setup>
import { NButton, NTag, NTooltip } from 'naive-ui'
import { GuardTag, MeCrud, MeQueryItem } from '@/components'
import { withPermission } from '@/directives'
import { formatDate } from '@/utils'
import api from './api'
import BackfillModal from './components/BackfillModal.vue'
import OrderModal from './components/OrderModal.vue'

const $table = ref(null)
const backfillModalRef = ref(null)
const orderModalRef = ref(null)

// 与后端 enum.BadgeType 的取值保持一致：1 总督、2 提督、3 舰长
const guardLevelOptions = ref([
  { label: '舰长', value: 3 },
  { label: '提督', value: 2 },
  { label: '总督', value: 1 },
])

const columns = [
  { title: '用户UID', key: 'uid', minWidth: 180, sorter: true },
  { title: '用户昵称', key: 'uname', minWidth: 140, sorter: true },
  {
    title: '大航海等级',
    key: 'guard_level',
    width: 120,
    render(row) {
      return h(GuardTag, { type: row.guard_level })
    },
  },
  { title: '起算月', key: 'cover_from_month', width: 120 },
  { title: '覆盖至月', key: 'cover_to_month', width: 120 },
  { title: '当月开通份数', key: 'open_num', width: 130 },
  {
    title: '当月最早开通',
    key: 'open_at',
    width: 180,
    render(row) {
      if (row.open_at)
        return row.open_at
      return h(NTooltip, { showArrow: false, trigger: 'hover' }, {
        trigger: () => h('span', { class: 'opacity-40' }, '往月延续'),
        default: () => '当月没有开通，名单是从前面的月份顺延过来的',
      })
    },
  },
  {
    title: '本月兑换',
    key: 'redeemed_products',
    width: 120,
    hideInExcel: true,
    render(row) {
      const list = row.redeemed_products || []
      if (!list.length)
        return h('span', { class: 'opacity-40' }, '未兑换')
      return h(NTooltip, { showArrow: false, trigger: 'hover' }, {
        trigger: () => h(NTag, { size: 'small', type: 'success', bordered: false }, { default: () => '已兑换' }),
        default: () => h('div', { class: 'space-y-4' }, list.map(item => h('div', null, `${item.product_name} ×${item.quantity}`))),
      })
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    align: 'right',
    fixed: 'right',
    hideInExcel: true,
    render(row) {
      return withPermission(
        h(NButton, {
          size: 'small',
          type: 'primary',
          secondary: true,
          onClick: () => orderModalRef.value?.open(row),
        }, { default: () => '帮用户下单' }),
        'PlaceOrderForUser',
      )
    },
  },
]

// 月份与等级是名单的唯一维度，都不能为空，默认落在当月 + 舰长
const query = ref({
  month: formatDate(new Date(), 'YYYY-MM'),
  guard_level: 3,
})

async function getData(params) {
  // 兜底：任一筛选条件为空时后端会直接报参数错误，这里先拦下来
  if (!params.month || !params.guard_level) {
    $message.warning('请先选择月份与大航海等级')
    return { data: { pageData: [], total: 0 } }
  }
  return api.getList(params)
}

function handleBackfill() {
  backfillModalRef.value?.open(query.value.guard_level)
}

function onBackfillSuccess() {
  $table.value?.handleSearch(true)
}

function onOrderSuccess() {
  // 下单成功后刷新，让「本月兑换」立刻变为已兑换
  $table.value?.handleSearch(true)
}

onMounted(() => {
  $table.value?.handleSearch()
})
</script>
