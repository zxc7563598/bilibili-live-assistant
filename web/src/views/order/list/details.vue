<template>
  <CommonPage back title="订单详情">
    <n-spin v-if="!loaded && !emptyText" class="block w-full py-40" size="large" />
    <n-empty v-else-if="emptyText" :description="emptyText" class="py-40" />
    <div v-else class="grid grid-cols-1 gap-16 lg:grid-cols-2">
      <section class="flex flex-col overflow-hidden card-border rounded-10 auto-bg shadow-[0_1px_3px_rgb(0_0_0/0.05)]">
        <header class="flex items-center justify-between gap-12 border-b border-light_border px-16 py-13 dark:border-dark_border">
          <div class="flex items-center gap-10">
            <div class="h-30 w-30 flex shrink-0 items-center justify-center rounded-8 bg-primary/10 text-15 text-primary">
              <i class="i-fe:file-text" />
            </div>
            <div class="flex flex-col gap-2">
              <span class="text-14 font-medium leading-19">基本信息</span>
              <span class="text-12 text-gray-400 leading-16 dark:text-gray-500">订单本身的标识与时间</span>
            </div>
          </div>
        </header>
        <div class="flex flex-col flex-1 gap-11 px-16 py-15">
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              订单编号
            </div>
            <div class="min-w-0 flex-1 break-all text-13 font-medium leading-20">
              {{ detail.order_sn }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              下单用户
            </div>
            <div class="min-w-0 flex flex-1 items-center gap-8 text-13 leading-20">
              <span class="break-all">{{ detail.uname || '未知昵称' }}</span>
              <span class="shrink-0 text-12 text-gray-400 dark:text-gray-500">UID {{ detail.uid }}</span>
              <NButton text type="primary" size="tiny" class="shrink-0" :disabled="!detail.user_id" @click="goUserDetail">
                查看用户
              </NButton>
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              商品ID / SKU
            </div>
            <div class="min-w-0 flex-1 break-all text-13 leading-20">
              {{ detail.product_id }} / {{ detail.product_sku_id || '—' }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              用户备注
            </div>
            <div class="min-w-0 flex-1 break-all text-13 leading-20">
              {{ dash(detail.remark) }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              下单时间
            </div>
            <div class="min-w-0 flex-1 text-13 leading-20">
              {{ dash(detail.created_at) }}
            </div>
          </div>
        </div>
      </section>
      <section class="flex flex-col overflow-hidden card-border rounded-10 auto-bg shadow-[0_1px_3px_rgb(0_0_0/0.05)]">
        <header class="flex items-center justify-between gap-12 border-b border-light_border px-16 py-13 dark:border-dark_border">
          <div class="flex items-center gap-10">
            <div class="h-30 w-30 flex shrink-0 items-center justify-center rounded-8 bg-primary/10 text-15 text-primary">
              <i class="i-fe:package" />
            </div>
            <div class="flex flex-col gap-2">
              <span class="text-14 font-medium leading-19">商品信息</span>
              <span class="text-12 text-gray-400 leading-16 dark:text-gray-500">下单时的商品与规格快照，后续商品改价不影响这里</span>
            </div>
          </div>
        </header>
        <div class="flex flex-col flex-1 px-16 py-15">
          <div class="flex items-start gap-14">
            <div v-if="detail.product_cover" class="h-96 w-96 shrink-0 overflow-hidden border border-light_border rounded-8 dark:border-dark_border">
              <n-image :src="detail.product_cover" :width="96" :height="96" object-fit="cover" class="block" />
            </div>
            <div v-else class="h-96 w-96 flex shrink-0 items-center justify-center border border-light_border rounded-8 text-20 text-gray-300 dark:border-dark_border dark:text-gray-600">
              <i class="i-fe:image" />
            </div>
            <div class="min-w-0 flex flex-col flex-1 gap-8">
              <div class="break-all text-13 font-medium leading-19">
                {{ detail.product_name }}
              </div>
              <div class="break-all text-12 text-gray-400 leading-17 dark:text-gray-500">
                规格：{{ formatProductSpecs(detail.product_spec_properties) || '无' }}
              </div>
            </div>
          </div>
          <div class="mt-auto border-t border-light_border pt-12 dark:border-dark_border">
            <div class="flex flex-wrap items-baseline gap-x-20 gap-y-8">
              <span class="flex items-baseline gap-5">
                <span class="text-12 text-gray-400 dark:text-gray-500">单价</span>
                <span class="text-13">{{ detail.price }}</span>
                <span class="text-12 text-gray-400 dark:text-gray-500">{{ creditLabel }}</span>
              </span>
              <span class="flex items-baseline gap-5">
                <span class="text-12 text-gray-400 dark:text-gray-500">数量</span>
                <span class="text-13">{{ detail.quantity }}</span>
              </span>
              <span class="flex items-baseline gap-5">
                <span class="text-12 text-gray-400 dark:text-gray-500">合计</span>
                <span class="text-15 text-primary font-medium">{{ totalAmount }}</span>
                <span class="text-12 text-gray-400 dark:text-gray-500">{{ creditLabel }}</span>
              </span>
            </div>
          </div>
        </div>
      </section>
      <section class="flex flex-col overflow-hidden card-border rounded-10 auto-bg shadow-[0_1px_3px_rgb(0_0_0/0.05)]">
        <header class="flex items-center justify-between gap-12 border-b border-light_border px-16 py-13 dark:border-dark_border">
          <div class="flex items-center gap-10">
            <div class="h-30 w-30 flex shrink-0 items-center justify-center rounded-8 bg-primary/10 text-15 text-primary">
              <i class="i-fe:map-pin" />
            </div>
            <div class="flex flex-col gap-2">
              <span class="text-14 font-medium leading-19">收货信息</span>
              <span class="text-12 text-gray-400 leading-16 dark:text-gray-500">{{ receiverTypeLabel }}</span>
            </div>
          </div>
          <NButton size="small" secondary type="primary" class="shrink-0" :disabled="!receiverEditable" @click="receiverModalRef?.open(detail)">
            变更收货信息
          </NButton>
        </header>
        <div class="flex flex-col flex-1 gap-11 px-16 py-15">
          <template v-if="isVirtual">
            <div class="flex items-start gap-12">
              <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
                收货邮箱
              </div>
              <div class="min-w-0 flex-1 break-all text-13 leading-20">
                {{ dash(detail.receiver_email) }}
              </div>
            </div>
          </template>
          <template v-else-if="isActual">
            <div class="flex items-start gap-12">
              <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
                收货人
              </div>
              <div class="min-w-0 flex-1 break-all text-13 leading-20">
                {{ dash(detail.receiver_name) }}
              </div>
            </div>
            <div class="flex items-start gap-12">
              <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
                手机号
              </div>
              <div class="min-w-0 flex-1 break-all text-13 leading-20">
                {{ dash(detail.receiver_phone) }}
              </div>
            </div>
            <div class="flex items-start gap-12">
              <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
                所在地区
              </div>
              <div class="min-w-0 flex-1 break-all text-13 leading-20">
                {{ dash(detail.receiver_region) }}
              </div>
            </div>
            <div class="flex items-start gap-12">
              <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
                详细地址
              </div>
              <div class="min-w-0 flex-1 break-all text-13 leading-20">
                {{ dash(detail.receiver_detail) }}
              </div>
            </div>
          </template>
          <div v-else class="text-13 text-red-500 leading-20">
            收货类型（{{ detail.receiver_type }}）无法识别，请检查订单数据
          </div>
        </div>
      </section>
      <section class="flex flex-col overflow-hidden card-border rounded-10 auto-bg shadow-[0_1px_3px_rgb(0_0_0/0.05)]">
        <header class="flex items-center justify-between gap-12 border-b border-light_border px-16 py-13 dark:border-dark_border">
          <div class="flex items-center gap-10">
            <div class="h-30 w-30 flex shrink-0 items-center justify-center rounded-8 bg-primary/10 text-15 text-primary">
              <i class="i-fe:truck" />
            </div>
            <div class="flex flex-col gap-2">
              <span class="text-14 font-medium leading-19">发货信息</span>
              <span class="text-12 text-gray-400 leading-16 dark:text-gray-500">变更发货状态会连带变更订单状态</span>
            </div>
          </div>
          <NButton size="small" secondary type="primary" class="shrink-0" @click="shipModalRef?.open(detail)">
            变更发货状态
          </NButton>
        </header>
        <div class="flex flex-col flex-1 gap-11 px-16 py-15">
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              发货状态
            </div>
            <div class="min-w-0 flex-1 text-13 leading-20">
              {{ getOptionsLabel(shipStatusOptions, detail.ship_status) }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              快递公司
            </div>
            <div class="min-w-0 flex-1 break-all text-13 leading-20">
              {{ isActual ? dash(detail.express_company) : '虚拟商品无物流' }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              快递单号
            </div>
            <div class="min-w-0 flex-1 break-all text-13 leading-20">
              {{ isActual ? dash(detail.express_no) : '虚拟商品无物流' }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              发货时间
            </div>
            <div class="min-w-0 flex-1 text-13 leading-20">
              {{ dash(detail.processed_at) }}
            </div>
          </div>
        </div>
      </section>
      <section class="flex flex-col overflow-hidden card-border rounded-10 auto-bg shadow-[0_1px_3px_rgb(0_0_0/0.05)] lg:col-span-2">
        <header class="flex items-center justify-between gap-12 border-b border-light_border px-16 py-13 dark:border-dark_border">
          <div class="flex items-center gap-10">
            <div class="h-30 w-30 flex shrink-0 items-center justify-center rounded-8 bg-primary/10 text-15 text-primary">
              <i class="i-fe:activity" />
            </div>
            <div class="flex flex-col gap-2">
              <span class="text-14 font-medium leading-19">订单状态</span>
              <span class="text-12 text-gray-400 leading-16 dark:text-gray-500">纯状态改写，不涉及退款与库存回滚</span>
            </div>
          </div>
          <NButton size="small" secondary type="primary" class="shrink-0" @click="orderModalRef?.open(detail)">
            变更订单状态
          </NButton>
        </header>
        <div class="grid grid-cols-1 flex-1 gap-x-24 gap-y-11 px-16 py-15 md:grid-cols-2">
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              订单状态
            </div>
            <div class="min-w-0 flex-1 text-13 font-medium leading-20">
              {{ getOptionsLabel(orderStatusOptions, detail.order_status) }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              支付状态
            </div>
            <div class="min-w-0 flex-1 text-13 leading-20">
              {{ getOptionsLabel(payStatusOptions, detail.pay_status) }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              支付时间
            </div>
            <div class="min-w-0 flex-1 text-13 leading-20">
              {{ dash(detail.pay_at) }}
            </div>
          </div>
          <div class="flex items-start gap-12">
            <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
              取消时间
            </div>
            <div class="min-w-0 flex-1 text-13 leading-20">
              {{ dash(detail.cancel_at) }}
            </div>
          </div>
        </div>
      </section>
    </div>
    <ShipStatusModal ref="shipModalRef" @success="onChanged" />
    <OrderStatusModal ref="orderModalRef" @success="onChanged" />
    <ReceiverModal ref="receiverModalRef" @success="onChanged" />
  </CommonPage>
</template>

<script setup>
import { NButton } from 'naive-ui'
import { formatProductSpecs, getOptionsLabel } from '@/utils'
import api from './api'
import OrderStatusModal from './components/OrderStatusModal.vue'
import ReceiverModal from './components/ReceiverModal.vue'
import ShipStatusModal from './components/ShipStatusModal.vue'

defineOptions({ name: 'OrderDetails' })

const ADDRESS_TYPE_VIRTUAL = 0
const ADDRESS_TYPE_ACTUAL = 1

const route = useRoute()
const router = useRouter()

const shipModalRef = ref(null)
const orderModalRef = ref(null)
const receiverModalRef = ref(null)

const orderID = computed(() => Number(route.query.id) || 0)
const detail = ref(null)
const loaded = ref(false)
const emptyText = ref('')

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

const isVirtual = computed(() => Number(detail.value?.receiver_type) === ADDRESS_TYPE_VIRTUAL)
const isActual = computed(() => Number(detail.value?.receiver_type) === ADDRESS_TYPE_ACTUAL)
const receiverEditable = computed(() => isVirtual.value || isActual.value)
const receiverTypeLabel = computed(() => {
  if (isVirtual.value)
    return '虚拟商品，只记录收货邮箱'
  if (isActual.value)
    return '实体商品，需要真实的收货地址'
  return '收货类型异常'
})

// price 是星光/积分的数量，不是金额分位，所以不做任何除法
const creditLabel = computed(() => (Number(detail.value?.credit_type) === 0 ? '星光' : '积分'))
const totalAmount = computed(() => {
  const d = detail.value
  if (!d)
    return 0
  return Number(d.price || 0) * Number(d.quantity || 0)
})

// 后端 timeutil.Format 对 0 值返回空串，这里统一成占位符，避免出现只有标签没内容的空行
function dash(value) {
  return value === '' || value === null || value === undefined ? '—' : value
}

function goUserDetail() {
  router.push({ path: '/shop/user/details', query: { id: detail.value.user_id } })
}

async function load() {
  // 先清空，避免切换订单时短暂显示上一个单的数据
  detail.value = null
  loaded.value = false
  emptyText.value = ''
  try {
    const res = await api.getDetails(orderID.value)
    detail.value = res.data
    loaded.value = true
  }
  catch {
    emptyText.value = '订单信息加载失败，请从订单列表重新进入'
  }
}

function onChanged() {
  load()
}

watch(orderID, (id) => {
  if (id <= 0)
    return
  load()
})

onMounted(() => {
  if (orderID.value > 0) {
    load()
    return
  }
  emptyText.value = '缺少订单信息，请从订单列表重新进入'
  $message.error('缺少订单信息，请从订单列表重新进入')
})
</script>
