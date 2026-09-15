<template>
  <n-modal v-model:show="visible" title="变更订单状态" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次变更的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的订单
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「订单状态」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            这里是纯状态改写：<span class="text-gray-700 font-medium dark:text-gray-200">不校验流转合法性</span>，状态之间任意互转都会成功。
          </div>
          <div>
            置为<span class="text-gray-700 font-medium dark:text-gray-200">已取消</span>会写入取消时间；由「已取消」改回其它状态会清空该时间。
          </div>
          <div>
            改订单状态<span class="text-gray-700 font-medium dark:text-gray-200">不会同步发货状态</span>。两者互相独立，需要一致时请分别变更。
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">变更状态不涉及退款与库存回滚</span>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          订单状态
        </div>
        <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
          <div v-for="item in orderStatusOptions" :key="item.value" class="cursor-pointer px-7 py-1.5 text-13 transition" :class="form.order_status === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.order_status = item.value">
            {{ item.label }}
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex">
        <NButton strong secondary type="primary" class="ml-auto" :loading="saving" :disabled="!changed" @click="save">
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

const orderStatusOptions = ref([
  { label: '待付款', value: 0 },
  { label: '待发货', value: 1 },
  { label: '待收货', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已取消', value: 4 },
  { label: '售后中', value: 5 },
])

const visible = ref(false)
const saving = ref(false)
const targetID = ref(0)
const targetLabel = ref('')
const originStatus = ref(0)

const form = ref({ order_status: 0 })

const changed = computed(() => Number(form.value.order_status) !== originStatus.value)

function open(order) {
  targetID.value = Number(order.id) || 0
  targetLabel.value = `${order.order_sn}（${order.uname || '未知昵称'}）`
  originStatus.value = Number(order.order_status) || 0
  form.value = { order_status: originStatus.value }
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = { order_status: 0 }
  originStatus.value = 0
  targetID.value = 0
  targetLabel.value = ''
}

function validate() {
  if (!targetID.value)
    return '订单信息异常，请刷新页面后重试'
  if (!orderStatusOptions.value.some(item => item.value === Number(form.value.order_status)))
    return '请选择订单状态'
  return ''
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.updateOrderStatus(targetID.value, Number(form.value.order_status))
    $message.success('订单状态变更成功', { key: 'order.order-status' })
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
