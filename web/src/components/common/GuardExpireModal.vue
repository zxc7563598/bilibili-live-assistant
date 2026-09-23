<template>
  <n-modal v-model:show="visible" title="变更大航海身份" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次变更的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的用户
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「大航海身份」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            身份由到期时间决定：某个档位的到期时间<span class="text-gray-700 font-medium dark:text-gray-200">大于当前时间</span>，用户就获得该档位身份。
          </div>
          <div>
            多个档位都还在有效期内时，<span class="text-gray-700 font-medium dark:text-gray-200">高等级覆盖低等级</span>，从上往下依次覆盖：总督覆盖提督，提督覆盖舰长。
          </div>
          <div>
            到期时间按<span class="text-gray-700 font-medium dark:text-gray-200">当天 0 点</span>计算，与开通大航海的计时方式一致。
          </div>
          <div>
            留空表示<span class="text-gray-700 font-medium dark:text-gray-200">清空</span>该档位的到期时间，原本有效的身份会立即失效。
          </div>
        </div>
      </div>
      <n-spin :show="loading">
        <div class="space-y-10">
          <div v-for="item in fields" :key="item.key" class="flex items-center gap-5">
            <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
              {{ item.label }}
            </div>
            <n-date-picker v-model:value="form[item.key]" type="date" clearable :disabled="loading" placeholder="留空表示清空该档位" class="w-full" />
          </div>
          <div class="flex items-center gap-5">
            <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
              生效身份
            </div>
            <div class="flex items-center gap-6 text-13 text-gray-500 dark:text-gray-400">
              保存后为
              <GuardTag :type="previewLevel" />
            </div>
          </div>
        </div>
      </n-spin>
    </div>
    <template #footer>
      <div class="flex">
        <NButton strong secondary type="primary" class="ml-auto" :loading="saving" :disabled="loading" @click="save">
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
import api from '@/api/liveuser'
import GuardTag from './GuardTag.vue'

const emit = defineEmits(['success'])

// 与后端三个到期时间字段一一对应，高等级在前，方便与上面的覆盖规则对照
const FIELDS = [
  { key: 'governor_expire_at', label: '总督到期时间', level: 1 },
  { key: 'admiral_expire_at', label: '提督到期时间', level: 2 },
  { key: 'captain_expire_at', label: '舰长到期时间', level: 3 },
]

const fields = FIELDS
const visible = ref(false)
const saving = ref(false)
const loading = ref(false)
// 每次打开时由调用方传入，避免组件里再缓存一份可能过期的用户ID
const targetUserID = ref(0)
const targetLabel = ref('')

function initForm() {
  return {
    governor_expire_at: null,
    admiral_expire_at: null,
    captain_expire_at: null,
  }
}
const form = ref(initForm())

// 预览保存后的身份，判定规则与后端 guardVipType 一致：有效期内的最高档
const previewLevel = computed(() => {
  const now = Date.now()
  const hit = FIELDS.find(item => form.value[item.key] > now)
  return hit ? hit.level : 0
})

// label 用于在弹窗里提示本次操作的对象，列表页传入用户昵称+UID 便于核对
function open(userID, label = '') {
  targetUserID.value = Number(userID) || 0
  targetLabel.value = label
  // 先清空再拉取，避免上一轮的数据闪现在新对象身上
  form.value = initForm()
  visible.value = true
  loadExpire()
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = initForm()
  targetUserID.value = 0
  targetLabel.value = ''
  loading.value = false
}

// 拉取该用户当前的三个到期时间回填
//
// 空表单直接提交会把已有身份全部清掉，所以必须等回填完成才允许保存。
async function loadExpire() {
  const userID = targetUserID.value
  if (!userID)
    return
  loading.value = true
  try {
    const res = await api.getGuardExpire(userID)
    // 加载期间可能已经切到别的用户或关掉了弹窗，这一份结果就不能再往表单里写
    if (targetUserID.value !== userID)
      return
    form.value = {
      governor_expire_at: toExpireMillis(res.data?.governor_expire_at),
      admiral_expire_at: toExpireMillis(res.data?.admiral_expire_at),
      captain_expire_at: toExpireMillis(res.data?.captain_expire_at),
    }
  }
  catch {
    // 失败提示由 http 拦截器统一处理；读不到就关掉，
    // 免得管理员在空表单上保存、把身份清空
    if (targetUserID.value === userID)
      close()
  }
  finally {
    if (targetUserID.value === userID)
      loading.value = false
  }
}

// 后端存 Unix 秒，日期选择器要毫秒；空值统一传 null 让选择器留空
function toExpireMillis(seconds) {
  return seconds ? seconds * 1000 : null
}

// 日期选择器给的是本地时间戳（当天 0 点），后端按秒存，空值统一传 null
function toExpireSeconds(value) {
  return typeof value === 'number' && value > 0 ? Math.floor(value / 1000) : null
}

async function save() {
  if (!targetUserID.value)
    return $message.warning('用户信息异常，请刷新列表后重试')

  saving.value = true
  try {
    await api.saveGuardExpire({
      user_id: targetUserID.value,
      governor_expire_at: toExpireSeconds(form.value.governor_expire_at),
      admiral_expire_at: toExpireSeconds(form.value.admiral_expire_at),
      captain_expire_at: toExpireSeconds(form.value.captain_expire_at),
    })
    $message.success('身份变更成功', { key: 'liveuser.guard' })
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
