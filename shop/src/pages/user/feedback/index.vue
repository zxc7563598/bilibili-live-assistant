<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import dialog from '@/utils/dialog'
import toast from '@/utils/toast'
import api from './api'

const router = useRouter()

const types = ['商品问题', '物流问题', '功能建议', '其他问题', '单纯发癫']
const type = ref('商品问题')
const content = ref('')
const contact = ref('')
const saveLoading = ref(false)

function submit() {
  if (type.value === '单纯发癫') {
    dialog.warning({
      title: '认真的？',
      message: `你确定吗？`,
      confirmText: '我不怕',
      confirmVariant: 'danger',
      onConfirm: () => {
        save()
      },
    })
  }
  else {
    save()
  }
}

function save() {
  saveLoading.value = true
  api.savedFeedback(type.value, content.value, contact.value).then((res) => {
    if (res.code === 0) {
      console.warn(res)
      toast.success('反馈成功，请等待主播与您联系')
      if (window.history.length > 1) {
        router.back()
      }
      else {
        router.replace('/profile')
      }
    }
    else {
      toast.error(res.msg)
    }
  }).finally(() => {
    saveLoading.value = false
  })
}
</script>

<template>
  <div class="min-h-dvh bg-bg pb-10">
    <AppNavBar title="投诉建议" />

    <main class="mx-auto w-full max-w-5xl space-y-4 px-4 pt-3">
      <div>
        <span class="mb-1.5 block text-sm font-medium text-fg-2">问题类型</span>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="t in types" :key="t" class="rounded-full border px-4 py-1.5 text-sm transition"
            :class="type === t ? 'border-primary bg-primary-soft text-primary' : 'border-line text-fg-2'"
            @click="type = t"
          >
            {{ t }}
          </button>
        </div>
      </div>

      <AppInput v-model="content" type="textarea" :rows="5" label="问题描述" placeholder="请描述你遇到的问题，我们会尽快处理" />
      <AppInput v-model="contact" label="联系方式" placeholder="手机号 / 邮箱（选填）" />

      <AppButton block size="lg" :loading="saveLoading" @click="submit">
        提交
      </AppButton>
    </main>
  </div>
</template>
