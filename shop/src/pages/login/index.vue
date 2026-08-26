<template>
  <div class="relative flex min-h-dvh items-center justify-center overflow-hidden px-5 py-8">
    <div class="absolute inset-0 -z-10 bg-bg">
      <img v-if="loginBg" :src="loginBg" alt="" class="absolute inset-0 h-full w-full object-cover">
      <template v-else>
        <div class="absolute -left-24 -top-24 h-80 w-80 rounded-full opacity-30 blur-3xl" style="background: color-mix(in srgb, var(--primary) 55%, transparent)" />
        <div class="absolute -right-20 top-1/3 h-96 w-96 rounded-full opacity-20 blur-3xl" style="background: color-mix(in srgb, var(--primary) 70%, transparent)" />
        <div class="absolute bottom-0 left-1/4 h-72 w-72 rounded-full opacity-20 blur-3xl" style="background: var(--primary)" />
      </template>
    </div>
    <div class="safe-top absolute inset-x-0 top-0 z-10 flex items-center justify-between px-4 pt-3">
      <div class="rounded-full bg-black/25 px-3 py-1 text-xs text-white backdrop-blur" />
      <button class="flex h-10 w-10 items-center justify-center rounded-full text-fg-2 glass press" aria-label="切换主题" @click="toggleTheme">
        <AppIcon :name="isDark ? 'sun' : 'moon'" :size="20" />
      </button>
    </div>
    <div class="glass-strong w-full max-w-sm rounded-3xl p-6">
      <div class="flex flex-col items-center">
        <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-on-primary shadow-(--shadow-btn)">
          <AppImage v-if="logo" :src="logo" alt="logo" class="w-28" />
          <AppIcon v-else name="points" :size="28" />
        </div>
        <h1 class="mt-3 text-xl font-bold">
          {{ title }}
        </h1>
        <p class="mt-1 text-sm text-fg-3">
          {{ slogan }}
        </p>
      </div>
      <div v-if="step === 1" class="mt-6 space-y-4">
        <label class="block">
          <span class="mb-1.5 block text-sm font-medium text-fg-2">B站UID</span>
          <input v-model="uid" type="number" inputmode="numeric" placeholder="请输入你的 UID" class="h-12 w-full rounded-2xl border border-line bg-surface px-4 text-[15px] outline-none transition placeholder:text-fg-3 focus:border-primary">
        </label>
        <AppButton block size="lg" :loading="nextLoading" @click="next">
          下一步
        </AppButton>
      </div>
      <div v-else class="mt-6 space-y-4">
        <p class="text-center text-sm text-fg-2">
          <AppIcon name="user" :size="16" class="mr-1 inline-block align-[-3px]" />{{ uid }} · {{ stepTitle }}
        </p>
        <label class="block">
          <span class="mb-1.5 block text-sm font-medium text-fg-2">{{ needSetup ? '设置密码' : '密码' }}</span>
          <input v-model="password" type="password" :placeholder="needSetup ? '请设置 6 位以上密码' : '请输入密码'" class="h-12 w-full rounded-2xl border border-line bg-surface px-4 text-[15px] outline-none transition placeholder:text-fg-3 focus:border-primary">
        </label>
        <label v-if="needSetup" class="block">
          <span class="mb-1.5 block text-sm font-medium text-fg-2">确认密码</span>
          <input v-model="confirmPwd" type="password" placeholder="请再次输入密码" class="h-12 w-full rounded-2xl border border-line bg-surface px-4 text-[15px] outline-none transition placeholder:text-fg-3 focus:border-primary">
        </label>
        <AppButton block size="lg" :loading="submitLoading" @click="submit">
          {{ needSetup ? '设置密码并登录' : '登录' }}
        </AppButton>
        <button class="w-full text-center text-sm text-fg-3" :disabled="submitLoading" @click="step = 1">
          返回上一步
        </button>
      </div>
      <p class="mt-6 text-center text-xs text-fg-3">
        登录即代表同意《用户协议》与《隐私政策》
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { isDark, toggleTheme } from '@/utils/theme'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()

// 页面信息
const loginBg = ref('')
const logo = ref('')
const title = ref('')
const slogan = ref('')

// 表单信息
const step = ref(1)
const uid = ref('')
const password = ref('')
const confirmPwd = ref('')
const needSetup = ref(false)
const nextLoading = ref(false)
const submitLoading = ref(false)

const stepTitle = computed(() => (needSetup.value ? '设置密码' : '输入密码'))

function next() {
  if (!uid.value) {
    toast.error('请输入 UID')
    return
  }
  nextLoading.value = true
  api.getAccount(uid.value).then((res) => {
    if (res.code === 0) {
      needSetup.value = res.data.account
      step.value = 2
    }
    else {
      toast.error(res.msg)
    }
  }).finally(() => {
    nextLoading.value = false
  })
}

function submit() {
  if (!password.value) {
    toast.error('请输入密码')
    return
  }
  if (needSetup.value && password.value !== confirmPwd.value) {
    toast.error('两次输入的密码不一致')
    return
  }
  // 执行登录
  submitLoading.value = true
  api.login(uid.value, password.value).then((res) => {
    if (res.code === 0) {
      toast.success('登录成功')
      router.replace('/')
    }
    else {
      toast.error(res.msg)
    }
  }).finally(() => {
    submitLoading.value = false
  })
}

onMounted(() => {
  api.getConfig().then(({ data }) => {
    logo.value = data.logo
    loginBg.value = data.loginBg
    title.value = data.title
    slogan.value = data.slogan
  })
})
</script>

<style scoped>
</style>
