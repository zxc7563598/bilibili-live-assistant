<template>
  <div class="min-h-dvh bg-bg pb-28">
    <AppNavBar title="收货地址" />
    <main class="mx-auto w-full max-w-5xl px-4 pt-3">
      <!-- 骨架屏 -->
      <div v-if="addressLoading" class="space-y-3">
        <div v-for="n in 2" :key="n" class="card space-y-3 p-4">
          <div class="flex items-center gap-3">
            <AppSkeleton class="h-10 w-10 rounded-full" />
            <div class="flex-1 space-y-2">
              <AppSkeleton class="h-4 w-1/2" />
              <AppSkeleton class="h-3 w-3/4" />
            </div>
          </div>
          <AppSkeleton class="h-3 w-full" />
        </div>
      </div>
      <!-- 地址列表 -->
      <TransitionGroup v-else name="list" tag="div" class="space-y-3">
        <div v-for="a in address" :key="a.id" class="card p-4">
          <div class="flex items-start gap-3">
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary">
              <AppIcon :name="a.type === 0 ? 'mail' : 'map-pin'" :size="20" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <p class="text-[15px] font-semibold">
                  {{ a.name }}
                </p>
                <Tag :color="a.type === 0 ? 'info' : 'primary'">
                  {{ a.type === 0 ? '虚拟' : '实体' }}
                </Tag>
                <Tag v-if="a.isDefault" color="success">
                  默认
                </Tag>
              </div>
              <p v-if="a.type === 0" class="mt-1 text-sm text-fg-2">
                {{ a.email }}
              </p>
              <p v-else class="mt-1 text-sm text-fg-2">
                {{ a.phone }}
              </p>
              <p v-if="a.type === 1" class="mt-0.5 text-sm text-fg-3">
                {{ a.region }} {{ a.detail }}
              </p>
            </div>
          </div>
          <div class="mt-3 flex items-center justify-end gap-3 border-t border-line pt-3 text-sm">
            <button class="flex items-center gap-1 text-fg-2 press" @click="router.push(`/user/address/edit/${a.id}`)">
              <AppIcon name="edit" :size="15" />编辑
            </button>
            <button class="flex items-center gap-1 text-danger press" @click="remove(a)">
              <AppIcon name="trash" :size="15" />删除
            </button>
          </div>
        </div>
      </TransitionGroup>
      <!-- 空态 -->
      <AppEmpty v-if="!addressLoading && !address.length" icon="map-pin" title="还没有收货地址" description="点击下方按钮新增" />
    </main>
    <div class="safe-bottom fixed inset-x-0 bottom-0 z-40 px-4 py-3">
      <div class="mx-auto w-full max-w-5xl">
        <AppButton block size="lg" @click="router.push('/user/address/edit')">
          <AppIcon name="plus" :size="18" />新增地址
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import dialog from '@/utils/dialog'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()
const address = ref([])
const addressLoading = ref(true)

function remove(a) {
  dialog.warning({
    title: '删除地址',
    message: `确认删除该${a.type === 0 ? '虚拟' : '实体'}地址？`,
    confirmText: '删除',
    confirmVariant: 'danger',
    onConfirm: () => {
      address.value = address.value.filter(x => x.id !== a.id)
      toast.success('地址已删除')
    },
  })
}

onMounted(() => {
  api.getAddressList().then((res) => {
    if (res.code === 0) {
      Object.assign(address.value, res.data.list)
    }
    else {
      toast.error(res.msg)
    }
  }).catch(() => {
    toast.error('加载失败，请重试')
  }).finally(() => {
    addressLoading.value = false
  })
})
</script>
