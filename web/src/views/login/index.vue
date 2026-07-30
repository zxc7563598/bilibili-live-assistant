<!-- Copyright © 2023 Ronnie Zhang (大脸怪). MIT License. -->

<template>
  <div class="wh-full flex-col bg-[url(@/assets/images/login_bg.webp)] bg-cover">
    <div class="m-auto max-w-700 min-w-345 f-c-c rounded-8 auto-bg bg-opacity-20 bg-cover p-12 card-shadow">
      <div class="hidden w-380 px-20 py-35 md:block">
        <img src="@/assets/images/login_banner.webp" class="w-full" alt="login_banner">
      </div>

      <div class="w-320 flex-col px-20 py-32">
        <h2 class="f-c-c text-24 text-#6a6a6a font-normal">
          <img src="@/assets/images/logo.png" class="mr-12 h-50">
          {{ title }}
        </h2>
        <n-input
          v-model:value="loginInfo.username" autofocus class="mt-32 h-40 items-center" placeholder="请输入用户名"
          :maxlength="20"
        >
          <template #prefix>
            <i class="i-fe:user mr-12 opacity-20" />
          </template>
        </n-input>
        <n-input
          v-model:value="loginInfo.password" class="mt-20 h-40 items-center" type="password"
          show-password-on="mousedown" placeholder="请输入密码" :maxlength="20"
        >
          <template #prefix>
            <i class="i-fe:lock mr-12 opacity-20" />
          </template>
        </n-input>

        <div v-if="isCaptcha" class="mt-20 flex items-center">
          <altcha-widget
            type="checkbox" language="zh-cn" :challenge="challenge" class="w-full"
            @statechange="onStateChange"
          />
        </div>

        <n-checkbox class="mt-20" :checked="isRemember" label="记住我" :on-update:checked="(val) => (isRemember = val)" />

        <div class="mt-20 flex items-center">
          <n-button class="h-40 flex-1 rounded-5 text-16" type="primary" ghost @click="quickLogin">
            一键体验
          </n-button>

          <n-button
            class="ml-32 h-40 flex-1 rounded-5 text-16" type="primary" :loading="loading"
            @click="handleLogin"
          >
            登录
          </n-button>
        </div>
      </div>
    </div>

    <TheFooter class="py-12" />
  </div>
</template>

<script setup>
import { useStorage } from '@vueuse/core'
import { useAuthStore } from '@/store'
import { lStorage } from '@/utils'
import api from './api'
import 'altcha'
import 'altcha/i18n/zh-cn'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const title = import.meta.env.VITE_TITLE
const challenge = `${import.meta.env.VITE_AXIOS_BASE_URL}/auth/altcha/challenge`
const isCaptcha = ref(true)

const loginInfo = ref({
  username: '',
  password: '',
  captcha: '',
})

const localLoginInfo = lStorage.get('loginInfo')
if (localLoginInfo) {
  loginInfo.value.username = localLoginInfo.username || ''
  loginInfo.value.password = localLoginInfo.password || ''
}

function quickLogin() {
  loginInfo.value.username = 'admin'
  loginInfo.value.password = '123456'
}

const isRemember = useStorage('isRemember', true)
const loading = ref(false)
async function handleLogin() {
  const { username, password, captcha } = loginInfo.value
  if (!username || !password)
    return $message.warning('请输入用户名和密码')
  if (isCaptcha.value && !captcha)
    return $message.warning('请完成验证')
  try {
    loading.value = true
    $message.loading('正在验证，请稍后...', { key: 'login' })
    const { data } = await api.login({ username, password: password.toString(), captcha })
    if (isRemember.value) {
      lStorage.set('loginInfo', { username, password })
    }
    else {
      lStorage.remove('loginInfo')
    }
    onLoginSuccess(data)
  }
  catch (error) {
    $message.destroy('login')
    console.error(error)
  }
  loading.value = false
}

async function onLoginSuccess(data = {}) {
  authStore.setToken(data.accessToken, data.refreshToken)
  $message.loading('登录中...', { key: 'login' })
  try {
    $message.success('登录成功', { key: 'login' })
    if (route.query.redirect) {
      const path = route.query.redirect
      delete route.query.redirect
      router.push({ path, query: route.query })
    }
    else {
      router.push('/')
    }
  }
  catch (error) {
    console.error(error)
    $message.destroy('login')
  }
}

function onStateChange(ev) {
  switch (ev.detail.state) {
    case 'verified':
      loginInfo.value.captcha = ev.detail.payload
      break
  }
}

onMounted(() => {
  api.isCaptcha().then(({ data }) => {
    isCaptcha.value = !!data?.enabled
  })
})
</script>
