<template>
  <div class="min-h-dvh bg-bg pb-28">
    <AppNavBar title="确认订单" />
    <section v-if="confirm.product.id > 0" class="mx-auto mt-3 w-full max-w-5xl px-4">
      <div v-if="!expired" class="card flex items-center gap-3 border-warning/40 bg-warning/10 p-4">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-warning/15 text-warning">
          <AppIcon name="clock" :size="20" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold text-warning">
            库存已锁定，请尽快兑换
          </p>
          <p class="mt-0.5 text-xs text-fg-3">
            超时后库存将自动释放，需重新购买
          </p>
        </div>
        <div class="shrink-0 rounded-full bg-warning/15 px-3 py-1.5 text-sm font-bold tabular-nums text-warning">
          {{ countdownText }}
        </div>
      </div>
      <div v-else class="card flex items-center gap-3 border-danger/40 bg-danger/10 p-4">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-danger/15 text-danger">
          <AppIcon name="close" :size="20" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold text-danger">
            支付超时，库存已释放
          </p>
          <p class="mt-0.5 text-xs text-fg-3">
            如需购买，请重新发起兑换
          </p>
        </div>
        <AppButton size="sm" variant="danger" :loading="reOrderLoading" @click="reOrder()">
          重新购买
        </AppButton>
      </div>
    </section>
    <button v-if="!addressLoading && selectedAddress.id" class="mx-auto block w-full max-w-5xl px-4 mt-3" @click="showAddressSheet = true">
      <div class="card flex w-full items-center gap-3 p-4 text-left">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary">
          <AppIcon :name="selectedAddress.type === 0 ? 'mail' : 'map-pin'" :size="20" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold">
            {{ selectedAddress.name }}
          </p>
          <p class="mt-0.5 truncate text-xs text-fg-3">
            {{ addressLine }}
          </p>
        </div>
        <AppIcon name="chevron-right" :size="18" class="text-fg-3" />
      </div>
    </button>
    <button v-if="!addressLoading && !selectedAddress.id" class="mx-auto block w-full max-w-5xl px-4 mt-3" @click="router.push('/user/address/edit')">
      <div class="card flex items-center justify-center gap-2 border-dashed p-4 text-primary">
        <AppIcon name="plus" :size="20" />
        <span class="text-sm font-medium">添加收货地址</span>
      </div>
    </button>
    <section v-if="confirm.product.id > 0" class="mx-auto mt-3 w-full max-w-5xl px-4">
      <div class="card flex items-center gap-3 p-4">
        <div class="w-20 shrink-0">
          <AppImage :src="confirm.product.cover" :label="confirm.product.name" ratio="1 / 1" rounded="rounded-lg" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="line-clamp-2 text-sm font-medium leading-snug">
            {{ confirm.product.name }}
          </p>
          <p class="mt-1 text-xs text-fg-3">
            规格：{{ confirm.product.sku.join('·') }}
          </p>
          <p class="mt-1 text-xs text-fg-3">
            数量：{{ confirm.product.num }}
          </p>
        </div>
        <div class="flex shrink-0 items-center gap-1" :class="confirm.product.type === 0 ? 'text-starlight' : 'text-primary'">
          <AppIcon :name="confirm.product.type === 0 ? 'star' : 'points'" :size="16" />
          <span class="font-bold tabular-nums">{{ confirm.product.amount }}</span>
        </div>
      </div>
    </section>
    <section v-if="confirm.product.id > 0" class="mx-auto mt-3 w-full max-w-5xl px-4">
      <div class="card divide-y divide-line p-4 text-sm">
        <div class="flex items-center justify-between py-2.5">
          <span class="text-fg-2">商品{{ confirm.product.type === 0 ? '星光' : '积分' }}</span>
          <span class="tabular-nums">{{ confirm.product.amount }}</span>
        </div>
        <div class="flex items-center justify-between py-2.5">
          <span class="text-fg-2">数量</span>
          <span class="tabular-nums">{{ confirm.product.num }}</span>
        </div>
        <div class="flex items-center justify-between py-2.5">
          <span class="text-fg-2">运费</span>
          <span class="text-success">由主播承担哦</span>
        </div>
        <div class="flex items-center justify-between py-2.5">
          <span class="text-fg-2">优惠</span>
          <span class="tabular-nums">-0</span>
        </div>
        <div class="flex items-center justify-between pt-3">
          <span class="font-semibold">合计</span>
          <span class="flex items-center gap-1" :class="confirm.product.type === 0 ? 'text-starlight' : 'text-primary'">
            <AppIcon :name="confirm.product.type === 0 ? 'star' : 'points'" :size="18" />
            <span class="text-lg font-bold tabular-nums">{{ confirm.product.amount * confirm.product.num }}</span>
          </span>
        </div>
      </div>
      <p class="mt-3 flex items-center gap-1.5 text-xs text-fg-3">
        <AppIcon name="info" :size="14" />兑换成功后{{ confirm.product.type === 0 ? '星光' : '积分' }}立即扣除，商品的发货时间需与主播进行确认
      </p>
    </section>
    <div v-if="confirm.id > 0" class="glass-strong safe-bottom fixed inset-x-0 bottom-0 z-40">
      <div class="mx-auto flex w-full max-w-5xl items-center gap-4 px-4 py-2.5">
        <div class="shrink-0">
          <p class="text-xs text-fg-3">
            需支付
          </p>
          <p class="mt-0.5 flex items-center gap-1" :class="confirm.product.type === 0 ? 'text-starlight' : 'text-primary'">
            <AppIcon :name="confirm.product.type === 0 ? 'star' : 'points'" :size="18" />
            <span class="text-xl font-bold tabular-nums">{{ confirm.product.amount * confirm.product.num }}</span>
          </p>
        </div>
        <AppButton size="md" class="flex-1" :loading="confirmPaymentLoading" :disabled="expired || !confirm.product.id" @click="confirmPayment">
          {{ expired ? '订单已超时' : '确认兑换' }}
        </AppButton>
      </div>
    </div>
    <AppBottomSheet v-model="showAddressSheet" title="选择收货地址">
      <div class="space-y-2">
        <button v-for="a in address" :key="a.id" class="flex w-full items-center gap-3 rounded-2xl border p-3 text-left transition" :class="selectedAddress.id === a.id ? 'border-primary bg-primary-soft' : 'border-line'" @click="selectedAddress = a; showAddressSheet = false">
          <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary">
            <AppIcon :name="a.type === 0 ? 'mail' : 'map-pin'" :size="18" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium">
              {{ a.name }}
              <Tag :color="a.type === 0 ? 'info' : 'primary'" class="ml-1">
                {{ a.type === 0 ? '虚拟' : '实体' }}
              </Tag>
            </p>
            <p class="mt-0.5 truncate text-xs text-fg-3">
              {{ a.type === 0 ? a.email : `${a.phone} · ${a.region} ${a.detail}` }}
            </p>
          </div>
          <AppIcon v-if="selectedAddress.id === a.id" name="check" :size="18" class="text-primary" />
        </button>
        <router-link to="/user/address" class="flex items-center justify-center gap-1 py-2 text-sm text-primary">
          <AppIcon name="plus" :size="16" />管理收货地址
        </router-link>
      </div>
    </AppBottomSheet>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()

const confirm = ref({
  id: 0,
  expireAt: 0, // 库存锁定截止时间（后端返回，毫秒时间戳），0 表示尚未返回
  product: {
    id: 0,
    name: '',
    cover: '',
    amount: 0,
    type: 0,
    sku: [],
    num: 0,
  },
})

const confirmPaymentLoading = ref(false)

// 倒计时：每秒刷新 now，剩余时间由 expireAt 推算（墙钟制，切后台不漂移）
const now = ref(Date.now())
let timer = null
const remaining = computed(() => Math.max(0, confirm.value.expireAt - now.value))
const expired = computed(() => confirm.value.expireAt > 0 && remaining.value <= 0)
const countdownText = computed(() => {
  const total = Math.floor(remaining.value / 1000)
  const pad = n => String(n).padStart(2, '0')
  if (total >= 3600) {
    const h = Math.floor(total / 3600)
    const m = Math.floor((total % 3600) / 60)
    const s = total % 60
    return `${pad(h)}:${pad(m)}:${pad(s)}`
  }
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${pad(m)}:${pad(s)}`
})
const reOrderLoading = ref(false)

const address = ref([])
const selectedAddress = ref({
  id: 0,
  type: 1,
  name: '',
  phone: '',
  region: '',
  detail: '',
  email: '',
  isDefault: true,
})
const showAddressSheet = ref(false)
const addressLoading = ref(true)

const addressLine = computed(() => {
  const a = selectedAddress.value
  if (a.type === 0)
    return `${a.name} · ${a.email}`
  return `${a.name} ${a.phone} · ${a.region} ${a.detail}`
})

function loadConfirm(reorder = false) {
  api.getConfirm().then((res) => {
    if (res.code === 0) {
      Object.assign(confirm.value, res.data)
      now.value = Date.now()
      if (reorder) {
        toast.success('已重新锁定库存，请尽快完成支付')
        reOrderLoading.value = false
      }
    }
    else {
      toast.error(res.msg)
    }
  })
}

function reOrder() {
  reOrderLoading.value = true
  api.reOrder(confirm.value.id).then((res) => {
    if (res.code === 0) {
      loadConfirm(true)
    }
    else {
      toast.error(res.msg)
    }
  })
}

function confirmPayment() {
  confirmPaymentLoading.value = true
  api.confirmPayment(confirm.value.id, selectedAddress.value.id).then((res) => {
    if (res.code === 0) {
      toast.success('支付成功')
      router.replace('/')
    }
    else {
      toast.error(res.msg)
    }
  }).finally(() => {
    confirmPaymentLoading.value = false
  })
}

onMounted(() => {
  timer = setInterval(() => {
    now.value = Date.now()
  }, 1000)
  api.getAddressList().then((res) => {
    if (res.code === 0) {
      Object.assign(address.value, res.data.list)
      address.value.forEach((item) => {
        if (item.isDefault) {
          Object.assign(selectedAddress.value, item)
        }
      })
      if (address.value.length && !address.value.some(item => item.isDefault)) {
        Object.assign(selectedAddress.value, address.value[0])
      }
    }
    else {
      toast.error(res.msg)
    }
    addressLoading.value = false
  })
  loadConfirm()
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>
