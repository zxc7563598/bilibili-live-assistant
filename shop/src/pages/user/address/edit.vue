<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="编辑地址" />
    <main v-if="!route.params.id || route.params.id === id" class="mx-auto w-full max-w-5xl space-y-4 px-4 pt-3">
      <AppSegmentedControl v-model="type" :options="typeOptions" class="w-full" />
      <template v-if="type === 1">
        <AppInput v-model="name" label="收件人" placeholder="请输入收件人姓名" />
        <AppInput v-model="phone" label="手机号" type="tel" placeholder="请输入手机号" />
        <AppSelect v-model="regionCode" :options="regions" label="所在地区" placeholder="省 / 市 / 区" />
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
        <AppSwitch v-model="isDefault" />
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

const id = ref(0)
const type = ref(1)
const name = ref('')
const phone = ref('')
const regionCode = ref([]) // 省市区 code 路径
const detail = ref('')
const email = ref('')
const isDefault = ref(true)
const saveLoading = ref(false)

const typeOptions = [
  { label: '实体地址', value: 1 },
  { label: '虚拟地址', value: 0 },
]

function save() {
  saveLoading.value = true
  api.savedAddress(id.value, type.value, name.value, phone.value, regionCode.value, detail.value, email.value, isDefault.value).then((res) => {
    if (res.code === 0) {
      console.warn(res)
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
  loadRegions()
  if (route.params.id && route.params.id > 0) {
    api.getAddressDetails(id).then((res) => {
      if (res.code === 0) {
        id.value = route.params.id
        type.value = res.data.type
        name.value = res.data.name
        phone.value = res.data.phone
        regionCode.value = res.data.regionCode ?? []
        detail.value = res.data.detail
        email.value = res.data.email
        isDefault.value = res.data.isDefault
      }
      else {
        toast.error(res.msg)
      }
    })
  }
})
</script>
