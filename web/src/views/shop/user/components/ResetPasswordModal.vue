<template>
  <n-modal v-model:show="visible" title="重置用户密码" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次重置的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的用户
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「重置密码」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            这里<span class="text-gray-700 font-medium dark:text-gray-200">不需要原密码</span>，保存后立即生效，该用户下一次登录就要使用新密码。
          </div>
          <div>
            密码最少 6 位、最多 72 位。请通过私聊等可信渠道告知用户。
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          新密码
        </div>
        <n-input v-model:value="newPassword" type="password" show-password-on="click" placeholder="最少 6 位，建议字母与数字组合" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          确认密码
        </div>
        <n-input v-model:value="confirmPassword" type="password" show-password-on="click" placeholder="请再次输入新密码" />
      </div>
    </div>
    <template #footer>
      <div class="flex">
        <NButton strong secondary type="primary" class="ml-auto" :loading="saving" @click="save">
          保存
        </NButton>
        <NButton strong secondary class="ml-10" :disabled="saving" @click="close">
          取消
        </NButton>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { NButton } from 'naive-ui'
import api from '../api'

const emit = defineEmits(['success'])

// bcrypt 最多处理 72 字节，超出后端会直接失败，这里提前拦住
const PASSWORD_MAX_BYTES = 72

const visible = ref(false)
const saving = ref(false)
const targetUserID = ref(0)
const targetLabel = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

function open(userID, label = '') {
  targetUserID.value = Number(userID) || 0
  targetLabel.value = label
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免密码在收起过程中被重置
function resetForm() {
  newPassword.value = ''
  confirmPassword.value = ''
  targetUserID.value = 0
  targetLabel.value = ''
}

function validate() {
  if (!targetUserID.value)
    return '用户信息异常，请刷新列表后重试'
  if (newPassword.value.length < 6)
    return '密码不能低于 6 位'
  // 汉字按 UTF-8 是 3 字节，按字节校验才能和后端 bcrypt 的限制对齐
  if (new TextEncoder().encode(newPassword.value).length > PASSWORD_MAX_BYTES)
    return '密码过长，请控制在 72 字节以内（约 24 个汉字）'
  if (newPassword.value !== confirmPassword.value)
    return '两次输入的密码不一致'
  return ''
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.resetPassword(targetUserID.value, newPassword.value)
    $message.success('密码重置成功', { key: 'shop.user.password' })
    close()
    emit('success')
  }
  catch {
    // 失败提示由 http 拦截器统一处理，这里吞掉 rejection 避免冒到控制台
  }
  finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>
