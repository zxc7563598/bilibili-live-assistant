<template>
  <CommonPage>
    <MeCrud ref="$table" v-model:query-items="query" :columns="columns" :get-data="api.getList" :scroll-x="1250" export-module="liveuser">
      <MeQueryItem label="用户UID">
        <n-input-number v-model:value="query.uid" :show-button="false" :precision="0" placeholder="用户完整UID" />
      </MeQueryItem>
      <MeQueryItem label="用户昵称">
        <n-input v-model:value="query.uname" placeholder="支持模糊搜索" />
      </MeQueryItem>
    </MeCrud>
    <BalanceModal ref="balanceModalRef" @success="onBalanceSuccess" />
    <ResetPasswordModal ref="passwordModalRef" />
    <GuardExpireModal ref="guardModalRef" @success="onGuardSuccess" />
  </CommonPage>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { GuardExpireModal, GuardTag, MeCrud, MeQueryItem } from '@/components'
import api from './api'
import BalanceModal from './components/BalanceModal.vue'
import ResetPasswordModal from './components/ResetPasswordModal.vue'

const router = useRouter()

const $table = ref(null)
const balanceModalRef = ref(null)
const passwordModalRef = ref(null)
const guardModalRef = ref(null)

// 列表里没有用户内部ID之外的标识，变更弹窗要靠昵称+UID 让管理员确认操作对象
function userLabel(row) {
  return `${row.uname || '未知昵称'}（UID ${row.uid}）`
}

const columns = [
  { title: '用户UID', key: 'uid', minWidth: 190, sorter: true },
  { title: '用户昵称', key: 'uname', minWidth: 140, sorter: true, render(row) {
    // 昵称前的身份标签点开即可变更身份，只影响展示，不单独占一列
    return h('div', { class: 'flex items-center' }, [
      h(GuardTag, {
        type: row.vip_type,
        class: 'mr-6',
        onClick: () => guardModalRef.value?.open(row.id, userLabel(row)),
      }),
      h('span', row.uname),
    ])
  } },
  { title: '积分', key: 'points', width: 120, sorter: true },
  { title: '星光', key: 'stars', width: 120, sorter: true },
  { title: '发送弹幕', key: 'total_danmu_count', width: 120, sorter: true },
  { title: '礼物金额', key: 'total_gift_amount', width: 120, sorter: true, render(row) {
    const total_gift_amount = (row.total_gift_amount / 100)
    return `¥${total_gift_amount.toFixed(2)}`
  } },
  {
    title: '操作',
    key: 'actions',
    width: 400,
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
            onClick: () => router.push({ path: '/shop/user/details', query: { id: row.id } }),
          },
          {
            default: () => '查看详情',
            icon: () => h('i', { class: 'i-fe:calendar text-14' }),
          },
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            class: 'ml-12px',
            secondary: true,
            onClick: () => balanceModalRef.value?.open(row.id, userLabel(row)),
          },
          {
            default: () => '变更积分/余额',
            icon: () => h('i', { class: 'i-fe:credit-card text-14' }),
          },
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            class: 'ml-12px',
            secondary: true,
            onClick: () => passwordModalRef.value?.open(row.id, userLabel(row)),
          },
          {
            default: () => '修改密码',
            icon: () => h('i', { class: 'i-fe:key text-14' }),
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

// 余额变了积分/星光两列就过期了，刷新当前页
function onBalanceSuccess() {
  $table.value?.handleSearch(true)
}

// 身份变了昵称前的标签就过期了，同样只刷新当前页
function onGuardSuccess() {
  $table.value?.handleSearch(true)
}

onMounted(() => {
  $table.value?.handleSearch()
})
</script>
