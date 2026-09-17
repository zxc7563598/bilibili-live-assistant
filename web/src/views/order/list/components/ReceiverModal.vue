<template>
  <n-modal v-model:show="visible" title="变更收货信息" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次变更的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的订单
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「变更收货信息」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            只修改<span class="text-gray-700 font-medium dark:text-gray-200">当前这一张订单里的收货快照</span>，不会动用户在积分商城「收货地址簿」里的原始地址。
          </div>
          <div>
            修改后<span class="text-gray-700 font-medium dark:text-gray-200">用户端立即可见</span>，请确认与用户沟通一致后再保存。
          </div>
          <div>
            收货类型由下单时的商品类型决定，<span class="text-gray-700 font-medium dark:text-gray-200">不能在这里切换</span>：虚拟商品只填邮箱，实体商品只填收货地址。
          </div>
        </div>
      </div>
      <template v-if="kind === 'virtual'">
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            收货邮箱
          </div>
          <n-input v-model:value="form.receiver_email" :maxlength="255" placeholder="虚拟商品发放凭据，例如：x@x.com" />
        </div>
      </template>
      <template v-else-if="kind === 'actual'">
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            收货人
          </div>
          <n-input v-model:value="form.receiver_name" :maxlength="100" placeholder="请输入收货人姓名" />
        </div>
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            手机号
          </div>
          <n-input v-model:value="form.receiver_phone" :maxlength="100" placeholder="请输入收货人手机号" />
        </div>
        <div class="flex items-center gap-5">
          <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
            所在地区
          </div>
          <RegionCascader v-model="form.receiver_region_codes" />
        </div>
        <div class="flex items-start gap-5">
          <div class="w-100 shrink-0 pt-6 text-right text-13 text-gray-500 dark:text-gray-400">
            详细地址
          </div>
          <n-input v-model:value="form.receiver_detail" type="textarea" :maxlength="255" :autosize="{ minRows: 2, maxRows: 4 }" show-count placeholder="街道、门牌号等" />
        </div>
      </template>
      <div v-else class="text-13 text-red-500">
        该订单的收货类型（{{ rawReceiverType }}）无法识别，暂不支持在此变更收货信息，请检查订单数据。
      </div>
    </div>
    <template #footer>
      <div class="flex">
        <NButton strong secondary type="primary" class="ml-auto" :loading="saving" :disabled="!canSave" @click="save">
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
import RegionCascader from './RegionCascader.vue'

const emit = defineEmits(['success'])

// 与后端 enum.AddressType 保持一致
const ADDRESS_TYPE_VIRTUAL = 0
const ADDRESS_TYPE_ACTUAL = 1

const visible = ref(false)
const saving = ref(false)
const targetID = ref(0)
const targetLabel = ref('')
const rawReceiverType = ref(ADDRESS_TYPE_VIRTUAL)

// 'virtual' | 'actual' | 'unknown'，unknown 时整个表单不可用
const kind = computed(() => {
  if (rawReceiverType.value === ADDRESS_TYPE_VIRTUAL)
    return 'virtual'
  if (rawReceiverType.value === ADDRESS_TYPE_ACTUAL)
    return 'actual'
  return 'unknown'
})

// 解析订单里存的地区 code
function parseRegionCodes(raw) {
  if (!raw || typeof raw !== 'string')
    return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed))
      return []
    return parsed.filter(item => typeof item === 'string' && item !== '')
  }
  catch {
    return []
  }
}

function initForm() {
  return {
    receiver_email: '',
    receiver_name: '',
    receiver_phone: '',
    receiver_region_codes: [],
    receiver_detail: '',
  }
}
const form = ref(initForm())
// 打开时的原始值，用于判断管理员是否真的改过
const origin = ref(initForm())

// 比对前先 trim：后端 strPtr 会去空格，不对齐的话「只多了个空格」会被判成有变更、
// 提交后数据却没有任何变化，看起来像保存失败
const changed = computed(() => {
  const f = form.value
  const o = origin.value
  if (kind.value === 'virtual')
    return (f.receiver_email ?? '').trim() !== o.receiver_email
  if (kind.value !== 'actual')
    return false
  return (f.receiver_name ?? '').trim() !== o.receiver_name
    || (f.receiver_phone ?? '').trim() !== o.receiver_phone
    || (f.receiver_detail ?? '').trim() !== o.receiver_detail
    || f.receiver_region_codes.join(',') !== o.receiver_region_codes.join(',')
})

// 收货类型异常时后端会直接返回 11303，这里把保存一并禁掉，不给必然失败的按钮
const canSave = computed(() => kind.value !== 'unknown' && changed.value)

function open(order) {
  targetID.value = Number(order.id) || 0
  targetLabel.value = `${order.order_sn}（${order.uname || '未知昵称'}）`
  rawReceiverType.value = Number(order.receiver_type)

  const snapshot = {
    receiver_email: order.receiver_email ?? '',
    receiver_name: order.receiver_name ?? '',
    receiver_phone: order.receiver_phone ?? '',
    receiver_region_codes: parseRegionCodes(order.receiver_region_code),
    receiver_detail: order.receiver_detail ?? '',
  }
  origin.value = { ...snapshot, receiver_region_codes: [...snapshot.receiver_region_codes] }
  form.value = { ...snapshot, receiver_region_codes: [...snapshot.receiver_region_codes] }
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = initForm()
  origin.value = initForm()
  targetID.value = 0
  targetLabel.value = ''
  rawReceiverType.value = ADDRESS_TYPE_VIRTUAL
}

// 校验顺序刻意与后端 service 的返回顺序对齐
function validate() {
  if (!targetID.value)
    return '订单信息异常，请刷新页面后重试'
  if (kind.value === 'unknown')
    return '该订单的收货类型无法识别，暂不支持变更收货信息'
  if (kind.value === 'virtual') {
    if (!(form.value.receiver_email ?? '').trim())
      return '请输入收货邮箱'
    return ''
  }
  if (!(form.value.receiver_name ?? '').trim())
    return '请输入收货人姓名'
  if (!(form.value.receiver_phone ?? '').trim())
    return '请输入收货人手机号'
  if (!form.value.receiver_region_codes.length)
    return '请选择省 / 市 / 区'
  if (!(form.value.receiver_detail ?? '').trim())
    return '请输入详细地址'
  return ''
}

function buildPayload() {
  const payload = { id: targetID.value }
  // 只放适用于该收货类型的键：后端对「传了不适用字段」会直接返回 11101
  if (kind.value === 'virtual') {
    payload.receiver_email = (form.value.receiver_email ?? '').trim()
    return payload
  }
  payload.receiver_name = (form.value.receiver_name ?? '').trim()
  payload.receiver_phone = (form.value.receiver_phone ?? '').trim()
  // 后端解析这个 JSON 数组后交给 region.Resolve 校验
  payload.receiver_region_code = JSON.stringify(form.value.receiver_region_codes)
  payload.receiver_detail = (form.value.receiver_detail ?? '').trim()
  return payload
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.updateReceiverInfo(buildPayload())
    $message.success('收货信息变更成功', { key: 'order.receiver' })
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
