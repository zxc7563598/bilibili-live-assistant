<template>
  <n-modal v-model:show="visible" title="补录大航海开通" preset="card" style="width: 680px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「补录开通」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            补录的是<span class="text-gray-700 font-medium dark:text-gray-200">系统漏记</span>的一次开通，会按直播链路的口径把礼物记录、累计消费金额、到期时间、开通台账与积分/星光奖励一并补齐。
          </div>
          <div>
            不会补发感谢弹幕，也不走黑名单赎回：补录的是历史，这两件事对历史没有意义。
          </div>
          <div>
            名单只按<span class="text-gray-700 font-medium dark:text-gray-200">开通月份与份数</span>算：这次开通算在开通当月，多出来的份数往后顺延（7 月开 3 份，7、8、9 三个月都在名单里）。开通时间具体到哪一天不影响名单，只影响礼物记录与到期时间。
          </div>
          <div>
            到期时间会按这次开通<span class="text-gray-700 font-medium dark:text-gray-200">往后叠加</span>，开通时间要填真实的开通时刻，填错会把用户的有效期算歪。
          </div>
        </div>
      </div>

      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          用户UID
        </div>
        <n-input-number v-model:value="form.uid" :show-button="false" :precision="0" placeholder="用户完整UID" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          用户昵称
        </div>
        <n-input v-model:value="form.uname" type="text" :maxlength="100" placeholder="用户开通时的昵称" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          房间ID
        </div>
        <n-input-number v-model:value="form.room_id" :show-button="false" :precision="0" placeholder="这次开通所在的房间ID" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          大航海等级
        </div>
        <n-select v-model:value="form.guard_level" :options="guardLevelOptions" :clearable="false" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          开通时间
        </div>
        <n-date-picker v-model:value="form.open_at" type="datetime" :clearable="false" placeholder="这次开通发生的时刻" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          开通数量
        </div>
        <n-input-number v-model:value="form.num" :show-button="false" :precision="0" placeholder="开通份数" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          开通价格（元）
        </div>
        <n-input-number v-model:value="form.price" :show-button="false" :precision="2" placeholder="单份价格，按元填写，舰长填 138" class="w-full" />
      </div>
      <div class="flex items-center gap-5">
        <div class="w-110 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          礼物ID
        </div>
        <n-input-number v-model:value="form.gift_id" :show-button="false" :precision="0" placeholder="可留空" class="w-full" />
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

// 与后端 enum.BadgeType 的取值保持一致：1 总督、2 提督、3 舰长
const GUARD_LEVELS = [1, 2, 3]

const visible = ref(false)
const saving = ref(false)

const guardLevelOptions = ref([
  { label: '舰长', value: 3 },
  { label: '提督', value: 2 },
  { label: '总督', value: 1 },
])

function initForm(guardLevel) {
  return {
    uid: null,
    uname: '',
    room_id: null,
    // 默认取列表当前筛选的等级
    guard_level: GUARD_LEVELS.includes(guardLevel) ? guardLevel : 3,
    open_at: null,
    num: 1,
    price: null,
    gift_id: null,
  }
}
const form = ref(initForm())

// guardLevel 由列表页传入，即当前筛选的等级
function open(guardLevel) {
  form.value = initForm(guardLevel)
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = initForm()
}

function validate() {
  const { uid, uname, room_id, guard_level, open_at, num, price, gift_id } = form.value
  if (!Number.isInteger(Number(uid)) || Number(uid) <= 0)
    return '请填写正确的用户UID'
  if (!String(uname ?? '').trim())
    return '请填写用户昵称'
  if (!Number.isInteger(Number(room_id)) || Number(room_id) <= 0)
    return '请填写正确的房间ID'
  if (!guardLevelOptions.value.some(item => item.value === Number(guard_level)))
    return '请选择大航海等级'
  if (!open_at)
    return '请选择开通时间'
  if (open_at > Date.now())
    return '开通时间不能晚于当前时间'
  if (!Number.isInteger(Number(num)) || Number(num) < 1)
    return '开通数量必须为大于 0 的整数'
  const priceValue = Number(price)
  if (!Number.isFinite(priceValue) || priceValue <= 0)
    return '请填写正确的开通价格'
  if (gift_id && (!Number.isInteger(Number(gift_id)) || Number(gift_id) < 0))
    return '礼物ID必须为大于 0 的整数，或留空'
  return ''
}

async function save() {
  const message = validate()
  if (message)
    return $message.warning(message)

  saving.value = true
  try {
    await api.backfill({
      uid: Number(form.value.uid),
      uname: String(form.value.uname ?? '').trim(),
      room_id: Number(form.value.room_id),
      guard_level: Number(form.value.guard_level),
      // 选择器给的是毫秒时间戳，后端按秒存
      open_at: Math.floor(form.value.open_at / 1000),
      num: Number(form.value.num),
      // 后端按分存，与礼物记录的 price 同口径
      price: Math.round(Number(form.value.price) * 100),
      gift_id: form.value.gift_id ? Number(form.value.gift_id) : 0,
    })
    $message.success('补录成功', { key: 'liveguard.backfill' })
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
