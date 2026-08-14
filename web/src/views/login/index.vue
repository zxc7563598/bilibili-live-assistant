<!-- Copyright © 2023 Ronnie Zhang (大脸怪). MIT License. -->
<template>
  <div class="login-page relative wh-full flex-col overflow-hidden">
    <div class="login-blob login-blob-1 -z-1" />
    <div class="login-blob login-blob-2 -z-1" />
    <div class="login-card relative m-auto max-w-[calc(100vw-32px)] w-345 flex items-stretch justify-center overflow-hidden rounded-16 card-shadow md:w-700">
      <div class="login-brand relative hidden w-320 flex-col shrink-0 justify-center px-32 py-40 md:flex">
        <img src="@/assets/images/logo.png" class="relative z-1 h-64 w-64" alt="logo">
        <h1 class="relative z-1 mt-16 text-26 text-white font-bold leading-tight">
          {{ title }}
        </h1>
        <p class="relative z-1 mt-8 text-14 text-white/65 leading-relaxed">
          B站直播机器人助手
        </p>
      </div>
      <div class="relative z-1 w-full flex flex-col px-24 py-32 md:flex-1 md:px-32 md:py-40">
        <div class="mb-28 flex items-center md:hidden">
          <div class="login-logo-tile h-44 w-44 f-c-c shrink-0 rounded-12">
            <img src="@/assets/images/logo.png" class="h-28 w-28" alt="logo">
          </div>
          <span class="ml-12 text-18 text-#333 font-bold dark:text-#eee">{{ title }}</span>
        </div>
        <div class="mb-28 hidden md:block">
          <h1 class="text-24 text-#333 font-bold dark:text-#eee">
            欢迎登录
          </h1>
          <p class="mt-6 text-14 text-#8a8a8a dark:text-#7d7d7d">
            欢迎回来，请登录您的账号
          </p>
        </div>
        <n-input v-model:value="loginInfo.username" autofocus size="large" placeholder="请输入用户名" :maxlength="20">
          <template #prefix>
            <i class="i-fe:user mr-12 opacity-20" />
          </template>
        </n-input>
        <n-input v-model:value="loginInfo.password" class="mt-16" size="large" type="password" show-password-on="mousedown" placeholder="请输入密码" :maxlength="20">
          <template #prefix>
            <i class="i-fe:lock mr-12 opacity-20" />
          </template>
        </n-input>
        <div v-if="isCaptcha" class="mt-16 flex items-center">
          <altcha-widget type="checkbox" language="zh-cn" :challenge="challenge" class="w-full" @statechange="onStateChange" />
        </div>
        <n-checkbox class="mt-16" :checked="isRemember" label="记住我" :on-update:checked="(val) => (isRemember = val)" />
        <div class="mt-24 flex items-center gap-16">
          <n-button class="h-40 flex-1 rounded-8 text-16" type="primary" ghost @click="quickLogin">
            一键体验
          </n-button>
          <n-button class="h-40 flex-1 rounded-8 text-16" type="primary" :loading="loading" @click="handleLogin">
            登录
          </n-button>
        </div>
      </div>
    </div>
    <div class="relative">
      <TheFooter class="py-12" />
    </div>
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

<style scoped>
.login-page {
  background-color: #f4f5f8;
  background-image:
    radial-gradient(1000px 520px at 110% -20%, rgba(var(--primary-color), 0.1), transparent 62%),
    radial-gradient(900px 480px at -15% 115%, rgba(var(--primary-color), 0.09), transparent 60%);
}
.dark .login-page {
  background-color: #18181c;
  background-image:
    radial-gradient(1000px 520px at 110% -20%, rgba(var(--primary-color), 0.14), transparent 62%),
    radial-gradient(900px 480px at -15% 115%, rgba(var(--primary-color), 0.12), transparent 60%);
}

.login-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(70px);
  pointer-events: none;
}
.login-blob-1 {
  top: -130px;
  right: -90px;
  width: 380px;
  height: 380px;
  background: rgba(var(--primary-color), 0.2);
}
.login-blob-2 {
  bottom: -110px;
  left: -70px;
  width: 320px;
  height: 320px;
  background: rgba(var(--primary-color), 0.14);
}
.dark .login-blob-1 {
  background: rgba(var(--primary-color), 0.18);
}
.dark .login-blob-2 {
  background: rgba(var(--primary-color), 0.12);
}

.login-card {
  background: rgba(255, 255, 255, 0.75);
  -webkit-backdrop-filter: blur(20px);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.65);
}
.dark .login-card {
  background: rgba(24, 24, 28, 0.62);
  border-color: rgba(255, 255, 255, 0.1);
}

.login-brand {
  overflow: hidden;
  background: linear-gradient(
    155deg,
    rgb(var(--primary-color)) 0%,
    color-mix(in srgb, rgb(var(--primary-color)) 55%, #101014) 100%
  );
}
.login-brand::before,
.login-brand::after {
  content: '';
  position: absolute;
  border-radius: 50%;
}
.login-brand::before {
  top: -70px;
  right: -55px;
  width: 210px;
  height: 210px;
  background: rgba(255, 255, 255, 0.1);
}
.login-brand::after {
  bottom: -75px;
  left: -60px;
  width: 230px;
  height: 230px;
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.login-logo-tile {
  background: linear-gradient(
    135deg,
    rgb(var(--primary-color)),
    color-mix(in srgb, rgb(var(--primary-color)) 55%, #101014)
  );
}
</style>
