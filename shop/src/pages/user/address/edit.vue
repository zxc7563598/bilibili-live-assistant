<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="编辑地址" />
    <!-- 编辑态骨架屏 -->
    <main v-if="loading" class="mx-auto w-full max-w-5xl space-y-4 px-4 pt-3">
      <AppSkeleton class="h-10 w-full rounded-full" />
      <AppSkeleton class="h-12 w-full rounded-2xl" />
      <AppSkeleton class="h-12 w-full rounded-2xl" />
      <AppSkeleton class="h-12 w-full rounded-2xl" />
      <AppSkeleton class="h-24 w-full rounded-2xl" />
      <div class="card flex items-center justify-between p-4">
        <AppSkeleton class="h-4 w-24" />
        <AppSkeleton class="h-7 w-12 rounded-full" />
      </div>
      <AppSkeleton class="h-13 w-full rounded-full" />
    </main>
    <main v-else class="mx-auto w-full max-w-5xl space-y-4 px-4 pt-3">
      <AppSegmentedControl v-if="paramType == null" v-model="type" :options="typeOptions" class="w-full" />
      <template v-if="type === 1">
        <AppInput v-model="name" label="收件人" placeholder="请输入收件人姓名" />
        <AppInput v-model="phone" label="手机号" type="tel" placeholder="请输入手机号" />
        <AppSelect v-model="region_code" :options="regions" label="所在地区" placeholder="省 / 市 / 区" @selected="onRegionSelected" />
        <AppInput v-model="detail" type="textarea" label="详细地址" placeholder="街道、门牌号等" />
      </template>
      <template v-else>
        <AppInput v-model="name" label="收件人" placeholder="请输入收件人姓名" />
        <AppInput v-model="email" label="电子邮箱" type="email" placeholder="用于接收虚拟商品" />
      </template>
      <div class="card flex items-center justify-between p-4">
        <div>
          <p class="text-[15px] font-medium">
            设为默认地址
          </p>
          <p class="mt-0.5 text-xs text-fg-3">
            下单时优先选择该地址
          </p>
        </div>
        <AppSwitch v-model="is_default" />
      </div>
      <AppButton block size="lg" :loading="saveLoading" @click="save">
        保存
      </AppButton>
    </main>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()
const route = useRoute()
const regions = ref([]) // 静态省市区树
const paramType = ref(null)

const loading = ref(false)
const id = ref(0)
const type = ref(1)
const name = ref('')
const phone = ref('')
const region = ref('')
const region_code = ref([])
const detail = ref('')
const email = ref('')
const is_default = ref(true)
const saveLoading = ref(false)

const typeOptions = [
  { label: '实体地址', value: 1 },
  { label: '虚拟地址', value: 0 },
]

function validate() {
  if (!name.value) {
    toast.error('请填写收件人')
    return false
  }
  if (type.value === 1) {
    if (!phone.value) {
      toast.error('请填写手机号')
      return false
    }
    if (!region_code.value.length) {
      toast.error('请选择所在地区')
      return false
    }
    if (!detail.value) {
      toast.error('请填写详细地址')
      return false
    }
  }
  else if (!email.value) {
    toast.error('请填写电子邮箱')
    return false
  }
  return true
}

function onRegionSelected({ display }) {
  region.value = display
}

// 后端对虚拟/已清空地区的地址返回空串 region_code，需容错为空数组，避免 JSON.parse('') 崩溃
function parseRegionCodes(src) {
  if (!src)
    return []
  try {
    const arr = JSON.parse(src)
    return Array.isArray(arr) ? arr : []
  }
  catch {
    return []
  }
}

function save() {
  if (!validate())
    return
  saveLoading.value = true
  api.savedAddress(
    id.value,
    name.value,
    phone.value,
    JSON.stringify(region_code.value),
    region.value,
    detail.value,
    email.value,
    type.value,
    is_default.value ? 1 : 0,
  ).then((res) => {
    if (res.code === 0) {
      toast.success(id.value > 0 ? '保存成功' : '添加成功')
      if (window.history.length > 1) {
        router.back()
      }
      else {
        router.replace('/user/address')
      }
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
  }).finally(() => {
    saveLoading.value = false
  })
}

// 省市区静态树懒加载
async function loadRegions() {
  if (!regions.value.length)
    regions.value = (await import('@/data/regions.json')).default
}

onMounted(() => {
  if (route.params.type) {
    paramType.value = Number(route.params.type)
    if (paramType.value != null) {
      type.value = paramType.value
    }
  }
  loadRegions()
  if (route.params.id && route.params.id > 0) {
    loading.value = true
    api.getAddressDetails(Number(route.params.id)).then((res) => {
      if (res.code === 0) {
        id.value = Number(route.params.id)
        type.value = res.data.type
        name.value = res.data.name
        phone.value = res.data.phone
        region_code.value = parseRegionCodes(res.data.region_code)
        region.value = res.data.region
        detail.value = res.data.detail
        email.value = res.data.email
        is_default.value = res.data.is_default
      }
      else {
        toast.error(res.msg)
      }
    }).catch(() => {
      toast.error('加载失败，请重试')
    }).finally(() => {
      loading.value = false
    })
  }
})
</script>
