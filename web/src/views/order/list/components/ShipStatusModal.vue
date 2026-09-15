<template>
  <n-modal v-model:show="visible" title="变更发货状态" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次变更的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的订单
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「发货状态」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            变更发货状态会<span class="text-gray-700 font-medium dark:text-gray-200">连带变更订单状态</span>，后台不校验流转合法性，任意状态互转都会成功。
          </div>
          <div>
            只补填快递信息、发货状态不变时，<span class="text-gray-700 font-medium dark:text-gray-200">订单状态与发货时间都保持原样</span>。
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">未发货</span>：订单回到「待发货」，并清空发货时间与快递信息；
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">已发货</span>：虚拟商品直接「完成」，实体商品进入「待收货」；
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">已签收</span>：订单一律置为「完成」。
          </div>
          <div>
            发货时间只在首次发货时写入，重复变更会保留原时间，不会把发货时间改晚。
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          发货状态
        </div>
        <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
          <div v-for="item in shipStatusOptions" :key="item.value" class="cursor-pointer px-7 py-1.5 text-13 transition" :class="form.ship_status === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.ship_status = item.value">
            {{ item.label }}
          </div>
        </div>
      </div>
      <template v-if="isActual">
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            快递公司
          </div>
          <n-input v-model:value="form.express_company" :maxlength="100" placeholder="例如：顺丰速运" />
        </div>
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            快递单号
          </div>
          <n-input v-model:value="form.express_no" :maxlength="100" placeholder="例如：SF1234567890" />
        </div>
        <div class="pl-105 text-12 text-gray-400 dark:text-gray-500">
          快递信息<span class="text-gray-600 dark:text-gray-300">留空表示不修改</span>，原值会保留（后端只在传入非空值时才覆盖）。
        </div>
      </template>
      <div v-else class="pl-105 text-12 text-gray-400 dark:text-gray-500">
        该订单是虚拟商品，没有物流环节，不接收快递信息。设为「已发货」或「已签收」都会直接把订单置为「完成」。
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

// 0 虚拟 / 1 实体，与后端 enum.AddressType 保持一致
const ADDRESS_TYPE_VIRTUAL = 0
const ADDRESS_TYPE_ACTUAL = 1

const shipStatusOptions = ref([
  { label: '未发货', value: 0 },
  { label: '已发货', value: 1 },
  { label: '已签收', value: 2 },
])

const visible = ref(false)
const saving = ref(false)
const targetID = ref(0)
const targetLabel = ref('')
// 打开时的原始值，用于判断管理员是否真的改过、以及对齐后端的合并语义
const origin = ref({ ship_status: 0, express_company: '', express_no: '' })
const receiverType = ref(ADDRESS_TYPE_VIRTUAL)

const isActual = computed(() => receiverType.value === ADDRESS_TYPE_ACTUAL)

function initForm() {
  return { ship_status: 0, express_company: '', express_no: '' }
}
const form = ref(initForm())

// 比对前先 trim：后端 strPtr 会去空格，不对齐的话「只多了个空格」会被判成有变更、
// 提交后数据却没有任何变化，看起来像保存失败
const changed = computed(() => {
  if (form.value.ship_status !== origin.value.ship_status)
    return true
  if (!isActual.value)
    return false
  return (form.value.express_company ?? '').trim() !== origin.value.express_company
    || (form.value.express_no ?? '').trim() !== origin.value.express_no
})

function open(order) {
  targetID.value = Number(order.id) || 0
  targetLabel.value = `${order.order_sn}（${order.uname || '未知昵称'}）`
  receiverType.value = Number(order.receiver_type)
  origin.value = {
    ship_status: Number(order.ship_status) || 0,
    express_company: order.express_company ?? '',
    express_no: order.express_no ?? '',
  }
  form.value = { ...origin.value }
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = initForm()
  origin.value = { ship_status: 0, express_company: '', express_no: '' }
  targetID.value = 0
  targetLabel.value = ''
  receiverType.value = ADDRESS_TYPE_VIRTUAL
}

function validate() {
  if (!targetID.value)
    return '订单信息异常，请刷新页面后重试'
  if (!shipStatusOptions.value.some(item => item.value === Number(form.value.ship_status)))
    return '请选择发货状态'
  if ((form.value.express_company ?? '').length > 100)
    return '快递公司不能超过 100 个字符'
  if ((form.value.express_no ?? '').length > 100)
    return '快递单号不能超过 100 个字符'
  return ''
}

function buildPayload() {
  const payload = {
    id: targetID.value,
    ship_status: Number(form.value.ship_status),
  }
  // 虚拟订单必须整键省略快递字段：后端按「非 nil」判定，传空串同样返回 11101
  if (isActual.value) {
    payload.express_company = (form.value.express_company ?? '').trim()
    payload.express_no = (form.value.express_no ?? '').trim()
  }
  return payload
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.updateShipStatus(buildPayload())
    $message.success('发货状态变更成功', { key: 'order.ship-status' })
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
