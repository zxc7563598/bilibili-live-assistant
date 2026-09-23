<template>
  <n-modal v-model:show="visible" title="变更限购档位" preset="card" style="width: 560px" :mask-closable="false" :on-after-leave="resetForm">
    <div class="space-y-10">
      <div v-if="targetLabel" class="text-highlight">
        本次变更的对象：<span class="text-primary font-medium">{{ targetLabel }}</span>，请确认是你要操作的商品
      </div>
      <div class="border border-gray-200 rounded-4 bg-gray-50 px-10 py-6 dark:border-gray-700 dark:bg-gray-800/50">
        <div class="text-13 text-gray-700 font-medium dark:text-gray-200">
          关于「限购」
        </div>
        <div class="mt-2 text-12 text-gray-500 space-y-10 dark:text-gray-400">
          <div>
            限购商品依然会正常展示，但需要用户持有<span class="text-gray-700 font-medium dark:text-gray-200">限制档位及以上</span>身份才允许下单。
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">不限制</span>：任何用户均可购买。
          </div>
          <div>
            <span class="text-gray-700 font-medium dark:text-gray-200">限制提督</span>：当前身份为提督或总督才允许购买，普通用户与舰长无法下单。
          </div>
        </div>
      </div>
      <div class="flex items-center gap-5">
        <div class="w-100 shrink-0 text-right text-13 text-gray-500 dark:text-gray-400">
          限购档位
        </div>
        <div class="inline-flex overflow-hidden border border-gray-200 rounded-6 dark:border-gray-700">
          <div v-for="item in minMemberLevelOptions" :key="item.value" class="cursor-pointer px-4 py-1.5 text-13 transition" :class="form.level === item.value ? 'bg-primary text-white' : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'" @click="form.level = item.value">
            {{ item.label }}
          </div>
        </div>
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
import { MIN_MEMBER_LEVEL_LABEL, MIN_MEMBER_LEVEL_OPTIONS } from '../limit-utils'

const emit = defineEmits(['success'])

const minMemberLevelOptions = MIN_MEMBER_LEVEL_OPTIONS
const visible = ref(false)
const saving = ref(false)
// 每次打开时由调用方传入，避免组件里再缓存一份可能过期的商品ID
const targetID = ref(0)
const targetLabel = ref('')

const form = ref({ level: 0 })

// 档位值与文案都直接取自列表行，不再请求详情：变更本身就是「设为某个值」，
// 行值即使略旧也无害
function open(row) {
  targetID.value = Number(row.id) || 0
  targetLabel.value = row.name || '未命名商品'
  // 列表里的未知档位一律按不限制展示，与标签组件的兜底保持一致
  form.value = { level: MIN_MEMBER_LEVEL_LABEL[row.min_member_level] ? row.min_member_level : 0 }
  visible.value = true
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免表单内容在收起过程中提前闪空
function resetForm() {
  form.value = { level: 0 }
  targetID.value = 0
  targetLabel.value = ''
}

async function save() {
  if (!targetID.value)
    return $message.warning('商品信息异常，请刷新列表后重试')

  saving.value = true
  try {
    await api.productMinMemberLevel(targetID.value, form.value.level)
    $message.success('限购变更成功', { key: 'product.limit' })
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
