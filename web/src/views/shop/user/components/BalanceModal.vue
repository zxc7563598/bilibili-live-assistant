<template>
  <n-modal v-model:show="visible" title="变更积分/余额" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次变更的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的用户
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「手动变更余额」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            保存后会写入一条资产流水，记录变动前后的数值、操作管理员和说明，可以在用户详情的「变更记录」里查到。
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">流水不可修改、不可删除</span>，请确认无误后再保存。
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">增加</span>：直接为用户加上对应数值；
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">减少</span>：从用户余额中扣掉，剩余不足时会被系统拒绝，不会扣成负数。
          </div>
          <div>
            说明会原样展示在积分商城的「账户记录」里，<span class="text-gray-700 font-medium dark:text-gray-200">用户可以看到</span>，请不要填写内部备注。
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          余额类型
        </div>
        <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
          <div v-for="item in creditTypeOptions" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="form.credit_type === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.credit_type = item.value">
            {{ item.label }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          变动类型
        </div>
        <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
          <div v-for="item in changeTypeOptions" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="form.change_type === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.change_type = item.value">
            {{ item.label }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          变动数值
        </div>
        <n-input-number v-model:value="form.change_amount" :show-button="false" placeholder="请输入大于 0 的整数" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          变动说明
        </div>
        <n-input v-model:value="form.remark" type="text" :maxlength="255" show-count placeholder="例如：活动补偿、误扣返还，可不填" />
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

// 与后端 enum.CreditType / enum.ChangeType 的取值保持一致
const CREDIT_TYPE = { stars: 0, points: 1 }
const CHANGE_TYPE = { reduce: 0, increase: 1 }

const visible = ref(false)
const saving = ref(false)
// 每次打开时由调用方传入，避免组件里再缓存一份可能过期的用户ID
const targetUserID = ref(0)
const targetLabel = ref('')

const creditTypeOptions = ref([
  { label: '星光', value: CREDIT_TYPE.stars },
  { label: '积分', value: CREDIT_TYPE.points },
])

const changeTypeOptions = ref([
  { label: '减少', value: CHANGE_TYPE.reduce },
  { label: '增加', value: CHANGE_TYPE.increase },
])

function initForm() {
  return {
    credit_type: CREDIT_TYPE.points,
    change_type: CHANGE_TYPE.increase,
    change_amount: null,
    remark: '',
  }
}
const form = ref(initForm())

// label 用于在弹窗里提示本次操作的对象，列表页传入用户昵称+UID 便于核对
function open(userID, label = '') {
  targetUserID.value = Number(userID) || 0
  targetLabel.value = label
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = initForm()
  targetUserID.value = 0
  targetLabel.value = ''
}

function validate() {
  if (!targetUserID.value)
    return '用户信息异常，请刷新列表后重试'
  if (!creditTypeOptions.value.some(item => item.value === Number(form.value.credit_type)))
    return '请选择余额类型'
  if (!changeTypeOptions.value.some(item => item.value === Number(form.value.change_type)))
    return '请选择变动类型'
  const amount = Number(form.value.change_amount)
  if (!form.value.change_amount || !Number.isInteger(amount) || amount <= 0)
    return '变动数值必须为大于 0 的整数'
  if (amount > Number.MAX_SAFE_INTEGER)
    return '变动数值过大，请确认后重新输入'
  if ((form.value.remark ?? '').length > 255)
    return '变动说明不能超过 255 个字符'
  return ''
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.saveAssets(
      targetUserID.value,
      Number(form.value.credit_type),
      Number(form.value.change_type),
      Number(form.value.change_amount),
      (form.value.remark ?? '').trim(),
    )
    $message.success('余额变更成功', { key: 'shop.user.balance' })
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
