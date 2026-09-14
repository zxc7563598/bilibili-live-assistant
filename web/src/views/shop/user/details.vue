<template>
  <CommonPage back title="用户详情">
    <template #action>
      <div class="flex items-center gap-8">
        <NButton secondary :disabled="!userID || !userDetail" @click="balanceModalRef?.open(userID, userLabel)">
          变更积分/余额
        </NButton>
        <NButton secondary :disabled="!userID || !userDetail" @click="passwordModalRef?.open(userID, userLabel)">
          修改密码
        </NButton>
      </div>
    </template>
    <AppCard v-if="stats.length" bordered bg="#fafafc dark:black" class="mb-30 min-h-60 rounded-4">
      <StatisticsCard :stats="stats" />
    </AppCard>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="getAssets" :scroll-x="900">
      <MeQueryItem label="余额类型">
        <n-select v-model:value="query.credit_type" clearable placeholder="全部类型" :options="creditTypeOptions" />
      </MeQueryItem>
    </MeCrud>
    <BalanceModal ref="balanceModalRef" @success="onBalanceSuccess" />
    <ResetPasswordModal ref="passwordModalRef" />
  </CommonPage>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { MeCrud, MeQueryItem } from '@/components'
import api from './api'
import BalanceModal from './components/BalanceModal.vue'
import ResetPasswordModal from './components/ResetPasswordModal.vue'

// 与后端 enum.CreditType / enum.ChangeType 的取值保持一致
const CREDIT_TYPE = { stars: 0, points: 1 }
const CHANGE_TYPE = { reduce: 0, increase: 1 }

const route = useRoute()
const $table = ref(null)
const balanceModalRef = ref(null)
const passwordModalRef = ref(null)

const userID = computed(() => Number(route.query.id) || 0)
// user_id 不放进 queryItems：它是页面上下文而不是筛选项，
// 放在这里会被「重置」清空，也会在切换用户时残留上一个用户的 ID
const query = ref({
  credit_type: null,
})

const creditTypeOptions = ref([
  { label: '星光', value: CREDIT_TYPE.stars },
  { label: '积分', value: CREDIT_TYPE.points },
])

const changeTypeOptions = ref([
  { label: '减少', value: CHANGE_TYPE.reduce },
  { label: '增加', value: CHANGE_TYPE.increase },
])

function getCreditTypeLabel(value) {
  const item = creditTypeOptions.value.find(item => item.value === Number(value))
  return item ? item.label : '未知'
}

function getChangeTypeLabel(value) {
  const item = changeTypeOptions.value.find(item => item.value === Number(value))
  return item ? item.label : '未知'
}

const columns = [
  { title: '发生时间', key: 'created_at', width: 180, sorter: true },
  { title: '余额类型', key: 'credit_type', width: 110, sorter: true, render(row) {
    return getCreditTypeLabel(row.credit_type)
  } },
  { title: '变动类型', key: 'change_type', width: 110, sorter: true, render(row) {
    return getChangeTypeLabel(row.change_type)
  } },
  { title: '变动数值', key: 'change_amount', width: 110, sorter: true, render(row) {
    const sign = Number(row.change_type) === CHANGE_TYPE.increase ? '+' : '-'
    return `${sign}${row.change_amount}`
  } },
  { title: '变动后数值', key: 'after_value', width: 130, sorter: true },
  { title: '说明', key: 'remark', minWidth: 200, sorter: true, ellipsis: { tooltip: true } },
]

const userDetail = ref(null)
// 变更弹窗里提示操作对象用，避免管理员在错误的人身上改余额
const userLabel = computed(() => {
  const detail = userDetail.value
  if (!detail)
    return ''
  return `${detail.name || '未知昵称'}（UID ${detail.uid}）`
})
const stats = computed(() => {
  const detail = userDetail.value
  if (!detail)
    return []
  return [
    { label: detail.name || '未知昵称', value: detail.uid, tooltip: '用户的 B站 昵称与 UID（账号）' },
    { label: '用户积分', value: detail.points, tooltip: '用户当前剩余的积分，可在积分商城兑换商品；每次变动都会在下方列表留下流水' },
    { label: '用户星光', value: detail.stars, tooltip: '用户当前剩余的星光，与积分相互独立，下单时按商品配置的类型扣减' },
  ]
})

// 列表查询：补上页面上下文里的 user_id
function getAssets(params = {}) {
  return api.getAssets({ ...params, user_id: userID.value })
}

// 余额变更成功后余额卡片和流水列表都过期了，一起刷新
function onBalanceSuccess() {
  getUserDetails()
  $table.value?.handleSearch(true)
}

async function getUserDetails() {
  // 先清空，避免切换用户时短暂显示上一个用户的余额
  userDetail.value = null
  try {
    const res = await api.getUserDetails(userID.value)
    userDetail.value = res.data
  }
  catch {
    // 失败提示由 http 拦截器统一处理，这里吞掉 rejection 避免冒到控制台
  }
}

// 同一路由只换 query.id 时组件会被复用，需要跟着新的用户重置并重新加载
watch(userID, (id) => {
  if (id <= 0)
    return
  query.value.credit_type = null
  getUserDetails()
  $table.value?.handleSearch()
})

onMounted(() => {
  if (userID.value > 0) {
    getUserDetails()
    $table.value?.handleSearch()
    return
  }
  $message.error('缺少用户信息，请从用户列表重新进入')
})
</script>
