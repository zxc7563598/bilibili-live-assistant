<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()

const oldPwd = ref('')
const newPwd = ref('')
const confirmPwd = ref('')
const saveLoading = ref(false)

function submit() {
  if (!oldPwd.value) {
    toast.error('请输入当前密码')
    return
  }
  if (!newPwd.value) {
    toast.error('请输入新密码')
    return
  }
  if (newPwd.value !== confirmPwd.value) {
    toast.error('两次输入的密码不一致')
    return
  }
  saveLoading.value = true
  api.savedPassword(oldPwd.value, newPwd.value).then((res) => {
    if (res.code === 0) {
      api.logout().then((res) => {
        if (res.code === 0) {
          toast.success('修改成功')
          router.replace('/login')
        }
        else {
          toast.error(res.msg)
        }
      }).catch(() => {
        toast.error('加载失败，请重试')
      })
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
</script>

<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="修改密码" />

    <main class="mx-auto w-full max-w-5xl space-y-4 px-4 pt-3">
      <AppInput v-model="oldPwd" type="password" label="当前密码" placeholder="请输入当前密码" />
      <AppInput v-model="newPwd" type="password" label="新密码" placeholder="请输入新密码" />
      <AppInput v-model="confirmPwd" type="password" label="确认新密码" placeholder="请再次输入新密码" />
      <AppButton block size="lg" :loading="saveLoading" @click="submit">
        确认修改
      </AppButton>
    </main>
  </div>
</template>
