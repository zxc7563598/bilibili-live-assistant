<template>
  <n-modal v-model:show="visible" title="投诉详情" preset="card" style="width: 640px" :mask-closable="false" :on-after-leave="resetForm">
    <n-spin v-if="loading" class="block w-full py-40" size="large" />
    <n-empty v-else-if="errorText" :description="errorText" class="py-40" />
    <div v-else-if="detail" class="flex flex-col gap-11">
      <div class="flex items-start gap-12">
        <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
          投诉用户
        </div>
        <div class="min-w-0 flex flex-1 items-center gap-8 text-13 leading-20">
          <span class="break-all">{{ detail.uname || '未知昵称' }}</span>
          <span class="shrink-0 text-12 text-gray-400 dark:text-gray-500">UID {{ detail.uid }}</span>
          <NButton text type="primary" size="tiny" class="shrink-0" :disabled="!detail.user_id" @click="goUserDetail">
            查看用户
          </NButton>
        </div>
      </div>
      <div class="flex items-start gap-12">
        <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
          问题类型
        </div>
        <div class="min-w-0 flex-1 break-all text-13 leading-20">
          {{ dash(detail.type) }}
        </div>
      </div>
      <div class="flex items-start gap-12">
        <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
          联系方式
        </div>
        <div class="min-w-0 flex-1 break-all text-13 leading-20">
          {{ dash(detail.contact) }}
        </div>
      </div>
      <div class="flex items-start gap-12">
        <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
          投诉时间
        </div>
        <div class="min-w-0 flex-1 text-13 leading-20">
          {{ dash(detail.created_at) }}
        </div>
      </div>
      <div class="flex items-start gap-12">
        <div class="w-88 shrink-0 text-13 text-gray-400 leading-20 dark:text-gray-500">
          投诉内容
        </div>
        <!-- 用户输入可能带换行，pre-wrap 原样保留，别压成一行 -->
        <div class="min-w-0 flex-1 whitespace-pre-wrap break-all text-13 leading-20">
          {{ dash(detail.content) }}
        </div>
      </div>
    </div>
    <template #footer>
      <div class="flex">
        <NButton strong secondary type="primary" class="ml-auto" @click="close">
          关闭
        </NButton>
      </div>
    </template>
  </n-modal>
</template>

<script setup>
import { NButton } from 'naive-ui'
import api from '../api'

const router = useRouter()

const visible = ref(false)
const loading = ref(false)
const targetID = ref(0)
const detail = ref(null)
const errorText = ref('')

// 详情请求的序号，只有最新一次请求的结果会被采用
let requestSeq = 0

// 列表行里没有正文，弹窗打开后再按 id 拉一次详情
function open(row) {
  targetID.value = Number(row.id) || 0
  visible.value = true
  load()
}
function close() {
  visible.value = false
}
// 关闭动画结束后再清空，避免内容在收起过程中提前闪空
function resetForm() {
  // 收起动画被打断（关掉后立刻又点开）时 after-leave 同样会触发，此时弹窗已重新打开、请求可能已经回来，
  // 清空会把刚渲染出来的详情抹掉，只留一个空白卡片
  if (visible.value)
    return
  loading.value = false
  targetID.value = 0
  detail.value = null
  errorText.value = ''
}

// 后端 timeutil.Format 对 0 值返回空串，这里统一成占位符，避免出现只有标签没内容的空行
function dash(value) {
  return value === '' || value === null || value === undefined ? '—' : value
}

function goUserDetail() {
  router.push({ path: '/shop/user/details', query: { id: detail.value.user_id } })
}

async function load() {
  // 先清空，避免切换投诉时短暂显示上一条的内容
  detail.value = null
  errorText.value = ''
  const id = targetID.value
  if (!id) {
    errorText.value = '投诉信息异常，请刷新列表后重试'
    return
  }
  // 连着点开不同投诉时，先发出的慢请求可能后返回；用序号丢弃过期结果，避免显示错记录的内容
  const seq = ++requestSeq
  loading.value = true
  try {
    const res = await api.getDetails(id)
    if (seq !== requestSeq)
      return
    detail.value = res.data
  }
  catch {
    if (seq !== requestSeq)
      return
    // 失败提示由 http 拦截器统一处理，这里只做弹窗内兜底
    errorText.value = '投诉详情加载失败，请关闭后重试'
  }
  finally {
    if (seq === requestSeq)
      loading.value = false
  }
}

defineExpose({ open })
</script>
